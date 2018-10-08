// Copyright 2017-2018 Stratumn SAS. All rights reserved.
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
	"errors"
	"testing"

	"github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-chainscript/chainscripttest"
	"github.com/stratumn/go-core/store"
	"github.com/stratumn/go-core/store/storetesting"
	"github.com/stratumn/go-core/types"
	"github.com/stretchr/testify/assert"
)

func TestBatch_CreateLink(t *testing.T) {
	ctx := context.Background()

	a := &storetesting.MockAdapter{}
	batch := NewBatch(ctx, a)

	l := chainscripttest.RandomLink(t)

	wantedErr := errors.New("error on MockCreateLink")
	a.MockCreateLink.Fn = func(link *chainscript.Link) (chainscript.LinkHash, error) { return nil, wantedErr }

	_, err := batch.CreateLink(ctx, l)
	assert.NoError(t, err)
	assert.Equal(t, 0, a.MockCreateLink.CalledCount)
	assert.Equal(t, 1, len(batch.Links))

	// Batch shouldn't do any kind of validation.
	l.Meta.MapId = ""
	_, err = batch.CreateLink(ctx, l)
	assert.NoError(t, err)
}

func TestBatch_GetSegment(t *testing.T) {
	ctx := context.Background()

	a := &storetesting.MockAdapter{}
	batch := NewBatch(ctx, a)

	storedLink := chainscripttest.RandomLink(t)
	storedLinkHash, _ := storedLink.Hash()
	batchLink1 := chainscripttest.RandomLink(t)
	batchLink2 := chainscripttest.RandomLink(t)

	batchLinkHash1, _ := batch.CreateLink(ctx, batchLink1)
	batchLinkHash2, _ := batch.CreateLink(ctx, batchLink2)

	notFoundErr := errors.New("Unit test error")
	a.MockGetSegment.Fn = func(linkHash chainscript.LinkHash) (*chainscript.Segment, error) {
		if bytes.Equal(storedLinkHash, linkHash) {
			return storedLink.Segmentify()
		}

		return nil, notFoundErr
	}

	var segment *chainscript.Segment
	var err error

	segment, err = batch.GetSegment(ctx, batchLinkHash1)
	assert.NoError(t, err, "batch.GetSegment()")
	assert.Equal(t, batchLink1, segment.Link)

	segment, err = batch.GetSegment(ctx, batchLinkHash2)
	assert.NoError(t, err, "batch.GetSegment()")
	assert.Equal(t, batchLink2, segment.Link)

	segment, err = batch.GetSegment(ctx, storedLinkHash)
	assert.NoError(t, err, "batch.GetSegment()")
	assert.Equal(t, storedLink, segment.Link)

	segment, err = batch.GetSegment(ctx, chainscripttest.RandomHash())
	assert.EqualError(t, err, notFoundErr.Error())
}

func TestBatch_FindSegments(t *testing.T) {
	ctx := context.Background()

	a := &storetesting.MockAdapter{}
	batch := NewBatch(ctx, a)

	storedLink := chainscripttest.RandomLink(t)
	storedLink.Meta.Process.Name = "Foo"
	storedSegment, _ := storedLink.Segmentify()

	l1 := chainscripttest.NewLinkBuilder(t).WithProcess("Foo").Build()
	l2 := chainscripttest.NewLinkBuilder(t).WithProcess("Bar").Build()

	batch.CreateLink(ctx, l1)
	batch.CreateLink(ctx, l2)

	notFoundErr := errors.New("Unit test error")
	a.MockFindSegments.Fn = func(filter *store.SegmentFilter) (*types.PaginatedSegments, error) {
		if filter.Process == "Foo" {
			return &types.PaginatedSegments{
				Segments:   types.SegmentSlice{storedSegment},
				TotalCount: 1,
			}, nil
		}
		if filter.Process == "Bar" {
			return &types.PaginatedSegments{}, nil
		}

		return nil, notFoundErr
	}

	var segments *types.PaginatedSegments
	var err error

	segments, err = batch.FindSegments(ctx, &store.SegmentFilter{Pagination: store.Pagination{Limit: store.DefaultLimit}, Process: "Foo"})
	assert.NoError(t, err, "batch.FindSegments()")
	assert.Len(t, segments.Segments, 2)
	assert.Equal(t, 2, segments.TotalCount)

	segments, err = batch.FindSegments(ctx, &store.SegmentFilter{Pagination: store.Pagination{Limit: store.DefaultLimit}, Process: "Bar"})
	assert.NoError(t, err, "batch.FindSegments()")
	assert.Len(t, segments.Segments, 1)
	assert.Equal(t, 1, segments.TotalCount)

	_, err = batch.FindSegments(ctx, &store.SegmentFilter{Pagination: store.Pagination{Limit: store.DefaultLimit}, Process: "NotFound"})
	assert.EqualError(t, err, notFoundErr.Error())
}

