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
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/stratumn/go-indigocore/bufferedbatch"
	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/types"
	"go.opencensus.io/trace"
)

const (
	// Name is the name set in the store's information.
	Name = "CouchDB"

	// Description is the description set in the store's information.
	Description = "Indigo's CouchDB Store"
)

// CouchStore is the type that implements github.com/stratumn/go-indigocore/store.Adapter.
type CouchStore struct {
	config     *Config
	eventChans []chan *store.Event
}

// CouchNotReadyError is returned when couchdb is not ready.
type CouchNotReadyError struct {
	originalError error
}

// Error implements error interface.
func (e *CouchNotReadyError) Error() string {
	return fmt.Sprintf("CouchDB not available: %v", e.originalError.Error())
}

// Config contains configuration options for the store.
type Config struct {
	// Address is CouchDB api end point.
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
		return nil, &CouchNotReadyError{originalError: err}
	}

	if couchResponseStatus.Ok == false {
		return nil, couchResponseStatus.error()
	}

	// required couchdb system database
	if err := couchstore.createDatabase("_users"); err != nil {
		return nil, err
	}

	// required couchdb system database
	if err := couchstore.createDatabase("_replicator"); err != nil {
		return nil, err
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

// GetInfo implements github.com/stratumn/go-indigocore/store.Adapter.GetInfo.
func (c *CouchStore) GetInfo(ctx context.Context) (interface{}, error) {
	ctx, span := trace.StartSpan(ctx, "couchstore/GetInfo")
	defer span.End()

	return &Info{
		Name:        Name,
		Description: Description,
		Version:     c.config.Version,
		Commit:      c.config.Commit,
	}, nil
}

// AddStoreEventChannel implements github.com/stratumn/go-indigocore/store.Adapter.AddStoreEventChannel
func (c *CouchStore) AddStoreEventChannel(eventChan chan *store.Event) {
	c.eventChans = append(c.eventChans, eventChan)
}

func (c *CouchStore) notifyEvent(event *store.Event) {
	for _, c := range c.eventChans {
		c <- event
	}
}

/********** Store writer implementation **********/

// CreateLink implements github.com/stratumn/go-indigocore/store.LinkWriter.CreateLink.
func (c *CouchStore) CreateLink(ctx context.Context, link *cs.Link) (*types.Bytes32, error) {
	ctx, span := trace.StartSpan(ctx, "couchstore/CreateLink")
	defer span.End()

	linkHash, err := c.createLink(link)
	if err != nil {
		return nil, err
	}

	linkEvent := store.NewSavedLinks(link)

	c.notifyEvent(linkEvent)

	return linkHash, nil
}

// AddEvidence implements github.com/stratumn/go-indigocore/store.EvidenceWriter.AddEvidence.
func (c *CouchStore) AddEvidence(ctx context.Context, linkHash *types.Bytes32, evidence *cs.Evidence) error {
	ctx, span := trace.StartSpan(ctx, "couchstore/AddEvidence")
	defer span.End()

	if err := c.addEvidence(linkHash.String(), evidence); err != nil {
		return err
	}

	evidenceEvent := store.NewSavedEvidences()
	evidenceEvent.AddSavedEvidence(linkHash, evidence)

	c.notifyEvent(evidenceEvent)

	return nil
}

/********** Store reader implementation **********/

// GetSegment implements github.com/stratumn/go-indigocore/store.Adapter.GetSegment.
func (c *CouchStore) GetSegment(ctx context.Context, linkHash *types.Bytes32) (*cs.Segment, error) {
	ctx, span := trace.StartSpan(ctx, "couchstore/GetSegment")
	defer span.End()

	linkDoc, err := c.getDocument(dbLink, linkHash.String())
	if err != nil || linkDoc == nil {
		return nil, err
	}
	return c.segmentify(ctx, linkDoc.Link), nil
}

// FindSegments implements github.com/stratumn/go-indigocore/store.Adapter.FindSegments.
func (c *CouchStore) FindSegments(ctx context.Context, filter *store.SegmentFilter) (cs.SegmentSlice, error) {
	ctx, span := trace.StartSpan(ctx, "couchstore/FindSegments")
	defer span.End()

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
		segments = append(segments, c.segmentify(ctx, doc.Link))
	}
	sort.Sort(segments)

	return segments, nil
}

// GetMapIDs implements github.com/stratumn/go-indigocore/store.Adapter.GetMapIDs.
func (c *CouchStore) GetMapIDs(ctx context.Context, filter *store.MapFilter) ([]string, error) {
	ctx, span := trace.StartSpan(ctx, "couchstore/GetMapIDs")
	defer span.End()

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

// GetEvidences implements github.com/stratumn/go-indigocore/store.EvidenceReader.GetEvidences.
func (c *CouchStore) GetEvidences(ctx context.Context, linkHash *types.Bytes32) (*cs.Evidences, error) {
	ctx, span := trace.StartSpan(ctx, "couchstore/GetEvidences")
	defer span.End()

	evidencesDoc, err := c.getDocument(dbEvidences, linkHash.String())
	if err != nil {
		return nil, err
	}
	if evidencesDoc == nil {
		return &cs.Evidences{}, nil
	}
	return evidencesDoc.Evidences, nil
}

/********** github.com/stratumn/go-indigocore/store.KeyValueStore implementation **********/

// SetValue implements github.com/stratumn/go-indigocore/store.KeyValueStore.SetValue.
func (c *CouchStore) SetValue(ctx context.Context, key, value []byte) error {
	ctx, span := trace.StartSpan(ctx, "couchstore/SetValue")
	defer span.End()

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

// GetValue implements github.com/stratumn/go-indigocore/store.Adapter.GetValue.
func (c *CouchStore) GetValue(ctx context.Context, key []byte) ([]byte, error) {
	ctx, span := trace.StartSpan(ctx, "couchstore/GetValue")
	defer span.End()

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

// DeleteValue implements github.com/stratumn/go-indigocore/store.Adapter.DeleteValue.
func (c *CouchStore) DeleteValue(ctx context.Context, key []byte) ([]byte, error) {
	ctx, span := trace.StartSpan(ctx, "couchstore/DeleteValue")
	defer span.End()

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

/********** github.com/stratumn/go-indigocore/store.Batch implementation **********/

// NewBatch implements github.com/stratumn/go-indigocore/store.Adapter.NewBatch.
func (c *CouchStore) NewBatch(ctx context.Context) (store.Batch, error) {
	return bufferedbatch.NewBatch(ctx, c), nil
}
