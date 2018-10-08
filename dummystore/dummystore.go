// Copyright 2017-2018 Stratumn SAS. All rights reserved.
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

	"github.com/pkg/errors"
	"github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-core/bufferedbatch"
	"github.com/stratumn/go-core/store"
	"github.com/stratumn/go-core/types"
)

const (
	// Name is the name set in the store's information.
	Name = "dummy"

	// Description is the description set in the store's information.
	Description = "Stratumn's Dummy Store"
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

// DummyStore is the type that implements github.com/stratumn/go-core/store.Adapter.
type DummyStore struct {
	config          *Config
	eventChans      []chan *store.Event
	links           linkMap           // maps link hashes to segments
	linksChildCount linkChildCountMap // maps link hashes to their children count
	evidences       evidenceMap       // maps link hashes to evidences
	values          valueMap          // maps keys to values
	maps            hashSetMap        // maps chains IDs to sets of link hashes
	mutex           sync.RWMutex      // simple global mutex
}

type linkMap map[string]*chainscript.Link
type linkChildCountMap map[string]int
type evidenceMap map[string]types.EvidenceSlice
type hashSet map[string]struct{}
type hashSetMap map[string]hashSet
type valueMap map[string][]byte

// New creates an instance of a DummyStore.
func New(config *Config) *DummyStore {
	return &DummyStore{
		config:          config,
		eventChans:      nil,
		links:           linkMap{},
		linksChildCount: linkChildCountMap{},
		evidences:       evidenceMap{},
		values:          valueMap{},
		maps:            hashSetMap{},
		mutex:           sync.RWMutex{},
	}
}

// GetInfo implements github.com/stratumn/go-core/store.Adapter.GetInfo.
func (a *DummyStore) GetInfo(ctx context.Context) (interface{}, error) {
	return &Info{
		Name:        Name,
		Description: Description,
		Version:     a.config.Version,
		Commit:      a.config.Commit,
	}, nil
}

// AddStoreEventChannel implements github.com/stratumn/go-core/store.Adapter.AddStoreEventChannel
func (a *DummyStore) AddStoreEventChannel(eventChan chan *store.Event) {
	a.eventChans = append(a.eventChans, eventChan)
}

/********** Store writer implementation **********/

// CreateLink implements github.com/stratumn/go-core/store.LinkWriter.CreateLink.
func (a *DummyStore) CreateLink(ctx context.Context, link *chainscript.Link) (chainscript.LinkHash, error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	return a.createLink(link)
}

func (a *DummyStore) createLink(link *chainscript.Link) (chainscript.LinkHash, error) {
	linkHash, err := link.Hash()
	if err != nil {
		return nil, err
	}

	linkHashStr := linkHash.String()
	_, ok := a.links[linkHashStr]
	if ok {
		return linkHash, nil
	}

	parentOk := a.canHaveNewChild(link.PrevLinkHash())
	if !parentOk {
		return linkHash, chainscript.ErrOutDegree
	}

	a.links[linkHashStr] = link
	a.incrementChildCount(link.PrevLinkHash())

	mapID := link.Meta.MapId
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

func (a *DummyStore) canHaveNewChild(linkHash chainscript.LinkHash) bool {
	if len(linkHash) == 0 {
		return true
	}

	linkHashStr := linkHash.String()
	link := a.links[linkHashStr]

	if link.Meta.OutDegree < 0 {
		return true
	}

	if link.Meta.OutDegree == 0 {
		return false
	}

	childCount, ok := a.linksChildCount[linkHashStr]
	if !ok {
		childCount = 0
	}

	return childCount < int(link.Meta.OutDegree)
}

func (a *DummyStore) incrementChildCount(linkHash chainscript.LinkHash) {
	linkHashStr := linkHash.String()
	_, ok := a.linksChildCount[linkHashStr]
	if !ok {
		a.linksChildCount[linkHashStr] = 0
	}

	a.linksChildCount[linkHashStr]++
}

// AddEvidence implements github.com/stratumn/go-core/store.EvidenceWriter.AddEvidence.
func (a *DummyStore) AddEvidence(ctx context.Context, linkHash chainscript.LinkHash, evidence *chainscript.Evidence) error {
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

func (a *DummyStore) addEvidence(linkHash string, evidence *chainscript.Evidence) error {
	currentEvidences := a.evidences[linkHash]
	if currentEvidences == nil {
		currentEvidences = types.EvidenceSlice{}
	}

	if err := currentEvidences.AddEvidence(evidence); err != nil {
		return err
	}

	a.evidences[linkHash] = currentEvidences

	return nil
}

/********** Store reader implementation **********/

// GetSegment implements github.com/stratumn/go-core/store.Adapter.GetSegment.
func (a *DummyStore) GetSegment(ctx context.Context, linkHash chainscript.LinkHash) (*chainscript.Segment, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	return a.getSegment(linkHash.String())
}

// GetSegment implements github.com/stratumn/go-core/store.Adapter.GetSegment.
func (a *DummyStore) getSegment(linkHash string) (*chainscript.Segment, error) {
	link, exists := a.links[linkHash]
	if !exists {
		return nil, nil
	}

	segment, err := link.Segmentify()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	evidences, exists := a.evidences[linkHash]
	if exists {
		segment.Meta.Evidences = evidences
	}

	return segment, nil
}

// FindSegments implements github.com/stratumn/go-core/store.Adapter.FindSegments.
func (a *DummyStore) FindSegments(ctx context.Context, filter *store.SegmentFilter) (*types.PaginatedSegments, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	var linkHashes = hashSet{}

	if len(filter.MapIDs) == 0 {
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

// GetMapIDs implements github.com/stratumn/go-core/store.Adapter.GetMapIDs.
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

// GetEvidences implements github.com/stratumn/go-core/store.EvidenceReader.GetEvidences.
func (a *DummyStore) GetEvidences(ctx context.Context, linkHash chainscript.LinkHash) (types.EvidenceSlice, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	evidences := a.evidences[linkHash.String()]
	return evidences, nil
}

/********** github.com/stratumn/go-core/store.KeyValueStore implementation **********/

// GetValue implements github.com/stratumn/go-core/store.KeyValueStore.GetValue.
func (a *DummyStore) GetValue(ctx context.Context, key []byte) ([]byte, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	return a.values[createKey(key)], nil
}

// SetValue implements github.com/stratumn/go-core/store.KeyValueStore.SetValue.
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

// DeleteValue implements github.com/stratumn/go-core/store.KeyValueStore.DeleteValue.
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

/********** github.com/stratumn/go-core/store.Batch implementation **********/

// NewBatch implements github.com/stratumn/go-core/store.Adapter.NewBatch.
func (a *DummyStore) NewBatch(ctx context.Context) (store.Batch, error) {
	return bufferedbatch.NewBatch(ctx, a), nil
}

/********** Utilities **********/

func (a *DummyStore) findHashesSegments(linkHashes hashSet, filter *store.SegmentFilter) (*types.PaginatedSegments, error) {
	segments := &types.PaginatedSegments{}

	for linkHash := range linkHashes {
		segment, err := a.getSegment(linkHash)
		if err != nil {
			return nil, err
		}

		if filter.Match(segment) {
			segments.Segments = append(segments.Segments, segment)
		}
	}
	segments.TotalCount = len(segments.Segments)

	segments.Segments.Sort(filter.Reverse)

	return filter.Pagination.PaginateSegments(segments), nil
}

func createKey(k []byte) string {
	return fmt.Sprintf("%x", k)
}
