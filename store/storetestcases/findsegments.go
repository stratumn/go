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
	"io/ioutil"
	"log"
	"math/rand"
	"sync/atomic"
	"testing"

	"github.com/stratumn/sdk/cs/cstesting"
	"github.com/stratumn/sdk/store"
	"github.com/stratumn/sdk/testutil"
)

// TestFindSegments tests what happens when you search with default pagination.
func (f Factory) TestFindSegments(t *testing.T) {
	a := f.initAdapter(t)
	defer f.free(a)

	for i := 0; i < store.DefaultLimit*2; i++ {
		a.SaveSegment(cstesting.RandomSegment())
	}

	slice, err := a.FindSegments(&store.Filter{
		Pagination: store.Pagination{
			Limit: store.DefaultLimit,
		},
	})
	if err != nil {
		t.Fatalf("a.FindSegments(): err: %s", err)
	}

	if got, want := len(slice), store.DefaultLimit; got != want {
		t.Errorf("len(slice) = %d want %d", got, want)
	}

	wantLTE := 100.0
	for _, s := range slice {
		got := s.Link.GetPriority()
		if got > wantLTE {
			t.Errorf("priority = %f want <= %f", got, wantLTE)
		}
		wantLTE = got
	}
}

// TestFindSegmentsPagination tests what happens when you search with
// pagination.
func (f Factory) TestFindSegmentsPagination(t *testing.T) {
	a := f.initAdapter(t)
	defer f.free(a)

	for i := 0; i < 100; i++ {
		a.SaveSegment(cstesting.RandomSegment())
	}

	limit := 10 + rand.Intn(10)
	slice, err := a.FindSegments(&store.Filter{
		Pagination: store.Pagination{
			Offset: rand.Intn(40),
			Limit:  limit,
		},
	})
	if err != nil {
		t.Fatalf("a.FindSegments(): err: %s", err)
	}

	if got, want := len(slice), limit; got != want {
		t.Errorf("len(slice) = %d want %d", got, want)
	}

	wantLTE := 100.0
	for _, s := range slice {
		got := s.Link.GetPriority()
		if got > wantLTE {
			t.Errorf("priority = %f want <= %f", got, wantLTE)
		}
		wantLTE = got
	}
}

// TestFindSegmentEmpty tests what happens when there are no matches.
func (f Factory) TestFindSegmentEmpty(t *testing.T) {
	a := f.initAdapter(t)
	defer f.free(a)

	for i := 0; i < 100; i++ {
		a.SaveSegment(cstesting.RandomSegment())
	}

	slice, err := a.FindSegments(&store.Filter{
		Tags: []string{"blablabla"},
	})
	if err != nil {
		t.Fatalf("a.FindSegments(): err: %s", err)
	}

	if got, want := len(slice), 0; got != want {
		t.Errorf("len(slice) = %d want %d", got, want)
	}
}

// TestFindSegmentsSingleTag tests what happens when you search with only one
// tag.
func (f Factory) TestFindSegmentsSingleTag(t *testing.T) {
	a := f.initAdapter(t)
	defer f.free(a)

	tag1 := testutil.RandomString(5)
	tag2 := testutil.RandomString(5)

	for i := 0; i < store.DefaultLimit; i++ {
		s := cstesting.RandomSegment()
		s.Link.Meta["tags"] = []interface{}{tag1, testutil.RandomString(5)}
		a.SaveSegment(s)
	}

	for i := 0; i < store.DefaultLimit; i++ {
		s := cstesting.RandomSegment()
		s.Link.Meta["tags"] = []interface{}{tag1, tag2, testutil.RandomString(5)}
		a.SaveSegment(s)
	}

	for i := 0; i < store.DefaultLimit; i++ {
		s := cstesting.RandomSegment()
		a.SaveSegment(s)
	}

	slice, err := a.FindSegments(&store.Filter{
		Pagination: store.Pagination{
			Limit: store.DefaultLimit * 3,
		},
		Tags: []string{tag1},
	})
	if err != nil {
		t.Fatalf("a.FindSegments(): err: %s", err)
	}

	if got, want := len(slice), store.DefaultLimit*2; got != want {
		t.Errorf("len(slice) = %d want %d", got, want)
	}
}

