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

package couchstore

import (
	"encoding/hex"
	"encoding/json"
	"sort"

	"github.com/pkg/errors"

	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/store"
	"github.com/stratumn/sdk/types"
)

const (
	// Name is the name set in the store's information.
	Name = "CouchDB"

	// Description is the description set in the store's information.
	Description = "Stratumn CouchDB Store"
)

// CouchStore is the type that implements github.com/stratumn/sdk/store.Adapter.
type CouchStore struct {
	config       *Config
	didSaveChans []chan *cs.Segment
	eventChans   []chan *store.Event
}

// Config contains configuration options for the store.
type Config struct {
	// Adress is CouchDB api end point.
	Address string

	// A version string that will be set in the store's information.
	Version string

	// A git commit hash that will be set in the store's information.
	Commit string
}

// Info is the info returned by GetInfo.
type Info struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Commit      string `json:"commit"`
}

// New creates an instance of a CouchStore.
func New(config *Config) (*CouchStore, error) {
	couchstore := &CouchStore{
		config: config,
	}
	_, couchResponseStatus, err := couchstore.get("/")
	if err != nil {
		return nil, errors.Errorf("No CouchDB running on %v", config.Address)

	}
	if couchResponseStatus.Ok == false {
		return nil, couchResponseStatus.error()
	}

	if err := couchstore.createDatabase(dbLink); err != nil {
		return nil, err
	}
	if err := couchstore.createDatabase(dbEvidences); err != nil {
		return nil, err
	}
	if err := couchstore.createDatabase(dbValue); err != nil {
		return nil, err
	}

	return couchstore, nil
}

// GetInfo implements github.com/stratumn/sdk/store.Adapter.GetInfo.
func (c *CouchStore) GetInfo() (interface{}, error) {
	return &Info{
		Name:        Name,
		Description: Description,
		Version:     c.config.Version,
		Commit:      c.config.Commit,
	}, nil
}

// AddDidSaveChannel implements github.com/stratumn/sdk/fossilizer.Store.AddDidSaveChannel.
func (c *CouchStore) AddDidSaveChannel(saveChan chan *cs.Segment) {
	c.didSaveChans = append(c.didSaveChans, saveChan)
}

// AddStoreEventChannel implements github.com/stratumn/sdk/store.AdapterV2.AddStoreEventChannel
func (c *CouchStore) AddStoreEventChannel(eventChan chan *store.Event) {
	c.eventChans = append(c.eventChans, eventChan)
}

func (c *CouchStore) notifyEvent(eventType store.EventType, details interface{}) {
	for _, c := range c.eventChans {
		c <- &store.Event{
			EventType: eventType,
			Details:   details,
		}
	}
}

/********** Store writer implementation **********/

// CreateLink implements github.com/stratumn/sdk/store.LinkWriter.CreateLink.
func (c *CouchStore) CreateLink(link *cs.Link) (*types.Bytes32, error) {
	linkHash, err := c.createLink(link)
	if err != nil {
		return nil, err
	}
	c.notifyEvent(store.SavedLink, link)
	return linkHash, nil
}

// AddEvidence implements github.com/stratumn/sdk/store.EvidenceWriter.AddEvidence.
func (c *CouchStore) AddEvidence(linkHash *types.Bytes32, evidence *cs.Evidence) error {
	if err := c.addEvidence(linkHash.String(), evidence); err != nil {
		return err
	}
	c.notifyEvent(store.SavedEvidence, evidence)
	return nil
}

// SaveSegment implements github.com/stratumn/sdk/store.Adapter.SaveSegment.
func (c *CouchStore) SaveSegment(segment *cs.Segment) error {
	linkHash, err := c.createLink(&segment.Link)
	if err != nil {
		return err
	}

	for _, evidence := range segment.Meta.Evidences {
		if err := c.AddEvidence(linkHash, evidence); err != nil {
			return err
		}
	}

	for _, ch := range c.didSaveChans {
		ch <- segment
	}

	return nil
}

// DeleteSegment implements github.com/stratumn/sdk/store.Adapter.DeleteSegment.
func (c *CouchStore) DeleteSegment(linkHash *types.Bytes32) (*cs.Segment, error) {
	segment, err := c.GetSegment(linkHash)
	if err != nil || segment == nil {
		return nil, err
	}

	_, evidenceErr := c.deleteDocument(dbEvidences, linkHash.String())
	segmentDoc, err := c.deleteDocument(dbLink, linkHash.String())
	if err != nil || segmentDoc == nil {
		return nil, err
	}

	return segment, evidenceErr
}

/********** Store reader implementation **********/

