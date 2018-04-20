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

	batchevidences "github.com/stratumn/go-indigocore/batchfossilizer/evidences"
	bcbatchevidences "github.com/stratumn/go-indigocore/bcbatchfossilizer/evidences"
	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/cs/cstesting"
	"github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var emptyPrevLinkHash = ""

func createLink(adapter store.Adapter, link *cs.Link, prepareLink func(l *cs.Link)) *cs.Link {
	if prepareLink != nil {
		prepareLink(link)
	}
	adapter.CreateLink(context.Background(), link)
	return link
}

func createRandomLink(adapter store.Adapter, prepareLink func(l *cs.Link)) *cs.Link {
	return createLink(adapter, cstesting.RandomLink(), prepareLink)
}

func createLinkBranch(adapter store.Adapter, parent *cs.Link, prepareLink func(l *cs.Link)) *cs.Link {
	return createLink(adapter, cstesting.NewLinkBuilder().Branch(parent).Build(), prepareLink)
}

func verifyPriorityOrdering(t *testing.T, slice cs.SegmentSlice) {
	wantLTE := 100.0
	for _, s := range slice {
		got := s.Link.Meta.Priority
		assert.True(t, got <= wantLTE, "Invalid priority")
		wantLTE = got
	}
}

func verifyResultsCount(t *testing.T, err error, slice cs.SegmentSlice, expectedCount int) {
	assert.NoError(t, err)
	assert.NotNil(t, slice)
	assert.Equal(t, expectedCount, len(slice), "Invalid number of results")
}

