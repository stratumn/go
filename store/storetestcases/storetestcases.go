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

	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/cs/cstesting"
	"github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/types"
	"github.com/stretchr/testify/assert"
)

// Factory wraps functions to allocate and free an adapter,
// and is used to run the tests on an adapter.
type Factory struct {
	// New creates an adapter.
	New func() (store.Adapter, error)

	// Free is an optional function to free an adapter.
	Free func(adapter store.Adapter)

	// NewKeyValueStore creates a KeyValueStore.
	// If your store implements the KeyValueStore interface,
	// you need to implement this method.
	NewKeyValueStore func() (store.KeyValueStore, error)

	// FreeKeyValueStore is an optional function to free
	// a KeyValueStore adapter.
	FreeKeyValueStore func(adapter store.KeyValueStore)
}

// RunKeyValueStoreTests runs all the tests for the key value store interface.
func (f Factory) RunKeyValueStoreTests(t *testing.T) {
	t.Run("TestKeyValueStore", f.TestKeyValueStore)
}

// RunStoreTests runs all the tests for the store adapter interface.
func (f Factory) RunStoreTests(t *testing.T) {
	t.Run("Test store events", f.TestStoreEvents)
	t.Run("Test store info", f.TestGetInfo)
	t.Run("Test finding segments", f.TestFindSegments)
	t.Run("Test getting map IDs", f.TestGetMapIDs)
	t.Run("Test getting segments", f.TestGetSegment)
	t.Run("Test creating links", f.TestCreateLink)
	t.Run("Test batch implementation", f.TestBatch)
	t.Run("Test evidence store", f.TestEvidenceStore)
}

// RunStoreBenchmarks runs all the benchmarks for the store adapter interface.
func (f Factory) RunStoreBenchmarks(b *testing.B) {
	b.Run("BenchmarkCreateLink", f.BenchmarkCreateLink)
	b.Run("BenchmarkCreateLinkParallel", f.BenchmarkCreateLinkParallel)

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
	b.Run("FindSegmentsPrevLinkHash10000ParalleRunBenchmarksl", f.BenchmarkFindSegmentsPrevLinkHash10000Parallel)
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
}

// RunKeyValueStoreBenchmarks runs all the benchmarks for the key-value store interface.
func (f Factory) RunKeyValueStoreBenchmarks(b *testing.B) {
	b.Run("GetValue", f.BenchmarkGetValue)
	b.Run("GetValueParallel", f.BenchmarkGetValueParallel)

	b.Run("SetValue", f.BenchmarkSetValue)
	b.Run("SetValueParallel", f.BenchmarkSetValueParallel)

	b.Run("DeleteValue", f.BenchmarkDeleteValue)
	b.Run("DeleteValueParallel", f.BenchmarkDeleteValueParallel)
}

