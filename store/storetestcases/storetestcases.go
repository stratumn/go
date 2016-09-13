// Copyright 2016 Stratumn SAS. All rights reserved.
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

	"github.com/stratumn/go/store"
)

// Factory wraps functions to allocate and free an adapter,
// and is used to run the tests on an adapter.
type Factory struct {
	// New creates an adapter.
	New func() (store.Adapter, error)

	// Free is an optional function to free an adapter.
	Free func(adapter store.Adapter)
}

// RunTests runs all the tests.
func (f Factory) RunTests(t *testing.T) {
	t.Run("DeleteSegment", f.TestDeleteSegment)
	t.Run("DeleteSegmentNotFound", f.TestDeleteSegmentNotFound)
	t.Run("FindSegments", f.TestFindSegments)
	t.Run("FindSegmentsPagination", f.TestFindSegmentsPagination)
	t.Run("FindSegmentEmpty", f.TestFindSegmentEmpty)
	t.Run("FindSegmentsSingleTag", f.TestFindSegmentsSingleTag)
	t.Run("FindSegmentsMultipleTags", f.TestFindSegmentsMultipleTags)
	t.Run("FindSegmentsMapID", f.TestFindSegmentsMapID)
	t.Run("FindSegmentsMapIDTags", f.TestFindSegmentsMapIDTags)
	t.Run("FindSegmentsMapIDNotFound", f.TestFindSegmentsMapIDNotFound)
	t.Run("FindSegmentsPrevLinkHash", f.TestFindSegmentsPrevLinkHash)
	t.Run("FindSegmentsPrevLinkHashTags", f.TestFindSegmentsPrevLinkHashTags)
	t.Run("FindSegmentsPrevLinkHashMapID", f.TestFindSegmentsPrevLinkHashMapID)
	t.Run("FindSegmentsPrevLinkHashNotFound", f.TestFindSegmentsPrevLinkHashNotFound)
	t.Run("GetInfo", f.TestGetInfo)
	t.Run("GetMapIDs", f.TestGetMapIDs)
	t.Run("GetMapIDsPagination", f.TestGetMapIDsPagination)
	t.Run("GetMapIDs_empty", f.TestGetMapIDsEmpty)
	t.Run("GetSegment", f.TestGetSegment)
	t.Run("GetSegmentUpdatedState", f.TestGetSegmentUpdatedState)
	t.Run("GetSegmentUpdatedMapID", f.TestGetSegmentUpdatedMapID)
	t.Run("GetSegmentNotFound", f.TestGetSegmentNotFound)
	t.Run("SaveSegment", f.TestSaveSegment)
	t.Run("SaveSegmentUpdatedState", f.TestSaveSegmentUpdatedState)
	t.Run("SaveSegmentUpdatedMapID", f.TestSaveSegmentUpdatedMapID)
	t.Run("SaveSegmentBranch", f.TestSaveSegmentBranch)
}

// RunBenchmarks runs all the benchmarks.
func (f Factory) RunBenchmarks(b *testing.B) {
	b.Run("DeleteSegment", f.BenchmarkDeleteSegment)
	b.Run("DeleteSegmentParallel", f.BenchmarkDeleteSegmentParallel)
	b.Run("GetSegment", f.BenchmarkGetSegment)
	b.Run("GetSegmentParallel", f.BenchmarkGetSegmentParallel)
	b.Run("SaveSegment", f.BenchmarkSaveSegment)
	b.Run("SaveSegmentParallel", f.BenchmarkSaveSegmentParallel)
	b.Run("SaveSegmentUpdatedState", f.BenchmarkSaveSegmentUpdatedState)
	b.Run("SaveSegmentUpdatedStateParallel", f.BenchmarkSaveSegmentUpdatedStateParallel)
	b.Run("SaveSegmentUpdatedMapID", f.BenchmarkSaveSegmentUpdatedMapID)
	b.Run("SaveSegmentUpdatedMapIDParallel", f.BenchmarkSaveSegmentUpdatedMapIDParallel)
}

func (f Factory) free(adapter store.Adapter) {
	if f.Free != nil {
		f.Free(adapter)
	}
}