// TestFindSegmentsMultipleTags tests what happens when you search with more
// than one tag.
func (f Factory) TestFindSegmentsMultipleTags(t *testing.T) {
	a := f.initAdapter(t)
	defer f.free(a)

	tag1 := testutil.RandomString(5)
	tag2 := testutil.RandomString(5)

	for i := 0; i < store.DefaultLimit; i++ {
		s := cstesting.RandomSegment()
		s.Link.Meta["tags"] = []interface{}{tag1, testutil.RandomString(5)}
		a.SaveSegment(s)
	}

	for i := 0; i < store.DefaultLimit; i++ {
		s := cstesting.RandomSegment()
		s.Link.Meta["tags"] = []interface{}{tag1, tag2, testutil.RandomString(5)}
		a.SaveSegment(s)
	}

	for i := 0; i < store.DefaultLimit; i++ {
		s := cstesting.RandomSegment()
		a.SaveSegment(s)
	}

	slice, err := a.FindSegments(&store.Filter{
		Pagination: store.Pagination{
			Limit: store.DefaultLimit * 3,
		},
		Tags: []string{tag2, tag1},
	})
	if err != nil {
		t.Fatalf("a.FindSegments(): err: %s", err)
	}

	if got, want := len(slice), store.DefaultLimit; got != want {
		t.Errorf("len(slice) = %d want %d", got, want)
	}
}

// TestFindSegmentsMapID tests whan happens when you search for an existing map
// ID.
func (f Factory) TestFindSegmentsMapID(t *testing.T) {
	a := f.initAdapter(t)
	defer f.free(a)

	for i := 0; i < 2; i++ {
		for j := 0; j < store.DefaultLimit; j++ {
			s := cstesting.RandomSegment()
			s.Link.Meta["mapId"] = fmt.Sprintf("map%d", i)
			a.SaveSegment(s)
		}
	}

	slice, err := a.FindSegments(&store.Filter{
		Pagination: store.Pagination{
			Limit: store.DefaultLimit * 2,
		},
		MapID: "map1",
	})
	if err != nil {
		t.Fatalf("a.FindSegments(): err: %s", err)
	}

	if got := slice; got == nil {
		t.Fatal("slice = nit want cs.SegmentSlice")
	}
	if got, want := len(slice), store.DefaultLimit; got != want {
		t.Errorf("len(slice) = %d want %d", got, want)
	}
}

// TestFindSegmentsMapIDTags tests whan happens when you search for an existing
// map ID and tags.
func (f Factory) TestFindSegmentsMapIDTags(t *testing.T) {
	a := f.initAdapter(t)
	defer f.free(a)

	tag1 := testutil.RandomString(5)

	for j := 0; j < store.DefaultLimit; j++ {
		s := cstesting.RandomSegment()
		s.Link.Meta["mapId"] = "map1"
		s.Link.Meta["tags"] = []interface{}{tag1, testutil.RandomString(5)}
		a.SaveSegment(s)
	}

	for j := 0; j < store.DefaultLimit; j++ {
		s := cstesting.RandomSegment()
		s.Link.Meta["mapId"] = "map1"
		s.Link.Meta["tags"] = []interface{}{testutil.RandomString(5)}
		a.SaveSegment(s)
	}

	for j := 0; j < store.DefaultLimit; j++ {
		s := cstesting.RandomSegment()
		s.Link.Meta["mapId"] = "map2"
		s.Link.Meta["tags"] = []interface{}{tag1, testutil.RandomString(5)}
		a.SaveSegment(s)
	}

	slice, err := a.FindSegments(&store.Filter{
		Pagination: store.Pagination{
			Limit: store.DefaultLimit * 3,
		},
		MapID: "map1",
		Tags:  []string{tag1},
	})
	if err != nil {
		t.Fatalf("a.FindSegments(): err: %s", err)
	}

	if got := slice; got == nil {
		t.Fatal("slice = nit want cs.SegmentSlice")
	}
	if got, want := len(slice), store.DefaultLimit; got != want {
		t.Errorf("len(slice) = %d want %d", got, want)
	}
}

