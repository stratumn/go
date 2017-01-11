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

// Package tmstore implements a store that saves all the segments in a
// tendermint app
package tmstore

import (
	"errors"

	"encoding/json"

	"github.com/stratumn/go/cs"
	"github.com/stratumn/go/store"
	"github.com/stratumn/go/tmpop"
	"github.com/stratumn/go/types"
)

const (
	// Name is the name set in the store's information.
	Name = "tm"

	// Description is the description set in the store's information.
	Description = "Stratumn TM Store"

	// DefaultEndpoint is the default Tendermint endpoint
	DefaultEndpoint = "tcp://127.0.0.1:46657"
)

// TMStore is the type that implements github.com/stratumn/go/store.Adapter.
type TMStore struct {
	config   *Config
	tmClient *TMClient
}

// Config contains configuration options for the store.
type Config struct {
	// A version string that will be set in the store's information.
	Version string

	// A git commit hash that will be set in the store's information.
	Commit string

	// Endoint used to communicate with Tendermint core
	Endpoint string
}

// Info is the info returned by GetInfo.
type Info struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	TMAppInfo   interface{} `json:"tmAppDescription"`
	Version     string      `json:"version"`
	Commit      string      `json:"commit"`
}

// New creates a new instance of a TMStore
func New(config *Config) *TMStore {
	client := NewTMClient(config.Endpoint)

	return &TMStore{config, client}
}

// GetInfo implements github.com/stratumn/go/store.Adapter.GetInfo.
func (t *TMStore) GetInfo() (interface{}, error) {
	info := &tmpop.Info{}
	err := t.sendQuery("GetInfo", nil, info)

	return &Info{
		Name:        Name,
		Description: Description,
		TMAppInfo:   info,
		Version:     t.config.Version,
		Commit:      t.config.Commit,
	}, err
}

// SaveSegment implements github.com/stratumn/go/store.Adapter.SaveSegment.
func (t *TMStore) SaveSegment(segment *cs.Segment) error {
	tx, err := json.Marshal(segment)
	if err != nil {
		return err
	}

	_, err = t.tmClient.BroadcastTxCommit(tx)
	return err
}

// GetSegment implements github.com/stratumn/go/store.Adapter.GetSegment.
func (t *TMStore) GetSegment(linkHash *types.Bytes32) (segment *cs.Segment, err error) {
	segment = &cs.Segment{}
	err = t.sendQuery("GetSegment", linkHash, segment)

	// Return nil when no segment has been found (and not an empty segment)
	if segment.IsEmpty() {
		segment = nil
	}
	return
}

// DeleteSegment implements github.com/stratumn/go/store.Adapter.DeleteSegment.
func (t *TMStore) DeleteSegment(linkHash *types.Bytes32) (segment *cs.Segment, err error) {
	segment = &cs.Segment{}
	err = t.sendQuery("DeleteSegment", linkHash, segment)

	// Return nil when no segment has been deleted (and not an empty segment)
	if segment.IsEmpty() {
		segment = nil
	}
	return
}

// FindSegments implements github.com/stratumn/go/store.Adapter.FindSegments.
func (t *TMStore) FindSegments(filter *store.Filter) (segmentSlice cs.SegmentSlice, err error) {
	segmentSlice = make(cs.SegmentSlice, 0)
	err = t.sendQuery("FindSegments", filter, &segmentSlice)
	return
}

// GetMapIDs implements github.com/stratumn/go/store.Adapter.GetMapIDs.
func (t *TMStore) GetMapIDs(pagination *store.Pagination) (ids []string, err error) {
	ids = make([]string, 0)
	err = t.sendQuery("GetMapIDs", pagination, &ids)
	return
}

func (t *TMStore) sendQuery(name string, args interface{}, result interface{}) error {
	query, err := tmpop.BuildQueryBinary(name, args)
	if err != nil {
		return err
	}

	res, err := t.tmClient.ABCIQuery(query)
	if err != nil {
		return err
	}
	if res.Result.IsErr() {
		return errors.New(res.Result.Error())
	}

	err = json.Unmarshal(res.Result.Data, result)

	return err
}
