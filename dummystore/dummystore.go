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
	"sort"
	"sync"

	"github.com/stratumn/go/cs"
	"github.com/stratumn/go/store"
	"github.com/stratumn/go/types"
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

// DummyStore is the type that implements github.com/stratumn/go/store.Adapter.
type DummyStore struct {
	config       *Config
	didSaveChans []chan *cs.Segment
	segments     segmentMap   // maps link hashes to segments
	maps         hashSetMap   // maps chains IDs to sets of link hashes
	mutex        sync.RWMutex // simple global mutex
}

type segmentMap map[string]*cs.Segment
type hashSet map[string]struct{}
type hashSetMap map[string]hashSet

// New creates an instance of a DummyStore.
func New(config *Config) *DummyStore {
	return &DummyStore{
		config,
		nil,
		segmentMap{},
		hashSetMap{},
		sync.RWMutex{},
	}
}

// GetInfo implements github.com/stratumn/go/store.Adapter.GetInfo.
func (a *DummyStore) GetInfo() (interface{}, error) {
	return &Info{
		Name:        Name,
		Description: Description,
		Version:     a.config.Version,
		Commit:      a.config.Commit,
	}, nil
}

// AddDidSaveChannel implements
// github.com/stratumn/go/fossilizer.Store.AddDidSaveChannel.
func (a *DummyStore) AddDidSaveChannel(saveChan chan *cs.Segment) {
	a.didSaveChans = append(a.didSaveChans, saveChan)
}

// SaveSegment implements github.com/stratumn/go/store.Adapter.SaveSegment.
func (a *DummyStore) SaveSegment(segment *cs.Segment) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

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

	// Send saved segment to all the save channels without blocking.
	go func(chans []chan *cs.Segment) {
		for _, c := range chans {
			c <- segment
		}
	}(a.didSaveChans)

	return nil
}

// GetSegment implements github.com/stratumn/go/store.Adapter.GetSegment.
func (a *DummyStore) GetSegment(linkHash *types.Bytes32) (*cs.Segment, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	return a.segments[linkHash.String()], nil
}

// DeleteSegment implements github.com/stratumn/go/store.Adapter.DeleteSegment.
func (a *DummyStore) DeleteSegment(linkHash *types.Bytes32) (*cs.Segment, error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	linkHashStr := linkHash.String()
	segment, exists := a.segments[linkHashStr]
	if !exists {
		return nil, nil
	}

	delete(a.segments, linkHashStr)
	delete(a.maps[segment.Link.GetMapID()], linkHashStr)

	return segment, nil
}

// FindSegments implements github.com/stratumn/go/store.Adapter.FindSegments.
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

// GetMapIDs implements github.com/stratumn/go/store.Adapter.GetMapIDs.
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
			tags := segment.Link.GetTags()
			if len(tags) > 0 {
				for _, tag := range filter.Tags {
					if !containsString(tags, tag) {
						continue HASH_LOOP
					}
				}
			} else {
				continue HASH_LOOP
			}
		}

		segments = append(segments, segment)
	}

	sort.Sort(segments)

	return filter.Pagination.PaginateSegments(segments), nil
}

func containsString(a []string, s string) bool {
	for _, v := range a {
		if v == s {
			return true
		}
	}
	return false
}