// TestFindSegmentsMapIDNotFound tests whan happens when you search for a
// nonexistent map ID.
func (f Factory) TestFindSegmentsMapIDNotFound(t *testing.T) {
	a := f.initAdapter(t)
	defer f.free(a)

	slice, err := a.FindSegments(&store.Filter{
		Pagination: store.Pagination{
			Limit: store.DefaultLimit,
		},
		MapID: testutil.RandomString(10),
	})
	if err != nil {
		t.Fatalf("a.FindSegments(): err: %s", err)
	}

	if got, want := len(slice), 0; got != want {
		t.Errorf("len(slice) = %d want %d", got, want)
	}
}

// TestFindSegmentsPrevLinkHash tests whan happens when you search for an
// existing previous link hash.
func (f Factory) TestFindSegmentsPrevLinkHash(t *testing.T) {
	a := f.initAdapter(t)
	defer f.free(a)

	s := cstesting.RandomSegment()
	a.SaveSegment(s)

	for i := 0; i < store.DefaultLimit; i++ {
		a.SaveSegment(cstesting.RandomBranch(s))
	}

	slice, err := a.FindSegments(&store.Filter{
		Pagination: store.Pagination{
			Limit: store.DefaultLimit * 2,
		},
		PrevLinkHash: s.GetLinkHash(),
	})
	if err != nil {
		t.Fatalf("a.FindSegments(): err: %s", err)
	}

	if got := slice; got == nil {
		t.Fatal("slice = nit want cs.SegmentSlice")
	}
	if got, want := len(slice), store.DefaultLimit; got != want {
		t.Errorf("len(slice) = %d want %d", got, want)
	}
}

// TestFindSegmentsPrevLinkHashTags tests whan happens when you search for a
// previous link hash and tags.
func (f Factory) TestFindSegmentsPrevLinkHashTags(t *testing.T) {
	a := f.initAdapter(t)
	defer f.free(a)

	s1 := cstesting.RandomSegment()
	a.SaveSegment(s1)
	tag1 := testutil.RandomString(5)

	for j := 0; j < store.DefaultLimit; j++ {
		s2 := cstesting.RandomBranch(s1)
		s2.Link.Meta["tags"] = []interface{}{tag1, testutil.RandomString(5)}
		a.SaveSegment(s2)
	}

	for j := 0; j < store.DefaultLimit; j++ {
		s2 := cstesting.RandomBranch(s1)
		s2.Link.Meta["tags"] = []interface{}{testutil.RandomString(5)}
		a.SaveSegment(s2)
	}

	for j := 0; j < store.DefaultLimit; j++ {
		s2 := cstesting.RandomSegment()
		s2.Link.Meta["tags"] = []interface{}{tag1, testutil.RandomString(5)}
		a.SaveSegment(s2)
	}

	slice, err := a.FindSegments(&store.Filter{
		Pagination: store.Pagination{
			Limit: store.DefaultLimit * 3,
		},
		PrevLinkHash: s1.GetLinkHash(),
		Tags:         []string{tag1},
	})
	if err != nil {
		t.Fatalf("a.FindSegments(): err: %s", err)
	}

	if got := slice; got == nil {
		t.Fatal("slice = nit want cs.SegmentSlice")
	}
	if got, want := len(slice), store.DefaultLimit; got != want {
		t.Errorf("len(slice) = %d want %d", got, want)
	}
}

