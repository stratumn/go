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

// Package filestore implements a store that saves all the segments to the file
// system.
//
// The segments are stored as JSON files named after the link hashes.
// It's a convenient store to use during the development of an agent.
// However, because it doesn't use an index, it's very slow, and shouldn't be
// used for production.
package filestore

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"sync"

	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/leveldbstore"
	"github.com/stratumn/sdk/store"
	"github.com/stratumn/sdk/types"
)

const (
	// Name is the name set in the store's information.
	Name = "file"

	// Description is the description set in the store's information.
	Description = "Stratumn File Store"

	// DefaultPath is the path where segments will be saved by default.
	DefaultPath = "/var/stratumn/filestore"
)

// FileStore is the type that implements github.com/stratumn/sdk/store.Adapter.
type FileStore struct {
	config       *Config
	didSaveChans []chan *cs.Segment
	eventChans   []chan *store.Event
	mutex        sync.RWMutex // simple global mutex
	kvDB         *leveldbstore.LevelDBStore
}

// Config contains configuration options for the store.
type Config struct {
	// A version string that will be set in the store's information.
	Version string

	// A git commit hash that will be set in the store's information.
	Commit string

	// Path where segments will be saved.
	Path string
}

// Info is the info returned by GetInfo.
type Info struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Commit      string `json:"commit"`
}

// New creates an instance of a FileStore.
func New(config *Config) (*FileStore, error) {
	kvStoreConfig := &leveldbstore.Config{
		Path: config.Path,
	}
	db, err := leveldbstore.New(kvStoreConfig)
	if err != nil {
		return nil, err
	}

	return &FileStore{config, nil, nil, sync.RWMutex{}, db}, nil
}

/********** Store adapter implementation **********/

// GetInfo implements github.com/stratumn/sdk/store.Adapter.GetInfo.
func (a *FileStore) GetInfo() (interface{}, error) {
	return &Info{
		Name:        Name,
		Description: Description,
		Version:     a.config.Version,
		Commit:      a.config.Commit,
	}, nil
}

// AddDidSaveChannel implements github.com/stratumn/sdk/store.Adapter.AddDidSaveChannel.
func (a *FileStore) AddDidSaveChannel(saveChan chan *cs.Segment) {
	a.didSaveChans = append(a.didSaveChans, saveChan)
}

// AddStoreEventChannel implements github.com/stratumn/sdk/store.AdapterV2.AddStoreEventChannel
func (a *FileStore) AddStoreEventChannel(eventChan chan *store.Event) {
	a.eventChans = append(a.eventChans, eventChan)
}

// NewBatch implements github.com/stratumn/sdk/store.Adapter.NewBatch.
func (a *FileStore) NewBatch() (store.Batch, error) {
	return NewBatch(a), nil
}

// NewBatchV2 implements github.com/stratumn/sdk/store.AdapterV2.NewBatchV2.
func (a *FileStore) NewBatchV2() (store.BatchV2, error) {
	return nil, nil
}

/********** Store writer implementation **********/

// CreateLink implements github.com/stratumn/sdk/store.LinkWriter.CreateLink.
func (a *FileStore) CreateLink(link *cs.Link) (*types.Bytes32, error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	return a.createLink(link)
}

func (a *FileStore) createLink(link *cs.Link) (*types.Bytes32, error) {
	linkHash, err := link.Hash()
	if err != nil {
		return nil, err
	}

	if err = a.initDir(); err != nil {
		return nil, err
	}

	js, err := json.MarshalIndent(link, "", "  ")
	if err != nil {
		return nil, err
	}

	linkPath := a.getLinkPath(linkHash)

	if err := ioutil.WriteFile(linkPath, js, 0644); err != nil {
		return nil, err
	}

	for _, c := range a.eventChans {
		c <- &store.Event{
			EventType: store.SavedLink,
			Details:   link,
		}
	}

	return linkHash, nil
}

// AddEvidence implements github.com/stratumn/sdk/store.EvidenceWriter.AddEvidence.
func (a *FileStore) AddEvidence(linkHash *types.Bytes32, evidence *cs.Evidence) error {
	currentEvidences, err := a.GetEvidences(linkHash)
	if err != nil {
		return err
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
	} else if err = currentEvidences.AddEvidence(*evidence); err != nil {
		return err
	}

	key := getEvidenceKey(linkHash)
	value, err := json.Marshal(currentEvidences)
	if err != nil {
		return err
	}

	if err = a.SetValue(key, value); err != nil {
		return err
	}

	for _, c := range a.eventChans {
		c <- &store.Event{
			EventType: store.SavedEvidence,
			Details:   evidence,
		}
	}

	return nil
}

// SaveSegment implements github.com/stratumn/sdk/store.Adapter.SaveSegment.
func (a *FileStore) SaveSegment(segment *cs.Segment) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	return a.saveSegment(segment)
}

func (a *FileStore) saveSegment(segment *cs.Segment) error {
	linkHash, err := a.createLink(&segment.Link)
	if err != nil {
		return err
	}

	for _, evidence := range segment.Meta.Evidences {
		if err := a.AddEvidence(linkHash, evidence); err != nil {
			return err
		}
	}

	for _, c := range a.didSaveChans {
		c <- segment
	}

	return nil
}

// DeleteSegment implements github.com/stratumn/sdk/store.Adapter.DeleteSegment.
func (a *FileStore) DeleteSegment(linkHash *types.Bytes32) (*cs.Segment, error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	return a.deleteSegment(linkHash)
}

