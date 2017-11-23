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
	"encoding/json"
	"io/ioutil"
	"log"
	"reflect"
	"sync/atomic"
	"testing"

	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/cs/cstesting"
	// import every type of evidence to see if we can deserialize all of them
	_ "github.com/stratumn/sdk/cs/evidences"
	"github.com/stratumn/sdk/testutil"
	"github.com/stratumn/sdk/types"
)

// TestGetSegment tests what happens when you get an existing segment.
func (f Factory) TestGetSegment(t *testing.T) {
	a := f.initAdapter(t)
	defer f.free(a)

	s1 := cstesting.RandomSegment()

	a.SaveSegment(s1)

	s2, err := a.GetSegment(s1.GetLinkHash())
	if err != nil {
		t.Fatalf("a.GetSegment(): err: %s", err)
	}

	if got := s2; got == nil {
		t.Fatal("s2 = nil want *cs.Segment")
	}

	s2.Meta.Evidences = nil
	if got, want := s2, s1; !reflect.DeepEqual(want, got) {
		gotJS, _ := json.MarshalIndent(got, "", "  ")
		wantJS, _ := json.MarshalIndent(want, "", "  ")
		t.Errorf("s2 = %s\n want%s", gotJS, wantJS)
	}
}

// TestGetSegmentV2 tests what happens when you get an existing segment.
func (f Factory) TestGetSegmentV2(t *testing.T) {
	a := f.initAdapterV2(t)
	defer f.freeV2(a)

	l1 := cstesting.RandomLink()
	linkHash, _ := a.CreateLink(l1)

	s2, err := a.GetSegment(linkHash)
	if err != nil {
		t.Fatalf("a.GetSegment(): err: %s", err)
	}

	if got := s2; got == nil {
		t.Fatal("s2 = nil want *cs.Segment")
	}

	if got, want := &s2.Link, l1; !reflect.DeepEqual(want, got) {
		gotJS, _ := json.MarshalIndent(got, "", "  ")
		wantJS, _ := json.MarshalIndent(want, "", "  ")
		t.Errorf("s2 = %s\n want%s", gotJS, wantJS)
	}
}

// TestGetSegmentUpdatedState tests what happens when you get a segment whose
// state was updated.
func (f Factory) TestGetSegmentUpdatedState(t *testing.T) {
	a := f.initAdapter(t)
	defer f.free(a)

	s1 := cstesting.RandomSegment()
	a.SaveSegment(s1)
	s2 := cstesting.ChangeSegmentState(s1)
	a.SaveSegment(s2)

	got, err := a.GetSegment(s1.GetLinkHash())
	if err != nil {
		t.Fatalf("a.GetSegment(): err: %s", err)
	}

	if got == nil {
		t.Fatal("s2 = nil want *cs.Segment")
	}

	got.Meta.Evidences = nil
	want := s1
	if !reflect.DeepEqual(want, got) {
		gotJS, _ := json.MarshalIndent(got, "", "  ")
		wantJS, _ := json.MarshalIndent(want, "", "  ")
		t.Errorf("got = %s\n want %s", gotJS, wantJS)
	}
}

// TestGetSegmentUpdatedStateV2 tests what happens when you get a segment whose
// state was updated.
func (f Factory) TestGetSegmentUpdatedStateV2(t *testing.T) {
	a := f.initAdapterV2(t)
	defer f.freeV2(a)

	l1 := cstesting.RandomLink()
	linkHash1, _ := a.CreateLink(l1)
	l2 := cstesting.ChangeLinkState(l1)
	a.CreateLink(l2)

	got, err := a.GetSegment(linkHash1)
	if err != nil {
		t.Fatalf("a.GetSegment(): err: %s", err)
	}

	if got == nil {
		t.Fatal("s2 = nil want *cs.Segment")
	}

	if got, want := &got.Link, l1; !reflect.DeepEqual(want, got) {
		gotJS, _ := json.MarshalIndent(got, "", "  ")
		wantJS, _ := json.MarshalIndent(want, "", "  ")
		t.Errorf("got = %s\n want %s", gotJS, wantJS)
	}
}

// TestGetSegmentUpdatedMapID tests what happens when you get a segment whose
// map ID was updated.
func (f Factory) TestGetSegmentUpdatedMapID(t *testing.T) {
	a := f.initAdapter(t)
	defer f.free(a)

	s1 := cstesting.RandomSegment()
	a.SaveSegment(s1)
	s2 := cstesting.ChangeSegmentMapID(s1)
	a.SaveSegment(s2)

	got, err := a.GetSegment(s1.GetLinkHash())
	if err != nil {
		t.Fatalf("a.GetSegment(): err: %s", err)
	}

	if got == nil {
		t.Fatal("s2 = nil want *cs.Segment")
	}

	got.Meta.Evidences = nil
	want := s1
	if !reflect.DeepEqual(want, got) {
		gotJS, _ := json.MarshalIndent(got, "", "  ")
		wantJS, _ := json.MarshalIndent(want, "", "  ")
		t.Errorf("s2 = %s\n want%s", gotJS, wantJS)
	}
}

// TestGetSegmentUpdatedMapIDV2 tests what happens when you get a segment whose
// map ID was updated.
func (f Factory) TestGetSegmentUpdatedMapIDV2(t *testing.T) {
	a := f.initAdapterV2(t)
	defer f.freeV2(a)

	l1 := cstesting.RandomLink()
	linkHash1, _ := a.CreateLink(l1)
	l2 := cstesting.ChangeLinkMapID(l1)
	a.CreateLink(l2)

	got, err := a.GetSegment(linkHash1)
	if err != nil {
		t.Fatalf("a.GetSegment(): err: %s", err)
	}

	if got == nil {
		t.Fatal("s2 = nil want *cs.Segment")
	}

	if got, want := &got.Link, l1; !reflect.DeepEqual(want, got) {
		gotJS, _ := json.MarshalIndent(got, "", "  ")
		wantJS, _ := json.MarshalIndent(want, "", "  ")
		t.Errorf("s2 = %s\n want%s", gotJS, wantJS)
	}
}

