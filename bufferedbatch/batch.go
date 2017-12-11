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

// Batch can be used as a base class for types
// that want to implement github.com/stratumn/sdk/store.Batch.
// All operations are stored in arrays and can be replayed.
// Only the Write method must be implemented.
type Batch struct {
	originalStore store.Adapter
	Links         []*cs.Link
}

// NewBatch creates a new Batch.
func NewBatch(a store.Adapter) *Batch {
	return &Batch{originalStore: a}
}

// CreateLink implements github.com/stratumn/sdk/store.LinkWriter.CreateLink.
func (b *Batch) CreateLink(link *cs.Link) (*types.Bytes32, error) {
	if err := link.Validate(b.GetSegment); err != nil {
		return nil, err
	}
	b.Links = append(b.Links, link)
	return link.Hash()
}

// GetSegment returns a segment from the cache or delegates the call to the store.
func (b *Batch) GetSegment(linkHash *types.Bytes32) (segment *cs.Segment, err error) {
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
func (b *Batch) FindSegments(filter *store.SegmentFilter) (cs.SegmentSlice, error) {
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
func (b *Batch) GetMapIDs(filter *store.MapFilter) ([]string, error) {
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

// Write implements github.com/stratumn/sdk/store.Batch.Write.
func (b *Batch) Write() (err error) {
	for _, link := range b.Links {
		_, err = b.originalStore.CreateLink(link)
		if err != nil {
			break
		}
	}

	return
}
