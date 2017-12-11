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

	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/cs/cstesting"
	"github.com/stratumn/sdk/store"
	"github.com/stratumn/sdk/testutil"
	"github.com/stratumn/sdk/types"
)

var emptyPrevLinkHash = ""

func createLink(adapter *store.Adapter, link *cs.Link, prepareLink func(l *cs.Link)) *cs.Link {
	if prepareLink != nil {
		prepareLink(link)
	}
	(*adapter).CreateLink(link)
	return link
}

func createRandomLink(adapter *store.Adapter, prepareLink func(l *cs.Link)) *cs.Link {
	return createLink(adapter, cstesting.RandomLink(), prepareLink)
}

func createLinkBranch(adapter *store.Adapter, parent *cs.Link, prepareLink func(l *cs.Link)) *cs.Link {
	return createLink(adapter, cstesting.RandomBranch(parent), prepareLink)
}

// TestFindSegments tests what happens when you search with default pagination.
func (f Factory) TestFindSegments(t *testing.T) {
	a := f.initAdapter(t)
	defer f.freeAdapter(a)

	for i := 0; i < store.DefaultLimit*2; i++ {
		createRandomLink(&a, nil)
	}

	slice, err := a.FindSegments(&store.SegmentFilter{
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
	defer f.freeAdapter(a)

	for i := 0; i < 100; i++ {
		createRandomLink(&a, nil)
	}

	limit := 10 + rand.Intn(10)
	slice, err := a.FindSegments(&store.SegmentFilter{
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
	defer f.freeAdapter(a)

	for i := 0; i < 100; i++ {
		createRandomLink(&a, nil)
	}

	slice, err := a.FindSegments(&store.SegmentFilter{
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
	defer f.freeAdapter(a)

	tag1 := testutil.RandomString(5)
	tag2 := testutil.RandomString(5)

	for i := 0; i < store.DefaultLimit; i++ {
		createRandomLink(&a, func(l *cs.Link) {
			l.Meta["tags"] = []interface{}{tag1, testutil.RandomString(5)}
		})
	}

	for i := 0; i < store.DefaultLimit; i++ {
		createRandomLink(&a, func(l *cs.Link) {
			l.Meta["tags"] = []interface{}{tag1, tag2, testutil.RandomString(5)}
		})
	}

	for i := 0; i < store.DefaultLimit; i++ {
		createRandomLink(&a, nil)
	}

	slice, err := a.FindSegments(&store.SegmentFilter{
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
	defer f.freeAdapter(a)

	tag1 := testutil.RandomString(5)
	tag2 := testutil.RandomString(5)

	for i := 0; i < store.DefaultLimit; i++ {
		createRandomLink(&a, func(l *cs.Link) {
			l.Meta["tags"] = []interface{}{tag1, testutil.RandomString(5)}
		})
	}

	for i := 0; i < store.DefaultLimit; i++ {
		createRandomLink(&a, func(l *cs.Link) {
			l.Meta["tags"] = []interface{}{tag1, tag2, testutil.RandomString(5)}
		})
	}

	for i := 0; i < store.DefaultLimit; i++ {
		createRandomLink(&a, nil)
	}

	slice, err := a.FindSegments(&store.SegmentFilter{
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
	defer f.freeAdapter(a)

	for i := 0; i < 2; i++ {
		for j := 0; j < store.DefaultLimit; j++ {
			createRandomLink(&a, func(l *cs.Link) {
				l.Meta["mapId"] = fmt.Sprintf("map%d", i)
			})
		}
	}

	slice, err := a.FindSegments(&store.SegmentFilter{
		Pagination: store.Pagination{
			Limit: store.DefaultLimit * 2,
		},
		MapIDs: []string{"map1"},
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

// TestFindSegmentsMapIDs tests whan happens when you search for several existing map
// IDs.
func (f Factory) TestFindSegmentsMapIDs(t *testing.T) {
	a := f.initAdapter(t)
	defer f.freeAdapter(a)

	for i := 0; i < 3; i++ {
		for j := 0; j < store.DefaultLimit; j++ {
			createRandomLink(&a, func(l *cs.Link) {
				l.Meta["mapId"] = fmt.Sprintf("map%d", i)
			})
		}
	}

	slice, err := a.FindSegments(&store.SegmentFilter{
		Pagination: store.Pagination{
			Limit: store.DefaultLimit * 3,
		},
		MapIDs: []string{"map1", "map2"},
	})
	if err != nil {
		t.Fatalf("a.FindSegments(): err: %s", err)
	}

	if got := slice; got == nil {
		t.Fatal("slice = nit want cs.SegmentSlice")
	}
	if got, want := len(slice), store.DefaultLimit*2; got != want {
		t.Errorf("len(slice) = %d want %d", got, want)
	}
}

// TestFindSegmentsMapIDTags tests whan happens when you search for an existing
// map ID and tags.
func (f Factory) TestFindSegmentsMapIDTags(t *testing.T) {
	a := f.initAdapter(t)
	defer f.freeAdapter(a)

	tag1 := testutil.RandomString(5)

	for j := 0; j < store.DefaultLimit; j++ {
		createRandomLink(&a, func(l *cs.Link) {
			l.Meta["mapId"] = "map1"
			l.Meta["tags"] = []interface{}{tag1, testutil.RandomString(5)}
		})
	}

	for j := 0; j < store.DefaultLimit; j++ {
		createRandomLink(&a, func(l *cs.Link) {
			l.Meta["mapId"] = "map1"
			l.Meta["tags"] = []interface{}{testutil.RandomString(5)}
		})
	}

	for j := 0; j < store.DefaultLimit; j++ {
		createRandomLink(&a, func(l *cs.Link) {
			l.Meta["mapId"] = "map2"
			l.Meta["tags"] = []interface{}{tag1, testutil.RandomString(5)}
		})
	}

	slice, err := a.FindSegments(&store.SegmentFilter{
		Pagination: store.Pagination{
			Limit: store.DefaultLimit * 3,
		},
		MapIDs: []string{"map1"},
		Tags:   []string{tag1},
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
	defer f.freeAdapter(a)

	slice, err := a.FindSegments(&store.SegmentFilter{
		Pagination: store.Pagination{
			Limit: store.DefaultLimit,
		},
		MapIDs: []string{testutil.RandomString(10)},
	})
	if err != nil {
		t.Fatalf("a.FindSegments(): err: %s", err)
	}

	if got, want := len(slice), 0; got != want {
		t.Errorf("len(slice) = %d want %d", got, want)
	}
}

// TestFindSegmentsLinkHashesMultiMatch tests searching for segments by a slice of
// linkHashes with multiple matches.
func (f Factory) TestFindSegmentsLinkHashesMultiMatch(t *testing.T) {
	a := f.initAdapter(t)
	defer f.freeAdapter(a)

	link1 := createRandomLink(&a, nil)
	link2 := createRandomLink(&a, nil)
	for j := 0; j < store.DefaultLimit; j++ {
		createRandomLink(&a, nil)
	}

	linkHash1, _ := link1.Hash()
	linkHash2, _ := link2.Hash()
	slice, err := a.FindSegments(&store.SegmentFilter{
		LinkHashes: []*types.Bytes32{
			linkHash1,
			testutil.RandomHash(),
			linkHash2,
		},
		Pagination: store.Pagination{
			Limit: store.DefaultLimit,
		},
	})
	if err != nil {
		t.Fatalf("a.FindSegments(): err: %s", err)
	}

	if got := slice; got == nil {
		t.Fatal("slice = nit want cs.SegmentSlice")
	}
	if got, want := len(slice), 2; got != want {
		t.Errorf("len(slice) = %d want %d", got, want)
	}
}

// TestFindSegmentsLinkHashesWithProcess tests matching a linkHash will fail
// if the provided process attribute does not match.
func (f Factory) TestFindSegmentsLinkHashesWithProcess(t *testing.T) {
	a := f.initAdapter(t)
	defer f.freeAdapter(a)

	link1 := createRandomLink(&a, nil)
	link2 := createRandomLink(&a, func(l *cs.Link) {
		l.Meta["process"] = "Baz"
	})
	for j := 0; j < store.DefaultLimit; j++ {
		createRandomLink(&a, nil)
	}

	linkHash1, _ := link1.Hash()
	linkHash2, _ := link2.Hash()
	slice, err := a.FindSegments(&store.SegmentFilter{
		LinkHashes: []*types.Bytes32{
			linkHash1,
			linkHash2,
		},
		Process: "Baz",
		Pagination: store.Pagination{
			Limit: store.DefaultLimit,
		},
	})
	if err != nil {
		t.Fatalf("a.FindSegments(): err: %s", err)
	}

	if got := slice; got == nil {
		t.Fatal("slice = nit want cs.SegmentSlice")
	}
	if got, want := len(slice), 1; got != want {
		t.Errorf("len(slice) = %d want %d", got, want)
	}
}

// TestFindSegmentsLinkHashesNoMatch tests searching for segments by a slice of
// linkHashes will return emtpy slice when there are no matches.
func (f Factory) TestFindSegmentsLinkHashesNoMatch(t *testing.T) {
	a := f.initAdapter(t)
	defer f.freeAdapter(a)

	for j := 0; j < store.DefaultLimit; j++ {
		createRandomLink(&a, nil)
	}

	slice, err := a.FindSegments(&store.SegmentFilter{
		LinkHashes: []*types.Bytes32{
			testutil.RandomHash(),
			testutil.RandomHash(),
		},
		Pagination: store.Pagination{
			Limit: store.DefaultLimit,
		},
	})
	if err != nil {
		t.Fatalf("a.FindSegments(): err: %s", err)
	}

	if got := slice; got == nil {
		t.Fatal("slice = nit want cs.SegmentSlice")
	}
	if got, want := len(slice), 0; got != want {
		t.Errorf("len(slice) = %d want %d", got, want)
	}
}

// TestFindSegmentsEmptyPrevLinkHash tests what happens when you search for an
// existing previous link hash.
func (f Factory) TestFindSegmentsEmptyPrevLinkHash(t *testing.T) {
	a := f.initAdapter(t)
	defer f.freeAdapter(a)

	l := createRandomLink(&a, func(l *cs.Link) {
		delete(l.Meta, "prevLinkHash")
	})

	for i := 0; i < store.DefaultLimit; i++ {
		createLinkBranch(&a, l, nil)
	}

	slice, err := a.FindSegments(&store.SegmentFilter{Pagination: store.Pagination{Limit: store.DefaultLimit}, PrevLinkHash: &emptyPrevLinkHash})
	if err != nil {
		t.Fatalf("a.FindSegments(): err: %s", err)
	}

	if got := slice; got == nil {
		t.Fatal("slice = nil want cs.SegmentSlice")
	}
	if got, want := len(slice), 1; got != want {
		t.Errorf("len(slice) = %d want %d", got, want)
	}
}

// TestFindSegmentsPrevLinkHash tests whan happens when you search for an
// existing previous link hash.
func (f Factory) TestFindSegmentsPrevLinkHash(t *testing.T) {
	a := f.initAdapter(t)
	defer f.freeAdapter(a)

	l := createRandomLink(&a, nil)

	for i := 0; i < store.DefaultLimit; i++ {
		createLinkBranch(&a, l, nil)
	}

	prevLinkHash, _ := l.HashString()
	slice, err := a.FindSegments(&store.SegmentFilter{
		Pagination: store.Pagination{
			Limit: store.DefaultLimit * 2,
		},
		PrevLinkHash: &prevLinkHash,
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
	defer f.freeAdapter(a)

	l1 := createRandomLink(&a, nil)
	tag1 := testutil.RandomString(5)

	for j := 0; j < store.DefaultLimit; j++ {
		createLinkBranch(&a, l1, func(l *cs.Link) {
			l.Meta["tags"] = []interface{}{tag1, testutil.RandomString(5)}
		})
	}

	for j := 0; j < store.DefaultLimit; j++ {
		createLinkBranch(&a, l1, func(l *cs.Link) {
			l.Meta["tags"] = []interface{}{testutil.RandomString(5)}
		})
	}

	for j := 0; j < store.DefaultLimit; j++ {
		createRandomLink(&a, func(l *cs.Link) {
			l.Meta["tags"] = []interface{}{tag1, testutil.RandomString(5)}
		})
	}

	prevLinkHash, _ := l1.HashString()
	slice, err := a.FindSegments(&store.SegmentFilter{
		Pagination: store.Pagination{
			Limit: store.DefaultLimit * 3,
		},
		PrevLinkHash: &prevLinkHash,
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

// TestFindSegmentsPrevLinkHashGoodMapID tests that map IDs match with
// segments found with the given previous link hash.
func (f Factory) TestFindSegmentsPrevLinkHashGoodMapID(t *testing.T) {
	a := f.initAdapter(t)
	defer f.freeAdapter(a)

	l1 := createRandomLink(&a, nil)
	var mapID1 = l1.Meta["mapId"].(string)
	l2 := createRandomLink(&a, nil)
	var mapID2 = l2.Meta["mapId"].(string)

	for j := 0; j < store.DefaultLimit; j++ {
		createLinkBranch(&a, l1, nil)
		createLinkBranch(&a, l2, nil)
		createRandomLink(&a, func(l *cs.Link) {
			l.Meta["mapId"] = mapID1
		})
		createRandomLink(&a, func(l *cs.Link) {
			l.Meta["mapId"] = mapID2
		})
	}

	prevLinkHash, _ := l1.HashString()
	slice, err := a.FindSegments(&store.SegmentFilter{
		Pagination: store.Pagination{
			Limit: store.DefaultLimit * 2,
		},
		PrevLinkHash: &prevLinkHash,
		MapIDs:       []string{mapID1, mapID2},
	})
	if err != nil {
		t.Fatalf("a.FindSegments(): err: %s", err)
	}

	if got := slice; got == nil {
		t.Fatal("slice = nil want cs.SegmentSlice")
	}
	if got, want := len(slice), store.DefaultLimit; got != want {
		t.Errorf("len(slice) = %d want %d", got, want)
	}
}

// TestFindSegmentsPrevLinkHashBadMapID tests that map IDs invalidate all
// segments found with the given previous link hash.
func (f Factory) TestFindSegmentsPrevLinkHashBadMapID(t *testing.T) {
	a := f.initAdapter(t)
	defer f.freeAdapter(a)

	l1 := createRandomLink(&a, nil)
	var mapID1 = l1.Meta["mapId"].(string)
	l2 := createRandomLink(&a, nil)
	var mapID2 = l2.Meta["mapId"].(string)

	for j := 0; j < store.DefaultLimit; j++ {
		createLinkBranch(&a, l1, nil)
		createLinkBranch(&a, l2, nil)
		createRandomLink(&a, func(l *cs.Link) {
			l.Meta["mapId"] = mapID1
		})
		createRandomLink(&a, func(l *cs.Link) {
			l.Meta["mapId"] = mapID2
		})
	}

	prevLinkHash, _ := l1.HashString()
	slice, err := a.FindSegments(&store.SegmentFilter{
		Pagination: store.Pagination{
			Limit: store.DefaultLimit * 2,
		},
		PrevLinkHash: &prevLinkHash,
		MapIDs:       []string{mapID2},
	})
	if err != nil {
		t.Fatalf("a.FindSegments(): err: %s", err)
	}

	if got := slice; got == nil {
		t.Fatal("slice = nil want cs.SegmentSlice")
	}
	if got, want := len(slice), 0; got != want {
		t.Errorf("len(slice) = %d want %d", got, want)
	}
}

// TestFindSegmentsPrevLinkHashNotFound tests whan happens when you search for a
// nonexistent previous link hash.
func (f Factory) TestFindSegmentsPrevLinkHashNotFound(t *testing.T) {
	a := f.initAdapter(t)
	defer f.freeAdapter(a)

	notFoundPrevLinkHash := testutil.RandomHash().String()
	slice, err := a.FindSegments(&store.SegmentFilter{
		Pagination: store.Pagination{
			Limit: store.DefaultLimit,
		},
		PrevLinkHash: &notFoundPrevLinkHash,
	})
	if err != nil {
		t.Fatalf("a.FindSegments(): err: %s", err)
	}

	if got, want := len(slice), 0; got != want {
		t.Errorf("len(slice) = %d want %d", got, want)
	}
}

// TestFindSegmentWithGoodProcess tests what happens when you search with a process name filter.
func (f Factory) TestFindSegmentWithGoodProcess(t *testing.T) {
	var processNames = [4]string{"Foo", "Bar", "Yin", "Yang"}

	a := f.initAdapter(t)
	defer f.freeAdapter(a)

	for i := 0; i < store.DefaultLimit; i++ {
		createRandomLink(&a, func(l *cs.Link) {
			l.Meta["process"] = processNames[i%len(processNames)]
		})
	}

	slice, err := a.FindSegments(&store.SegmentFilter{
		Pagination: store.Pagination{
			Limit: store.DefaultLimit,
		},
		Process: processNames[0],
	})
	if err != nil {
		t.Fatalf("a.FindSegments(): err: %s", err)
	}

	if got, want := len(slice), store.DefaultLimit/len(processNames); got != want {
		t.Errorf("len(slice) = %d want %d", got, want)
	}
}

// TestFindSegmentWithBadProcess tests what happens when you search with an unexisting process name.
func (f Factory) TestFindSegmentWithBadProcess(t *testing.T) {
	var processNames = [2]string{"Foo", "Bar"}

	a := f.initAdapter(t)
	defer f.freeAdapter(a)

	for i := 0; i < store.DefaultLimit*2; i++ {
		createRandomLink(&a, func(l *cs.Link) {
			l.Meta["process"] = processNames[i%2]
		})
	}

	slice, err := a.FindSegments(&store.SegmentFilter{
		Pagination: store.Pagination{
			Limit: store.DefaultLimit,
		},
		Process: "Baz",
	})
	if err != nil {
		t.Fatalf("a.FindSegments(): err: %s", err)
	}

	if got, want := len(slice), 0; got != want {
		t.Errorf("len(slice) = %d want %d", got, want)
	}
}

// BenchmarkFindSegments benchmarks finding segments.
func (f Factory) BenchmarkFindSegments(b *testing.B, numLinks int, createLinkFunc CreateLinkFunc, filterFunc FilterFunc) {
	a := f.initAdapterB(b)
	defer f.freeAdapter(a)

	for i := 0; i < numLinks; i++ {
		a.CreateLink(createLinkFunc(b, numLinks, i))
	}

	filters := make([]*store.SegmentFilter, b.N)
	for i := 0; i < b.N; i++ {
		filters[i] = filterFunc(b, numLinks, i)
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
	f.BenchmarkFindSegments(b, 100, RandomLink, RandomFilterOffset)
}

// BenchmarkFindSegments1000 benchmarks finding segments within 1000 segments.
func (f Factory) BenchmarkFindSegments1000(b *testing.B) {
	f.BenchmarkFindSegments(b, 1000, RandomLink, RandomFilterOffset)
}

// BenchmarkFindSegments10000 benchmarks finding segments within 10000 segments.
func (f Factory) BenchmarkFindSegments10000(b *testing.B) {
	f.BenchmarkFindSegments(b, 10000, RandomLink, RandomFilterOffset)
}

// BenchmarkFindSegmentsMapID100 benchmarks finding segments with a map ID
// within 100 segments.
func (f Factory) BenchmarkFindSegmentsMapID100(b *testing.B) {
	f.BenchmarkFindSegments(b, 100, RandomLinkMapID, RandomFilterOffsetMapID)
}

// BenchmarkFindSegmentsMapID1000 benchmarks finding segments with a map ID
// within 1000 segments.
func (f Factory) BenchmarkFindSegmentsMapID1000(b *testing.B) {
	f.BenchmarkFindSegments(b, 1000, RandomLinkMapID, RandomFilterOffsetMapID)
}

// BenchmarkFindSegmentsMapID10000 benchmarks finding segments with a map ID
// within 10000 segments.
func (f Factory) BenchmarkFindSegmentsMapID10000(b *testing.B) {
	f.BenchmarkFindSegments(b, 10000, RandomLinkMapID, RandomFilterOffsetMapID)
}

// BenchmarkFindSegmentsMapIDs100 benchmarks finding segments with several map IDs
// within 100 segments.
func (f Factory) BenchmarkFindSegmentsMapIDs100(b *testing.B) {
	f.BenchmarkFindSegments(b, 100, RandomLinkMapID, RandomFilterOffsetMapIDs)
}

// BenchmarkFindSegmentsMapIDs1000 benchmarks finding segments with several map IDs
// within 1000 segments.
func (f Factory) BenchmarkFindSegmentsMapIDs1000(b *testing.B) {
	f.BenchmarkFindSegments(b, 1000, RandomLinkMapID, RandomFilterOffsetMapIDs)
}

// BenchmarkFindSegmentsMapIDs10000 benchmarks finding segments with several map IDs
// within 10000 segments.
func (f Factory) BenchmarkFindSegmentsMapIDs10000(b *testing.B) {
	f.BenchmarkFindSegments(b, 10000, RandomLinkMapID, RandomFilterOffsetMapIDs)
}

// BenchmarkFindSegmentsPrevLinkHash100 benchmarks finding segments with
// previous link hash within 100 segments.
func (f Factory) BenchmarkFindSegmentsPrevLinkHash100(b *testing.B) {
	f.BenchmarkFindSegments(b, 100, RandomLinkPrevLinkHash, RandomFilterOffsetPrevLinkHash)
}

// BenchmarkFindSegmentsPrevLinkHash1000 benchmarks finding segments with
// previous link hash within 1000 segments.
func (f Factory) BenchmarkFindSegmentsPrevLinkHash1000(b *testing.B) {
	f.BenchmarkFindSegments(b, 1000, RandomLinkPrevLinkHash, RandomFilterOffsetPrevLinkHash)
}

// BenchmarkFindSegmentsPrevLinkHash10000 benchmarks finding segments with
// previous link hash within 10000 segments.
func (f Factory) BenchmarkFindSegmentsPrevLinkHash10000(b *testing.B) {
	f.BenchmarkFindSegments(b, 10000, RandomLinkPrevLinkHash, RandomFilterOffsetPrevLinkHash)
}

// BenchmarkFindSegmentsTags100 benchmarks finding segments with tags within 100
// segments.
func (f Factory) BenchmarkFindSegmentsTags100(b *testing.B) {
	f.BenchmarkFindSegments(b, 100, RandomLinkTags, RandomFilterOffsetTags)
}

// BenchmarkFindSegmentsTags1000 benchmarks finding segments with tags within
// 1000 segments.
func (f Factory) BenchmarkFindSegmentsTags1000(b *testing.B) {
	f.BenchmarkFindSegments(b, 1000, RandomLinkTags, RandomFilterOffsetTags)
}

// BenchmarkFindSegmentsTags10000 benchmarks finding segments with tags within
// 10000 segments.
func (f Factory) BenchmarkFindSegmentsTags10000(b *testing.B) {
	f.BenchmarkFindSegments(b, 10000, RandomLinkTags, RandomFilterOffsetTags)
}

// BenchmarkFindSegmentsMapIDTags100 benchmarks finding segments with map ID and
// tags within 100 segments.
func (f Factory) BenchmarkFindSegmentsMapIDTags100(b *testing.B) {
	f.BenchmarkFindSegments(b, 100, RandomLinkMapIDTags, RandomFilterOffsetMapIDTags)
}

// BenchmarkFindSegmentsMapIDTags1000 benchmarks finding segments with map ID
// and tags within 1000 segments.
func (f Factory) BenchmarkFindSegmentsMapIDTags1000(b *testing.B) {
	f.BenchmarkFindSegments(b, 1000, RandomLinkMapIDTags, RandomFilterOffsetMapIDTags)
}

// BenchmarkFindSegmentsMapIDTags10000 benchmarks finding segments with map ID
// and tags within 10000 segments.
func (f Factory) BenchmarkFindSegmentsMapIDTags10000(b *testing.B) {
	f.BenchmarkFindSegments(b, 10000, RandomLinkMapIDTags, RandomFilterOffsetMapIDTags)
}

// BenchmarkFindSegmentsPrevLinkHashTags100 benchmarks finding segments with
// previous link hash and tags within 100 segments.
func (f Factory) BenchmarkFindSegmentsPrevLinkHashTags100(b *testing.B) {
	f.BenchmarkFindSegments(b, 100, RandomLinkPrevLinkHashTags, RandomFilterOffsetPrevLinkHashTags)
}

// BenchmarkFindSegmentsPrevLinkHashTags1000 benchmarks finding segments with
// previous link hash and tags within 1000 segments.
func (f Factory) BenchmarkFindSegmentsPrevLinkHashTags1000(b *testing.B) {
	f.BenchmarkFindSegments(b, 1000, RandomLinkPrevLinkHashTags, RandomFilterOffsetPrevLinkHashTags)
}

// BenchmarkFindSegmentsPrevLinkHashTags10000 benchmarks finding segments with
// previous link hash and tags within 10000 segments.
func (f Factory) BenchmarkFindSegmentsPrevLinkHashTags10000(b *testing.B) {
	f.BenchmarkFindSegments(b, 10000, RandomLinkPrevLinkHashTags, RandomFilterOffsetPrevLinkHashTags)
}

// BenchmarkFindSegmentsParallel benchmarks finding segments.
func (f Factory) BenchmarkFindSegmentsParallel(b *testing.B, numLinks int, createLinkFunc CreateLinkFunc, filterFunc FilterFunc) {
	a := f.initAdapterB(b)
	defer f.freeAdapter(a)

	for i := 0; i < numLinks; i++ {
		a.CreateLink(createLinkFunc(b, numLinks, i))
	}

	filters := make([]*store.SegmentFilter, b.N)
	for i := 0; i < b.N; i++ {
		filters[i] = filterFunc(b, numLinks, i)
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
	f.BenchmarkFindSegmentsParallel(b, 100, RandomLink, RandomFilterOffset)
}

// BenchmarkFindSegments1000Parallel benchmarks finding segments within 1000
// segments.
func (f Factory) BenchmarkFindSegments1000Parallel(b *testing.B) {
	f.BenchmarkFindSegmentsParallel(b, 1000, RandomLink, RandomFilterOffset)
}

// BenchmarkFindSegments10000Parallel benchmarks finding segments within 10000
// segments.
func (f Factory) BenchmarkFindSegments10000Parallel(b *testing.B) {
	f.BenchmarkFindSegmentsParallel(b, 10000, RandomLink, RandomFilterOffset)
}

// BenchmarkFindSegmentsMapID100Parallel benchmarks finding segments with a map
// ID within 100 segments.
func (f Factory) BenchmarkFindSegmentsMapID100Parallel(b *testing.B) {
	f.BenchmarkFindSegmentsParallel(b, 100, RandomLinkMapID, RandomFilterOffsetMapID)
}

// BenchmarkFindSegmentsMapID1000Parallel benchmarks finding segments with a map
// ID within 1000 segments.
func (f Factory) BenchmarkFindSegmentsMapID1000Parallel(b *testing.B) {
	f.BenchmarkFindSegmentsParallel(b, 1000, RandomLinkMapID, RandomFilterOffsetMapID)
}

// BenchmarkFindSegmentsMapID10000Parallel benchmarks finding segments with a
// map ID within 10000 segments.
func (f Factory) BenchmarkFindSegmentsMapID10000Parallel(b *testing.B) {
	f.BenchmarkFindSegmentsParallel(b, 10000, RandomLinkMapID, RandomFilterOffsetMapID)
}

// BenchmarkFindSegmentsMapIDs100Parallel benchmarks finding segments with several map
// ID within 100 segments.
func (f Factory) BenchmarkFindSegmentsMapIDs100Parallel(b *testing.B) {
	f.BenchmarkFindSegmentsParallel(b, 100, RandomLinkMapID, RandomFilterOffsetMapIDs)
}

// BenchmarkFindSegmentsMapIDs1000Parallel benchmarks finding segments with several map
// ID within 1000 segments.
func (f Factory) BenchmarkFindSegmentsMapIDs1000Parallel(b *testing.B) {
	f.BenchmarkFindSegmentsParallel(b, 1000, RandomLinkMapID, RandomFilterOffsetMapIDs)
}

// BenchmarkFindSegmentsMapIDs10000Parallel benchmarks finding segments with several
// map ID within 10000 segments.
func (f Factory) BenchmarkFindSegmentsMapIDs10000Parallel(b *testing.B) {
	f.BenchmarkFindSegmentsParallel(b, 10000, RandomLinkMapID, RandomFilterOffsetMapIDs)
}

// BenchmarkFindSegmentsPrevLinkHash100Parallel benchmarks finding segments with
// a previous link hash within 100 segments.
func (f Factory) BenchmarkFindSegmentsPrevLinkHash100Parallel(b *testing.B) {
	f.BenchmarkFindSegmentsParallel(b, 100, RandomLinkPrevLinkHash, RandomFilterOffsetPrevLinkHash)
}

// BenchmarkFindSegmentsPrevLinkHash1000Parallel benchmarks finding segments
// with a previous link hash within 1000 segments.
func (f Factory) BenchmarkFindSegmentsPrevLinkHash1000Parallel(b *testing.B) {
	f.BenchmarkFindSegmentsParallel(b, 1000, RandomLinkPrevLinkHash, RandomFilterOffsetPrevLinkHash)
}

// BenchmarkFindSegmentsPrevLinkHash10000Parallel benchmarks finding segments
// with a previous link hash within 10000 segments.
func (f Factory) BenchmarkFindSegmentsPrevLinkHash10000Parallel(b *testing.B) {
	f.BenchmarkFindSegmentsParallel(b, 10000, RandomLinkPrevLinkHash, RandomFilterOffsetPrevLinkHash)
}

// BenchmarkFindSegmentsTags100Parallel benchmarks finding segments with tags
// within 100 segments.
func (f Factory) BenchmarkFindSegmentsTags100Parallel(b *testing.B) {
	f.BenchmarkFindSegmentsParallel(b, 100, RandomLinkTags, RandomFilterOffsetTags)
}

// BenchmarkFindSegmentsTags1000Parallel benchmarks finding segments with tags
// within 1000 segments.
func (f Factory) BenchmarkFindSegmentsTags1000Parallel(b *testing.B) {
	f.BenchmarkFindSegmentsParallel(b, 1000, RandomLinkTags, RandomFilterOffsetTags)
}

// BenchmarkFindSegmentsTags10000Parallel benchmarks finding segments with tags
// within 10000 segments.
func (f Factory) BenchmarkFindSegmentsTags10000Parallel(b *testing.B) {
	f.BenchmarkFindSegmentsParallel(b, 10000, RandomLinkTags, RandomFilterOffsetTags)
}

// BenchmarkFindSegmentsMapIDTags100Parallel benchmarks finding segments with
// map ID and tags within 100 segments.
func (f Factory) BenchmarkFindSegmentsMapIDTags100Parallel(b *testing.B) {
	f.BenchmarkFindSegmentsParallel(b, 100, RandomLinkMapIDTags, RandomFilterOffsetMapIDTags)
}

// BenchmarkFindSegmentsMapIDTags1000Parallel benchmarks finding segments with
// map ID and tags within 1000 segments.
func (f Factory) BenchmarkFindSegmentsMapIDTags1000Parallel(b *testing.B) {
	f.BenchmarkFindSegmentsParallel(b, 1000, RandomLinkMapIDTags, RandomFilterOffsetMapIDTags)
}

// BenchmarkFindSegmentsMapIDTags10000Parallel benchmarks finding segments with
// map ID and tags within 10000 segments.
func (f Factory) BenchmarkFindSegmentsMapIDTags10000Parallel(b *testing.B) {
	f.BenchmarkFindSegmentsParallel(b, 10000, RandomLinkMapIDTags, RandomFilterOffsetMapIDTags)
}

// BenchmarkFindSegmentsPrevLinkHashTags100Parallel benchmarks finding segments
// with map ID and tags within 100 segments.
func (f Factory) BenchmarkFindSegmentsPrevLinkHashTags100Parallel(b *testing.B) {
	f.BenchmarkFindSegmentsParallel(b, 100, RandomLinkPrevLinkHashTags, RandomFilterOffsetPrevLinkHashTags)
}

// BenchmarkFindSegmentsPrevLinkHashTags1000Parallel benchmarks finding segments
// with map ID and tags within 1000 segments.
func (f Factory) BenchmarkFindSegmentsPrevLinkHashTags1000Parallel(b *testing.B) {
	f.BenchmarkFindSegmentsParallel(b, 1000, RandomLinkPrevLinkHashTags, RandomFilterOffsetPrevLinkHashTags)
}

// BenchmarkFindSegmentsPrevLinkHashTags10000Parallel benchmarks finding
// segments with map ID and tags within 10000 segments.
func (f Factory) BenchmarkFindSegmentsPrevLinkHashTags10000Parallel(b *testing.B) {
	f.BenchmarkFindSegmentsParallel(b, 10000, RandomLinkPrevLinkHashTags, RandomFilterOffsetPrevLinkHashTags)
}
