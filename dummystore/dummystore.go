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
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/stratumn/go-indigocore/bufferedbatch"
	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/types"
)

const (
	// Name is the name set in the store's information.
	Name = "dummy"

	// Description is the description set in the store's information.
	Description = "Indigo's Dummy Store"
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

// DummyStore is the type that implements github.com/stratumn/go-indigocore/store.Adapter.
type DummyStore struct {
	config     *Config
	eventChans []chan *store.Event
	links      linkMap      // maps link hashes to segments
	evidences  evidenceMap  // maps link hashes to evidences
	values     valueMap     // maps keys to values
	maps       hashSetMap   // maps chains IDs to sets of link hashes
	mutex      sync.RWMutex // simple global mutex
}

type linkMap map[string]*cs.Link
type evidenceMap map[string]*cs.Evidences
type hashSet map[string]struct{}
type hashSetMap map[string]hashSet
type valueMap map[string][]byte

// New creates an instance of a DummyStore.
func New(config *Config) *DummyStore {
	return &DummyStore{
		config,
		nil,
		linkMap{},
		evidenceMap{},
		valueMap{},
		hashSetMap{},
		sync.RWMutex{},
	}
}

// GetInfo implements github.com/stratumn/go-indigocore/store.Adapter.GetInfo.
func (a *DummyStore) GetInfo(ctx context.Context) (interface{}, error) {
	return &Info{
		Name:        Name,
		Description: Description,
		Version:     a.config.Version,
		Commit:      a.config.Commit,
	}, nil
}

// AddStoreEventChannel implements github.com/stratumn/go-indigocore/store.Adapter.AddStoreEventChannel
func (a *DummyStore) AddStoreEventChannel(eventChan chan *store.Event) {
	a.eventChans = append(a.eventChans, eventChan)
}

/********** Store writer implementation **********/

// CreateLink implements github.com/stratumn/go-indigocore/store.LinkWriter.CreateLink.
func (a *DummyStore) CreateLink(ctx context.Context, link *cs.Link) (*types.Bytes32, error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	return a.createLink(link)
}

func (a *DummyStore) createLink(link *cs.Link) (*types.Bytes32, error) {
	linkHash, err := link.Hash()
	if err != nil {
		return nil, err
	}

	linkHashStr := linkHash.String()
	a.links[linkHashStr] = link

	mapID := link.Meta.MapID
	_, exists := a.maps[mapID]
	if !exists {
		a.maps[mapID] = hashSet{}
	}

	a.maps[mapID][linkHashStr] = struct{}{}

	linkEvent := store.NewSavedLinks(link)

	for _, c := range a.eventChans {
		c <- linkEvent
	}

	return linkHash, nil
}

// AddEvidence implements github.com/stratumn/go-indigocore/store.EvidenceWriter.AddEvidence.
func (a *DummyStore) AddEvidence(ctx context.Context, linkHash *types.Bytes32, evidence *cs.Evidence) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if err := a.addEvidence(linkHash.String(), evidence); err != nil {
		return err
	}

	evidenceEvent := store.NewSavedEvidences()
	evidenceEvent.AddSavedEvidence(linkHash, evidence)

	for _, c := range a.eventChans {
		c <- evidenceEvent
	}

	return nil
}

func (a *DummyStore) addEvidence(linkHash string, evidence *cs.Evidence) error {
	currentEvidences := a.evidences[linkHash]
	if currentEvidences == nil {
		currentEvidences = &cs.Evidences{}
	}

	if err := currentEvidences.AddEvidence(*evidence); err != nil {
		return err
	}

	a.evidences[linkHash] = currentEvidences

	return nil
}

/********** Store reader implementation **********/

// GetSegment implements github.com/stratumn/go-indigocore/store.Adapter.GetSegment.
func (a *DummyStore) GetSegment(ctx context.Context, linkHash *types.Bytes32) (*cs.Segment, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	return a.getSegment(linkHash.String())
}

// GetSegment implements github.com/stratumn/go-indigocore/store.Adapter.GetSegment.
func (a *DummyStore) getSegment(linkHash string) (*cs.Segment, error) {
	link, exists := a.links[linkHash]
	if !exists {
		return nil, nil
	}

	segment := &cs.Segment{
		Link: *link,
		Meta: cs.SegmentMeta{
			Evidences: cs.Evidences{},
			LinkHash:  linkHash,
		},
	}

	evidences, exists := a.evidences[linkHash]
	if exists {
		segment.Meta.Evidences = *evidences
	}

	return segment, nil
}

// FindSegments implements github.com/stratumn/go-indigocore/store.Adapter.FindSegments.
func (a *DummyStore) FindSegments(ctx context.Context, filter *store.SegmentFilter) (cs.SegmentSlice, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	var linkHashes = hashSet{}

	if len(filter.MapIDs) == 0 || filter.PrevLinkHash != nil {
		for linkHash := range a.links {
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

// GetMapIDs implements github.com/stratumn/go-indigocore/store.Adapter.GetMapIDs.
func (a *DummyStore) GetMapIDs(ctx context.Context, filter *store.MapFilter) ([]string, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	mapIDs := make([]string, 0, len(a.maps))
	for mapID, linkHashes := range a.maps {
		for linkHash := range linkHashes {
			if link, exists := a.links[linkHash]; exists && filter.MatchLink(link) {
				mapIDs = append(mapIDs, mapID)
				break
			}
		}
	}

	sort.Strings(mapIDs)
	return filter.Pagination.PaginateStrings(mapIDs), nil
}

// GetEvidences implements github.com/stratumn/go-indigocore/store.EvidenceReader.GetEvidences.
func (a *DummyStore) GetEvidences(ctx context.Context, linkHash *types.Bytes32) (*cs.Evidences, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	evidences, _ := a.evidences[linkHash.String()]
	return evidences, nil
}

/********** github.com/stratumn/go-indigocore/store.KeyValueStore implementation **********/

// GetValue implements github.com/stratumn/go-indigocore/store.KeyValueStore.GetValue.
func (a *DummyStore) GetValue(ctx context.Context, key []byte) ([]byte, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	return a.values[createKey(key)], nil
}

// SetValue implements github.com/stratumn/go-indigocore/store.KeyValueStore.SetValue.
func (a *DummyStore) SetValue(ctx context.Context, key, value []byte) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	return a.setValue(key, value)
}

func (a *DummyStore) setValue(key, value []byte) error {
	k := createKey(key)
	a.values[k] = value

	return nil
}

// DeleteValue implements github.com/stratumn/go-indigocore/store.KeyValueStore.DeleteValue.
func (a *DummyStore) DeleteValue(ctx context.Context, key []byte) ([]byte, error) {
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

/********** github.com/stratumn/go-indigocore/store.Batch implementation **********/

// NewBatch implements github.com/stratumn/go-indigocore/store.Adapter.NewBatch.
func (a *DummyStore) NewBatch(ctx context.Context) (store.Batch, error) {
	return bufferedbatch.NewBatch(ctx, a), nil
}

/********** Utilities **********/

func (a *DummyStore) findHashesSegments(linkHashes hashSet, filter *store.SegmentFilter) (cs.SegmentSlice, error) {
	var segments cs.SegmentSlice

	for linkHash := range linkHashes {
		segment, err := a.getSegment(linkHash)
		if err != nil {
			return nil, err
		}

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