// TestFindSegmentsPrevLinkHashMapID tests that the map ID is ignored if a
// previous link hash is given.
func (f Factory) TestFindSegmentsPrevLinkHashMapID(t *testing.T) {
	a := f.initAdapter(t)
	defer f.free(a)

	s1 := cstesting.RandomSegment()
	a.SaveSegment(s1)

	for j := 0; j < store.DefaultLimit; j++ {
		s2 := cstesting.RandomBranch(s1)
		a.SaveSegment(s2)
	}

	for j := 0; j < store.DefaultLimit; j++ {
		s2 := cstesting.RandomSegment()
		a.SaveSegment(s2)
	}

	slice, err := a.FindSegments(&store.Filter{
		Pagination: store.Pagination{
			Limit: store.DefaultLimit * 2,
		},
		PrevLinkHash: s1.GetLinkHash(),
		MapID:        "test",
	})
	if err != nil {
		t.Fatalf("a.FindSegments(): err: %s", err)
	}

	if got := slice; got == nil {
		t.Fatal("slice = nit want cs.SegmentSlice")
	}
	if got, want := len(slice), store.DefaultLimit; got != want {
		t.Errorf("len(slice) = %d want %d", got, want)
	}
}

// TestFindSegmentsPrevLinkHashNotFound tests whan happens when you search for a
// nonexistent previous link hash.
func (f Factory) TestFindSegmentsPrevLinkHashNotFound(t *testing.T) {
	a := f.initAdapter(t)
	defer f.free(a)

	slice, err := a.FindSegments(&store.Filter{
		Pagination: store.Pagination{
			Limit: store.DefaultLimit,
		},
		PrevLinkHash: testutil.RandomHash(),
	})
	if err != nil {
		t.Fatalf("a.FindSegments(): err: %s", err)
	}

	if got, want := len(slice), 0; got != want {
		t.Errorf("len(slice) = %d want %d", got, want)
	}
}

// BenchmarkFindSegments benchmarks finding segments.
func (f Factory) BenchmarkFindSegments(b *testing.B, numSegments int, segmentFunc SegmentFunc, filterFunc FilterFunc) {
	a := f.initAdapterB(b)
	defer f.free(a)

	for i := 0; i < numSegments; i++ {
		a.SaveSegment(segmentFunc(b, numSegments, i))
	}

	filters := make([]*store.Filter, b.N)
	for i := 0; i < b.N; i++ {
		filters[i] = filterFunc(b, numSegments, i)
	}

	b.ResetTimer()
	log.SetOutput(ioutil.Discard)

	for i := 0; i < b.N; i++ {
		if s, err := a.FindSegments(filters[i]); err != nil {
			b.Fatal(err)
		} else if s == nil {
			b.Error("s = nil want cs.SegmentSlice")
		}
	}
}

// BenchmarkFindSegments100 benchmarks finding segments within 100 segments.
func (f Factory) BenchmarkFindSegments100(b *testing.B) {
	f.BenchmarkFindSegments(b, 100, RandomSegment, RandomFilterOffset)
}

// BenchmarkFindSegments1000 benchmarks finding segments within 1000 segments.
func (f Factory) BenchmarkFindSegments1000(b *testing.B) {
	f.BenchmarkFindSegments(b, 1000, RandomSegment, RandomFilterOffset)
}

// BenchmarkFindSegments10000 benchmarks finding segments within 10000 segments.
func (f Factory) BenchmarkFindSegments10000(b *testing.B) {
	f.BenchmarkFindSegments(b, 10000, RandomSegment, RandomFilterOffset)
}

// BenchmarkFindSegmentsMapID100 benchmarks finding segments with a map ID
// within 100 segments.
func (f Factory) BenchmarkFindSegmentsMapID100(b *testing.B) {
	f.BenchmarkFindSegments(b, 100, RandomSegmentMapID, RandomFilterOffsetMapID)
}

// BenchmarkFindSegmentsMapID1000 benchmarks finding segments with a map ID
// within 1000 segments.
func (f Factory) BenchmarkFindSegmentsMapID1000(b *testing.B) {
	f.BenchmarkFindSegments(b, 1000, RandomSegmentMapID, RandomFilterOffsetMapID)
}

