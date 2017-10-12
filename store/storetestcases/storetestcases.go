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

// Package storetestcases defines test cases to test stores.
package storetestcases

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/cs/cstesting"
	"github.com/stratumn/sdk/store"
	"github.com/stratumn/sdk/types"
)

// Factory wraps functions to allocate and free an adapter, and is used to run
// the tests on an adapter.
type Factory struct {
	// New creates an adapter.
	New func() (store.Adapter, error)

	// Free is an optional function to free an adapter.
	Free func(adapter store.Adapter)
}

// RunTests runs all the tests.
func (f Factory) RunTests(t *testing.T) {
	t.Run("AddDidSaveChannel", f.TestAddDidSaveChannel)
	t.Run("BatchDeleteSegment", f.TestBatchDeleteSegment)
	t.Run("BatchDeleteValue", f.TestBatchDeleteValue)
	t.Run("TestBatchFindSegments", f.TestBatchFindSegments)
	t.Run("TestBatchGetMapIDs", f.TestBatchGetMapIDs)
	t.Run("BatchSaveSegment", f.TestBatchSaveSegment)
	t.Run("BatchSaveValue", f.TestBatchSaveValue)
	t.Run("BatchWriteDeleteValue", f.TestBatchWriteDeleteValue)
	t.Run("BatchWriteDeleteSegment", f.TestBatchWriteDeleteSegment)
	t.Run("BatchWriteSaveSegment", f.TestBatchWriteSaveSegment)
	t.Run("BatchWriteSaveValue", f.TestBatchWriteSaveValue)
	t.Run("DeleteSegment", f.TestDeleteSegment)
	t.Run("DeleteSegmentNotFound", f.TestDeleteSegmentNotFound)
	t.Run("DeleteValue", f.TestDeleteValue)
	t.Run("DeleteValueNotFound", f.TestDeleteValueNotFound)
	t.Run("FindSegments", f.TestFindSegments)
	t.Run("FindSegmentsPagination", f.TestFindSegmentsPagination)
	t.Run("FindSegmentEmpty", f.TestFindSegmentEmpty)
	t.Run("FindSegmentsSingleTag", f.TestFindSegmentsSingleTag)
	t.Run("FindSegmentsMultipleTags", f.TestFindSegmentsMultipleTags)
	t.Run("FindSegmentsMapID", f.TestFindSegmentsMapID)
	t.Run("FindSegmentsMapIDs", f.TestFindSegmentsMapIDs)
	t.Run("FindSegmentsMapIDTags", f.TestFindSegmentsMapIDTags)
	t.Run("FindSegmentsMapIDNotFound", f.TestFindSegmentsMapIDNotFound)
	t.Run("TestFindSegmentsEmptyPrevLinkHash", f.TestFindSegmentsEmptyPrevLinkHash)
	t.Run("FindSegmentsPrevLinkHash", f.TestFindSegmentsPrevLinkHash)
	t.Run("FindSegmentsPrevLinkHashTags", f.TestFindSegmentsPrevLinkHashTags)
	t.Run("FindSegmentsPrevLinkHashGoodMapID", f.TestFindSegmentsPrevLinkHashGoodMapID)
	t.Run("FindSegmentsPrevLinkHashBadMapID", f.TestFindSegmentsPrevLinkHashBadMapID)
	t.Run("FindSegmentsPrevLinkHashNotFound", f.TestFindSegmentsPrevLinkHashNotFound)
	t.Run("TestFindSegmentWithGoodProcess", f.TestFindSegmentWithGoodProcess)
	t.Run("TestFindSegmentWithBadProcess", f.TestFindSegmentWithBadProcess)
	t.Run("GetInfo", f.TestGetInfo)
	t.Run("GetMapIDs", f.TestGetMapIDs)
	t.Run("GetMapIDsPagination", f.TestGetMapIDsPagination)
	t.Run("GetMapIDs_empty", f.TestGetMapIDsEmpty)
	t.Run("GetMapIDsByProcess", f.TestGetMapIDsByProcess)
	t.Run("GetSegment", f.TestGetSegment)
	t.Run("GetSegmentUpdatedState", f.TestGetSegmentUpdatedState)
	t.Run("GetSegmentUpdatedMapID", f.TestGetSegmentUpdatedMapID)
	t.Run("GetSegmentNotFound", f.TestGetSegmentNotFound)
	t.Run("GetValue", f.TestGetValue)
	t.Run("GetValueNotFound", f.TestGetValueNotFound)
	t.Run("SaveSegment", f.TestSaveSegment)
	t.Run("SaveValue", f.TestSaveValue)
	t.Run("SaveSegmentUpdatedState", f.TestSaveSegmentUpdatedState)
	t.Run("SaveSegmentUpdatedMapID", f.TestSaveSegmentUpdatedMapID)
	t.Run("SaveSegmentBranch", f.TestSaveSegmentBranch)
}

