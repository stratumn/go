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

	docType = "doc"
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

func (es *ESStore) createIndex(indexName string) error {
	exists, err := es.client.IndexExists(indexName).Do(*es.context)
	if err != nil {
		return err
	}
	if !exists {
		// TODO: pass mapping through BodyString.
		createIndex, err := es.client.CreateIndex(indexName).Do(*es.context)
		if err != nil {
			return err
		}
		if !createIndex.Acknowledged {
			// Not acknowledged.
			return fmt.Errorf("Error creating %s index", indexName)
		}
	}

	return nil
}

func (es *ESStore) deleteIndex(indexName string) error {
	del, err := es.client.DeleteIndex(indexName).Do(*es.context)
	if err != nil {
		return err
	}

	if !del.Acknowledged {
		return fmt.Errorf("Index %s was not deleted", indexName)
	}

	return nil
}

func (es *ESStore) notifyEvent(event *store.Event) {
	for _, c := range es.eventChans {
		c <- event
	}
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
		return nil, fmt.Errorf("Link is immutable, %s already exists", linkHashStr)
	}

	return linkHash, es.indexDocument(linksIndex, linkHashStr, link)
}

func (es *ESStore) hasDocument(indexName, id string) (bool, error) {
	return es.client.Exists().Index(indexName).Type(docType).Id(id).Do(*es.context)
}

func (es *ESStore) indexDocument(indexName, id string, doc interface{}) error {
	_, err := es.client.Index().Index(indexName).Type(docType).Id(id).BodyJson(doc).Do(*es.context)
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
	get, err := es.client.Get().Index(indexName).Type(docType).Id(id).Do(*es.context)
	if err != nil {
		return nil, err
	}
	if !get.Found {
		return nil, nil
	}

	return get.Source, nil
}

func (es *ESStore) deleteDocument(indexName, id string) error {
	_, err := es.client.Delete().Index(indexName).Type(docType).Id(id).Do(*es.context)
	return err
}

func (es *ESStore) getLink(id string) (*cs.Link, error) {
	var link cs.Link
	jsn, err := es.getDocument(linksIndex, id)
	if err != nil {
		return nil, err
	}
	if jsn == nil {
		return nil, nil
	}
	err = json.Unmarshal(*jsn, &link)
	return &link, err
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

func (es *ESStore) segmentify(link *cs.Link) *cs.Segment {
	segment := link.Segmentify()

	evidences, err := es.GetEvidences(segment.Meta.GetLinkHash())
	if evidences != nil && err == nil {
		segment.Meta.Evidences = *evidences
	}
	return segment
}

func (es *ESStore) getMapIDs(filter *store.MapFilter) ([]string, error) {
	// Flush to make sure the documents got written.
	_, err := es.client.Flush().Index(linksIndex).Do(*es.context)
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
	if len(filter.Process) > 0 {
		q := elastic.
			NewBoolQuery().
			Filter(elastic.
				NewTermQuery("meta.process.keyword", filter.Process))
		svc.Query(q)
	}

	// run search.
	sr, err := svc.Do(*es.context)
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

func (es *ESStore) findSegments(filter *store.SegmentFilter) (cs.SegmentSlice, error) {
	// Flush to make sure the documents got written.
	_, err := es.client.Flush().Index(linksIndex).Do(*es.context)
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

	// prepare filter queries.
	filterQueries := []elastic.Query{}

	// prevLinkHash filter.
	if filter.PrevLinkHash != nil {
		if len(*filter.PrevLinkHash) > 0 {
			q := elastic.NewTermQuery("meta.prevLinkHash.keyword", *filter.PrevLinkHash)
			filterQueries = append(filterQueries, q)
		} else {
			q := elastic.NewBoolQuery().MustNot(elastic.NewExistsQuery("meta.prevLinkHash"))
			filterQueries = append(filterQueries, q)
		}

	}

	// process filter.
	if len(filter.Process) > 0 {
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

	// make final query.
	q := elastic.NewBoolQuery().Filter(filterQueries...)

	// run search.
	sr, err := svc.Query(q).Do(*es.context)
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
		res = append(res, es.segmentify(&link))
	}

	sort.Sort(res)

	return res, nil
}
