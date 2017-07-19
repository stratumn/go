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

	"github.com/stratumn/sdk/cs/cstesting"
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

	delete(s2.Meta, "evidence")

	if got, want := s2, s1; !reflect.DeepEqual(want, got) {
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
	s1 = cstesting.ChangeSegmentState(s1)
	a.SaveSegment(s1)

	s2, err := a.GetSegment(s1.GetLinkHash())
	if err != nil {
		t.Fatalf("a.GetSegment(): err: %s", err)
	}

	if got := s2; got == nil {
		t.Fatal("s2 = nil want *cs.Segment")
	}

	delete(s2.Meta, "evidence")

	if got, want := s2, s1; !reflect.DeepEqual(want, got) {
		gotJS, _ := json.MarshalIndent(got, "", "  ")
		wantJS, _ := json.MarshalIndent(want, "", "  ")
		t.Errorf("s2 = %s\n want%s", gotJS, wantJS)
	}
}

// TestGetSegmentUpdatedMapID tests what happens when you get a segment whose
// map ID was updated.
func (f Factory) TestGetSegmentUpdatedMapID(t *testing.T) {
	a := f.initAdapter(t)
	defer f.free(a)

	s1 := cstesting.RandomSegment()
	a.SaveSegment(s1)
	s1 = cstesting.ChangeSegmentMapID(s1)
	a.SaveSegment(s1)

	s2, err := a.GetSegment(s1.GetLinkHash())
	if err != nil {
		t.Fatalf("a.GetSegment(): err: %s", err)
	}

	if got := s2; got == nil {
		t.Fatal("s2 = nil want *cs.Segment")
	}

	delete(s2.Meta, "evidence")

	if got, want := s2, s1; !reflect.DeepEqual(want, got) {
		gotJS, _ := json.MarshalIndent(got, "", "  ")
		wantJS, _ := json.MarshalIndent(want, "", "  ")
		t.Errorf("s2 = %s\n want%s", gotJS, wantJS)
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