// GetSegment implements github.com/stratumn/sdk/store.Adapter.GetSegment.
func (c *CouchStore) GetSegment(linkHash *types.Bytes32) (*cs.Segment, error) {
	linkDoc, err := c.getDocument(dbLink, linkHash.String())
	if err != nil || linkDoc == nil {
		return nil, err
	}
	return c.segmentify(linkDoc.Link), nil
}

// FindSegments implements github.com/stratumn/sdk/store.Adapter.FindSegments.
func (c *CouchStore) FindSegments(filter *store.SegmentFilter) (cs.SegmentSlice, error) {
	queryBytes, err := NewSegmentQuery(filter)
	if err != nil {
		return nil, err
	}

	body, couchResponseStatus, err := c.post("/"+dbLink+"/_find", queryBytes)
	if err != nil {
		return nil, err
	}

	if couchResponseStatus.Ok == false {
		return nil, couchResponseStatus.error()
	}

	couchFindResponse := &CouchFindResponse{}
	if err := json.Unmarshal(body, couchFindResponse); err != nil {
		return nil, err
	}

	segments := cs.SegmentSlice{}
	for _, doc := range couchFindResponse.Docs {
		segments = append(segments, c.segmentify(doc.Link))
	}
	sort.Sort(segments)

	return segments, nil
}

// GetMapIDs implements github.com/stratumn/sdk/store.Adapter.GetMapIDs.
func (c *CouchStore) GetMapIDs(filter *store.MapFilter) ([]string, error) {
	queryBytes, err := NewMapQuery(filter)
	if err != nil {
		return nil, err
	}

	body, couchResponseStatus, err := c.post("/"+dbLink+"/_find", queryBytes)
	if err != nil {
		return nil, err
	}

	if couchResponseStatus.Ok == false {
		return nil, couchResponseStatus.error()
	}

	couchFindResponse := &CouchFindResponse{}
	if err := json.Unmarshal(body, couchFindResponse); err != nil {
		return nil, err
	}

	mapIDs := []string{}
	for _, mapDoc := range couchFindResponse.Docs {
		mapIDs = append(mapIDs, mapDoc.ID)
	}

	return mapIDs, nil
}

// GetEvidences implements github.com/stratumn/sdk/store.EvidenceReader.GetEvidences.
func (c *CouchStore) GetEvidences(linkHash *types.Bytes32) (*cs.Evidences, error) {
	evidencesDoc, err := c.getDocument(dbEvidences, linkHash.String())
	if err != nil {
		return nil, err
	}
	if evidencesDoc == nil {
		return &cs.Evidences{}, nil
	}
	return evidencesDoc.Evidences, nil
}

/********** github.com/stratumn/sdk/store.KeyValueStore implementation **********/

// SaveValue implements github.com/stratumn/sdk/store.Adapter.SaveValue.
func (c *CouchStore) SaveValue(key, value []byte) error {
	hexKey := hex.EncodeToString(key)
	valueDoc, err := c.getDocument(dbValue, hexKey)
	if err != nil {
		return err
	}

	newValueDoc := Document{
		Value: value,
	}

	if valueDoc != nil {
		newValueDoc.Revision = valueDoc.Revision
	}

	return c.saveDocument(dbValue, hexKey, newValueDoc)
}

// GetValue implements github.com/stratumn/sdk/store.Adapter.GetValue.
func (c *CouchStore) GetValue(key []byte) ([]byte, error) {
	hexKey := hex.EncodeToString(key)
	valueDoc, err := c.getDocument(dbValue, hexKey)
	if err != nil {
		return nil, err
	}

	if valueDoc == nil {
		return nil, nil
	}

	return valueDoc.Value, nil
}

// DeleteValue implements github.com/stratumn/sdk/store.Adapter.DeleteValue.
func (c *CouchStore) DeleteValue(key []byte) ([]byte, error) {
	hexKey := hex.EncodeToString(key)
	valueDoc, err := c.deleteDocument(dbValue, hexKey)
	if err != nil {
		return nil, err
	}

	if valueDoc == nil {
		return nil, nil
	}

	return valueDoc.Value, nil
}

/********** github.com/stratumn/sdk/store.Batch implementation **********/

// NewBatch implements github.com/stratumn/sdk/store.Adapter.NewBatch.
func (c *CouchStore) NewBatch() (store.Batch, error) {
	return NewBatch(c), nil
}

// NewBatchV2 implements github.com/stratumn/sdk/store.AdapterV2.NewBatchV2.
func (c *CouchStore) NewBatchV2() (store.BatchV2, error) {
	return nil, nil
}
