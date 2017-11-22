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
	"github.com/stratumn/sdk/store"
	"github.com/stratumn/sdk/types"
	"github.com/tendermint/tmlibs/db"
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
	kvDB         db.DB
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
	db, err := db.NewGoLevelDB("keyvalue-store", config.Path)
	if err != nil {
		return nil, err
	}

	return &FileStore{config, nil, nil, sync.RWMutex{}, db}, nil
}

/********** github.com/stratumn/sdk/store.Adapter implementation **********/

// GetInfo implements github.com/stratumn/sdk/store.Adapter.GetInfo.
func (a *FileStore) GetInfo() (interface{}, error) {
	return &Info{
		Name:        Name,
		Description: Description,
		Version:     a.config.Version,
		Commit:      a.config.Commit,
	}, nil
}

// AddDidSaveChannel implements
// github.com/stratumn/sdk/store.Adapter.AddDidSaveChannel.
func (a *FileStore) AddDidSaveChannel(saveChan chan *cs.Segment) {
	a.didSaveChans = append(a.didSaveChans, saveChan)
}

// AddStoreEventChannel implements
// github.com/stratumn/sdk/store.AdapterV2.AddStoreEventChannel
func (a *FileStore) AddStoreEventChannel(eventChan chan *store.Event) {
	a.eventChans = append(a.eventChans, eventChan)
}

// NewBatch implements github.com/stratumn/sdk/store.Adapter.NewBatch.
func (a *FileStore) NewBatch() (store.Batch, error) {
	return NewBatch(a), nil
}

/********** github.com/stratumn/sdk/store.Adapter.Writer implementation **********/

// SaveSegment implements github.com/stratumn/sdk/store.Adapter.SaveSegment.
func (a *FileStore) SaveSegment(segment *cs.Segment) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	return a.saveSegment(segment)
}

func (a *FileStore) saveSegment(segment *cs.Segment) (err error) {
	if err = a.initDir(); err != nil {
		return err
	}

	curr, err := a.getSegment(segment.GetLinkHash())
	if err != nil {
		return err
	}
	if curr != nil {
		if segment, err = curr.MergeMeta(segment); err != nil {
			return err
		}
	}

	js, err := json.MarshalIndent(segment, "", "  ")
	if err != nil {
		return err
	}

	segmentPath := a.getSegmentPath(segment.GetLinkHashString())

	if err := ioutil.WriteFile(segmentPath, js, 0644); err != nil {
		return err
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

	if err = os.Remove(a.getSegmentPath(linkHash.String())); err != nil {
		return nil, err
	}

	return segment, err
}

/********** github.com/stratumn/sdk/store.Adapter.Reader implementation **********/

// GetSegment implements github.com/stratumn/sdk/store.Adapter.GetSegment.
func (a *FileStore) GetSegment(linkHash *types.Bytes32) (*cs.Segment, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	return a.getSegment(linkHash)
}

// FindSegments implements github.com/stratumn/sdk/store.Adapter.FindSegments.
func (a *FileStore) FindSegments(filter *store.SegmentFilter) (cs.SegmentSlice, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

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

// GetMapIDs implements github.com/stratumn/sdk/store.Adapter.GetMapIDs.
func (a *FileStore) GetMapIDs(filter *store.MapFilter) ([]string, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

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

/********** github.com/stratumn/sdk/store.KeyValueStore implementation **********/

// SaveValue implements github.com/stratumn/sdk/store.Adapter.SaveValue.
func (a *FileStore) SaveValue(key []byte, value []byte) error {
	return a.SetValue(key, value)
}

// SetValue implements github.com/stratumn/sdk/store.KeyValueStore.SetValue.
func (a *FileStore) SetValue(key []byte, value []byte) error {
	a.kvDB.Set(key, value)
	return nil
}

// GetValue implements github.com/stratumn/sdk/store.Adapter.GetValue
// and github.com/stratumn/sdk/store.KeyValueStore.GetValue.
func (a *FileStore) GetValue(key []byte) ([]byte, error) {
	return a.kvDB.Get(key), nil
}

// DeleteValue implements github.com/stratumn/sdk/store.Adapter.DeleteValue
// and github.com/stratumn/sdk/store.KeyValueStore.DeleteValue.
func (a *FileStore) DeleteValue(key []byte) ([]byte, error) {
	v := a.kvDB.Get(key)

	if v != nil {
		a.kvDB.Delete(key)
		return v, nil
	}

	return nil, nil
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

func (a *FileStore) getSegment(linkHash *types.Bytes32) (*cs.Segment, error) {
	file, err := os.Open(a.getSegmentPath(linkHash.String()))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	defer file.Close()

	var segment cs.Segment
	if err = json.NewDecoder(file).Decode(&segment); err != nil {
		return nil, err
	}

	return &segment, nil
}

func (a *FileStore) getSegmentPath(linkHash string) string {
	return path.Join(a.config.Path, linkHash+".json")
}

var segmentFileRegepx = regexp.MustCompile("(.*)\\.json$")

func (a *FileStore) forEach(fn func(*cs.Segment) error) error {
	files, err := ioutil.ReadDir(a.config.Path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	for _, file := range files {
		name := file.Name()
		if segmentFileRegepx.MatchString(name) {
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