// TestGetSegmentWithEvidences tests what happens when you update the state of a
// segment.
func (f Factory) TestGetSegmentWithEvidences(t *testing.T) {
	a := f.initAdapter(t)
	defer f.free(a)

	e1 := cs.Evidence{Backend: "TMPop", Provider: "1"}
	e2 := cs.Evidence{Backend: "dummy", Provider: "2"}
	e3 := cs.Evidence{Backend: "batch", Provider: "3"}
	e4 := cs.Evidence{Backend: "bcbatch", Provider: "4"}
	e5 := cs.Evidence{Backend: "generic", Provider: "5"}

	s := cstesting.RandomSegment()
	s.Meta.AddEvidence(e1)
	s.Meta.AddEvidence(e2)
	s.Meta.AddEvidence(e3)
	s.Meta.AddEvidence(e4)
	s.Meta.AddEvidence(e5)

	if err := a.SaveSegment(s); err != nil {
		t.Fatalf("a.SaveSegment(): err: %s", err)
	}
	got, err := a.GetSegment(s.GetLinkHash())
	if err != nil {
		t.Fatalf("a.GetSegment(): err: %s", err)
	}

	if got == nil {
		t.Fatal("s2 = nil want *cs.Segment")
	}
}

// TestGetSegmentWithEvidencesV2 tests what happens when you add
// evidence to a segment.
func (f Factory) TestGetSegmentWithEvidencesV2(t *testing.T) {
	a := f.initAdapterV2(t)
	defer f.freeV2(a)

	e1 := cs.Evidence{Backend: "TMPop", Provider: "1"}
	e2 := cs.Evidence{Backend: "dummy", Provider: "2"}
	e3 := cs.Evidence{Backend: "batch", Provider: "3"}
	e4 := cs.Evidence{Backend: "bcbatch", Provider: "4"}
	e5 := cs.Evidence{Backend: "generic", Provider: "5"}
	evidences := []cs.Evidence{e1, e2, e3, e4, e5}

	l := cstesting.RandomLink()
	linkHash, err := a.CreateLink(l)
	if err != nil {
		t.Fatalf("a.CreateLink(): err: %s", err)
	}

	for _, e := range evidences {
		if err = a.AddEvidence(linkHash, &e); err != nil {
			t.Fatalf("a.AddEvidence(): err: %s", err)
		}
	}

	got, err := a.GetSegment(linkHash)
	if err != nil {
		t.Fatalf("a.GetSegment(): err: %s", err)
	}
	if got == nil {
		t.Fatal("s2 = nil want *cs.Segment")
	}
	if len(got.Meta.Evidences) != 5 {
		t.Fatalf("Invalid number of evidences: got %d, expected %d",
			len(got.Meta.Evidences), 5)
	}
}

// TestGetSegmentNotFound tests what happens when you get a nonexistent segment.
func (f Factory) TestGetSegmentNotFound(t *testing.T) {
	a := f.initAdapter(t)
	defer f.free(a)

	s, err := a.GetSegment(testutil.RandomHash())
	if err != nil {
		t.Fatalf("a.GetSegment(): err: %s", err)
	}

	if got := s; got != nil {
		gotJS, _ := json.MarshalIndent(got, "", "  ")
		t.Errorf("s = %s\n want nil", gotJS)
	}
}

// TestGetSegmentNotFoundV2 tests what happens when you get a nonexistent segment.
func (f Factory) TestGetSegmentNotFoundV2(t *testing.T) {
	a := f.initAdapterV2(t)
	defer f.freeV2(a)

	s, err := a.GetSegment(testutil.RandomHash())
	if err != nil {
		t.Fatalf("a.GetSegment(): err: %s", err)
	}

	if got := s; got != nil {
		gotJS, _ := json.MarshalIndent(got, "", "  ")
		t.Errorf("s = %s\n want nil", gotJS)
	}
}

// BenchmarkGetSegment benchmarks getting existing segments.
func (f Factory) BenchmarkGetSegment(b *testing.B) {
	a := f.initAdapterB(b)
	defer f.free(a)

	linkHashes := make([]*types.Bytes32, b.N)
	for i := 0; i < b.N; i++ {
		s := cstesting.RandomSegment()
		a.SaveSegment(s)
		linkHashes[i] = s.GetLinkHash()
	}

	b.ResetTimer()
	log.SetOutput(ioutil.Discard)

	for i := 0; i < b.N; i++ {
		if s, err := a.GetSegment(linkHashes[i]); err != nil {
			b.Fatal(err)
		} else if s == nil {
			b.Error("s = nil want *cs.Segment")
		}
	}
}

// BenchmarkGetSegmentParallel benchmarks getting existing segments in parallel.
func (f Factory) BenchmarkGetSegmentParallel(b *testing.B) {
	a := f.initAdapterB(b)
	defer f.free(a)

	linkHashes := make([]*types.Bytes32, b.N)
	for i := 0; i < b.N; i++ {
		s := cstesting.RandomSegment()
		a.SaveSegment(s)
		linkHashes[i] = s.GetLinkHash()
	}

	var counter uint64

	b.ResetTimer()
	log.SetOutput(ioutil.Discard)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			i := atomic.AddUint64(&counter, 1) - 1
			if s, err := a.GetSegment(linkHashes[i]); err != nil {
				b.Error(err)
			} else if s == nil {
				b.Error("s = nil want *cs.Segment")
			}
		}
	})
}
