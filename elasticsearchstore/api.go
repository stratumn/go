// Copyright 2017 Stratumn SAS. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package elasticsearchstore

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/olivere/elastic"
	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/types"
)

const (
	linksIndex     = "links"
	evidencesIndex = "evidences"
	valuesIndex    = "values"

	// This is the mapping for the links index.
	// We voluntarily disable indexing of the following fields:
	// meta.inputs, meta.refs, state, signatures
	linksMapping = `{
		"mappings": {
			"_doc": {
				"properties": {
					"meta": {
						"properties":{
							"mapId": {
								"type": "text",
								"fields": {
									"keyword": {
										"type": "keyword"
									}
								}
							},
							"process": {
								"type": "text",
								"fields": {
									"keyword": {
										"type": "keyword"
									}
								}
							},
							"action": {
								"type": "text",
								"fields": {
									"keyword": {
										"type": "keyword"
									}
								}
							},
							"type": {
								"type": "text",
								"fields": {
									"keyword": {
										"type": "keyword"
									}
								}
							},
							"inputs": {
								"enabled": false
							},
							"tags": {
								"type": "text",
								"fields": {
									"keyword": {
										"type": "keyword"
									}
								}
							},
							"priority": {
								"type": "double"
							},
							"prevLinkHash": {
								"type": "text",
								"fields": {
									"keyword": {
										"type": "keyword"
									}
								}
							},
							"refs": {
								"enabled": false
							},
							"data":{
								"enabled": false
							}
						}
					},
					"state": {
						"enabled": false
					},
					"signatures": {
						"enabled": false
					},
					"stateTokens": {
						"type": "text"
					}
				}
			}
		}
	}`

	// this is a generic mapping used for evidences and values index,
	// where we do not require indexing to be enabled.
	noMapping = `{
		"mappings": {
			"_doc": { 
				"enabled": false
			}
		}
	}`

	docType = "_doc"
)

// Evidences is a wrapper around cs.Evidences for json ElasticSearch serialization compliance.
// Elastic Search does not allow indexing of arrays directly.
type Evidences struct {
	Evidences *cs.Evidences `json:"evidences,omitempty"`
}

// Value is a wrapper struct for the value of the keyvalue store part.
// Elastic only accepts json structured objects.
type Value struct {
	Value []byte `json:"value,omitempty"`
}

type csLink cs.Link
type linkDoc struct {
	cs.Link
	StateTokens []string `json:"stateTokens"`
}

// SearchQuery contains pagination and query string information.
type SearchQuery struct {
	store.SegmentFilter
	Query string
}

func (es *ESStore) createIndex(indexName, mapping string) error {
	ctx := context.TODO()
	exists, err := es.client.IndexExists(indexName).Do(ctx)
	if err != nil {
		return err
	}
	if !exists {
		// TODO: pass mapping through BodyString.
		createIndex, err := es.client.CreateIndex(indexName).BodyString(mapping).Do(ctx)
		if err != nil {
			return err
		}
		if !createIndex.Acknowledged {
			// Not acknowledged.
			return fmt.Errorf("error creating %s index", indexName)
		}
	}

	return nil
}

func (es *ESStore) createLinksIndex() error {
	return es.createIndex(linksIndex, linksMapping)
}

func (es *ESStore) createEvidencesIndex() error {
	return es.createIndex(evidencesIndex, noMapping)
}

func (es *ESStore) createValuesIndex() error {
	return es.createIndex(valuesIndex, noMapping)
}

func (es *ESStore) deleteIndex(indexName string) error {
	ctx := context.TODO()
	del, err := es.client.DeleteIndex(indexName).Do(ctx)
	if err != nil {
		return err
	}

	if !del.Acknowledged {
		return fmt.Errorf("index %s was not deleted", indexName)
	}

	return nil
}

func (es *ESStore) deleteAllIndex() error {
	if err := es.deleteIndex(linksIndex); err != nil {
		return err
	}

	if err := es.deleteIndex(evidencesIndex); err != nil {
		return err
	}

	return es.deleteIndex(valuesIndex)
}

func (es *ESStore) notifyEvent(event *store.Event) {
	for _, c := range es.eventChans {
		c <- event
	}
}

// only extract leaves that are strings.
func (o *linkDoc) extractTokens(obj interface{}) {
	switch value := obj.(type) {
	case string:
		o.StateTokens = append(o.StateTokens, value)
	case map[string]interface{}:
		for _, v := range value {
			o.extractTokens(v)
		}
	case []interface{}:
		for _, v := range value {
			o.extractTokens(v)
		}
	case float64:
	case bool:
	default:
		return
	}
}

