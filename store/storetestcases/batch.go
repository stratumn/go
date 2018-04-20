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

package storetestcases

import (
	"context"
	"fmt"
	"testing"

	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/cs/cstesting"
	"github.com/stratumn/go-indigocore/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func initBatch(t *testing.T, a store.Adapter) store.Batch {
	b, err := a.NewBatch(context.Background())
	require.NoError(t, err, "a.NewBatch()")
	assert.NotNil(t, b, "Batch should not be nil")
	return b
}

// TestBatch runs all tests for the store.Batch interface
func (f Factory) TestBatch(t *testing.T) {
	ctx := context.Background()
	a := f.initAdapter(t)
	defer f.freeAdapter(a)

	// Initialize the adapter with a few links with specific map ids
	for i := 0; i < 6; i++ {
		link := cstesting.NewLinkBuilder().WithMapID(fmt.Sprintf("map%d", i%3)).Build()
		a.CreateLink(ctx, link)
	}

	t.Run("CreateLink should not write to underlying store", func(t *testing.T) {
		ctx = context.Background()
		b := initBatch(t, a)

		link := cstesting.RandomLink()
		linkHash, err := b.CreateLink(ctx, link)
		assert.NoError(t, err, "b.CreateLink()")

		found, err := a.GetSegment(ctx, linkHash)
		assert.NoError(t, err, "a.GetSegment()")
		assert.Nil(t, found, "Link should not be found in adapter until Write is called")
	})

	t.Run("Write should write to the underlying store", func(t *testing.T) {
		ctx = context.Background()
		b := initBatch(t, a)

		link := cstesting.RandomLink()
		linkHash, err := b.CreateLink(ctx, link)
		assert.NoError(t, err, "b.CreateLink()")

		err = b.Write(ctx)
		assert.NoError(t, err, "b.Write()")

		found, err := a.GetSegment(ctx, linkHash)
		assert.NoError(t, err, "a.GetSegment()")
		require.NotNil(t, found, "a.GetSegment()")
		assert.EqualValues(t, *link, found.Link, "Link should be found in adapter after a Write")
	})

	t.Run("Finding segments should find in both batch and underlying store", func(t *testing.T) {
		ctx = context.Background()
		b := initBatch(t, a)

		var segs cs.SegmentSlice
		var err error
		segs, err = b.FindSegments(ctx, &store.SegmentFilter{Pagination: store.Pagination{Limit: store.DefaultLimit}})
		assert.NoError(t, err, "b.FindSegments()")
		adapterLinksCount := len(segs)

		b.CreateLink(ctx, cstesting.RandomLink())
		b.CreateLink(ctx, cstesting.RandomLink())

		segs, err = b.FindSegments(ctx, &store.SegmentFilter{Pagination: store.Pagination{Limit: store.DefaultLimit}})
		assert.NoError(t, err, "b.FindSegments()")
		assert.Equal(t, adapterLinksCount+2, len(segs), "Invalid number of segments found")
	})

	t.Run("Finding maps should find in both batch and underlying store", func(t *testing.T) {
		ctx = context.Background()
		b := initBatch(t, a)

		mapIDs, err := b.GetMapIDs(ctx, &store.MapFilter{Pagination: store.Pagination{Limit: store.DefaultLimit}})
		assert.NoError(t, err, "b.GetMapIDs()")
		adapterMapIdsCount := len(mapIDs)

		for _, mapID := range []string{"map42", "map43"} {
			link := cstesting.NewLinkBuilder().WithMapID(mapID).Build()
			b.CreateLink(ctx, link)
		}

		mapIDs, err = b.GetMapIDs(ctx, &store.MapFilter{Pagination: store.Pagination{Limit: store.DefaultLimit}})
		assert.NoError(t, err, "b.GetMapIDs()")
		assert.Equal(t, adapterMapIdsCount+2, len(mapIDs), "Invalid number of maps")

		want := map[string]interface{}{"map0": nil, "map1": nil, "map2": nil, "map42": nil, "map43": nil}
		got := make(map[string]interface{}, len(mapIDs))
		for _, mapID := range mapIDs {
			got[mapID] = nil
		}

		for mapID := range want {
			_, exist := got[mapID]
			assert.True(t, exist, "Missing map: %s", mapID)
		}
	})
}