// RunBenchmarks runs all the benchmarks.
func (f Factory) RunBenchmarks(b *testing.B) {
	b.Run("DeleteSegment", f.BenchmarkDeleteSegment)
	b.Run("DeleteSegmentParallel", f.BenchmarkDeleteSegmentParallel)
	b.Run("DeleteValue", f.BenchmarkDeleteValue)
	b.Run("DeleteValueParallel", f.BenchmarkDeleteValueParallel)
	b.Run("FindSegments100", f.BenchmarkFindSegments100)
	b.Run("FindSegments1000", f.BenchmarkFindSegments1000)
	b.Run("FindSegments10000", f.BenchmarkFindSegments10000)
	b.Run("FindSegmentsMapID100", f.BenchmarkFindSegmentsMapID100)
	b.Run("FindSegmentsMapID1000", f.BenchmarkFindSegmentsMapID1000)
	b.Run("FindSegmentsMapID10000", f.BenchmarkFindSegmentsMapID10000)
	b.Run("FindSegmentsMapIDs100", f.BenchmarkFindSegmentsMapIDs100)
	b.Run("FindSegmentsMapIDs1000", f.BenchmarkFindSegmentsMapIDs1000)
	b.Run("FindSegmentsMapIDs10000", f.BenchmarkFindSegmentsMapIDs10000)
	b.Run("FindSegmentsPrevLinkHash100", f.BenchmarkFindSegmentsPrevLinkHash100)
	b.Run("FindSegmentsPrevLinkHash1000", f.BenchmarkFindSegmentsPrevLinkHash1000)
	b.Run("FindSegmentsPrevLinkHash10000", f.BenchmarkFindSegmentsPrevLinkHash10000)
	b.Run("FindSegmentsTags100", f.BenchmarkFindSegmentsTags100)
	b.Run("FindSegmentsTags1000", f.BenchmarkFindSegmentsTags1000)
	b.Run("FindSegmentsTags10000", f.BenchmarkFindSegmentsTags10000)
	b.Run("FindSegmentsMapIDTags100", f.BenchmarkFindSegmentsMapIDTags100)
	b.Run("FindSegmentsMapIDTags1000", f.BenchmarkFindSegmentsMapIDTags1000)
	b.Run("FindSegmentsMapIDTags10000", f.BenchmarkFindSegmentsMapIDTags10000)
	b.Run("FindSegmentsPrevLinkHashTags100", f.BenchmarkFindSegmentsPrevLinkHashTags100)
	b.Run("FindSegmentsPrevLinkHashTags1000", f.BenchmarkFindSegmentsPrevLinkHashTags1000)
	b.Run("FindSegmentsPrevLinkHashTags10000", f.BenchmarkFindSegmentsPrevLinkHashTags10000)
	b.Run("FindSegments100Parallel", f.BenchmarkFindSegments100Parallel)
	b.Run("FindSegments1000Parallel", f.BenchmarkFindSegments1000Parallel)
	b.Run("FindSegments10000Parallel", f.BenchmarkFindSegments10000Parallel)
	b.Run("FindSegmentsMapID100Parallel", f.BenchmarkFindSegmentsMapID100Parallel)
	b.Run("FindSegmentsMapID1000Parallel", f.BenchmarkFindSegmentsMapID1000Parallel)
	b.Run("FindSegmentsMapID10000Parallel", f.BenchmarkFindSegmentsMapID10000Parallel)
	b.Run("FindSegmentsMapIDs100Parallel", f.BenchmarkFindSegmentsMapIDs100Parallel)
	b.Run("FindSegmentsMapIDs1000Parallel", f.BenchmarkFindSegmentsMapIDs1000Parallel)
	b.Run("FindSegmentsMapIDs10000Parallel", f.BenchmarkFindSegmentsMapIDs10000Parallel)
	b.Run("FindSegmentsPrevLinkHash100Parallel", f.BenchmarkFindSegmentsPrevLinkHash100Parallel)
	b.Run("FindSegmentsPrevLinkHash1000Parallel", f.BenchmarkFindSegmentsPrevLinkHash1000Parallel)
	b.Run("FindSegmentsPrevLinkHash10000Parallel", f.BenchmarkFindSegmentsPrevLinkHash10000Parallel)
	b.Run("FindSegmentsTags100Parallel", f.BenchmarkFindSegmentsTags100Parallel)
	b.Run("FindSegmentsTags1000Parallel", f.BenchmarkFindSegmentsTags1000Parallel)
	b.Run("FindSegmentsTags10000Parallel", f.BenchmarkFindSegmentsTags10000Parallel)
	b.Run("FindSegmentsMapIDTags100Parallel", f.BenchmarkFindSegmentsMapIDTags100Parallel)
	b.Run("FindSegmentsMapIDTags1000Parallel", f.BenchmarkFindSegmentsMapIDTags1000Parallel)
	b.Run("FindSegmentsMapIDTags10000Parallel", f.BenchmarkFindSegmentsMapIDTags10000Parallel)
	b.Run("FindSegmentsPrevLinkHashTags100Parallel", f.BenchmarkFindSegmentsPrevLinkHashTags100Parallel)
	b.Run("FindSegmentsPrevLinkHashTags1000Parallel", f.BenchmarkFindSegmentsPrevLinkHashTags1000Parallel)
	b.Run("FindSegmentsPrevLinkHashTags10000Parallel", f.BenchmarkFindSegmentsPrevLinkHashTags10000Parallel)
	b.Run("GetMapIDs100", f.BenchmarkGetMapIDs100)
	b.Run("GetMapIDs1000", f.BenchmarkGetMapIDs1000)
	b.Run("GetMapIDs10000", f.BenchmarkGetMapIDs10000)
	b.Run("GetMapIDs100Parallel", f.BenchmarkGetMapIDs100Parallel)
	b.Run("GetMapIDs1000Parallel", f.BenchmarkGetMapIDs1000Parallel)
	b.Run("GetMapIDs10000Parallel", f.BenchmarkGetMapIDs10000Parallel)
	b.Run("GetSegment", f.BenchmarkGetSegment)
	b.Run("GetSegmentParallel", f.BenchmarkGetSegmentParallel)
	b.Run("GetValue", f.BenchmarkGetValue)
	b.Run("GetValueParallel", f.BenchmarkGetValueParallel)
	b.Run("SaveSegment", f.BenchmarkSaveSegment)
	b.Run("SaveSegmentParallel", f.BenchmarkSaveSegmentParallel)
	b.Run("SaveSegmentUpdatedState", f.BenchmarkSaveSegmentUpdatedState)
	b.Run("SaveSegmentUpdatedStateParallel", f.BenchmarkSaveSegmentUpdatedStateParallel)
	b.Run("SaveSegmentUpdatedMapID", f.BenchmarkSaveSegmentUpdatedMapID)
	b.Run("SaveSegmentUpdatedMapIDParallel", f.BenchmarkSaveSegmentUpdatedMapIDParallel)
	b.Run("SaveValue", f.BenchmarkSaveValue)
	b.Run("SaveValueParallel", f.BenchmarkSaveValueParallel)
}