func TestBatch_GetMapIDs(t *testing.T) {
	ctx := context.Background()

	a := &storetesting.MockAdapter{}
	batch := NewBatch(ctx, a)

	storedLink1 := chainscripttest.RandomLink(t)
	storedLink1.Meta.MapId = "Foo1"
	storedLink1.Meta.Process.Name = "FooProcess"
	storedLink2 := chainscripttest.RandomLink(t)
	storedLink2.Meta.MapId = "Bar"
	storedLink2.Meta.Process.Name = "BarProcess"

	batchLink1 := chainscripttest.RandomLink(t)
	batchLink1.Meta.MapId = "Foo2"
	batchLink1.Meta.Process.Name = "FooProcess"
	batchLink2 := chainscripttest.RandomLink(t)
	batchLink2.Meta.MapId = "Yin"
	batchLink2.Meta.Process.Name = "YinProcess"

	batch.CreateLink(ctx, batchLink1)
	batch.CreateLink(ctx, batchLink2)

	a.MockGetMapIDs.Fn = func(filter *store.MapFilter) ([]string, error) {
		if filter.Process == storedLink1.Meta.Process.Name {
			return []string{storedLink1.Meta.MapId}, nil
		}
		if filter.Process == storedLink2.Meta.Process.Name {
			return []string{storedLink2.Meta.MapId}, nil
		}

		return []string{
			storedLink1.Meta.MapId,
			storedLink2.Meta.MapId,
		}, nil
	}

	var mapIDs []string
	var err error

	mapIDs, err = batch.GetMapIDs(ctx, &store.MapFilter{Pagination: store.Pagination{Limit: store.DefaultLimit}})
	assert.NoError(t, err, "batch.GetMapIDs()")
	assert.Equal(t, 4, len(mapIDs))

	processFilter := &store.MapFilter{
		Process:    "FooProcess",
		Pagination: store.Pagination{Limit: store.DefaultLimit},
	}
	mapIDs, err = batch.GetMapIDs(ctx, processFilter)
	assert.NoError(t, err, "batch.GetMapIDs()")
	assert.Equal(t, 2, len(mapIDs))

	for _, mapID := range []string{
		storedLink1.Meta.MapId,
		batchLink1.Meta.MapId,
	} {
		assert.True(t, mapIDs[0] == mapID || mapIDs[1] == mapID)
	}
}

func TestBatch_GetMapIDsWithStoreReturningAnErrorOnGetMapIDs(t *testing.T) {
	ctx := context.Background()

	a := &storetesting.MockAdapter{}
	batch := NewBatch(ctx, a)

	wantedMapIds := []string{"Foo", "Bar"}
	notFoundErr := errors.New("Unit test error")
	a.MockGetMapIDs.Fn = func(filter *store.MapFilter) ([]string, error) {
		return wantedMapIds, notFoundErr
	}

	mapIDs, err := batch.GetMapIDs(ctx, &store.MapFilter{})
	assert.EqualError(t, err, notFoundErr.Error(), "batch.GetMapIDs()")
	assert.Equal(t, len(wantedMapIds), len(mapIDs))
	assert.Equal(t, wantedMapIds, mapIDs)
}

func TestBatch_WriteLink(t *testing.T) {
	ctx := context.Background()

	a := &storetesting.MockAdapter{}
	l := chainscripttest.RandomLink(t)

	batch := NewBatch(ctx, a)

	_, err := batch.CreateLink(ctx, l)
	assert.NoError(t, err, "batch.CreateLink()")

	err = batch.Write(ctx)
	assert.NoError(t, err, "batch.Write()")
	assert.Equal(t, 1, a.MockCreateLink.CalledCount)
	assert.Equal(t, l, a.MockCreateLink.LastCalledWith)
}

func TestBatch_WriteLinkWithFailure(t *testing.T) {
	ctx := context.Background()

	a := &storetesting.MockAdapter{}
	mockError := errors.New("Error")

	la := chainscripttest.RandomLink(t)
	lb := chainscripttest.RandomLink(t)

	a.MockCreateLink.Fn = func(l *chainscript.Link) (chainscript.LinkHash, error) {
		if l == la {
			return nil, mockError
		}
		return l.Hash()
	}

	batch := NewBatch(ctx, a)

	_, err := batch.CreateLink(ctx, la)
	assert.NoError(t, err, "batch.CreateLink()")

	_, err = batch.CreateLink(ctx, lb)
	assert.NoError(t, err, "batch.CreateLink()")

	err = batch.Write(ctx)
	assert.EqualError(t, err, mockError.Error())
}
