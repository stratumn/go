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
	"encoding/json"
	"io/ioutil"
	"log"
	"reflect"
	"sync/atomic"
	"testing"

	"github.com/stratumn/go/cs/cstesting"
	"github.com/stratumn/go/testutil"
	"github.com/stratumn/go/types"
)

// TestDeleteSegment tests what happens when you delete an existing segments.
func (f Factory) TestDeleteSegment(t *testing.T) {
	a, err := f.New()
	if err != nil {
		t.Fatalf("f.New(): err: %s", err)
	}
	if a == nil {
		t.Fatal("a = nil want store.Adapter")
	}
	defer f.free(a)

	s1 := cstesting.RandomSegment()
	a.SaveSegment(s1)

	linkHash, err := types.NewBytes32FromString(s1.Meta["linkHash"].(string))
	if err != nil {
		t.Fatalf("types.NewBytes32FromString(): err: %s", err)
	}

	s2, err := a.DeleteSegment(linkHash)
	if err != nil {
		t.Fatalf("a.DeleteSegment(): err: %s", err)
	}

	if got := s2; got == nil {
		t.Error("s2 = nil want *cs.Segment")
	}
	if got, want := s2, s1; !reflect.DeepEqual(want, got) {
		gotJS, _ := json.MarshalIndent(got, "", "  ")
		wantJS, _ := json.MarshalIndent(got, "", "  ")
		t.Errorf("s2 = %s\n want%s", gotJS, wantJS)
	}

	s2, err = a.GetSegment(linkHash)
	if err != nil {
		t.Fatalf("a.GetSegment(): err: %s", err)
	}
	if got := s2; got != nil {
		gotJS, _ := json.MarshalIndent(got, "", "  ")
		t.Errorf("s2 = %s\n want nil", gotJS)
	}
}

// TestDeleteSegmentNotFound tests what happens when you delete a nonexistent segment.
func (f Factory) TestDeleteSegmentNotFound(t *testing.T) {
	a, err := f.New()
	if err != nil {
		t.Fatalf("f.New(): err: %s", err)
	}
	if a == nil {
		t.Fatal("a = nil want store.Adapter")
	}
	defer f.free(a)

	s, err := a.DeleteSegment(testutil.RandomHash())
	if err != nil {
		t.Fatalf("a.DeleteSegment(): err: %s", err)
	}

	if got := s; got != nil {
		gotJS, _ := json.MarshalIndent(got, "", "  ")
		t.Errorf("s = %s\n want nil", gotJS)
	}
}

// BenchmarkDeleteSegment benchmarks deleting existing segments.
func (f Factory) BenchmarkDeleteSegment(b *testing.B) {
	a, err := f.New()
	if err != nil {
		b.Fatalf("f.New(): err: %s", err)
	}
	if a == nil {
		b.Fatal("a = nil want store.Adapter")
	}
	defer f.free(a)

	linkHashes := make([]*types.Bytes32, b.N)
	for i := 0; i < b.N; i++ {
		s := cstesting.RandomSegment()
		a.SaveSegment(s)
		linkHashes[i], _ = types.NewBytes32FromString(s.Meta["linkHash"].(string))
	}

	b.ResetTimer()
	log.SetOutput(ioutil.Discard)

	for i := 0; i < b.N; i++ {
		if s, err := a.DeleteSegment(linkHashes[i]); err != nil {
			b.Error(err)
		} else if s == nil {
			b.Error("s = nil want *cs.Segment")
		}
	}
}

// BenchmarkDeleteSegmentParallel benchmarks deleting existing segments in parallel.
func (f Factory) BenchmarkDeleteSegmentParallel(b *testing.B) {
	a, err := f.New()
	if err != nil {
		b.Fatalf("f.New(): err: %s", err)
	}
	if a == nil {
		b.Fatal("a = nil want store.Adapter")
	}
	defer f.free(a)

	linkHashes := make([]*types.Bytes32, b.N)
	for i := 0; i < b.N; i++ {
		s := cstesting.RandomSegment()
		a.SaveSegment(s)
		linkHashes[i], _ = types.NewBytes32FromString(s.Meta["linkHash"].(string))
	}

	var counter uint64

	b.ResetTimer()
	log.SetOutput(ioutil.Discard)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			i := atomic.AddUint64(&counter, 1) - 1
			if s, err := a.DeleteSegment(linkHashes[i]); err != nil {
				b.Error(err)
			} else if s == nil {
				b.Error("s = nil want *cs.Segment")
			}
		}
	})
}