func (f Factory) free(adapter store.Adapter) {
	if f.Free != nil {
		f.Free(adapter)
	}
}

// SegmentFunc is a type for a function that creates a segment for benchmarks.
type SegmentFunc func(b *testing.B, numSegments, i int) *cs.Segment

// RandomSegment is a SegmentFunc that create a random segment.
func RandomSegment(b *testing.B, numSegments, i int) *cs.Segment {
	return cstesting.RandomSegment()
}

// RandomSegmentMapID is a SegmentFunc that create a random segment with map ID.
// The map ID will be one of ten possible values.
func RandomSegmentMapID(b *testing.B, numSegments, i int) *cs.Segment {
	s := cstesting.RandomSegment()
	s.Link.Meta["mapId"] = fmt.Sprintf("%d", i%10)
	return s
}

// RandomSegmentPrevLinkHash is a SegmentFunc that create a random segment with
// previous link hash.
// The previous link hash will be one of ten possible values.
func RandomSegmentPrevLinkHash(b *testing.B, numSegments, i int) *cs.Segment {
	s := cstesting.RandomSegment()
	s.Link.Meta["prevLinkHash"] = fmt.Sprintf("00000000000000000000000000000000000000000000000000000000000000%2d", i%10)
	return s
}

// RandomSegmentTags is a SegmentFunc that create a random segment with tags.
// The tags will contain one of ten possible values.
func RandomSegmentTags(b *testing.B, numSegments, i int) *cs.Segment {
	s := cstesting.RandomSegment()
	s.Link.Meta["tags"] = []interface{}{fmt.Sprintf("%d", i%10)}
	return s
}

