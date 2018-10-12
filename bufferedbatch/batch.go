// Copyright 2016-2018 Stratumn SAS. All rights reserved.
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
	"bytes"
	"context"

	"github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-core/monitoring"
	"github.com/stratumn/go-core/store"
	"github.com/stratumn/go-core/types"

	"go.opencensus.io/stats"
	"go.opencensus.io/tag"
	"go.opencensus.io/trace"
)

// Batch can be used as a base class for types
// that want to implement github.com/stratumn/go-core/store.Batch.
// All operations are stored in arrays and can be replayed.
// Only the Write method must be implemented.
type Batch struct {
	originalStore store.Adapter
	Links         []*chainscript.Link
}

// NewBatch creates a new Batch.
func NewBatch(ctx context.Context, a store.Adapter) *Batch {
	stats.Record(ctx, batchCount.M(1))
	return &Batch{originalStore: a}
}

// CreateLink implements github.com/stratumn/go-core/store.LinkWriter.CreateLink.
func (b *Batch) CreateLink(ctx context.Context, link *chainscript.Link) (_ chainscript.LinkHash, err error) {
	_, span := trace.StartSpan(ctx, "bufferedbatch/CreateLink")
	defer func() {
		monitoring.SetSpanStatusAndEnd(span, err)
	}()

	if link.Meta.OutDegree >= 0 {
		return nil, types.WrapError(store.ErrOutDegreeNotSupported, monitoring.Unimplemented, store.Component, "could not create link")
	}

	b.Links = append(b.Links, link)
	return link.Hash()
}

// GetSegment returns a segment from the cache or delegates the call to the store.
func (b *Batch) GetSegment(ctx context.Context, linkHash chainscript.LinkHash) (segment *chainscript.Segment, err error) {
	ctx, span := trace.StartSpan(ctx, "bufferedbatch/GetSegment")
	defer func() {
		monitoring.SetSpanStatusAndEnd(span, err)
	}()

	for _, link := range b.Links {
		lh, err := link.Hash()
		if err != nil {
			return nil, types.WrapError(err, monitoring.InvalidArgument, store.Component, "could not hash link")
		}

		if bytes.Equal(lh, linkHash) {
			segment, err = link.Segmentify()
			if err != nil {
				return nil, types.WrapError(err, monitoring.InvalidArgument, store.Component, "could not segmentify")
			}

			break
		}
	}

	if segment != nil {
		return segment, nil
	}

	return b.originalStore.GetSegment(ctx, linkHash)
}

// FindSegments returns the union of segments in the store and not committed yet.
func (b *Batch) FindSegments(ctx context.Context, filter *store.SegmentFilter) (_ *types.PaginatedSegments, err error) {
	ctx, span := trace.StartSpan(ctx, "bufferedbatch/FindSegments")
	defer func() {
		monitoring.SetSpanStatusAndEnd(span, err)
	}()

	segments, err := b.originalStore.FindSegments(ctx, filter)
	if err != nil {
		return nil, err
	}

	for _, link := range b.Links {
		if filter.MatchLink(link) {
			segment, err := link.Segmentify()
			if err != nil {
				return nil, types.WrapError(err, monitoring.InvalidArgument, store.Component, "could not segmentify")
			}

			segments.Segments = append(segments.Segments, segment)
			segments.TotalCount++
		}
	}

	return filter.Pagination.PaginateSegments(segments), nil
}

// GetMapIDs returns the union of mapIds in the store and not committed yet.
func (b *Batch) GetMapIDs(ctx context.Context, filter *store.MapFilter) (_ []string, err error) {
	ctx, span := trace.StartSpan(ctx, "bufferedbatch/GetMapIDs")
	defer func() {
		monitoring.SetSpanStatusAndEnd(span, err)
	}()

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
			mapIDs[link.Meta.MapId]++
		}
	}

	ids := make([]string, 0, len(mapIDs))
	for k := range mapIDs {
		ids = append(ids, k)
	}

	return filter.Pagination.PaginateStrings(ids), err
}

// Write implements github.com/stratumn/go-core/store.Batch.Write.
func (b *Batch) Write(ctx context.Context) (err error) {
	ctx, span := trace.StartSpan(ctx, "bufferedbatch/Write")
	defer func() {
		monitoring.SetSpanStatusAndEnd(span, err)
	}()

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
