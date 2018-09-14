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

	"github.com/olivere/elastic"
	"github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/types"
)

const (
	linksIndex     = "links"
	evidencesIndex = "evidences"
	valuesIndex    = "values"

	// This is the mapping for the links index.
	// We voluntarily disable indexing of some fields, such as:
	// meta.refs, meta.data, data, signatures, etc.
	linksMapping = `{
		"mappings": {
			"_doc": {
				"properties": {
					"version": {
						"enabled": false
					},
					"meta": {
						"properties": {
							"clientId": {
								"enabled": false
							},
							"mapId": {
								"type": "text",
								"fields": {
									"keyword": {
										"type": "keyword"
									}
								}
							},
							"process": {
								"properties": {
									"name": {
										"type": "text",
										"fields": {
											"keyword": {
												"type": "keyword"
											}
										}
									},
									"state": {
										"type": "text",
										"fields": {
											"keyword": {
												"type": "keyword"
											}
										}
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
							"priority": {
								"enabled": false
							},
							"prevLinkHash": {
								"enabled": false
							},
							"step": {
								"type": "text",
								"fields": {
									"keyword": {
										"type": "keyword"
									}
								}
							},
							"tags": {
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
							"data": {
								"enabled": false
							}
						}
					},
					"data": {
						"enabled": false
					},
					"signatures": {
						"enabled": false
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
					"dataTokens": {
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

// Evidences is a wrapper around types.EvidenceSlice for json ElasticSearch serialization compliance.
// Elastic Search does not allow indexing of arrays directly.
type Evidences struct {
	Evidences types.EvidenceSlice `json:"evidences,omitempty"`
}

// Value is a wrapper struct for the value of the keyvalue store part.
// Elastic only accepts json structured objects.
type Value struct {
	Value []byte `json:"value,omitempty"`
}

type linkDoc struct {
	chainscript.Link
	Priority     float64  `json:"priority"`
	PrevLinkHash string   `json:"prevLinkHash"`
	DataTokens   []string `json:"dataTokens"`
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
		o.DataTokens = append(o.DataTokens, value)
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

func fromLink(link *chainscript.Link) (*linkDoc, error) {
	doc := linkDoc{
		Link:       *link,
		Priority:   link.Meta.Priority,
		DataTokens: []string{},
	}

	if len(link.PrevLinkHash()) > 0 {
		doc.PrevLinkHash = link.PrevLinkHash().String()
	}

	if len(link.Data) > 0 {
		var strData string
		err := link.StructurizeData(&strData)
		if err == nil {
			doc.extractTokens(strData)
		}

		var objData map[string]interface{}
		err = link.StructurizeData(&objData)
		if err == nil {
			doc.extractTokens(objData)
		}
	}

	return &doc, nil
}

func (es *ESStore) createLink(ctx context.Context, link *chainscript.Link) (chainscript.LinkHash, error) {
	linkHash, err := link.Hash()
	if err != nil {
		return nil, err
	}
	linkHashStr := linkHash.String()

	has, err := es.hasDocument(ctx, linksIndex, linkHashStr)
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

	return linkHash, es.indexDocument(ctx, linksIndex, linkHashStr, linkDoc)
}

func (es *ESStore) hasDocument(ctx context.Context, indexName, id string) (bool, error) {
	return es.client.Exists().Index(indexName).Type(docType).Id(id).Do(ctx)
}

func (es *ESStore) indexDocument(ctx context.Context, indexName, id string, doc interface{}) error {
	_, err := es.client.Index().Index(indexName).Type(docType).Id(id).BodyJson(doc).Do(ctx)
	return err
}

func (es *ESStore) getDocument(ctx context.Context, indexName, id string) (*json.RawMessage, error) {
	has, err := es.hasDocument(ctx, indexName, id)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}

	get, err := es.client.Get().Index(indexName).Type(docType).Id(id).Do(ctx)
	if err != nil {
		return nil, err
	}
	if !get.Found {
		return nil, nil
	}

	return get.Source, nil
}

func (es *ESStore) deleteDocument(ctx context.Context, indexName, id string) error {
	_, err := es.client.Delete().Index(indexName).Type(docType).Id(id).Do(ctx)
	return err
}

func (es *ESStore) getLink(ctx context.Context, id string) (*chainscript.Link, error) {
	var link linkDoc
	jsn, err := es.getDocument(ctx, linksIndex, id)
	if err != nil {
		return nil, err
	}
	if jsn == nil {
		return nil, nil
	}
	err = json.Unmarshal(*jsn, &link)
	return &link.Link, err
}

func (es *ESStore) getEvidences(ctx context.Context, id string) (types.EvidenceSlice, error) {
	jsn, err := es.getDocument(ctx, evidencesIndex, id)
	if err != nil {
		return nil, err
	}
	evidences := Evidences{Evidences: types.EvidenceSlice{}}
	if jsn != nil {
		err = json.Unmarshal(*jsn, &evidences)
	}
	return evidences.Evidences, err
}

func (es *ESStore) addEvidence(ctx context.Context, linkHash string, evidence *chainscript.Evidence) error {
	currentDoc, err := es.getEvidences(ctx, linkHash)
	if err != nil {
		return err
	}

	if err := currentDoc.AddEvidence(evidence); err != nil {
		return err
	}

	evidences := Evidences{
		Evidences: currentDoc,
	}

	return es.indexDocument(ctx, evidencesIndex, linkHash, &evidences)
}

func (es *ESStore) getValue(ctx context.Context, key string) ([]byte, error) {
	var value Value
	jsn, err := es.getDocument(ctx, valuesIndex, key)
	if err != nil {
		return nil, err
	}
	if jsn != nil {
		err = json.Unmarshal(*jsn, &value)
	}
	return value.Value, err
}

func (es *ESStore) setValue(ctx context.Context, key string, value []byte) error {
	v := Value{
		Value: value,
	}
	return es.indexDocument(ctx, valuesIndex, key, v)
}

func (es *ESStore) deleteValue(ctx context.Context, key string) ([]byte, error) {
	value, err := es.getValue(ctx, key)
	if err != nil {
		return nil, err
	}

	if value == nil {
		return nil, nil
	}

	return value, es.deleteDocument(ctx, valuesIndex, key)
}

func (es *ESStore) segmentify(ctx context.Context, link *chainscript.Link) *chainscript.Segment {
	segment, err := link.Segmentify()
	if err != nil {
		panic(err)
	}

	evidences, err := es.GetEvidences(ctx, segment.LinkHash())
	if evidences != nil && err == nil {
		segment.Meta.Evidences = evidences
	}
	return segment
}

func (es *ESStore) getMapIDs(ctx context.Context, filter *store.MapFilter) ([]string, error) {
	// Flush to make sure the documents got written.
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

	filterQueries := []elastic.Query{}

	if filter.Process != "" {
		q := elastic.NewTermQuery("meta.process.name.keyword", filter.Process)
		filterQueries = append(filterQueries, q)
	}

	if filter.Prefix != "" {
		q := elastic.NewPrefixQuery("meta.mapId.keyword", filter.Prefix)
		filterQueries = append(filterQueries, q)
	}
	if filter.Suffix != "" {
		// There is no efficient way to do suffix filter in ES better than regex filter.
		q := elastic.NewRegexpQuery("meta.mapId", fmt.Sprintf(".*%s", filter.Suffix))
		filterQueries = append(filterQueries, q)
	}

	if len(filterQueries) > 0 {
		svc.Query(elastic.NewBoolQuery().Filter(filterQueries...))
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
	if filter.WithoutParent {
		q := elastic.NewTermQuery("prevLinkHash.keyword", "")
		filterQueries = append(filterQueries, q)
	} else if len(filter.PrevLinkHash) > 0 {
		q := elastic.NewTermQuery("prevLinkHash.keyword", filter.PrevLinkHash.String())
		filterQueries = append(filterQueries, q)
	}

	// process filter.
	if filter.Process != "" {
		q := elastic.NewTermQuery("meta.process.name.keyword", filter.Process)
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
		lhs := make([]string, len(filter.LinkHashes))
		for i, lh := range filter.LinkHashes {
			lhs[i] = lh.String()
		}

		q := elastic.NewIdsQuery(docType).Ids(lhs...)
		filterQueries = append(filterQueries, q)
	}

	return filterQueries
}

func (es *ESStore) genericSearch(ctx context.Context, filter *store.SegmentFilter, q elastic.Query) (*types.PaginatedSegments, error) {
	// Flush to make sure the documents got written.
	_, err := es.client.Flush().Index(linksIndex).Do(ctx)
	if err != nil {
		return nil, err
	}

	// prepare search service.
	svc := es.client.
		Search(linksIndex).
		Type(docType)

	// add pagination.
	svc = svc.
		Sort("priority", filter.Reverse).
		From(filter.Pagination.Offset).
		Size(filter.Pagination.Limit)

	// run search.
	sr, err := svc.Query(q).Do(ctx)
	if err != nil {
		return nil, err
	}

	// populate SegmentSlice.
	if sr == nil || sr.TotalHits() == 0 {
		return &types.PaginatedSegments{}, nil
	}

	res := &types.PaginatedSegments{
		Segments:   types.SegmentSlice{},
		TotalCount: int(sr.TotalHits()),
	}

	for _, hit := range sr.Hits.Hits {
		var link chainscript.Link
		if err := json.Unmarshal(*hit.Source, &link); err != nil {
			return nil, err
		}
		res.Segments = append(res.Segments, es.segmentify(ctx, &link))
	}

	res.Segments.Sort(filter.Reverse)

	return res, nil
}

func (es *ESStore) findSegments(ctx context.Context, filter *store.SegmentFilter) (*types.PaginatedSegments, error) {
	// prepare query.
	q := elastic.NewBoolQuery().Filter(makeFilterQueries(filter)...)

	// run search.
	return es.genericSearch(ctx, filter, q)
}

func (es *ESStore) simpleSearchQuery(ctx context.Context, query *SearchQuery) (*types.PaginatedSegments, error) {
	// prepare Query.
	q := elastic.NewBoolQuery().
		// add filter queries.
		Filter(makeFilterQueries(&query.SegmentFilter)...).
		// add simple search query.
		Must(elastic.NewSimpleQueryStringQuery(query.Query))

	// run search.
	return es.genericSearch(ctx, &query.SegmentFilter, q)
}

func (es *ESStore) multiMatchQuery(ctx context.Context, query *SearchQuery) (*types.PaginatedSegments, error) {
	// fields to search through: all meta + dataTokens.
	fields := []string{
		"meta.mapId",
		"meta.process.name",
		"meta.action",
		"meta.step",
		"meta.tags",
		"prevLinkHash",
		"dataTokens",
	}

	// prepare Query.
	q := elastic.NewMultiMatchQuery(query.Query, fields...).Type("best_fields")

	// run search.
	return es.genericSearch(ctx, &query.SegmentFilter, q)
}
