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

// Package dummystore implements a store that saves all the segments in memory.
//
// It can be used for testing, but it's unoptimized and not designed for
// production.
package dummystore

import (
	"fmt"
	"sort"
	"sync"

	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/store"
	"github.com/stratumn/sdk/types"
)

const (
	// Name is the name set in the store's information.
	Name = "dummy"

	// Description is the description set in the store's information.
	Description = "Stratumn Dummy Store"
)

// Config contains configuration options for the store.
type Config struct {
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

// DummyStore is the type that implements github.com/stratumn/sdk/store.Adapter.
type DummyStore struct {
	config       *Config
	didSaveChans []chan *cs.Segment
	eventChans   []chan *store.Event
	segments     segmentMap   // maps link hashes to segments
	values       valueMap     // maps keys to values
	maps         hashSetMap   // maps chains IDs to sets of link hashes
	mutex        sync.RWMutex // simple global mutex
}

type segmentMap map[string]*cs.Segment
type hashSet map[string]struct{}
type hashSetMap map[string]hashSet
type valueMap map[string][]byte

// New creates an instance of a DummyStore.
func New(config *Config) *DummyStore {
	return &DummyStore{
		config,
		nil,
		nil,
		segmentMap{},
		valueMap{},
		hashSetMap{},
		sync.RWMutex{},
	}
}

// GetInfo implements github.com/stratumn/sdk/store.Adapter.GetInfo.
func (a *DummyStore) GetInfo() (interface{}, error) {
	return &Info{
		Name:        Name,
		Description: Description,
		Version:     a.config.Version,
		Commit:      a.config.Commit,
	}, nil
}

// AddDidSaveChannel implements
// github.com/stratumn/sdk/fossilizer.Store.AddDidSaveChannel.
func (a *DummyStore) AddDidSaveChannel(saveChan chan *cs.Segment) {
	a.didSaveChans = append(a.didSaveChans, saveChan)
}

// AddStoreEventChannel implements github.com/stratumn/sdk/store.AdapterV2.AddStoreEventChannel
func (a *DummyStore) AddStoreEventChannel(eventChan chan *store.Event) {
	a.eventChans = append(a.eventChans, eventChan)
}

/********** Store writer implementation **********/

// CreateLink implements github.com/stratumn/sdk/store.LinkWriter.CreateLink.
func (a *DummyStore) CreateLink(link *cs.Link) (*types.Bytes32, error) {
	return nil, nil
}

// AddEvidence implements github.com/stratumn/sdk/store.EvidenceWriter.AddEvidence.
func (a *DummyStore) AddEvidence(linkHash *types.Bytes32, evidence *cs.Evidence) error {
	return nil
}

// SaveSegment implements github.com/stratumn/sdk/store.Adapter.SaveSegment.
func (a *DummyStore) SaveSegment(segment *cs.Segment) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	return a.saveSegment(segment)
}

func (a *DummyStore) saveSegment(segment *cs.Segment) (err error) {
	linkHashStr := segment.GetLinkHashString()
	curr := a.segments[linkHashStr]
	mapID := segment.Link.GetMapID()

	if curr != nil {
		if segment, err = curr.MergeMeta(segment); err != nil {
			return err
		}
	}

	_, exists := a.maps[mapID]
	if !exists {
		a.maps[mapID] = hashSet{}
	}

	a.segments[linkHashStr] = segment
	a.maps[mapID][linkHashStr] = struct{}{}

	for _, c := range a.didSaveChans {
		c <- segment
	}

	return nil
}

// DeleteSegment implements github.com/stratumn/sdk/store.Adapter.DeleteSegment.
func (a *DummyStore) DeleteSegment(linkHash *types.Bytes32) (*cs.Segment, error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	return a.deleteSegment(linkHash)
}

func (a *DummyStore) deleteSegment(linkHash *types.Bytes32) (*cs.Segment, error) {
	linkHashStr := linkHash.String()
	segment, exists := a.segments[linkHashStr]
	if !exists {
		return nil, nil
	}

	delete(a.segments, linkHashStr)
	delete(a.maps[segment.Link.GetMapID()], linkHashStr)
	if len(a.maps[segment.Link.GetMapID()]) == 0 {
		delete(a.maps, segment.Link.GetMapID())
	}

	return segment, nil
}

