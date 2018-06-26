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
	"context"

	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/monitoring"
	"github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/types"

	"go.opencensus.io/stats"
	"go.opencensus.io/tag"
	"go.opencensus.io/trace"
)

// Batch can be used as a base class for types
// that want to implement github.com/stratumn/go-indigocore/store.Batch.
// All operations are stored in arrays and can be replayed.
// Only the Write method must be implemented.
type Batch struct {
	originalStore store.Adapter
	Links         []*cs.Link
}

// NewBatch creates a new Batch.
func NewBatch(ctx context.Context, a store.Adapter) *Batch {
	stats.Record(ctx, batchCount.M(1))
	return &Batch{originalStore: a}
}

// CreateLink implements github.com/stratumn/go-indigocore/store.LinkWriter.CreateLink.
func (b *Batch) CreateLink(ctx context.Context, link *cs.Link) (_ *types.Bytes32, err error) {
	_, span := trace.StartSpan(ctx, "bufferedbatch/CreateLink")
	defer monitoring.SetSpanStatusAndEnd(span, err)

	b.Links = append(b.Links, link)
	return link.Hash()
}

// GetSegment returns a segment from the cache or delegates the call to the store.
func (b *Batch) GetSegment(ctx context.Context, linkHash *types.Bytes32) (segment *cs.Segment, err error) {
	ctx, span := trace.StartSpan(ctx, "bufferedbatch/GetSegment")
	defer monitoring.SetSpanStatusAndEnd(span, err)

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

	return b.originalStore.GetSegment(ctx, linkHash)
}

// FindSegments returns the union of segments in the store and not committed yet.
func (b *Batch) FindSegments(ctx context.Context, filter *store.SegmentFilter) (_ cs.SegmentSlice, err error) {
	ctx, span := trace.StartSpan(ctx, "bufferedbatch/FindSegments")
	defer monitoring.SetSpanStatusAndEnd(span, err)

	segments, err := b.originalStore.FindSegments(ctx, filter)
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
func (b *Batch) GetMapIDs(ctx context.Context, filter *store.MapFilter) (_ []string, err error) {
	ctx, span := trace.StartSpan(ctx, "bufferedbatch/GetMapIDs")
	defer monitoring.SetSpanStatusAndEnd(span, err)

	tmpMapIDs, err := b.originalStore.GetMapIDs(ctx, filter)
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
			mapIDs[link.Meta.MapID]++
		}
	}

	ids := make([]string, 0, len(mapIDs))
	for k := range mapIDs {
		ids = append(ids, k)
	}

	return filter.Pagination.PaginateStrings(ids), err
}

// Write implements github.com/stratumn/go-indigocore/store.Batch.Write.
func (b *Batch) Write(ctx context.Context) (err error) {
	ctx, span := trace.StartSpan(ctx, "bufferedbatch/Write")
	defer monitoring.SetSpanStatusAndEnd(span, err)

	stats.Record(ctx, linksPerBatch.M(int64(len(b.Links))))

	for _, link := range b.Links {
		_, err = b.originalStore.CreateLink(ctx, link)
		if err != nil {
			break
		}
	}

	if err == nil {
		ctx, _ = tag.New(ctx, tag.Upsert(writeStatus, "success"))
	} else {
		ctx, _ = tag.New(ctx, tag.Upsert(writeStatus, "failure"))
	}

	stats.Record(ctx, writeCount.M(1))

	return
}