func fromLink(link *cs.Link) (*linkDoc, error) {
	doc := linkDoc{
		Link:        *link,
		StateTokens: []string{},
	}

	doc.extractTokens(link.State)

	return &doc, nil
}

func (es *ESStore) createLink(link *cs.Link) (*types.Bytes32, error) {
	linkHash, err := link.Hash()
	if err != nil {
		return nil, err
	}
	linkHashStr := linkHash.String()

	has, err := es.hasDocument(linksIndex, linkHashStr)
	if err != nil {
		return nil, err
	}

	if has {
		return nil, fmt.Errorf("link is immutable, %s already exists", linkHashStr)
	}

	linkDoc, err := fromLink(link)
	if err != nil {
		return nil, err
	}

	return linkHash, es.indexDocument(linksIndex, linkHashStr, linkDoc)
}

func (es *ESStore) hasDocument(indexName, id string) (bool, error) {
	ctx := context.TODO()
	return es.client.Exists().Index(indexName).Type(docType).Id(id).Do(ctx)
}

func (es *ESStore) indexDocument(indexName, id string, doc interface{}) error {
	ctx := context.TODO()
	_, err := es.client.Index().Index(indexName).Type(docType).Id(id).BodyJson(doc).Do(ctx)
	return err
}

func (es *ESStore) getDocument(indexName, id string) (*json.RawMessage, error) {
	has, err := es.hasDocument(indexName, id)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	ctx := context.TODO()
	get, err := es.client.Get().Index(indexName).Type(docType).Id(id).Do(ctx)
	if err != nil {
		return nil, err
	}
	if !get.Found {
		return nil, nil
	}

	return get.Source, nil
}

func (es *ESStore) deleteDocument(indexName, id string) error {
	ctx := context.TODO()
	_, err := es.client.Delete().Index(indexName).Type(docType).Id(id).Do(ctx)
	return err
}

func (es *ESStore) getLink(id string) (*cs.Link, error) {
	var link linkDoc
	jsn, err := es.getDocument(linksIndex, id)
	if err != nil {
		return nil, err
	}
	if jsn == nil {
		return nil, nil
	}
	err = json.Unmarshal(*jsn, &link)
	return &link.Link, err
}

func (es *ESStore) getEvidences(id string) (*cs.Evidences, error) {
	jsn, err := es.getDocument(evidencesIndex, id)
	if err != nil {
		return nil, err
	}
	evidences := Evidences{Evidences: &cs.Evidences{}}
	if jsn != nil {
		err = json.Unmarshal(*jsn, &evidences)
	}
	return evidences.Evidences, err
}

func (es *ESStore) addEvidence(linkHash string, evidence *cs.Evidence) error {
	currentDoc, err := es.getEvidences(linkHash)
	if err != nil {
		return err
	}

	if err := currentDoc.AddEvidence(*evidence); err != nil {
		return err
	}

	evidences := Evidences{
		Evidences: currentDoc,
	}

	return es.indexDocument(evidencesIndex, linkHash, &evidences)
}

func (es *ESStore) getValue(key string) ([]byte, error) {
	var value Value
	jsn, err := es.getDocument(valuesIndex, key)
	if err != nil {
		return nil, err
	}
	if jsn != nil {
		err = json.Unmarshal(*jsn, &value)
	}
	return value.Value, err
}

func (es *ESStore) setValue(key string, value []byte) error {
	v := Value{
		Value: value,
	}
	return es.indexDocument(valuesIndex, key, v)
}

func (es *ESStore) deleteValue(key string) ([]byte, error) {
	value, err := es.getValue(key)
	if err != nil {
		return nil, err
	}

	if value == nil {
		return nil, nil
	}

	return value, es.deleteDocument(valuesIndex, key)
}

func (es *ESStore) segmentify(ctx context.Context, link *cs.Link) *cs.Segment {
	segment := link.Segmentify()

	evidences, err := es.GetEvidences(ctx, segment.Meta.GetLinkHash())
	if evidences != nil && err == nil {
		segment.Meta.Evidences = *evidences
	}
	return segment
}

