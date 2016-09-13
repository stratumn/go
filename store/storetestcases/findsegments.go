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

package storetestcases

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stratumn/go/cs/cstesting"
	"github.com/stratumn/go/store"
	"github.com/stratumn/go/testutil"
)

// TestFindSegments tests what happens when you search with default pagination.
func (f Factory) TestFindSegments(t *testing.T) {
	a, err := f.New()
	if err != nil {
		t.Fatalf("f.New(): err: %s", err)
	}
	if a == nil {
		t.Fatal("a = nil want store.Adapter")
	}
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

// TestFindSegmentsPagination tests what happens when you search with pagination.
func (f Factory) TestFindSegmentsPagination(t *testing.T) {
	a, err := f.New()
	if err != nil {
		t.Fatalf("f.New(): err: %s", err)
	}
	if a == nil {
		t.Fatal("a = nil want store.Adapter")
	}
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

	if got, want := len(slice), limit; want != got {
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
	a, err := f.New()
	if err != nil {
		t.Fatalf("f.New(): err: %s", err)
	}
	if a == nil {
		t.Fatal("a = nil want store.Adapter")
	}
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

	if got, want := len(slice), 0; want != got {
		t.Errorf("len(slice) = %d want %d", got, want)
	}
}

// TestFindSegmentsSingleTag tests what happens when you search with only one tag.
func (f Factory) TestFindSegmentsSingleTag(t *testing.T) {
	a, err := f.New()
	if err != nil {
		t.Fatalf("f.New(): err: %s", err)
	}
	if a == nil {
		t.Fatal("a = nil want store.Adapter")
	}
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

	slice, err := a.FindSegments(&store.Filter{
		Pagination: store.Pagination{
			Limit: store.DefaultLimit * 2,
		},
		Tags: []string{tag1},
	})
	if err != nil {
		t.Fatalf("a.FindSegments(): err: %s", err)
	}

	if got, want := len(slice), store.DefaultLimit*2; want != got {
		t.Errorf("len(slice) = %d want %d", got, want)
	}
}

// TestFindSegmentsMultipleTags tests what happens when you search with more than one tag.
func (f Factory) TestFindSegmentsMultipleTags(t *testing.T) {
	a, err := f.New()
	if err != nil {
		t.Fatalf("f.New(): err: %s", err)
	}
	if a == nil {
		t.Fatal("a = nil want store.Adapter")
	}
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

	slice, err := a.FindSegments(&store.Filter{
		Pagination: store.Pagination{
			Limit: store.DefaultLimit * 2,
		},
		Tags: []string{tag2, tag1},
	})
	if err != nil {
		t.Fatalf("a.FindSegments(): err: %s", err)
	}

	if got, want := len(slice), store.DefaultLimit; want != got {
		t.Errorf("len(slice) = %d want %d", got, want)
	}
}

// TestFindSegmentsMapID tests whan happens when you search for an existing map ID.
func (f Factory) TestFindSegmentsMapID(t *testing.T) {
	a, err := f.New()
	if err != nil {
		t.Fatalf("f.New(): err: %s", err)
	}
	if a == nil {
		t.Fatal("a = nil want store.Adapter")
	}
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
	if got, want := len(slice), store.DefaultLimit; want != got {
		t.Errorf("len(slice) = %d want %d", got, want)
	}
}

// TestFindSegmentsMapIDNotFound tests whan happens when you search for a nonexistent map ID.
func (f Factory) TestFindSegmentsMapIDNotFound(t *testing.T) {
	a, err := f.New()
	if err != nil {
		t.Fatalf("f.New(): err: %s", err)
	}
	if a == nil {
		t.Fatal("a = nil want store.Adapter")
	}
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

	if got, want := len(slice), 0; want != got {
		t.Errorf("len(slice) = %d want %d", got, want)
	}
}

// TestFindSegmentsPrevLinkHash tests whan happens when you search for an existing previous link hash.
func (f Factory) TestFindSegmentsPrevLinkHash(t *testing.T) {
	a, err := f.New()
	if err != nil {
		t.Fatalf("f.New(): err: %s", err)
	}
	if a == nil {
		t.Fatal("a = nil want store.Adapter")
	}
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
	if got, want := len(slice), store.DefaultLimit; want != got {
		t.Errorf("len(slice) = %d want %d", got, want)
	}
}

// TestFindSegmentsPrevLinkHashNotFound tests whan happens when you search for a nonexistent previous link hash.
func (f Factory) TestFindSegmentsPrevLinkHashNotFound(t *testing.T) {
	a, err := f.New()
	if err != nil {
		t.Fatalf("f.New(): err: %s", err)
	}
	if a == nil {
		t.Fatal("a = nil want store.Adapter")
	}
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

	if got, want := len(slice), 0; want != got {
		t.Errorf("len(slice) = %d want %d", got, want)
	}
}
