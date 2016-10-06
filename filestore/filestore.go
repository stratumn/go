// Copyright 2016 Stratumn SAS. All rights reserved.
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

	"github.com/stratumn/go/cs"
	"github.com/stratumn/go/store"
	"github.com/stratumn/go/types"
)

const (
	// Name is the name set in the store's information.
	Name = "file"

	// Description is the description set in the store's information.
	Description = "Stratumn File Store"

	// DefaultPath is the path where segments will be saved by default.
	DefaultPath = "/var/stratumn/filestore"
)

// FileStore is the type that implements github.com/stratumn/go/store.Adapter.
type FileStore struct {
	config *Config
	mutex  sync.RWMutex // simple global mutex
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
func New(config *Config) *FileStore {
	return &FileStore{config, sync.RWMutex{}}
}

// GetInfo implements github.com/stratumn/go/store.Adapter.GetInfo.
func (a *FileStore) GetInfo() (interface{}, error) {
	return &Info{
		Name:        Name,
		Description: Description,
		Version:     a.config.Version,
		Commit:      a.config.Commit,
	}, nil
}

// SaveSegment implements github.com/stratumn/go/store.Adapter.SaveSegment.
func (a *FileStore) SaveSegment(segment *cs.Segment) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	js, err := json.MarshalIndent(segment, "", "  ")
	if err != nil {
		return err
	}

	if err = os.MkdirAll(a.config.Path, 0755); err != nil {
		if !os.IsExist(err) {
			return err
		}
	}

	segmentPath := path.Join(a.config.Path, segment.Meta["linkHash"].(string)+".json")
	return ioutil.WriteFile(segmentPath, js, 0644)
}

// GetSegment implements github.com/stratumn/go/store.Adapter.GetSegment.
func (a *FileStore) GetSegment(linkHash *types.Bytes32) (*cs.Segment, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	return a.getSegment(linkHash)
}

// DeleteSegment implements github.com/stratumn/go/store.Adapter.DeleteSegment.
func (a *FileStore) DeleteSegment(linkHash *types.Bytes32) (*cs.Segment, error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	segment, err := a.getSegment(linkHash)
	if segment == nil {
		return segment, err
	}

	if err = os.Remove(path.Join(a.config.Path, linkHash.String()+".json")); err != nil {
		return nil, err
	}

	return segment, err
}

// FindSegments implements github.com/stratumn/go/store.Adapter.FindSegments.
func (a *FileStore) FindSegments(filter *store.Filter) (cs.SegmentSlice, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	var segments cs.SegmentSlice

	a.forEach(func(segment *cs.Segment) error {
		if filter.PrevLinkHash != nil {
			prevLinkHash := segment.Link.GetPrevLinkHash()
			if prevLinkHash == nil || *filter.PrevLinkHash != *prevLinkHash {
				return nil
			}
		} else if filter.MapID != "" && filter.MapID != segment.Link.GetMapID() {
			return nil
		}

		if len(filter.Tags) > 0 {
			tags := segment.Link.GetTags()
			if len(tags) > 0 {
				for _, tag := range filter.Tags {
					if !containsString(tags, tag) {
						return nil
					}
				}
			} else {
				return nil
			}
		}

		segments = append(segments, segment)

		return nil
	})

	sort.Sort(segments)

	return paginateSegments(segments, &filter.Pagination), nil
}

// GetMapIDs implements github.com/stratumn/go/store.Adapter.GetMapIDs.
func (a *FileStore) GetMapIDs(pagination *store.Pagination) ([]string, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	set := map[string]struct{}{}
	a.forEach(func(segment *cs.Segment) error {
		set[segment.Link.GetMapID()] = struct{}{}
		return nil
	})

	var mapIDs []string
	for mapID := range set {
		mapIDs = append(mapIDs, mapID)
	}

	sort.Strings(mapIDs)
	return paginateStrings(mapIDs, pagination), nil
}

func (a *FileStore) getSegment(linkHash *types.Bytes32) (*cs.Segment, error) {
	file, err := os.Open(path.Join(a.config.Path, linkHash.String()+".json"))
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

func containsString(a []string, s string) bool {
	for _, v := range a {
		if v == s {
			return true
		}
	}

	return false
}

func paginateStrings(a []string, p *store.Pagination) []string {
	l := len(a)
	if p.Offset >= l {
		return []string{}
	}

	end := min(l, p.Offset+p.Limit)
	return a[p.Offset:end]
}

func paginateSegments(a cs.SegmentSlice, p *store.Pagination) cs.SegmentSlice {
	l := len(a)
	if p.Offset >= l {
		return cs.SegmentSlice{}
	}

	end := min(l, p.Offset+p.Limit)
	return a[p.Offset:end]
}

// Min of two ints, duh.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