// RandomSegmentMapIDTags is a SegmentFunc that create a random segment with map
// ID and tags.
// The map ID will be one of ten possible values.
// The tags will contain one of ten possible values.
func RandomSegmentMapIDTags(b *testing.B, numSegments, i int) *cs.Segment {
	s := cstesting.RandomSegment()
	s.Link.Meta["mapId"] = fmt.Sprintf("%d", i%10)
	s.Link.Meta["tags"] = []interface{}{fmt.Sprintf("%d", i%10)}
	return s
}

// RandomSegmentPrevLinkHashTags is a SegmentFunc that create a random segment
// with previous link hash and tags.
// The previous link hash will be one of ten possible values.
// The tags will contain one of ten possible values.
func RandomSegmentPrevLinkHashTags(b *testing.B, numSegments, i int) *cs.Segment {
	s := cstesting.RandomSegment()
	s.Link.Meta["prevLinkHash"] = fmt.Sprintf("00000000000000000000000000000000000000000000000000000000000000%2d", i%10)
	s.Link.Meta["tags"] = []interface{}{fmt.Sprintf("%d", i%10)}
	return s
}

// MapFilterFunc is a type for a function that creates a mapId filter for
// benchmarks.
type MapFilterFunc func(b *testing.B, numSegments, i int) *store.MapFilter

// RandomPaginationOffset is a a PaginationFunc that create a pagination with a random offset.
func RandomPaginationOffset(b *testing.B, numSegments, i int) *store.MapFilter {
	return &store.MapFilter{
		Pagination: store.Pagination{
			Offset: rand.Int() % numSegments,
			Limit:  store.DefaultLimit,
		},
	}
}

// FilterFunc is a type for a function that creates a filter for benchmarks.
type FilterFunc func(b *testing.B, numSegments, i int) *store.SegmentFilter

// RandomFilterOffset is a a FilterFunc that create a filter with a random
// offset.
func RandomFilterOffset(b *testing.B, numSegments, i int) *store.SegmentFilter {
	return &store.SegmentFilter{
		Pagination: store.Pagination{
			Offset: rand.Int() % numSegments,
			Limit:  store.DefaultLimit,
		},
	}
}

// RandomFilterOffsetMapID is a a FilterFunc that create a filter with a random
// offset and map ID.
// The map ID will be one of ten possible values.
func RandomFilterOffsetMapID(b *testing.B, numSegments, i int) *store.SegmentFilter {
	return &store.SegmentFilter{
		Pagination: store.Pagination{
			Offset: rand.Int() % numSegments,
			Limit:  store.DefaultLimit,
		},
		MapIDs: []string{fmt.Sprintf("%d", i%10)},
	}
}

