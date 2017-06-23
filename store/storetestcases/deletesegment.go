// Copyright 2017 Stratumn SAS. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

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

// TestDeleteSegment tests what happens when you delete an existing segments.
func (f Factory) TestDeleteSegment(t *testing.T) {
	a := f.initAdapter(t)
	defer f.free(a)

	s1 := cstesting.RandomSegment()
	a.SaveSegment(s1)

	linkHash := s1.GetLinkHash()
	s2, err := a.DeleteSegment(linkHash)
	if err != nil {
		t.Fatalf("a.DeleteSegment(): err: %s", err)
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

	s2, err = a.GetSegment(linkHash)
	if err != nil {
		t.Fatalf("a.GetSegment(): err: %s", err)
	}
	if got := s2; got != nil {
		gotJS, _ := json.MarshalIndent(got, "", "  ")
		t.Errorf("s2 = %s\n want nil", gotJS)
	}
}

// TestDeleteSegmentNotFound tests what happens when you delete a nonexistent
// segment.
func (f Factory) TestDeleteSegmentNotFound(t *testing.T) {
	a := f.initAdapter(t)
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
		if s, err := a.DeleteSegment(linkHashes[i]); err != nil {
			b.Error(err)
		} else if s == nil {
			b.Error("s = nil want *cs.Segment")
		}
	}
}

// BenchmarkDeleteSegmentParallel benchmarks deleting existing segments in
// parallel.
func (f Factory) BenchmarkDeleteSegmentParallel(b *testing.B) {
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
			if s, err := a.DeleteSegment(linkHashes[i]); err != nil {
				b.Error(err)
			} else if s == nil {
				b.Error("s = nil want *cs.Segment")
			}
		}
	})
}