// TestFindSegments tests what happens when you search for segments with various filters.
func (f Factory) TestFindSegments(t *testing.T) {
	a := f.initAdapter(t)
	defer f.freeAdapter(a)

	// Setup a test fixture with segments matching different types of filters

	testPageSize := 3
	segmentsTotalCount := 8

	createRandomLink(a, func(l *cs.Link) {
		l.Meta.MapID = "map1"
		l.Meta.PrevLinkHash = ""
		l.Meta.Process = "Foo"
	})

	createRandomLink(a, func(l *cs.Link) {
		l.Meta.Tags = []string{"tag1", "tag42"}
		l.Meta.MapID = "map2"
	})

	createRandomLink(a, func(l *cs.Link) {
		l.Meta.Tags = []string{"tag2"}
	})

	link4 := createRandomLink(a, nil)
	linkHash4, _ := link4.Hash()

	createLinkBranch(a, link4, func(l *cs.Link) {
		l.Meta.Tags = []string{"tag1", testutil.RandomString(5)}
		l.Meta.MapID = "map1"
	})

	link6 := createRandomLink(a, func(l *cs.Link) {
		l.Meta.Tags = []string{"tag2", "tag42"}
		l.Meta.Process = "Foo"
		l.Meta.PrevLinkHash = ""
	})
	linkHash6, _ := link6.Hash()

	createRandomLink(a, func(l *cs.Link) {
		l.Meta.MapID = "map2"
	})

	createLinkBranch(a, link4, nil)

	t.Run("Should order by priority", func(t *testing.T) {
		ctx := context.Background()
		slice, err := a.FindSegments(ctx, &store.SegmentFilter{
			Pagination: store.Pagination{
				Limit: testPageSize,
			},
		})
		verifyResultsCount(t, err, slice, testPageSize)
		verifyPriorityOrdering(t, slice)
	})

	t.Run("Should support pagination", func(t *testing.T) {
		ctx := context.Background()
		slice, err := a.FindSegments(ctx, &store.SegmentFilter{
			Pagination: store.Pagination{
				Offset: testPageSize,
				Limit:  testPageSize,
			},
		})
		verifyResultsCount(t, err, slice, testPageSize)
		verifyPriorityOrdering(t, slice)
	})

	t.Run("Should return no results for invalid tag filter", func(t *testing.T) {
		ctx := context.Background()
		slice, err := a.FindSegments(ctx, &store.SegmentFilter{
			Tags: []string{"blablabla"},
		})
		verifyResultsCount(t, err, slice, 0)
	})

	t.Run("Supports tags filtering", func(t *testing.T) {
		ctx := context.Background()
		slice, err := a.FindSegments(ctx, &store.SegmentFilter{
			Pagination: store.Pagination{
				Limit: segmentsTotalCount,
			},
			Tags: []string{"tag1"},
		})
		verifyResultsCount(t, err, slice, 2)
	})

	t.Run("Supports filtering on multiple tags", func(t *testing.T) {
		ctx := context.Background()
		slice, err := a.FindSegments(ctx, &store.SegmentFilter{
			Pagination: store.Pagination{
				Limit: segmentsTotalCount,
			},
			Tags: []string{"tag2", "tag42"},
		})
		verifyResultsCount(t, err, slice, 1)
	})

	t.Run("Supports filtering on map ID", func(t *testing.T) {
		ctx := context.Background()
		slice, err := a.FindSegments(ctx, &store.SegmentFilter{
			Pagination: store.Pagination{
				Limit: segmentsTotalCount,
			},
			MapIDs: []string{"map1"},
		})
		verifyResultsCount(t, err, slice, 2)
	})

	t.Run("Supports filtering on multiple map IDs", func(t *testing.T) {
		ctx := context.Background()
		slice, err := a.FindSegments(ctx, &store.SegmentFilter{
			Pagination: store.Pagination{
				Limit: segmentsTotalCount,
			},
			MapIDs: []string{"map1", "map2"},
		})
		verifyResultsCount(t, err, slice, 4)
	})

	t.Run("Supports filtering on map ID and tag at the same time", func(t *testing.T) {
		ctx := context.Background()
		slice, err := a.FindSegments(ctx, &store.SegmentFilter{
			Pagination: store.Pagination{
				Limit: segmentsTotalCount,
			},
			MapIDs: []string{"map1"},
			Tags:   []string{"tag1"},
		})
		verifyResultsCount(t, err, slice, 1)
	})

	t.Run("Returns no results for map ID not found", func(t *testing.T) {
		ctx := context.Background()
		slice, err := a.FindSegments(ctx, &store.SegmentFilter{
			Pagination: store.Pagination{
				Limit: segmentsTotalCount,
			},
			MapIDs: []string{"yolo42000"},
		})
		verifyResultsCount(t, err, slice, 0)
	})

	t.Run("Supports filtering on link hashes", func(t *testing.T) {
		ctx := context.Background()
		slice, err := a.FindSegments(ctx, &store.SegmentFilter{
			LinkHashes: []string{
				linkHash4.String(),
				testutil.RandomHash().String(),
				linkHash6.String(),
			},
			Pagination: store.Pagination{
				Limit: segmentsTotalCount,
			},
		})
		verifyResultsCount(t, err, slice, 2)
	})

	t.Run("Supports filtering on link hash and process at the same time", func(t *testing.T) {
		ctx := context.Background()
		slice, err := a.FindSegments(ctx, &store.SegmentFilter{
			LinkHashes: []string{
				linkHash4.String(),
				linkHash6.String(),
			},
			Process: "Foo",
			Pagination: store.Pagination{
				Limit: segmentsTotalCount,
			},
		})
		verifyResultsCount(t, err, slice, 1)
	})

	t.Run("Should return no results for unknown link hashes", func(t *testing.T) {
		ctx := context.Background()
		slice, err := a.FindSegments(ctx, &store.SegmentFilter{
			LinkHashes: []string{
				testutil.RandomHash().String(),
				testutil.RandomHash().String(),
			},
			Pagination: store.Pagination{
				Limit: segmentsTotalCount,
			},
		})
		verifyResultsCount(t, err, slice, 0)
	})

	t.Run("Supports filtering for segments with empty previous link hash", func(t *testing.T) {
		ctx := context.Background()
		slice, err := a.FindSegments(ctx, &store.SegmentFilter{
			Pagination:   store.Pagination{Limit: segmentsTotalCount},
			PrevLinkHash: &emptyPrevLinkHash,
		})
		verifyResultsCount(t, err, slice, 2)
	})

	t.Run("Supports filtering by previous link hash", func(t *testing.T) {
		ctx := context.Background()
		prevLinkHash := linkHash4.String()
		slice, err := a.FindSegments(ctx, &store.SegmentFilter{
			Pagination: store.Pagination{
				Limit: segmentsTotalCount,
			},
			PrevLinkHash: &prevLinkHash,
		})
		verifyResultsCount(t, err, slice, 2)
	})

	t.Run("Supports filtering by previous link hash and tags at the same time", func(t *testing.T) {
		ctx := context.Background()
		prevLinkHash := linkHash4.String()
		slice, err := a.FindSegments(ctx, &store.SegmentFilter{
			Pagination: store.Pagination{
				Limit: segmentsTotalCount,
			},
			PrevLinkHash: &prevLinkHash,
			Tags:         []string{"tag1"},
		})
		verifyResultsCount(t, err, slice, 1)
	})

	t.Run("Supports filtering by previous link hash and tags at the same time", func(t *testing.T) {
		ctx := context.Background()
		prevLinkHash := linkHash4.String()
		slice, err := a.FindSegments(ctx, &store.SegmentFilter{
			Pagination: store.Pagination{
				Limit: segmentsTotalCount,
			},
			PrevLinkHash: &prevLinkHash,
			MapIDs:       []string{"map1", "map2"},
		})
		verifyResultsCount(t, err, slice, 1)
	})

	t.Run("Returns no result when filtering on good previous link hash but invalid map ID", func(t *testing.T) {
		ctx := context.Background()
		prevLinkHash := linkHash4.String()
		slice, err := a.FindSegments(ctx, &store.SegmentFilter{
			Pagination: store.Pagination{
				Limit: segmentsTotalCount,
			},
			PrevLinkHash: &prevLinkHash,
			MapIDs:       []string{"map2"},
		})
		verifyResultsCount(t, err, slice, 0)
	})

	t.Run("Returns no result for previous link hash not found", func(t *testing.T) {
		ctx := context.Background()
		notFoundPrevLinkHash := testutil.RandomHash().String()
		slice, err := a.FindSegments(ctx, &store.SegmentFilter{
			Pagination: store.Pagination{
				Limit: segmentsTotalCount,
			},
			PrevLinkHash: &notFoundPrevLinkHash,
		})
		verifyResultsCount(t, err, slice, 0)
	})

	t.Run("Supports filtering by process", func(t *testing.T) {
		ctx := context.Background()
		slice, err := a.FindSegments(ctx, &store.SegmentFilter{
			Pagination: store.Pagination{
				Limit: segmentsTotalCount,
			},
			Process: "Foo",
		})
		verifyResultsCount(t, err, slice, 2)
	})

	t.Run("Returns no result for process not found", func(t *testing.T) {
		ctx := context.Background()
		slice, err := a.FindSegments(ctx, &store.SegmentFilter{
			Pagination: store.Pagination{
				Limit: segmentsTotalCount,
			},
			Process: "Bar",
		})
		verifyResultsCount(t, err, slice, 0)
	})

	t.Run("Returns its evidences", func(t *testing.T) {
		ctx := context.Background()
		e1 := cs.Evidence{Backend: "dummy", Provider: "1", Proof: &cs.GenericProof{}}
		e2 := cs.Evidence{Backend: "batch", Provider: "2", Proof: &batchevidences.BatchProof{}}
		e3 := cs.Evidence{Backend: "bcbatch", Provider: "3", Proof: &bcbatchevidences.BcBatchProof{}}
		e4 := cs.Evidence{Backend: "generic", Provider: "4"}
		testEvidences := []cs.Evidence{e1, e2, e3, e4}

		for _, e := range testEvidences {
			err := a.AddEvidence(ctx, linkHash4, &e)
			assert.NoError(t, err, "a.AddEvidence()")
		}

		got, err := a.FindSegments(ctx, &store.SegmentFilter{
			Pagination: store.Pagination{
				Limit: segmentsTotalCount,
			},
			LinkHashes: []string{
				linkHash4.String(),
			},
		})
		assert.NoError(t, err, "a.FindSegments()")
		assert.NotNil(t, got)
		require.Len(t, got, 1)
		assert.True(t, len(got[0].Meta.Evidences) >= 4)
	})

}

// BenchmarkFindSegments benchmarks finding segments.
func (f Factory) BenchmarkFindSegments(b *testing.B, numLinks int, createLinkFunc CreateLinkFunc, filterFunc FilterFunc) {
	a := f.initAdapterB(b)
	defer f.freeAdapter(a)

	for i := 0; i < numLinks; i++ {
		a.CreateLink(context.Background(), createLinkFunc(b, numLinks, i))
	}

	filters := make([]*store.SegmentFilter, b.N)
	for i := 0; i < b.N; i++ {
		filters[i] = filterFunc(b, numLinks, i)
	}

	b.ResetTimer()
	log.SetOutput(ioutil.Discard)

	for i := 0; i < b.N; i++ {
		if s, err := a.FindSegments(context.Background(), filters[i]); err != nil {
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
		a.CreateLink(context.Background(), createLinkFunc(b, numLinks, i))
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
			if s, err := a.FindSegments(context.Background(), filters[i]); err != nil {
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
