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
	"context"
	"io/ioutil"
	"log"
	"sync/atomic"
	"testing"

	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/cs/cstesting"
	"github.com/stratumn/go-indigocore/testutil"
	// import every type of evidence to see if we can deserialize all of them
	_ "github.com/stratumn/go-indigocore/cs/evidences"
	"github.com/stratumn/go-indigocore/types"
	"github.com/stretchr/testify/assert"
)

// TestGetSegment tests what happens when you get a segment.
func (f Factory) TestGetSegment(t *testing.T) {
	a := f.initAdapter(t)
	defer f.freeAdapter(a)

	link := cstesting.RandomLink()
	linkHash, _ := a.CreateLink(context.Background(), link)

	link2 := cstesting.ChangeState(link)
	linkHash2, _ := a.CreateLink(context.Background(), link2)

	t.Run("Getting an existing segment should work", func(t *testing.T) {
		ctx := context.Background()
		s, err := a.GetSegment(ctx, linkHash)
		assert.NoError(t, err)
		assert.NotNil(t, s, "Segment should be found")
		assert.EqualValues(t, link, &s.Link, "Invalid link")
		gotHash, err := s.Link.Hash()
		assert.NoError(t, err, "Hash should be computed")
		assert.EqualValues(t, linkHash, gotHash, "Invalid linkHash")
	})

	t.Run("Getting an updated segment should work", func(t *testing.T) {
		ctx := context.Background()
		got, err := a.GetSegment(ctx, linkHash2)
		assert.NoError(t, err)
		assert.NotNil(t, got, "Segment should be found")
		assert.EqualValues(t, link2, &got.Link, "Invalid link")
		gotHash, err := got.Link.Hash()
		assert.NoError(t, err, "Hash should be computed")
		assert.EqualValues(t, linkHash2, gotHash, "Invalid linkHash")
	})

	t.Run("Getting an unknown segment should return nil", func(t *testing.T) {
		ctx := context.Background()
		s, err := a.GetSegment(ctx, testutil.RandomHash())
		assert.NoError(t, err)
		assert.Nil(t, s)
	})

	t.Run("Getting a segment should return its evidences", func(t *testing.T) {
		ctx := context.Background()
		e1 := cs.Evidence{Backend: "TMPop", Provider: "1"}
		e2 := cs.Evidence{Backend: "dummy", Provider: "2"}
		e3 := cs.Evidence{Backend: "batch", Provider: "3"}
		e4 := cs.Evidence{Backend: "bcbatch", Provider: "4"}
		e5 := cs.Evidence{Backend: "generic", Provider: "5"}
		evidences := []cs.Evidence{e1, e2, e3, e4, e5}

		for _, e := range evidences {
			err := a.AddEvidence(ctx, linkHash2, &e)
			assert.NoError(t, err, "a.AddEvidence()")
		}

		got, err := a.GetSegment(ctx, linkHash2)
		assert.NoError(t, err, "a.GetSegment()")
		assert.NotNil(t, got)
		assert.Len(t, got.Meta.Evidences, 5, "Invalid number of evidences")
	})
}

// BenchmarkGetSegment benchmarks getting existing segments.
func (f Factory) BenchmarkGetSegment(b *testing.B) {
	a := f.initAdapterB(b)
	defer f.freeAdapter(a)

	linkHashes := make([]*types.Bytes32, b.N)
	for i := 0; i < b.N; i++ {
		l := cstesting.RandomLink()
		linkHash, _ := a.CreateLink(context.Background(), l)
		linkHashes[i] = linkHash
	}

	b.ResetTimer()
	log.SetOutput(ioutil.Discard)

	for i := 0; i < b.N; i++ {
		if s, err := a.GetSegment(context.Background(), linkHashes[i]); err != nil {
			b.Fatal(err)
		} else if s == nil {
			b.Error("s = nil want *cs.Segment")
		}
	}
}

// BenchmarkGetSegmentParallel benchmarks getting existing segments in parallel.
func (f Factory) BenchmarkGetSegmentParallel(b *testing.B) {
	a := f.initAdapterB(b)
	defer f.freeAdapter(a)

	linkHashes := make([]*types.Bytes32, b.N)
	for i := 0; i < b.N; i++ {
		l := cstesting.RandomLink()
		linkHash, _ := a.CreateLink(context.Background(), l)
		linkHashes[i] = linkHash
	}

	var counter uint64

	b.ResetTimer()
	log.SetOutput(ioutil.Discard)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			i := atomic.AddUint64(&counter, 1) - 1
			if s, err := a.GetSegment(context.Background(), linkHashes[i]); err != nil {
				b.Error(err)
			} else if s == nil {
				b.Error("s = nil want *cs.Segment")
			}
		}
	})
}