func (a *FileStore) deleteSegment(linkHash *types.Bytes32) (*cs.Segment, error) {
	segment, err := a.getSegment(linkHash)
	if segment == nil {
		return segment, err
	}

	if err = os.Remove(a.getLinkPath(linkHash)); err != nil {
		return nil, err
	}

	return segment, err
}

/********** Store reader implementation **********/

// GetSegment implements github.com/stratumn/sdk/store.Adapter.GetSegment
// and github.com/stratumn/sdk/store.SegmentReader.GetSegment.
func (a *FileStore) GetSegment(linkHash *types.Bytes32) (*cs.Segment, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	return a.getSegment(linkHash)
}

func (a *FileStore) getSegment(linkHash *types.Bytes32) (*cs.Segment, error) {
	link, err := a.getLink(linkHash)
	if err != nil || link == nil {
		return nil, err
	}

	evidences, err := a.GetEvidences(linkHash)
	if err != nil {
		return nil, err
	}

	return &cs.Segment{
		Link: *link,
		Meta: cs.SegmentMeta{
			Evidences: *evidences,
			LinkHash:  linkHash.String(),
		},
	}, nil
}

// FindSegments implements github.com/stratumn/sdk/store.Adapter.FindSegments
// and github.com/stratumn/sdk/store.SegmentReader.FindSegments.
func (a *FileStore) FindSegments(filter *store.SegmentFilter) (cs.SegmentSlice, error) {
	var segments cs.SegmentSlice

	a.forEach(func(segment *cs.Segment) error {
		if filter.Match(segment) {
			segments = append(segments, segment)
		}
		return nil
	})

	sort.Sort(segments)

	return filter.Pagination.PaginateSegments(segments), nil
}

// GetMapIDs implements github.com/stratumn/sdk/store.Adapter.GetMapIDs
// and github.com/stratumn/sdk/store.SegmentReader.GetMapIDs.
func (a *FileStore) GetMapIDs(filter *store.MapFilter) ([]string, error) {
	set := map[string]struct{}{}
	a.forEach(func(segment *cs.Segment) error {
		if filter.Match(segment) {
			set[segment.Link.GetMapID()] = struct{}{}
		}
		return nil
	})

	var mapIDs []string
	for mapID := range set {
		mapIDs = append(mapIDs, mapID)
	}

	sort.Strings(mapIDs)
	return filter.Pagination.PaginateStrings(mapIDs), nil
}

// GetEvidences implements github.com/stratumn/sdk/store.EvidenceReader.GetEvidences.
func (a *FileStore) GetEvidences(linkHash *types.Bytes32) (*cs.Evidences, error) {
	key := getEvidenceKey(linkHash)
	evidencesData, err := a.GetValue(key)
	if err != nil {
		return nil, err
	}

	evidences := cs.Evidences{}
	if evidencesData != nil && len(evidencesData) > 0 {
		if err := json.Unmarshal(evidencesData, &evidences); err != nil {
			return nil, err
		}
	}

	return &evidences, nil
}

func getEvidenceKey(linkHash *types.Bytes32) []byte {
	return []byte("evidences:" + linkHash.String())
}

func (a *FileStore) getLink(linkHash *types.Bytes32) (*cs.Link, error) {
	file, err := os.Open(a.getLinkPath(linkHash))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	defer file.Close()

	var link cs.Link
	if err = json.NewDecoder(file).Decode(&link); err != nil {
		return nil, err
	}

	return &link, nil
}

/********** github.com/stratumn/sdk/store.KeyValueStore implementation **********/

// SaveValue implements github.com/stratumn/sdk/store.Adapter.SaveValue.
func (a *FileStore) SaveValue(key []byte, value []byte) error {
	return a.kvDB.SetValue(key, value)
}

// SetValue implements github.com/stratumn/sdk/store.KeyValueStore.SetValue.
func (a *FileStore) SetValue(key []byte, value []byte) error {
	return a.kvDB.SetValue(key, value)
}

// GetValue implements github.com/stratumn/sdk/store.Adapter.GetValue
// and github.com/stratumn/sdk/store.KeyValueStore.GetValue.
func (a *FileStore) GetValue(key []byte) ([]byte, error) {
	return a.kvDB.GetValue(key)
}

// DeleteValue implements github.com/stratumn/sdk/store.Adapter.DeleteValue
// and github.com/stratumn/sdk/store.KeyValueStore.DeleteValue.
func (a *FileStore) DeleteValue(key []byte) ([]byte, error) {
	return a.kvDB.DeleteValue(key)
}

/********** Utilities **********/

func (a *FileStore) initDir() error {
	if err := os.MkdirAll(a.config.Path, 0755); err != nil {
		if !os.IsExist(err) {
			return err
		}
	}
	return nil
}

func (a *FileStore) getLinkPath(linkHash *types.Bytes32) string {
	return path.Join(a.config.Path, linkHash.String()+".json")
}

var linkFileRegex = regexp.MustCompile("(.*)\\.json$")

func (a *FileStore) forEach(fn func(*cs.Segment) error) error {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	files, err := ioutil.ReadDir(a.config.Path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	for _, file := range files {
		name := file.Name()
		if linkFileRegex.MatchString(name) {
			linkHashStr := name[:len(name)-5]
			linkHash, err := types.NewBytes32FromString(linkHashStr)
			if err != nil {
				return err
			}

			segment, err := a.getSegment(linkHash)
			if err != nil {
				return err
			}
			if segment == nil {
				return fmt.Errorf("could not find segment %q", filepath.Base(name))
			}
			if err = fn(segment); err != nil {
				return err
			}
		}
	}

	return nil
}
