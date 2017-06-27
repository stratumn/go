// Copyright 2017 Stratumn SAS. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

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

// SaveSegment implements github.com/stratumn/sdk/store.Adapter.SaveSegment.
func (a *DummyStore) SaveSegment(segment *cs.Segment) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	return a.saveSegment(segment)
}

func (a *DummyStore) saveSegment(segment *cs.Segment) error {
	linkHashStr := segment.GetLinkHashString()
	curr := a.segments[linkHashStr]
	mapID := segment.Link.GetMapID()

	if curr != nil {
		currMapID := curr.Link.GetMapID()
		if currMapID != mapID {
			delete(a.maps[currMapID], linkHashStr)
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

// GetSegment implements github.com/stratumn/sdk/store.Adapter.GetSegment.
func (a *DummyStore) GetSegment(linkHash *types.Bytes32) (*cs.Segment, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	return a.segments[linkHash.String()], nil
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

// FindSegments implements github.com/stratumn/sdk/store.Adapter.FindSegments.
func (a *DummyStore) FindSegments(filter *store.Filter) (cs.SegmentSlice, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	var (
		linkHashes hashSet
		exists     bool
	)

	if filter.MapID == "" || filter.PrevLinkHash != nil {
		linkHashes = hashSet{}
		for linkHash := range a.segments {
			linkHashes[linkHash] = struct{}{}
		}
	} else {
		linkHashes, exists = a.maps[filter.MapID]
		if !exists {
			return cs.SegmentSlice{}, nil
		}
	}

	return a.findHashesSegments(linkHashes, filter)
}

// GetMapIDs implements github.com/stratumn/sdk/store.Adapter.GetMapIDs.
func (a *DummyStore) GetMapIDs(pagination *store.Pagination) ([]string, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	mapIDs := make([]string, len(a.maps))
	i := 0
	for mapID := range a.maps {
		mapIDs[i] = mapID
		i++
	}

	sort.Strings(mapIDs)
	return pagination.PaginateStrings(mapIDs), nil
}

// GetValue implements github.com/stratumn/sdk/store.Adapter.GetValue.
func (a *DummyStore) GetValue(key []byte) ([]byte, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	return a.values[createKey(key)], nil
}

// SaveValue implements github.com/stratumn/sdk/store.Adapter.SaveValue.
func (a *DummyStore) SaveValue(key, value []byte) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	return a.saveValue(key, value)
}

func (a *DummyStore) saveValue(key, value []byte) error {
	k := createKey(key)
	a.values[k] = value

	return nil
}

// DeleteValue implements github.com/stratumn/sdk/store.Adapter.DeleteValue.
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

// NewBatch implements github.com/stratumn/sdk/store.Adapter.NewBatch.
func (a *DummyStore) NewBatch() (store.Batch, error) {
	return NewBatch(a), nil
}

func (a *DummyStore) findHashesSegments(linkHashes hashSet, filter *store.Filter) (cs.SegmentSlice, error) {
	var segments cs.SegmentSlice

HASH_LOOP:
	for linkHash := range linkHashes {
		segment := a.segments[linkHash]

		if filter.PrevLinkHash != nil {
			prevLinkHash := segment.Link.GetPrevLinkHash()
			if prevLinkHash == nil || *filter.PrevLinkHash != *prevLinkHash {
				continue
			}
		}

		if len(filter.Tags) > 0 {
			tags := segment.Link.GetTagMap()
			for _, tag := range filter.Tags {
				if _, ok := tags[tag]; !ok {
					continue HASH_LOOP
				}
			}
		}

		segments = append(segments, segment)
	}

	sort.Sort(segments)

	return filter.Pagination.PaginateSegments(segments), nil
}

func createKey(k []byte) string {
	return fmt.Sprintf("%x", k)
}