// BenchmarkFindSegmentsMapID10000 benchmarks finding segments with a map ID
// within 10000 segments.
func (f Factory) BenchmarkFindSegmentsMapID10000(b *testing.B) {
	f.BenchmarkFindSegments(b, 10000, RandomSegmentMapID, RandomFilterOffsetMapID)
}

// BenchmarkFindSegmentsPrevLinkHash100 benchmarks finding segments with
// previous link hash within 100 segments.
func (f Factory) BenchmarkFindSegmentsPrevLinkHash100(b *testing.B) {
	f.BenchmarkFindSegments(b, 100, RandomSegmentPrevLinkHash, RandomFilterOffsetPrevLinkHash)
}

// BenchmarkFindSegmentsPrevLinkHash1000 benchmarks finding segments with
// previous link hash within 1000 segments.
func (f Factory) BenchmarkFindSegmentsPrevLinkHash1000(b *testing.B) {
	f.BenchmarkFindSegments(b, 1000, RandomSegmentPrevLinkHash, RandomFilterOffsetPrevLinkHash)
}

// BenchmarkFindSegmentsPrevLinkHash10000 benchmarks finding segments with
// previous link hash within 10000 segments.
func (f Factory) BenchmarkFindSegmentsPrevLinkHash10000(b *testing.B) {
	f.BenchmarkFindSegments(b, 10000, RandomSegmentPrevLinkHash, RandomFilterOffsetPrevLinkHash)
}

// BenchmarkFindSegmentsTags100 benchmarks finding segments with tags within 100
// segments.
func (f Factory) BenchmarkFindSegmentsTags100(b *testing.B) {
	f.BenchmarkFindSegments(b, 100, RandomSegmentTags, RandomFilterOffsetTags)
}

// BenchmarkFindSegmentsTags1000 benchmarks finding segments with tags within
// 1000 segments.
func (f Factory) BenchmarkFindSegmentsTags1000(b *testing.B) {
	f.BenchmarkFindSegments(b, 1000, RandomSegmentTags, RandomFilterOffsetTags)
}

// BenchmarkFindSegmentsTags10000 benchmarks finding segments with tags within
// 10000 segments.
func (f Factory) BenchmarkFindSegmentsTags10000(b *testing.B) {
	f.BenchmarkFindSegments(b, 10000, RandomSegmentTags, RandomFilterOffsetTags)
}

// BenchmarkFindSegmentsMapIDTags100 benchmarks finding segments with map ID and
// tags within 100 segments.
func (f Factory) BenchmarkFindSegmentsMapIDTags100(b *testing.B) {
	f.BenchmarkFindSegments(b, 100, RandomSegmentMapIDTags, RandomFilterOffsetMapIDTags)
}

// BenchmarkFindSegmentsMapIDTags1000 benchmarks finding segments with map ID
// and tags within 1000 segments.
func (f Factory) BenchmarkFindSegmentsMapIDTags1000(b *testing.B) {
	f.BenchmarkFindSegments(b, 1000, RandomSegmentMapIDTags, RandomFilterOffsetMapIDTags)
}

// BenchmarkFindSegmentsMapIDTags10000 benchmarks finding segments with map ID
// and tags within 10000 segments.
func (f Factory) BenchmarkFindSegmentsMapIDTags10000(b *testing.B) {
	f.BenchmarkFindSegments(b, 10000, RandomSegmentMapIDTags, RandomFilterOffsetMapIDTags)
}

// BenchmarkFindSegmentsPrevLinkHashTags100 benchmarks finding segments with
// previous link hash and tags within 100 segments.
func (f Factory) BenchmarkFindSegmentsPrevLinkHashTags100(b *testing.B) {
	f.BenchmarkFindSegments(b, 100, RandomSegmentPrevLinkHashTags, RandomFilterOffsetPrevLinkHashTags)
}

// BenchmarkFindSegmentsPrevLinkHashTags1000 benchmarks finding segments with
// previous link hash and tags within 1000 segments.
func (f Factory) BenchmarkFindSegmentsPrevLinkHashTags1000(b *testing.B) {
	f.BenchmarkFindSegments(b, 1000, RandomSegmentPrevLinkHashTags, RandomFilterOffsetPrevLinkHashTags)
}

