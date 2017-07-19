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
	"io/ioutil"
	"log"
	"sync/atomic"
	"testing"

	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/cs/cstesting"
)

// TestSaveSegment tests what happens when you save a new segment.
func (f Factory) TestSaveSegment(t *testing.T) {
	a := f.initAdapter(t)
	defer f.free(a)

	s := cstesting.RandomSegment()
	if err := a.SaveSegment(s); err != nil {
		t.Fatalf("a.SaveSegment(): err: %s", err)
	}
}

// TestSaveSegmentUpdatedState tests what happens when you update the state of a
// segment.
func (f Factory) TestSaveSegmentUpdatedState(t *testing.T) {
	a := f.initAdapter(t)
	defer f.free(a)

	s := cstesting.RandomSegment()
	if err := a.SaveSegment(s); err != nil {
		t.Fatalf("a.SaveSegment(): err: %s", err)
	}

	s = cstesting.ChangeSegmentState(s)
	if err := a.SaveSegment(s); err != nil {
		t.Fatalf("a.SaveSegment(): err: %s", err)
	}
}

// TestSaveSegmentUpdatedMapID tests what happens when you update the map ID of
// a segment.
func (f Factory) TestSaveSegmentUpdatedMapID(t *testing.T) {
	a := f.initAdapter(t)
	defer f.free(a)

	s1 := cstesting.RandomSegment()
	if err := a.SaveSegment(s1); err != nil {
		t.Fatalf("a.SaveSegment(): err: %s", err)
	}

	s2 := cstesting.ChangeSegmentMapID(s1)
	if err := a.SaveSegment(s2); err != nil {
		t.Fatalf("a.SaveSegment(): err: %s", err)
	}
}

// TestSaveSegmentBranch tests what happens when you save a segment with a
// previous link hash.
func (f Factory) TestSaveSegmentBranch(t *testing.T) {
	a := f.initAdapter(t)
	defer f.free(a)

	s := cstesting.RandomSegment()
	if err := a.SaveSegment(s); err != nil {
		t.Fatalf("a.SaveSegment(): err: %s", err)
	}

	s = cstesting.RandomBranch(s)
	if err := a.SaveSegment(s); err != nil {
		t.Fatalf("a.SaveSegment(): err: %s", err)
	}
}

// BenchmarkSaveSegment benchmarks saving new segments.
func (f Factory) BenchmarkSaveSegment(b *testing.B) {
	a := f.initAdapterB(b)
	defer f.free(a)

	slice := make([]*cs.Segment, b.N)
	for i := 0; i < b.N; i++ {
		slice[i] = cstesting.RandomSegment()
	}

	b.ResetTimer()
	log.SetOutput(ioutil.Discard)

	for i := 0; i < b.N; i++ {
		if err := a.SaveSegment(slice[i]); err != nil {
			b.Error(err)
		}
	}
}

// BenchmarkSaveSegmentParallel benchmarks saving new segments in parallel.
func (f Factory) BenchmarkSaveSegmentParallel(b *testing.B) {
	a := f.initAdapterB(b)
	defer f.free(a)

	slice := make([]*cs.Segment, b.N)
	for i := 0; i < b.N; i++ {
		slice[i] = cstesting.RandomSegment()
	}

	var counter uint64

	b.ResetTimer()
	log.SetOutput(ioutil.Discard)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			i := atomic.AddUint64(&counter, 1) - 1
			if err := a.SaveSegment(slice[i]); err != nil {
				b.Error(err)
			}
		}
	})
}

// BenchmarkSaveSegmentUpdatedState benchmarks updating segments states.
func (f Factory) BenchmarkSaveSegmentUpdatedState(b *testing.B) {
	a := f.initAdapterB(b)
	defer f.free(a)

	slice := make([]*cs.Segment, b.N)
	for i := 0; i < b.N; i++ {
		s := cstesting.RandomSegment()
		a.SaveSegment(s)
		slice[i] = cstesting.ChangeSegmentState(s)
	}

	b.ResetTimer()
	log.SetOutput(ioutil.Discard)

	for i := 0; i < b.N; i++ {
		if err := a.SaveSegment(slice[i]); err != nil {
			b.Error(err)
		}
	}
}

// BenchmarkSaveSegmentUpdatedStateParallel benchmarks updating segments states
// in parallel.
func (f Factory) BenchmarkSaveSegmentUpdatedStateParallel(b *testing.B) {
	a := f.initAdapterB(b)
	defer f.free(a)

	slice := make([]*cs.Segment, b.N)
	for i := 0; i < b.N; i++ {
		segment := cstesting.RandomSegment()
		a.SaveSegment(segment)
		slice[i] = cstesting.ChangeSegmentState(segment)
	}
	var counter uint64

	b.ResetTimer()
	log.SetOutput(ioutil.Discard)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			i := atomic.AddUint64(&counter, 1) - 1
			if err := a.SaveSegment(slice[i]); err != nil {
				b.Error(err)
			}
		}
	})
}

// BenchmarkSaveSegmentUpdatedMapID benchmarks updating segment map IDs.
func (f Factory) BenchmarkSaveSegmentUpdatedMapID(b *testing.B) {
	a := f.initAdapterB(b)
	defer f.free(a)

	slice := make([]*cs.Segment, b.N)
	for i := 0; i < b.N; i++ {
		segment := cstesting.RandomSegment()
		a.SaveSegment(segment)
		slice[i] = cstesting.ChangeSegmentMapID(segment)
	}

	b.ResetTimer()
	log.SetOutput(ioutil.Discard)

	for i := 0; i < b.N; i++ {
		if err := a.SaveSegment(slice[i]); err != nil {
			b.Error(err)
		}
	}
}

// BenchmarkSaveSegmentUpdatedMapIDParallel benchmarks updating segment map IDs
// in parallel.
func (f Factory) BenchmarkSaveSegmentUpdatedMapIDParallel(b *testing.B) {
	a := f.initAdapterB(b)
	defer f.free(a)

	slice := make([]*cs.Segment, b.N)
	for i := 0; i < b.N; i++ {
		segment := cstesting.RandomSegment()
		a.SaveSegment(segment)
		slice[i] = cstesting.ChangeSegmentMapID(segment)
	}
	var counter uint64

	b.ResetTimer()
	log.SetOutput(ioutil.Discard)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			i := atomic.AddUint64(&counter, 1) - 1
			if err := a.SaveSegment(slice[i]); err != nil {
				b.Error(err)
			}
		}
	})
}
