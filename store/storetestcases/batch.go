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
	"fmt"
	"testing"

	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/cs/cstesting"
	"github.com/stratumn/sdk/store"
	"github.com/stretchr/testify/assert"
)

// TestBatchCreateLink tests what happens
// when you create a link in a Batch
func (f Factory) TestBatchCreateLink(t *testing.T) {
	a, b := f.initBatch(t)
	defer f.freeAdapter(a)

	link := cstesting.RandomLink()
	linkHash, err := b.CreateLink(link)
	assert.NoError(t, err, "b.CreateLink()")

	found, err := a.GetSegment(linkHash)
	assert.NoError(t, err, "a.GetSegment()")
	assert.Nil(t, found, "Link should not be found in adapter until Write is called")
}

// TestBatchWriteCreateLink tests what happens when you write a Batch with a created link.
func (f Factory) TestBatchWriteCreateLink(t *testing.T) {
	a, b := f.initBatch(t)
	defer f.freeAdapter(a)

	link := cstesting.RandomLink()
	linkHash, err := b.CreateLink(link)
	assert.NoError(t, err, "b.CreateLink()")

	err = b.Write()
	assert.NoError(t, err, "b.Write()")

	found, err := a.GetSegment(linkHash)
	assert.NoError(t, err, "a.GetSegment()")
	assert.EqualValues(t, *link, found.Link, "Link should be found in adapter after a Write")
}

// TestBatchFindSegments tests what happens when you find segments in batch and store.
func (f Factory) TestBatchFindSegments(t *testing.T) {
	a, b := f.initBatch(t)
	defer f.freeAdapter(a)

	var nbLinks = store.DefaultLimit / 2
	for i := 0; i < nbLinks; i++ {
		a.CreateLink(cstesting.RandomLink())
	}

	var segs cs.SegmentSlice
	var err error
	segs, err = b.FindSegments(&store.SegmentFilter{Pagination: store.Pagination{Limit: store.DefaultLimit}})
	assert.NoError(t, err, "b.FindSegments()")
	assert.Equal(t, nbLinks, len(segs), "Invalid number of segments found")

	b.CreateLink(cstesting.RandomLink())
	b.CreateLink(cstesting.RandomLink())

	segs, err = b.FindSegments(&store.SegmentFilter{Pagination: store.Pagination{Limit: store.DefaultLimit}})
	assert.NoError(t, err, "b.FindSegments()")
	assert.Equal(t, nbLinks+2, len(segs), "Invalid number of segments found")
}

// TestBatchGetMapIDs tests what happens when you get mapIds in batch and store.
func (f Factory) TestBatchGetMapIDs(t *testing.T) {
	a, b := f.initBatch(t)
	defer f.freeAdapter(a)

	segsByMapID := make(map[string]cs.SegmentSlice, 3)

	for i := 0; i < 6*store.DefaultLimit; i++ {
		link := cstesting.RandomLink()
		mapID := fmt.Sprintf("map%d", i%3)
		link.Meta["mapId"] = mapID
		if i < 3 {
			segsByMapID[mapID] = make(cs.SegmentSlice, 0, 2*store.DefaultLimit)
		}
		segsByMapID[mapID] = append(segsByMapID[mapID], link.Segmentify())
		a.CreateLink(link)
	}

	mapIDs, err := b.GetMapIDs(&store.MapFilter{Pagination: store.Pagination{Limit: store.DefaultLimit}})
	assert.NoError(t, err, "b.GetMapIDs()")
	assert.Equal(t, len(segsByMapID), len(mapIDs), "Invalid number of maps")

	for _, mapID := range []string{"map42", "map43"} {
		link := cstesting.RandomLink()
		link.Meta["mapId"] = mapID
		b.CreateLink(link)
	}

	mapIDs, err = b.GetMapIDs(&store.MapFilter{Pagination: store.Pagination{Limit: store.DefaultLimit}})
	assert.NoError(t, err, "b.GetMapIDs()")
	assert.Equal(t, len(segsByMapID)+2, len(mapIDs), "Invalid number of maps")

	want := map[string]interface{}{"map0": nil, "map1": nil, "map2": nil, "map42": nil, "map43": nil}
	got := make(map[string]interface{}, len(mapIDs))
	for _, mapID := range mapIDs {
		got[mapID] = nil
	}

	assert.Equal(t, len(want), len(got), "Invalid maps returned")
	for mapID := range got {
		_, exist := want[mapID]
		assert.True(t, exist, "Missing map: %s", mapID)
	}
}

func (f Factory) initBatch(t *testing.T) (store.Adapter, store.Batch) {
	a := f.initAdapter(t)

	b, err := a.NewBatch()
	if err != nil {
		t.Fatalf("a.NewBatch(): err: %s", err)
	}
	if b == nil {
		t.Fatal("b = nil want store.Batch")
	}

	return a, b
}