func (f Factory) initAdapter(t *testing.T) store.Adapter {
	a, err := f.New()
	assert.NoError(t, err, "f.New()")
	assert.NotNil(t, a, "Store.Adapter")
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

func (f Factory) freeAdapter(adapter store.Adapter) {
	if f.Free != nil {
		f.Free(adapter)
	}
}

func (f Factory) initKeyValueStore(t *testing.T) store.KeyValueStore {
	a, err := f.NewKeyValueStore()
	assert.NoError(t, err, "f.NewKeyValueStore()")
	assert.NotNil(t, a, "Store.KeyValueStore")
	return a
}

func (f Factory) initKeyValueStoreB(b *testing.B) store.KeyValueStore {
	a, err := f.NewKeyValueStore()
	if err != nil {
		b.Fatalf("f.NewKeyValueStore(): err: %s", err)
	}
	if a == nil {
		b.Fatal("a = nil want store.KeyValueStore")
	}
	return a
}

func (f Factory) freeKeyValueStore(adapter store.KeyValueStore) {
	if f.FreeKeyValueStore != nil {
		f.FreeKeyValueStore(adapter)
	}
}

// CreateLinkFunc is a type for a function that creates a link for benchmarks.
type CreateLinkFunc func(b *testing.B, numLinks, i int) *cs.Link

// RandomLink is a CreateLinkFunc that creates a random segment.
func RandomLink(b *testing.B, numLinks, i int) *cs.Link {
	return cstesting.RandomLink()
}

// RandomLinkMapID is a CreateLinkFunc that creates a random link with map ID.
// The map ID will be one of ten possible values.
func RandomLinkMapID(b *testing.B, numLinks, i int) *cs.Link {
	l := cstesting.RandomLink()
	l.Meta.MapID = fmt.Sprintf("%d", i%10)
	return l
}

// RandomLinkPrevLinkHash is a CreateLinkFunc that creates a random link with
// previous link hash.
// The previous link hash will be one of ten possible values.
func RandomLinkPrevLinkHash(b *testing.B, numLinks, i int) *cs.Link {
	l := cstesting.RandomLink()
	l.Meta.PrevLinkHash = fmt.Sprintf("000000000000000000000000000000000000000000000000000000000000000%d", i%10)
	return l
}

// RandomLinkTags is a CreateLinkFunc that creates a random link with tags.
// The tags will contain one of ten possible values.
func RandomLinkTags(b *testing.B, numLinks, i int) *cs.Link {
	l := cstesting.RandomLink()
	l.Meta.Tags = []string{fmt.Sprintf("%d", i%10)}
	return l
}

// RandomLinkMapIDTags is a CreateLinkFunc that creates a random link with map
// ID and tags.
// The map ID will be one of ten possible values.
// The tags will contain one of ten possible values.
func RandomLinkMapIDTags(b *testing.B, numLinks, i int) *cs.Link {
	l := cstesting.RandomLink()
	l.Meta.MapID = fmt.Sprintf("%d", i%10)
	l.Meta.Tags = []string{fmt.Sprintf("%d", i%10)}
	return l
}

// RandomLinkPrevLinkHashTags is a CreateLinkFunc that creates a random link
// with previous link hash and tags.
// The previous link hash will be one of ten possible values.
// The tags will contain one of ten possible values.
func RandomLinkPrevLinkHashTags(b *testing.B, numLinks, i int) *cs.Link {
	l := cstesting.RandomLink()
	l.Meta.PrevLinkHash = fmt.Sprintf("000000000000000000000000000000000000000000000000000000000000000%d", i%10)
	l.Meta.Tags = []string{fmt.Sprintf("%d", i%10)}
	return l
}

// MapFilterFunc is a type for a function that creates a mapId filter for
// benchmarks.
type MapFilterFunc func(b *testing.B, numLinks, i int) *store.MapFilter

// RandomPaginationOffset is a a PaginationFunc that create a pagination with a random offset.
func RandomPaginationOffset(b *testing.B, numLinks, i int) *store.MapFilter {
	return &store.MapFilter{
		Pagination: store.Pagination{
			Offset: rand.Int() % numLinks,
			Limit:  store.DefaultLimit,
		},
	}
}

// FilterFunc is a type for a function that creates a filter for benchmarks.
type FilterFunc func(b *testing.B, numLinks, i int) *store.SegmentFilter

// RandomFilterOffset is a a FilterFunc that create a filter with a random
// offset.
func RandomFilterOffset(b *testing.B, numLinks, i int) *store.SegmentFilter {
	return &store.SegmentFilter{
		Pagination: store.Pagination{
			Offset: rand.Int() % numLinks,
			Limit:  store.DefaultLimit,
		},
	}
}

// RandomFilterOffsetMapID is a a FilterFunc that create a filter with a random
// offset and map ID.
// The map ID will be one of ten possible values.
func RandomFilterOffsetMapID(b *testing.B, numLinks, i int) *store.SegmentFilter {
	return &store.SegmentFilter{
		Pagination: store.Pagination{
			Offset: rand.Int() % numLinks,
			Limit:  store.DefaultLimit,
		},
		MapIDs: []string{fmt.Sprintf("%d", i%10)},
	}
}

// RandomFilterOffsetMapIDs is a a FilterFunc that create a filter with a random
// offset and 2 map IDs.
// The map ID will be one of ten possible values.
func RandomFilterOffsetMapIDs(b *testing.B, numLinks, i int) *store.SegmentFilter {
	return &store.SegmentFilter{
		Pagination: store.Pagination{
			Offset: rand.Int() % numLinks,
			Limit:  store.DefaultLimit,
		},
		MapIDs: []string{fmt.Sprintf("%d", i%10), fmt.Sprintf("%d", (i+1)%10)},
	}
}

// RandomFilterOffsetPrevLinkHash is a a FilterFunc that create a filter with a
// random offset and previous link hash.
// The previous link hash will be one of ten possible values.
func RandomFilterOffsetPrevLinkHash(b *testing.B, numLinks, i int) *store.SegmentFilter {
	prevLinkHash, _ := types.NewBytes32FromString(fmt.Sprintf("000000000000000000000000000000000000000000000000000000000000000%d", i%10))
	prevLinkHashStr := prevLinkHash.String()
	return &store.SegmentFilter{
		Pagination: store.Pagination{
			Offset: rand.Int() % numLinks,
			Limit:  store.DefaultLimit,
		},
		PrevLinkHash: &prevLinkHashStr,
	}
}

// RandomFilterOffsetTags is a a FilterFunc that create a filter with a random
// offset and map ID.
// The tags will be one of fifty possible combinations.
func RandomFilterOffsetTags(b *testing.B, numLinks, i int) *store.SegmentFilter {
	return &store.SegmentFilter{
		Pagination: store.Pagination{
			Offset: rand.Int() % numLinks,
			Limit:  store.DefaultLimit,
		},
		Tags: []string{fmt.Sprintf("%d", i%5), fmt.Sprintf("%d", i%10)},
	}
}

// RandomFilterOffsetMapIDTags is a a FilterFunc that create a filter with a
// random offset and map ID and tags.
// The map ID will be one of ten possible values.
// The tags will be one of fifty possible combinations.
func RandomFilterOffsetMapIDTags(b *testing.B, numLinks, i int) *store.SegmentFilter {
	return &store.SegmentFilter{
		Pagination: store.Pagination{
			Offset: rand.Int() % numLinks,
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
func RandomFilterOffsetPrevLinkHashTags(b *testing.B, numLinks, i int) *store.SegmentFilter {
	prevLinkHash, _ := types.NewBytes32FromString(fmt.Sprintf("000000000000000000000000000000000000000000000000000000000000000%d", i%10))
	prevLinkHashStr := prevLinkHash.String()
	return &store.SegmentFilter{
		Pagination: store.Pagination{
			Offset: rand.Int() % numLinks,
			Limit:  store.DefaultLimit,
		},
		PrevLinkHash: &prevLinkHashStr,
		Tags:         []string{fmt.Sprintf("%d", i%5), fmt.Sprintf("%d", i%10)},
	}
}
