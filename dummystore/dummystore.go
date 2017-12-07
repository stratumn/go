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

	"github.com/stratumn/sdk/bufferedbatch"
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
	links        linkMap      // maps link hashes to segments
	evidences    evidenceMap  // maps link hashes to evidences
	values       valueMap     // maps keys to values
	maps         hashSetMap   // maps chains IDs to sets of link hashes
	mutex        sync.RWMutex // simple global mutex
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
		nil,
		linkMap{},
		evidenceMap{},
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

	mapID := link.GetMapID()
	_, exists := a.maps[mapID]
	if !exists {
		a.maps[mapID] = hashSet{}
	}

	a.maps[mapID][linkHashStr] = struct{}{}

	linkEvent := store.NewSavedLinks()
	linkEvent.AddSavedLink(link)

	for _, c := range a.eventChans {
		c <- linkEvent
	}

	return linkHash, nil
}

// AddEvidence implements github.com/stratumn/sdk/store.EvidenceWriter.AddEvidence.
func (a *DummyStore) AddEvidence(linkHash *types.Bytes32, evidence *cs.Evidence) error {
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

	// If we already have an evidence for that provider, it means
	// we're in the case where we go from a PENDING evidence to a
	// COMPLETE one. This won't be necessary after the store interface
	// update, but meanwhile we need to correctly update the existing
	// evidence.
	previousEvidence := currentEvidences.GetEvidence(evidence.Provider)
	if previousEvidence != nil {
		if previousEvidence.State == cs.PendingEvidence {
			previousEvidence.State = evidence.State
			previousEvidence.Proof = evidence.Proof
		}
	} else if err := currentEvidences.AddEvidence(*evidence); err != nil {
		return err
	}

	a.evidences[linkHash] = currentEvidences

	return nil
}

// SaveSegment implements github.com/stratumn/sdk/store.Adapter.SaveSegment.
func (a *DummyStore) SaveSegment(segment *cs.Segment) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	return a.saveSegment(segment)
}

func (a *DummyStore) saveSegment(segment *cs.Segment) error {
	linkHash, err := a.createLink(&segment.Link)
	if err != nil {
		return err
	}

	for _, evidence := range segment.Meta.Evidences {
		if err := a.addEvidence(linkHash.String(), evidence); err != nil {
			return err
		}
	}

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
	segment, err := a.getSegment(linkHashStr)
	if err != nil || segment == nil {
		return nil, err
	}

	delete(a.links, linkHashStr)
	_, exists := a.evidences[linkHashStr]
	if exists {
		delete(a.evidences, linkHashStr)
	}

	mapID := segment.Link.GetMapID()
	delete(a.maps[mapID], linkHashStr)
	if len(a.maps[mapID]) == 0 {
		delete(a.maps, mapID)
	}

	return segment, nil
}

/********** Store reader implementation **********/

// GetSegment implements github.com/stratumn/sdk/store.Adapter.GetSegment.
func (a *DummyStore) GetSegment(linkHash *types.Bytes32) (*cs.Segment, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	return a.getSegment(linkHash.String())
}

// GetSegment implements github.com/stratumn/sdk/store.Adapter.GetSegment.
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

// FindSegments implements github.com/stratumn/sdk/store.Adapter.FindSegments.
func (a *DummyStore) FindSegments(filter *store.SegmentFilter) (cs.SegmentSlice, error) {
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

// GetMapIDs implements github.com/stratumn/sdk/store.Adapter.GetMapIDs.
func (a *DummyStore) GetMapIDs(filter *store.MapFilter) ([]string, error) {
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

// GetEvidences implements github.com/stratumn/sdk/store.EvidenceReader.GetEvidences.
func (a *DummyStore) GetEvidences(linkHash *types.Bytes32) (*cs.Evidences, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	evidences, _ := a.evidences[linkHash.String()]
	return evidences, nil
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
	return bufferedbatch.NewBatchV2(a), nil
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