// BenchmarkFindSegmentsPrevLinkHashTags10000 benchmarks finding segments with
// previous link hash and tags within 10000 segments.
func (f Factory) BenchmarkFindSegmentsPrevLinkHashTags10000(b *testing.B) {
	f.BenchmarkFindSegments(b, 10000, RandomSegmentPrevLinkHashTags, RandomFilterOffsetPrevLinkHashTags)
}

// BenchmarkFindSegmentsParallel benchmarks finding segments.
func (f Factory) BenchmarkFindSegmentsParallel(b *testing.B, numSegments int, segmentFunc SegmentFunc, filterFunc FilterFunc) {
	a := f.initAdapterB(b)
	defer f.free(a)

	for i := 0; i < numSegments; i++ {
		a.SaveSegment(segmentFunc(b, numSegments, i))
	}

	filters := make([]*store.Filter, b.N)
	for i := 0; i < b.N; i++ {
		filters[i] = filterFunc(b, numSegments, i)
	}

	var counter uint64

	b.ResetTimer()
	log.SetOutput(ioutil.Discard)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			i := int(atomic.AddUint64(&counter, 1) - 1)
			if s, err := a.FindSegments(filters[i]); err != nil {
				b.Error(err)
			} else if s == nil {
				b.Error("s = nil want cs.SegmentSlice")
			}
		}
	})
}

// BenchmarkFindSegments100Parallel benchmarks finding segments within 100
// segments.
func (f Factory) BenchmarkFindSegments100Parallel(b *testing.B) {
	f.BenchmarkFindSegmentsParallel(b, 100, RandomSegment, RandomFilterOffset)
}

// BenchmarkFindSegments1000Parallel benchmarks finding segments within 1000
// segments.
func (f Factory) BenchmarkFindSegments1000Parallel(b *testing.B) {
	f.BenchmarkFindSegmentsParallel(b, 1000, RandomSegment, RandomFilterOffset)
}

// BenchmarkFindSegments10000Parallel benchmarks finding segments within 10000
// segments.
func (f Factory) BenchmarkFindSegments10000Parallel(b *testing.B) {
	f.BenchmarkFindSegmentsParallel(b, 10000, RandomSegment, RandomFilterOffset)
}

// BenchmarkFindSegmentsMapID100Parallel benchmarks finding segments with a map
// ID within 100 segments.
func (f Factory) BenchmarkFindSegmentsMapID100Parallel(b *testing.B) {
	f.BenchmarkFindSegmentsParallel(b, 100, RandomSegmentMapID, RandomFilterOffsetMapID)
}

// BenchmarkFindSegmentsMapID1000Parallel benchmarks finding segments with a map
// ID within 1000 segments.
func (f Factory) BenchmarkFindSegmentsMapID1000Parallel(b *testing.B) {
	f.BenchmarkFindSegmentsParallel(b, 1000, RandomSegmentMapID, RandomFilterOffsetMapID)
}

// BenchmarkFindSegmentsMapID10000Parallel benchmarks finding segments with a
// map ID within 10000 segments.
func (f Factory) BenchmarkFindSegmentsMapID10000Parallel(b *testing.B) {
	f.BenchmarkFindSegmentsParallel(b, 10000, RandomSegmentMapID, RandomFilterOffsetMapID)
}

// BenchmarkFindSegmentsPrevLinkHash100Parallel benchmarks finding segments with
// a previous link hash within 100 segments.
func (f Factory) BenchmarkFindSegmentsPrevLinkHash100Parallel(b *testing.B) {
	f.BenchmarkFindSegmentsParallel(b, 100, RandomSegmentPrevLinkHash, RandomFilterOffsetPrevLinkHash)
}

// BenchmarkFindSegmentsPrevLinkHash1000Parallel benchmarks finding segments
// with a previous link hash within 1000 segments.
func (f Factory) BenchmarkFindSegmentsPrevLinkHash1000Parallel(b *testing.B) {
	f.BenchmarkFindSegmentsParallel(b, 1000, RandomSegmentPrevLinkHash, RandomFilterOffsetPrevLinkHash)
}