// RandomFilterOffsetMapIDs is a a FilterFunc that create a filter with a random
// offset and 2 map IDs.
// The map ID will be one of ten possible values.
func RandomFilterOffsetMapIDs(b *testing.B, numSegments, i int) *store.SegmentFilter {
	return &store.SegmentFilter{
		Pagination: store.Pagination{
			Offset: rand.Int() % numSegments,
			Limit:  store.DefaultLimit,
		},
		MapIDs: []string{fmt.Sprintf("%d", i%10), fmt.Sprintf("%d", (i+1)%10)},
	}
}

// RandomFilterOffsetPrevLinkHash is a a FilterFunc that create a filter with a
// random offset and previous link hash.
// The previous link hash will be one of ten possible values.
func RandomFilterOffsetPrevLinkHash(b *testing.B, numSegments, i int) *store.SegmentFilter {
	prevLinkHash, _ := types.NewBytes32FromString(fmt.Sprintf("00000000000000000000000000000000000000000000000000000000000000%2d", i%10))
	prevLinkHashStr := prevLinkHash.String()
	return &store.SegmentFilter{
		Pagination: store.Pagination{
			Offset: rand.Int() % numSegments,
			Limit:  store.DefaultLimit,
		},
		PrevLinkHash: &prevLinkHashStr,
	}
}

// RandomFilterOffsetTags is a a FilterFunc that create a filter with a random
// offset and map ID.
// The tags will be one of fifty possible combinations.
func RandomFilterOffsetTags(b *testing.B, numSegments, i int) *store.SegmentFilter {
	return &store.SegmentFilter{
		Pagination: store.Pagination{
			Offset: rand.Int() % numSegments,
			Limit:  store.DefaultLimit,
		},
		Tags: []string{fmt.Sprintf("%d", i%5), fmt.Sprintf("%d", i%10)},
	}
}

// RandomFilterOffsetMapIDTags is a a FilterFunc that create a filter with a
// random offset and map ID and tags.
// The map ID will be one of ten possible values.
// The tags will be one of fifty possible combinations.
func RandomFilterOffsetMapIDTags(b *testing.B, numSegments, i int) *store.SegmentFilter {
	return &store.SegmentFilter{
		Pagination: store.Pagination{
			Offset: rand.Int() % numSegments,
			Limit:  store.DefaultLimit,
		},
		MapIDs: []string{fmt.Sprintf("%d", i%10)},
		Tags:   []string{fmt.Sprintf("%d", i%5), fmt.Sprintf("%d", i%10)},
	}
}

// RandomFilterOffsetPrevLinkHashTags is a a FilterFunc that create a filter
// with a random offset and previous link hash and tags.
// The previous link hash will be one of ten possible values.
// The tags will be one of fifty possible combinations.
func RandomFilterOffsetPrevLinkHashTags(b *testing.B, numSegments, i int) *store.SegmentFilter {
	prevLinkHash, _ := types.NewBytes32FromString(fmt.Sprintf("00000000000000000000000000000000000000000000000000000000000000%2d", i%10))
	prevLinkHashStr := prevLinkHash.String()
	return &store.SegmentFilter{
		Pagination: store.Pagination{
			Offset: rand.Int() % numSegments,
			Limit:  store.DefaultLimit,
		},
		PrevLinkHash: &prevLinkHashStr,
		Tags:         []string{fmt.Sprintf("%d", i%5), fmt.Sprintf("%d", i%10)},
	}
}

func (f Factory) initAdapter(t *testing.T) store.Adapter {
	a, err := f.New()
	if err != nil {
		t.Fatalf("f.New(): err: %s", err)
	}
	if a == nil {
		t.Fatal("a = nil want store.Adapter")
	}
	return a
}

func (f Factory) initAdapterB(b *testing.B) store.Adapter {
	a, err := f.New()
	if err != nil {
		b.Fatalf("f.New(): err: %s", err)
	}
	if a == nil {
		b.Fatal("a = nil want store.Adapter")
	}
	return a
}
