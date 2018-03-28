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
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"sync"

	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/leveldbstore"
	"github.com/stratumn/go-indigocore/monitoring"
	"github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/types"
	"go.opencensus.io/trace"
)

const (
	// Name is the name set in the store's information.
	Name = "file"

	// Description is the description set in the store's information.
	Description = "Indigo's File Store"

	// DefaultPath is the path where segments will be saved by default.
	DefaultPath = "/var/stratumn/filestore"
)

// FileStore is the type that implements github.com/stratumn/go-indigocore/store.Adapter.
type FileStore struct {
	config     *Config
	eventChans []chan *store.Event
	mutex      sync.RWMutex // simple global mutex
	kvDB       *leveldbstore.LevelDBStore
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

	return &FileStore{config, nil, sync.RWMutex{}, db}, nil
}

/********** Store adapter implementation **********/

// GetInfo implements github.com/stratumn/go-indigocore/store.Adapter.GetInfo.
func (a *FileStore) GetInfo(ctx context.Context) (_ interface{}, err error) {
	ctx, span := trace.StartSpan(ctx, "filestore/GetInfo")
	defer monitoring.SetSpanStatusAndEnd(span, err)

	return &Info{
		Name:        Name,
		Description: Description,
		Version:     a.config.Version,
		Commit:      a.config.Commit,
	}, nil
}

// AddStoreEventChannel implements github.com/stratumn/go-indigocore/store.Adapter.AddStoreEventChannel
func (a *FileStore) AddStoreEventChannel(eventChan chan *store.Event) {
	a.eventChans = append(a.eventChans, eventChan)
}

// NewBatch implements github.com/stratumn/go-indigocore/store.Adapter.NewBatch.
func (a *FileStore) NewBatch(ctx context.Context) (store.Batch, error) {
	return NewBatch(ctx, a), nil
}

/********** Store writer implementation **********/

// CreateLink implements github.com/stratumn/go-indigocore/store.LinkWriter.CreateLink.
func (a *FileStore) CreateLink(ctx context.Context, link *cs.Link) (_ *types.Bytes32, err error) {
	ctx, span := trace.StartSpan(ctx, "filestore/CreateLink")
	defer monitoring.SetSpanStatusAndEnd(span, err)

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

	linkEvent := store.NewSavedLinks(link)

	for _, c := range a.eventChans {
		c <- linkEvent
	}

	return linkHash, nil
}

// AddEvidence implements github.com/stratumn/go-indigocore/store.EvidenceWriter.AddEvidence.
func (a *FileStore) AddEvidence(ctx context.Context, linkHash *types.Bytes32, evidence *cs.Evidence) (err error) {
	ctx, span := trace.StartSpan(ctx, "filestore/AddEvidence")
	defer monitoring.SetSpanStatusAndEnd(span, err)

	currentEvidences, err := a.GetEvidences(ctx, linkHash)
	if err != nil {
		return err
	}

	if err = currentEvidences.AddEvidence(*evidence); err != nil {
		return err
	}

	key := getEvidenceKey(linkHash)
	value, err := json.Marshal(currentEvidences)
	if err != nil {
		return err
	}

	if err = a.SetValue(ctx, key, value); err != nil {
		return err
	}

	evidenceEvent := store.NewSavedEvidences()
	evidenceEvent.AddSavedEvidence(linkHash, evidence)

	for _, c := range a.eventChans {
		c <- evidenceEvent
	}

	return nil
}

/********** Store reader implementation **********/

// GetSegment implements github.com/stratumn/go-indigocore/store.SegmentReader.GetSegment.
func (a *FileStore) GetSegment(ctx context.Context, linkHash *types.Bytes32) (_ *cs.Segment, err error) {
	ctx, span := trace.StartSpan(ctx, "filestore/GetSegment")
	defer monitoring.SetSpanStatusAndEnd(span, err)

	a.mutex.RLock()
	defer a.mutex.RUnlock()

	return a.getSegment(ctx, linkHash)
}

func (a *FileStore) getSegment(ctx context.Context, linkHash *types.Bytes32) (_ *cs.Segment, err error) {
	link, err := a.getLink(linkHash)
	if err != nil || link == nil {
		return nil, err
	}

	evidences, err := a.GetEvidences(ctx, linkHash)
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

// FindSegments implements github.com/stratumn/go-indigocore/store.SegmentReader.FindSegments.
func (a *FileStore) FindSegments(ctx context.Context, filter *store.SegmentFilter) (_ cs.SegmentSlice, err error) {
	ctx, span := trace.StartSpan(ctx, "filestore/FindSegments")
	defer monitoring.SetSpanStatusAndEnd(span, err)

	var segments cs.SegmentSlice

	a.forEach(ctx, func(segment *cs.Segment) error {
		if filter.Match(segment) {
			segments = append(segments, segment)
		}
		return nil
	})

	sort.Sort(segments)

	return filter.Pagination.PaginateSegments(segments), nil
}

// GetMapIDs implements github.com/stratumn/go-indigocore/store.SegmentReader.GetMapIDs.
func (a *FileStore) GetMapIDs(ctx context.Context, filter *store.MapFilter) (_ []string, err error) {
	ctx, span := trace.StartSpan(ctx, "filestore/GetMapIDs")
	defer monitoring.SetSpanStatusAndEnd(span, err)

	set := map[string]struct{}{}
	a.forEach(ctx, func(segment *cs.Segment) error {
		if filter.Match(segment) {
			set[segment.Link.Meta.MapID] = struct{}{}
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

// GetEvidences implements github.com/stratumn/go-indigocore/store.EvidenceReader.GetEvidences.
func (a *FileStore) GetEvidences(ctx context.Context, linkHash *types.Bytes32) (_ *cs.Evidences, err error) {
	ctx, span := trace.StartSpan(ctx, "filestore/GetEvidences")
	defer monitoring.SetSpanStatusAndEnd(span, err)

	key := getEvidenceKey(linkHash)
	evidencesData, err := a.GetValue(ctx, key)
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

/********** github.com/stratumn/go-indigocore/store.KeyValueStore implementation **********/

// SetValue implements github.com/stratumn/go-indigocore/store.KeyValueStore.SetValue.
func (a *FileStore) SetValue(ctx context.Context, key []byte, value []byte) (err error) {
	ctx, span := trace.StartSpan(ctx, "filestore/SetValue")
	defer monitoring.SetSpanStatusAndEnd(span, err)

	return a.kvDB.SetValue(ctx, key, value)
}

// GetValue implements github.com/stratumn/go-indigocore/store.KeyValueStore.GetValue.
func (a *FileStore) GetValue(ctx context.Context, key []byte) (_ []byte, err error) {
	ctx, span := trace.StartSpan(ctx, "filestore/GetValue")
	defer monitoring.SetSpanStatusAndEnd(span, err)

	return a.kvDB.GetValue(ctx, key)
}

// DeleteValue implements github.com/stratumn/go-indigocore/store.KeyValueStore.DeleteValue.
func (a *FileStore) DeleteValue(ctx context.Context, key []byte) (_ []byte, err error) {
	ctx, span := trace.StartSpan(ctx, "filestore/DeleteValue")
	defer monitoring.SetSpanStatusAndEnd(span, err)

	return a.kvDB.DeleteValue(ctx, key)
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

func (a *FileStore) forEach(ctx context.Context, fn func(*cs.Segment) error) error {
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

			segment, err := a.getSegment(ctx, linkHash)
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