// BenchmarkFindSegmentsPrevLinkHash10000Parallel benchmarks finding segments
// with a previous link hash within 10000 segments.
func (f Factory) BenchmarkFindSegmentsPrevLinkHash10000Parallel(b *testing.B) {
	f.BenchmarkFindSegmentsParallel(b, 10000, RandomSegmentPrevLinkHash, RandomFilterOffsetPrevLinkHash)
}

// BenchmarkFindSegmentsTags100Parallel benchmarks finding segments with tags
// within 100 segments.
func (f Factory) BenchmarkFindSegmentsTags100Parallel(b *testing.B) {
	f.BenchmarkFindSegmentsParallel(b, 100, RandomSegmentTags, RandomFilterOffsetTags)
}

// BenchmarkFindSegmentsTags1000Parallel benchmarks finding segments with tags
// within 1000 segments.
func (f Factory) BenchmarkFindSegmentsTags1000Parallel(b *testing.B) {
	f.BenchmarkFindSegmentsParallel(b, 1000, RandomSegmentTags, RandomFilterOffsetTags)
}

// BenchmarkFindSegmentsTags10000Parallel benchmarks finding segments with tags
// within 10000 segments.
func (f Factory) BenchmarkFindSegmentsTags10000Parallel(b *testing.B) {
	f.BenchmarkFindSegmentsParallel(b, 10000, RandomSegmentTags, RandomFilterOffsetTags)
}

// BenchmarkFindSegmentsMapIDTags100Parallel benchmarks finding segments with
// map ID and tags within 100 segments.
func (f Factory) BenchmarkFindSegmentsMapIDTags100Parallel(b *testing.B) {
	f.BenchmarkFindSegmentsParallel(b, 100, RandomSegmentMapIDTags, RandomFilterOffsetMapIDTags)
}

// BenchmarkFindSegmentsMapIDTags1000Parallel benchmarks finding segments with
// map ID and tags within 1000 segments.
func (f Factory) BenchmarkFindSegmentsMapIDTags1000Parallel(b *testing.B) {
	f.BenchmarkFindSegmentsParallel(b, 1000, RandomSegmentMapIDTags, RandomFilterOffsetMapIDTags)
}

// BenchmarkFindSegmentsMapIDTags10000Parallel benchmarks finding segments with
// map ID and tags within 10000 segments.
func (f Factory) BenchmarkFindSegmentsMapIDTags10000Parallel(b *testing.B) {
	f.BenchmarkFindSegmentsParallel(b, 10000, RandomSegmentMapIDTags, RandomFilterOffsetMapIDTags)
}

// BenchmarkFindSegmentsPrevLinkHashTags100Parallel benchmarks finding segments
// with map ID and tags within 100 segments.
func (f Factory) BenchmarkFindSegmentsPrevLinkHashTags100Parallel(b *testing.B) {
	f.BenchmarkFindSegmentsParallel(b, 100, RandomSegmentPrevLinkHashTags, RandomFilterOffsetPrevLinkHashTags)
}

// BenchmarkFindSegmentsPrevLinkHashTags1000Parallel benchmarks finding segments
// with map ID and tags within 1000 segments.
func (f Factory) BenchmarkFindSegmentsPrevLinkHashTags1000Parallel(b *testing.B) {
	f.BenchmarkFindSegmentsParallel(b, 1000, RandomSegmentPrevLinkHashTags, RandomFilterOffsetPrevLinkHashTags)
}

// BenchmarkFindSegmentsPrevLinkHashTags10000Parallel benchmarks finding
// segments with map ID and tags within 10000 segments.
func (f Factory) BenchmarkFindSegmentsPrevLinkHashTags10000Parallel(b *testing.B) {
	f.BenchmarkFindSegmentsParallel(b, 10000, RandomSegmentPrevLinkHashTags, RandomFilterOffsetPrevLinkHashTags)
}
