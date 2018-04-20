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
	"errors"
	"testing"

	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/cs/cstesting"
	"github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/store/storetesting"
	"github.com/stratumn/go-indigocore/testutil"
	"github.com/stratumn/go-indigocore/types"
	"github.com/stretchr/testify/assert"
)

func TestBatch_CreateLink(t *testing.T) {
	ctx := context.Background()

	a := &storetesting.MockAdapter{}
	batch := NewBatch(ctx, a)

	l := cstesting.RandomLink()

	wantedErr := errors.New("error on MockCreateLink")
	a.MockCreateLink.Fn = func(link *cs.Link) (*types.Bytes32, error) { return nil, wantedErr }

	_, err := batch.CreateLink(ctx, l)
	assert.NoError(t, err)
	assert.Equal(t, 0, a.MockCreateLink.CalledCount)
	assert.Equal(t, 1, len(batch.Links))

	// Batch shouldn't do any kind of validation.
	l.Meta.MapID = ""
	_, err = batch.CreateLink(ctx, l)
	assert.NoError(t, err)
}

func TestBatch_GetSegment(t *testing.T) {
	ctx := context.Background()

	a := &storetesting.MockAdapter{}
	batch := NewBatch(ctx, a)

	storedLink := cstesting.RandomLink()
	storedLinkHash, _ := storedLink.Hash()
	batchLink1 := cstesting.RandomLink()
	batchLink2 := cstesting.RandomLink()

	batchLinkHash1, _ := batch.CreateLink(ctx, batchLink1)
	batchLinkHash2, _ := batch.CreateLink(ctx, batchLink2)

	notFoundErr := errors.New("Unit test error")
	a.MockGetSegment.Fn = func(linkHash *types.Bytes32) (*cs.Segment, error) {
		if *storedLinkHash == *linkHash {
			return storedLink.Segmentify(), nil
		}

		return nil, notFoundErr
	}

	var segment *cs.Segment
	var err error

	segment, err = batch.GetSegment(ctx, batchLinkHash1)
	assert.NoError(t, err, "batch.GetSegment()")
	assert.Equal(t, batchLink1, &segment.Link)

	segment, err = batch.GetSegment(ctx, batchLinkHash2)
	assert.NoError(t, err, "batch.GetSegment()")
	assert.Equal(t, batchLink2, &segment.Link)

	segment, err = batch.GetSegment(ctx, storedLinkHash)
	assert.NoError(t, err, "batch.GetSegment()")
	assert.Equal(t, storedLink, &segment.Link)

	segment, err = batch.GetSegment(ctx, testutil.RandomHash())
	assert.EqualError(t, err, notFoundErr.Error())
}

func TestBatch_FindSegments(t *testing.T) {
	ctx := context.Background()

	a := &storetesting.MockAdapter{}
	batch := NewBatch(ctx, a)

	storedLink := cstesting.RandomLink()
	storedLink.Meta.Process = "Foo"
	l1 := cstesting.NewLinkBuilder().WithProcess("Foo").Build()
	l2 := cstesting.NewLinkBuilder().WithProcess("Bar").Build()

	batch.CreateLink(ctx, l1)
	batch.CreateLink(ctx, l2)

	notFoundErr := errors.New("Unit test error")
	a.MockFindSegments.Fn = func(filter *store.SegmentFilter) (cs.SegmentSlice, error) {
		if filter.Process == "Foo" {
			return cs.SegmentSlice{storedLink.Segmentify()}, nil
		}
		if filter.Process == "Bar" {
			return cs.SegmentSlice{}, nil
		}

		return nil, notFoundErr
	}

	var segments cs.SegmentSlice
	var err error

	segments, err = batch.FindSegments(ctx, &store.SegmentFilter{Pagination: store.Pagination{Limit: store.DefaultLimit}, Process: "Foo"})
	assert.NoError(t, err, "batch.FindSegments()")
	assert.Equal(t, 2, len(segments))

	segments, err = batch.FindSegments(ctx, &store.SegmentFilter{Pagination: store.Pagination{Limit: store.DefaultLimit}, Process: "Bar"})
	assert.NoError(t, err, "batch.FindSegments()")
	assert.Equal(t, 1, len(segments))

	_, err = batch.FindSegments(ctx, &store.SegmentFilter{Pagination: store.Pagination{Limit: store.DefaultLimit}, Process: "NotFound"})
	assert.EqualError(t, err, notFoundErr.Error())
}

func TestBatch_GetMapIDs(t *testing.T) {
	ctx := context.Background()

	a := &storetesting.MockAdapter{}
	batch := NewBatch(ctx, a)

	storedLink1 := cstesting.RandomLink()
	storedLink1.Meta.MapID = "Foo1"
	storedLink1.Meta.Process = "FooProcess"
	storedLink2 := cstesting.RandomLink()
	storedLink2.Meta.MapID = "Bar"
	storedLink2.Meta.Process = "BarProcess"

	batchLink1 := cstesting.RandomLink()
	batchLink1.Meta.MapID = "Foo2"
	batchLink1.Meta.Process = "FooProcess"
	batchLink2 := cstesting.RandomLink()
	batchLink2.Meta.MapID = "Yin"
	batchLink2.Meta.Process = "YinProcess"

	batch.CreateLink(ctx, batchLink1)
	batch.CreateLink(ctx, batchLink2)

	a.MockGetMapIDs.Fn = func(filter *store.MapFilter) ([]string, error) {
		if filter.Process == storedLink1.Meta.Process {
			return []string{storedLink1.Meta.MapID}, nil
		}
		if filter.Process == storedLink2.Meta.Process {
			return []string{storedLink2.Meta.MapID}, nil
		}

		return []string{
			storedLink1.Meta.MapID,
			storedLink2.Meta.MapID,
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
		storedLink1.Meta.MapID,
		batchLink1.Meta.MapID,
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
	l := cstesting.RandomLink()

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

	la := cstesting.RandomLink()
	lb := cstesting.RandomLink()

	a.MockCreateLink.Fn = func(l *cs.Link) (*types.Bytes32, error) {
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
