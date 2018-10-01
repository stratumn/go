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
	"testing"

	"github.com/stratumn/go-core/store"
	"github.com/stretchr/testify/require"
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
	t.Run("Test adapter config", f.TestAdapterConfig)
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
	require.NoError(t, err, "f.New()")
	require.NotNil(t, a, "Store.Adapter")
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
	require.NoError(t, err, "f.NewKeyValueStore()")
	require.NotNil(t, a, "Store.KeyValueStore")
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