func (es *ESStore) getMapIDs(filter *store.MapFilter) ([]string, error) {
	// Flush to make sure the documents got written.
	ctx := context.TODO()
	_, err := es.client.Flush().Index(linksIndex).Do(ctx)
	if err != nil {
		return nil, err
	}

	// prepare search service.
	svc := es.client.
		Search().
		Index(linksIndex).
		Type(docType)

	// add aggregation for map ids.
	a := elastic.
		NewTermsAggregation().
		Field("meta.mapId.keyword").
		Order("_key", true)
	svc.Aggregation("mapIds", a)

	// add process filtering.
	if filter.Process != "" {
		q := elastic.
			NewBoolQuery().
			Filter(elastic.
				NewTermQuery("meta.process.keyword", filter.Process))
		svc.Query(q)
	}

	// run search.
	sr, err := svc.Do(ctx)
	if err != nil {
		return nil, err
	}

	// construct result using pagination.
	res := []string{}
	if agg, found := sr.Aggregations.Terms("mapIds"); found {
		for _, bucket := range agg.Buckets {
			res = append(res, bucket.Key.(string))
		}
	}
	return filter.PaginateStrings(res), nil
}

func makeFilterQueries(filter *store.SegmentFilter) []elastic.Query {
	// prepare filter queries.
	filterQueries := []elastic.Query{}

	// prevLinkHash filter.
	if filter.PrevLinkHash != nil {
		q := elastic.NewTermQuery("meta.prevLinkHash.keyword", *filter.PrevLinkHash)
		filterQueries = append(filterQueries, q)
	}

	// process filter.
	if filter.Process != "" {
		q := elastic.NewTermQuery("meta.process.keyword", filter.Process)
		filterQueries = append(filterQueries, q)
	}

	// mapIds filter.
	if len(filter.MapIDs) > 0 {
		termQueries := []elastic.Query{}
		for _, x := range filter.MapIDs {
			q := elastic.NewTermQuery("meta.mapId.keyword", x)
			termQueries = append(termQueries, q)
		}
		shouldQuery := elastic.NewBoolQuery().Should(termQueries...)
		filterQueries = append(filterQueries, shouldQuery)
	}

	// tags filter.
	if len(filter.Tags) > 0 {
		termQueries := []elastic.Query{}
		for _, x := range filter.Tags {
			q := elastic.NewTermQuery("meta.tags.keyword", x)
			termQueries = append(termQueries, q)
		}
		shouldQuery := elastic.NewBoolQuery().Must(termQueries...)
		filterQueries = append(filterQueries, shouldQuery)
	}

	// linkHashes filter.
	if len(filter.LinkHashes) > 0 {
		q := elastic.NewIdsQuery(docType).Ids(filter.LinkHashes...)
		filterQueries = append(filterQueries, q)
	}

	return filterQueries
}

func (es *ESStore) genericSearch(filter *store.SegmentFilter, q elastic.Query) (cs.SegmentSlice, error) {
	// Flush to make sure the documents got written.
	ctx := context.TODO()
	_, err := es.client.Flush().Index(linksIndex).Do(ctx)
	if err != nil {
		return nil, err
	}

	// prepare search service.
	svc := es.client.
		Search().
		Index(linksIndex).
		Type(docType)

	// add pagination.
	svc = svc.
		From(filter.Pagination.Offset).
		Size(filter.Pagination.Limit)

	// run search.
	sr, err := svc.Query(q).Do(ctx)
	if err != nil {
		return nil, err
	}

	// populate SegmentSlice.
	res := cs.SegmentSlice{}
	if sr == nil || sr.TotalHits() == 0 {
		return res, nil
	}

	for _, hit := range sr.Hits.Hits {
		var link cs.Link
		if err := json.Unmarshal(*hit.Source, &link); err != nil {
			return nil, err
		}
		res = append(res, es.segmentify(ctx, &link))
	}

	sort.Sort(res)

	return res, nil
}

func (es *ESStore) findSegments(filter *store.SegmentFilter) (cs.SegmentSlice, error) {
	// prepare query.
	q := elastic.NewBoolQuery().Filter(makeFilterQueries(filter)...)

	// run search.
	return es.genericSearch(filter, q)
}

func (es *ESStore) simpleSearchQuery(query *SearchQuery) (cs.SegmentSlice, error) {
	// prepare Query.
	q := elastic.NewBoolQuery().
		// add filter queries.
		Filter(makeFilterQueries(&query.SegmentFilter)...).
		// add simple search query.
		Must(elastic.NewSimpleQueryStringQuery(query.Query))

	// run search.
	return es.genericSearch(&query.SegmentFilter, q)
}

func (es *ESStore) multiMatchQuery(query *SearchQuery) (cs.SegmentSlice, error) {
	// fields to search through: all meta + stateTokens.
	fields := []string{
		"meta.mapId",
		"meta.process",
		"meta.action",
		"meta.type",
		"meta.tags",
		"meta.prevLinkHash",
		"stateTokens",
	}

	// prepare Query.
	q := elastic.NewMultiMatchQuery(query.Query, fields...).Type("best_fields")

	// run search.
	return es.genericSearch(&query.SegmentFilter, q)
}