/********** Store reader implementation **********/

// GetSegment implements github.com/stratumn/sdk/store.Adapter.GetSegment.
func (a *DummyStore) GetSegment(linkHash *types.Bytes32) (*cs.Segment, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	return a.segments[linkHash.String()], nil
}

// FindSegments implements github.com/stratumn/sdk/store.Adapter.FindSegments.
func (a *DummyStore) FindSegments(filter *store.SegmentFilter) (cs.SegmentSlice, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	var linkHashes = hashSet{}

	if len(filter.MapIDs) == 0 || filter.PrevLinkHash != nil {
		for linkHash := range a.segments {
			linkHashes[linkHash] = struct{}{}
		}
	} else {
		for _, mapID := range filter.MapIDs {
			l, e := a.maps[mapID]
			if e {
				for k, v := range l {
					linkHashes[k] = v
				}
			}
		}
	}

	return a.findHashesSegments(linkHashes, filter)
}

// GetMapIDs implements github.com/stratumn/sdk/store.Adapter.GetMapIDs.
func (a *DummyStore) GetMapIDs(filter *store.MapFilter) ([]string, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	mapIDs := make([]string, 0, len(a.maps))
	for mapID, linkHashes := range a.maps {
		for linkHash := range linkHashes {
			if segment, exist := a.segments[linkHash]; exist && filter.Match(segment) {
				mapIDs = append(mapIDs, mapID)
				break
			}
		}
	}

	sort.Strings(mapIDs)
	return filter.Pagination.PaginateStrings(mapIDs), nil
}

// GetEvidences implements github.com/stratumn/sdk/store.EvidenceReader.GetEvidences.
func (a *DummyStore) GetEvidences(linkHash *types.Bytes32) (*cs.Evidences, error) {
	return nil, nil
}

/********** github.com/stratumn/sdk/store.KeyValueStore implementation **********/

// GetValue implements github.com/stratumn/sdk/store.KeyValueStore.GetValue.
func (a *DummyStore) GetValue(key []byte) ([]byte, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	return a.values[createKey(key)], nil
}

// SetValue implements github.com/stratumn/sdk/store.KeyValueStore.SetValue.
func (a *DummyStore) SetValue(key, value []byte) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	return a.setValue(key, value)
}

// SaveValue implements github.com/stratumn/sdk/store.Adapter.SaveValue.
func (a *DummyStore) SaveValue(key, value []byte) error {
	return a.SetValue(key, value)
}

func (a *DummyStore) setValue(key, value []byte) error {
	k := createKey(key)
	a.values[k] = value

	return nil
}

// DeleteValue implements github.com/stratumn/sdk/store.KeyValueStore.DeleteValue.
func (a *DummyStore) DeleteValue(key []byte) ([]byte, error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	return a.deleteValue(key)
}

func (a *DummyStore) deleteValue(key []byte) ([]byte, error) {
	k := createKey(key)

	value, exists := a.values[k]
	if !exists {
		return nil, nil
	}

	delete(a.values, k)

	return value, nil
}

/********** github.com/stratumn/sdk/store.Batch implementation **********/

// NewBatch implements github.com/stratumn/sdk/store.Adapter.NewBatch.
func (a *DummyStore) NewBatch() (store.Batch, error) {
	return NewBatch(a), nil
}

// NewBatchV2 implements github.com/stratumn/sdk/store.AdapterV2.NewBatchV2.
func (a *DummyStore) NewBatchV2() (store.BatchV2, error) {
	return nil, nil
}

/********** Utilities **********/

func (a *DummyStore) findHashesSegments(linkHashes hashSet, filter *store.SegmentFilter) (cs.SegmentSlice, error) {
	var segments cs.SegmentSlice

	for linkHash := range linkHashes {
		segment := a.segments[linkHash]
		if filter.Match(segment) {
			segments = append(segments, segment)
		}
	}

	sort.Sort(segments)

	return filter.Pagination.PaginateSegments(segments), nil
}

func createKey(k []byte) string {
	return fmt.Sprintf("%x", k)
}
