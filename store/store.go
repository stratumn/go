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
	"context"
	"strings"

	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/types"
)

const (
	// DefaultLimit is the default pagination limit.
	DefaultLimit = 20

	// MaxLimit is the maximum pagination limit.
	MaxLimit = 200
)

// SegmentReader is the interface for reading Segments from a store.
type SegmentReader interface {
	// Get a segment by link hash. Returns nil if no match is found.
	// Will return link and evidences (if there are some in that store).
	GetSegment(ctx context.Context, linkHash *types.Bytes32) (*cs.Segment, error)

	// Find segments. Returns an empty slice if there are no results.
	// Will return links and evidences (if there are some).
	FindSegments(ctx context.Context, filter *SegmentFilter) (*cs.PaginatedSegments, error)

	// Get all the existing map IDs.
	GetMapIDs(ctx context.Context, filter *MapFilter) ([]string, error)
}

// LinkWriter is the interface for writing links to a store.
// Links are immutable and cannot be deleted.
type LinkWriter interface {
	// Create the immutable part of a segment.
	// The input link is expected to be valid.
	// Returns the link hash or an error.
	CreateLink(ctx context.Context, link *cs.Link) (*types.Bytes32, error)
}

// EvidenceReader is the interface for reading segment evidence from a store.
type EvidenceReader interface {
	// Get the evidences for a segment.
	// Can return a nil error with an empty evidence slice if
	// the segment currently doesn't have evidence.
	GetEvidences(ctx context.Context, linkHash *types.Bytes32) (*cs.Evidences, error)
}

// EvidenceWriter is the interface for adding evidence to a segment in a store.
type EvidenceWriter interface {
	// Add an evidence to a segment.
	AddEvidence(ctx context.Context, linkHash *types.Bytes32, evidence *cs.Evidence) error
}

// EvidenceStore is the interface for storing and reading segment evidence.
type EvidenceStore interface {
	EvidenceReader
	EvidenceWriter
}

// Batch represents a database transaction.
type Batch interface {
	SegmentReader
	LinkWriter

	// Write definitely writes the content of the Batch
	Write(ctx context.Context) error
}

// Adapter is the minimal interface that all stores should implement.
// Then a store may optionally implement the KeyValueStore interface.
type Adapter interface {
	SegmentReader
	LinkWriter
	EvidenceStore

	// Returns arbitrary information about the adapter.
	GetInfo(ctx context.Context) (interface{}, error)

	// Adds a channel that receives events from the store.
	AddStoreEventChannel(chan *Event)

	// Creates a new Batch
	NewBatch(ctx context.Context) (Batch, error)
}

// KeyValueReader is the interface for reading key-value pairs.
type KeyValueReader interface {
	GetValue(ctx context.Context, key []byte) ([]byte, error)
}

// KeyValueWriter is the interface for writing key-value pairs.
type KeyValueWriter interface {
	SetValue(ctx context.Context, key []byte, value []byte) error
	DeleteValue(ctx context.Context, key []byte) ([]byte, error)
}

// KeyValueStore is the interface for a key-value store.
// Some stores will implement this interface, but not all.
type KeyValueStore interface {
	KeyValueReader
	KeyValueWriter
}

// Pagination contains pagination options.
type Pagination struct {
	// Index of the first entry.
	Offset int `json:"offset" url:"offset"`

	// Maximum number of entries.
	Limit int `json:"limit" url:"limit"`
}

// SegmentFilter contains filtering options for segments.
// If PrevLinkHash is not nil, MapID is ignored because a previous link hash
// implies the map ID of the previous segment.
type SegmentFilter struct {
	Pagination `json:"pagination"`

	// Map IDs the segments must have.
	MapIDs []string `json:"mapIds" url:"mapIds,brackets"`

	// Process name is optionnal.
	Process string `json:"process" url:"-"`

	// A previous link hash the segments must have.
	// nil makes this attribute as optional
	// empty string is to search Segments without parent
	PrevLinkHash *string `json:"prevLinkHash" url:"prevLinkHash"`

	// A slice of linkHashes to search Segments.
	// This attribute is optional.
	LinkHashes []string `json:"linkHashes" url:"linkHashes,brackets"`

	// A slice of tags the segments must all contain.
	Tags []string `json:"tags" url:"tags,brackets"`

	// Flag to reverse segment ordering.
	Reverse bool `json:"reverse" url:"reverse"`
}

// MapFilter contains filtering options for segments.
type MapFilter struct {
	Pagination `json:"pagination"`

	// Filter to get maps with IDs starting with a given prefix.
	Prefix string `json:"prefix" url:"prefix"`

	// Filter to get maps with IDs ending with a given suffix.
	Suffix string `json:"suffix" url:"suffix"`

	// Process name is optionnal.
	Process string `json:"process" url:"-"`
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
func (p *Pagination) PaginateSegments(a *cs.PaginatedSegments) *cs.PaginatedSegments {
	l := len(a.Segments)
	if p.Offset >= l {
		return &cs.PaginatedSegments{
			Segments:   cs.SegmentSlice{},
			TotalCount: a.TotalCount,
		}
	}

	end := min(l, p.Offset+p.Limit)
	return &cs.PaginatedSegments{
		Segments:   a.Segments[p.Offset:end],
		TotalCount: a.TotalCount,
	}
}

// Min of two ints, duh.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Match checks if segment matches with filter
func (filter SegmentFilter) Match(segment *cs.Segment) bool {
	if segment == nil {
		return false
	}

	return filter.MatchLink(&segment.Link)
}

// MatchLink checks if link matches with filter
func (filter SegmentFilter) MatchLink(link *cs.Link) bool {
	if link == nil {
		return false
	}

	if filter.PrevLinkHash != nil {
		prevLinkHash := link.Meta.GetPrevLinkHash()
		if *filter.PrevLinkHash == "" {
			if prevLinkHash != nil {
				return false
			}
		} else {
			filterPrevLinkHash, err := types.NewBytes32FromString(*filter.PrevLinkHash)
			if err != nil || prevLinkHash == nil || *filterPrevLinkHash != *prevLinkHash {
				return false
			}
		}
	}

	if len(filter.LinkHashes) > 0 {
		lh, _ := link.HashString()
		var match bool
		for _, linkHash := range filter.LinkHashes {
			if linkHash == lh {
				match = true
				break
			}
		}
		if !match {
			return false
		}
	}

	if filter.Process != "" && filter.Process != link.Meta.Process {
		return false
	}

	if len(filter.MapIDs) > 0 {
		var match = false
		mapID := link.Meta.MapID
		for _, filterMapIDs := range filter.MapIDs {
			match = match || filterMapIDs == mapID
		}
		if !match {
			return false
		}
	}

	if len(filter.Tags) > 0 {
		tags := link.Meta.GetTagMap()
		for _, tag := range filter.Tags {
			if _, ok := tags[tag]; !ok {
				return false
			}
		}
	}
	return true
}

// Match checks if segment matches with filter
func (filter MapFilter) Match(segment *cs.Segment) bool {
	if segment == nil {
		return false
	}

	return filter.MatchLink(&segment.Link)
}

// MatchLink checks if link matches with filter
func (filter MapFilter) MatchLink(link *cs.Link) bool {
	if link == nil {
		return false
	}
	if filter.Process != "" && filter.Process != link.Meta.Process {
		return false
	}
	if filter.Prefix != "" && !strings.HasPrefix(link.Meta.MapID, filter.Prefix) {
		return false
	}
	if filter.Suffix != "" && !strings.HasSuffix(link.Meta.MapID, filter.Suffix) {
		return false
	}
	return true
}
