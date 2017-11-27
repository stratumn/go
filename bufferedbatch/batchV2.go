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

package bufferedbatch

import (
	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/store"
	"github.com/stratumn/sdk/types"
)

// BatchV2 can be used as a base class for types
// that want to implement github.com/stratumn/sdk/store.BatchV2.
// All operations are stored in arrays and can be replayed.
// Only the WriteV2 method must be implemented.
type BatchV2 struct {
	originalStore store.AdapterV2
	Links         []*cs.Link
}

// NewBatchV2 creates a new BatchV2.
func NewBatchV2(a store.AdapterV2) *BatchV2 {
	return &BatchV2{originalStore: a}
}

// CreateLink implements github.com/stratumn/sdk/store.LinkWriter.CreateLink.
func (b *BatchV2) CreateLink(link *cs.Link) (*types.Bytes32, error) {
	segment := link.Segmentify()
	if err := segment.Validate(b.GetSegment); err != nil {
		return nil, err
	}
	b.Links = append(b.Links, link)
	return segment.GetLinkHash(), nil
}

// GetSegment returns a segment from the cache or delegates the call to the store.
func (b *BatchV2) GetSegment(linkHash *types.Bytes32) (segment *cs.Segment, err error) {
	for _, link := range b.Links {
		lh, err := link.Hash()
		if err != nil {
			return nil, err
		}

		if *lh == *linkHash {
			segment = link.Segmentify()
		}
	}

	if segment != nil {
		return segment, nil
	}

	return b.originalStore.GetSegment(linkHash)
}

// FindSegments returns the union of segments in the store and not committed yet.
func (b *BatchV2) FindSegments(filter *store.SegmentFilter) (cs.SegmentSlice, error) {
	segments, err := b.originalStore.FindSegments(filter)
	if err != nil {
		return segments, err
	}

	for _, link := range b.Links {
		if filter.MatchLink(link) {
			segments = append(segments, link.Segmentify())
		}
	}

	return filter.Pagination.PaginateSegments(segments), nil
}

// GetMapIDs returns the union of mapIds in the store and not committed yet.
func (b *BatchV2) GetMapIDs(filter *store.MapFilter) ([]string, error) {
	tmpMapIDs, err := b.originalStore.GetMapIDs(filter)
	if err != nil {
		return tmpMapIDs, err
	}
	mapIDs := make(map[string]int, len(tmpMapIDs))
	for _, m := range tmpMapIDs {
		mapIDs[m] = 0
	}

	// Apply uncommitted links
	for _, link := range b.Links {
		if filter.MatchLink(link) {
			mapID := link.Meta["mapId"].(string)
			mapIDs[mapID]++
		}
	}

	ids := make([]string, 0, len(mapIDs))
	for k := range mapIDs {
		ids = append(ids, k)
	}

	return filter.Pagination.PaginateStrings(ids), err
}

// WriteV2 implements github.com/stratumn/sdk/store.BatchV2.WriteV2.
func (b *BatchV2) WriteV2() (err error) {
	for _, link := range b.Links {
		_, err = b.originalStore.CreateLink(link)
		if err != nil {
			break
		}
	}

	return
}
