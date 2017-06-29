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

// Package store defines types to implement a store.
package store

import (
	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/types"
)

const (
	// DefaultLimit is the default pagination limit.
	DefaultLimit = 20

	// MaxLimit is the maximum pagination limit.
	MaxLimit = 200
)

// Writer is the interface that wraps the Write methods of a Store.
type Writer interface {
	// Creates or updates a segment. Segments passed to this method are
	// assumed to be valid.
	SaveSegment(segment *cs.Segment) error

	// Deletes a segment by link hash. Returns the removed segment or nil
	// if not found.
	DeleteSegment(linkHash *types.Bytes32) (*cs.Segment, error)

	// Saves a value at a key.
	SaveValue(key []byte, value []byte) error

	// Deletes a value at a key.
	DeleteValue(key []byte) ([]byte, error)
}

// Reader is the interface that wraps the Read methods of a Store.
type Reader interface {
	// Get a segment by link hash. Returns nil if no match is found.
	GetSegment(linkHash *types.Bytes32) (*cs.Segment, error)

	// Find segments. Returns an empty slice if there are no results.
	FindSegments(filter *Filter) (cs.SegmentSlice, error)

	// Get all the existing map IDs.
	GetMapIDs(pagination *Pagination) ([]string, error)

	// Gets a value at a key
	GetValue(key []byte) ([]byte, error)
}

// Batch represents a database transaction
type Batch interface {
	Reader
	Writer

	// Write definitely writes the content of the Batch
	Write() error
}

// Adapter must be implemented by a store.
type Adapter interface {
	Reader
	Writer

	// Returns arbitrary information about the adapter.
	GetInfo() (interface{}, error)

	// Adds a channel that receives segments whenever they are saved.
	AddDidSaveChannel(chan *cs.Segment)

	// Creates a new Batch
	NewBatch() (Batch, error)
}

// Pagination contains pagination options.
type Pagination struct {
	// Index of the first entry.
	Offset int `json:"offset"`

	// Maximum number of entries.
	Limit int `json:"limit"`
}

// Filter contains filtering options for segments.
// If PrevLinkHash is not nil, MapID is ignored because a previous link hash
// implies the map ID of the previous segment.
type Filter struct {
	Pagination `json:"pagination"`

	// A map ID the segments must have.
	MapID string `json:"mapId"`

	// A previous link hash the segments must have.
	PrevLinkHash *types.Bytes32 `json:"prevLinkHash"`

	// A slice of tags the segments must all contain.
	Tags []string `json:"tags"`
}

// PaginateStrings paginates a list of strings
func (p *Pagination) PaginateStrings(a []string) []string {
	l := len(a)
	if p.Offset >= l {
		return []string{}
	}

	end := min(l, p.Offset+p.Limit)
	return a[p.Offset:end]
}

// PaginateSegments paginate a list of segments
func (p *Pagination) PaginateSegments(a cs.SegmentSlice) cs.SegmentSlice {
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
