// Copyright 2016-2018 Stratumn SAS. All rights reserved.
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
	"sort"
	"sync/atomic"
	"testing"

	"github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-chainscript/chainscripttest"
	"github.com/stratumn/go-core/store"
	"github.com/stratumn/go-core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var emptyPrevLinkHash = ""
var priorities = sort.Float64Slice{}

func createLink(t *testing.T, adapter store.Adapter, link *chainscript.Link, prepareLink func(l *chainscript.Link)) *chainscript.Link {
	if prepareLink != nil {
		prepareLink(link)
	}

	_, err := adapter.CreateLink(context.Background(), link)
	require.NoError(t, err, "adapter.CreateLink()")

	priorities = append(priorities, link.Meta.Priority)

	return link
}

func createRandomLink(t *testing.T, adapter store.Adapter, prepareLink func(l *chainscript.Link)) *chainscript.Link {
	return createLink(t, adapter, chainscripttest.RandomLink(t), prepareLink)
}

func createLinkBranch(t *testing.T, adapter store.Adapter, parent *chainscript.Link, prepareLink func(l *chainscript.Link)) *chainscript.Link {
	return createLink(t, adapter, chainscripttest.NewLinkBuilder(t).Branch(t, parent).Build(), prepareLink)
}

func verifyPriorities(t *testing.T, segments *types.PaginatedSegments, offset int, reverse bool) {
	sort.Sort(sort.Reverse(priorities))
	want, op := 100.0, "<="
	f := func(got, want float64) func() bool { return func() bool { return got <= want } }
	if reverse {
		sort.Sort(priorities)
		want, op = -100.0, ">="
		f = func(got, want float64) func() bool { return func() bool { return got >= want } }
	}
	for i, s := range segments.Segments {
		assert.Equal(t, priorities[i+offset], s.Link.Meta.Priority)
		got := s.Link.Meta.Priority
		assert.Conditionf(t, f(got, want), "slice#%d: priority = %v want %s %v", i, got, op, want)
		want = got
	}
}

func verifyPriorityOrdering(t *testing.T, segments *types.PaginatedSegments, offset int) {
	verifyPriorities(t, segments, offset, false)
}

func verifyReversePriorityOrdering(t *testing.T, segments *types.PaginatedSegments, offset int) {
	verifyPriorities(t, segments, offset, true)
}

func verifyResultsCountWithTotalCount(t *testing.T, err error, segments *types.PaginatedSegments, expectedCount, expectedTotalCount int) {
	require.NoError(t, err)
	require.NotNil(t, segments)

	require.Len(t, segments.Segments, expectedCount, "Invalid number of results")
	assert.Equal(t, expectedTotalCount, segments.TotalCount, "Invalid number of results before pagination")
	assert.Conditionf(t, func() bool { return len(segments.Segments) <= segments.TotalCount }, "Invalid total count of results. got: %d / expected less than %d", len(segments.Segments), segments.TotalCount)
}

func verifyResultsCount(t *testing.T, err error, segments *types.PaginatedSegments, expectedCount int) {
	verifyResultsCountWithTotalCount(t, err, segments, expectedCount, expectedCount)
}

// TestFindSegments tests what happens when you search for segments with various filters.
func (f Factory) TestFindSegments(t *testing.T) {
	a := f.initAdapter(t)
	defer f.freeAdapter(a)

	// Setup a test fixture with segments matching different types of filters

	testPageSize := 3
	segmentsTotalCount := 8

	createRandomLink(t, a, func(l *chainscript.Link) {
		l.Meta.MapId = "map1"
		l.Meta.Process.Name = "Foo"
		l.Meta.Step = "propose"
	})

	createRandomLink(t, a, func(l *chainscript.Link) {
		l.Meta.Tags = []string{"tag1", "tag42"}
		l.Meta.MapId = "map2"
	})

	createRandomLink(t, a, func(l *chainscript.Link) {
		l.Meta.Tags = []string{"tag2"}
	})

	link4 := createRandomLink(t, a, nil)
	linkHash4, _ := link4.Hash()

	createLinkBranch(t, a, link4, func(l *chainscript.Link) {
		l.Meta.Tags = []string{"tag1", chainscripttest.RandomString(5)}
		l.Meta.Step = "propose"
	})

	link6 := createRandomLink(t, a, func(l *chainscript.Link) {
		l.Meta.Tags = []string{"tag2", "tag42"}
		l.Meta.Process.Name = "Foo"
	})
	linkHash6, _ := link6.Hash()

	createRandomLink(t, a, func(l *chainscript.Link) {
		l.Meta.MapId = "map2"
	})

	createLinkBranch(t, a, link4, nil)

	t.Run("Should order by priority", func(t *testing.T) {
		ctx := context.Background()
		segments, err := a.FindSegments(ctx, &store.SegmentFilter{
			Pagination: store.Pagination{
				Limit: testPageSize,
			},
		})
		verifyResultsCountWithTotalCount(t, err, segments, testPageSize, segmentsTotalCount)
		verifyPriorityOrdering(t, segments, 0)
	})

	t.Run("Should reverse order by priority", func(t *testing.T) {
		ctx := context.Background()
		segments, err := a.FindSegments(ctx, &store.SegmentFilter{
			Pagination: store.Pagination{
				Limit: testPageSize,
			},
			Reverse: true,
		})
		verifyResultsCountWithTotalCount(t, err, segments, testPageSize, segmentsTotalCount)
		verifyReversePriorityOrdering(t, segments, 0)
	})

	t.Run("Should support pagination", func(t *testing.T) {
		ctx := context.Background()
		segments, err := a.FindSegments(ctx, &store.SegmentFilter{
			Pagination: store.Pagination{
				Offset: testPageSize,
				Limit:  testPageSize,
			},
		})
		verifyResultsCountWithTotalCount(t, err, segments, testPageSize, segmentsTotalCount)
		verifyPriorityOrdering(t, segments, testPageSize)
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

	t.Run("Supports filtering on step", func(t *testing.T) {
		ctx := context.Background()
		slice, err := a.FindSegments(ctx, &store.SegmentFilter{
			Pagination: store.Pagination{
				Limit: segmentsTotalCount,
			},
			Step: "propose",
		})
		verifyResultsCount(t, err, slice, 2)
	})

	t.Run("Supports filtering on step and map ID", func(t *testing.T) {
		ctx := context.Background()
		slice, err := a.FindSegments(ctx, &store.SegmentFilter{
			Pagination: store.Pagination{
				Limit: segmentsTotalCount,
			},
			Process: link4.Meta.Process.Name,
			MapIDs:  []string{link4.Meta.MapId},
			Step:    "propose",
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
		verifyResultsCount(t, err, slice, 1)
	})

	t.Run("Supports filtering on multiple map IDs", func(t *testing.T) {
		ctx := context.Background()
		slice, err := a.FindSegments(ctx, &store.SegmentFilter{
			Pagination: store.Pagination{
				Limit: segmentsTotalCount,
			},
			MapIDs: []string{"map1", "map2"},
		})
		verifyResultsCount(t, err, slice, 3)
	})

	t.Run("Supports filtering on map ID and tag at the same time", func(t *testing.T) {
		ctx := context.Background()
		slice, err := a.FindSegments(ctx, &store.SegmentFilter{
			Pagination: store.Pagination{
				Limit: segmentsTotalCount,
			},
			MapIDs: []string{"map2"},
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
			LinkHashes: []chainscript.LinkHash{
				linkHash4,
				chainscripttest.RandomHash(),
				linkHash6,
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
			LinkHashes: []chainscript.LinkHash{
				linkHash4,
				linkHash6,
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
			LinkHashes: []chainscript.LinkHash{
				chainscripttest.RandomHash(),
				chainscripttest.RandomHash(),
			},
			Pagination: store.Pagination{
				Limit: segmentsTotalCount,
			},
		})
		verifyResultsCount(t, err, slice, 0)
	})

	t.Run("Supports filtering for segments without parents", func(t *testing.T) {
		ctx := context.Background()
		slice, err := a.FindSegments(ctx, &store.SegmentFilter{
			Pagination:    store.Pagination{Limit: segmentsTotalCount},
			WithoutParent: true,
		})
		verifyResultsCount(t, err, slice, 6)
	})

	t.Run("Supports filtering by previous link hash", func(t *testing.T) {
		ctx := context.Background()
		slice, err := a.FindSegments(ctx, &store.SegmentFilter{
			Pagination: store.Pagination{
				Limit: segmentsTotalCount,
			},
			PrevLinkHash: linkHash4,
		})
		verifyResultsCount(t, err, slice, 2)
	})

	t.Run("Supports filtering by previous link hash and tags at the same time", func(t *testing.T) {
		ctx := context.Background()
		slice, err := a.FindSegments(ctx, &store.SegmentFilter{
			Pagination: store.Pagination{
				Limit: segmentsTotalCount,
			},
			PrevLinkHash: linkHash4,
			Tags:         []string{"tag1"},
		})
		verifyResultsCount(t, err, slice, 1)
	})

	t.Run("Supports filtering by previous link hash and mapIDs at the same time", func(t *testing.T) {
		ctx := context.Background()
		slice, err := a.FindSegments(ctx, &store.SegmentFilter{
			Pagination: store.Pagination{
				Limit: segmentsTotalCount,
			},
			PrevLinkHash: linkHash4,
			MapIDs:       []string{link4.Meta.MapId, "map2"},
		})
		verifyResultsCount(t, err, slice, 2)
	})

	t.Run("Returns no result when filtering on good previous link hash but invalid map ID", func(t *testing.T) {
		ctx := context.Background()
		slice, err := a.FindSegments(ctx, &store.SegmentFilter{
			Pagination: store.Pagination{
				Limit: segmentsTotalCount,
			},
			PrevLinkHash: linkHash4,
			MapIDs:       []string{"map2"},
		})
		verifyResultsCount(t, err, slice, 0)
	})

	t.Run("Returns no result for previous link hash not found", func(t *testing.T) {
		ctx := context.Background()
		notFoundPrevLinkHash := chainscripttest.RandomHash()
		slice, err := a.FindSegments(ctx, &store.SegmentFilter{
			Pagination: store.Pagination{
				Limit: segmentsTotalCount,
			},
			PrevLinkHash: notFoundPrevLinkHash,
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
		e1, _ := chainscript.NewEvidence("1.0.0", "dummy", "1", []byte{1})
		e2, _ := chainscript.NewEvidence("1.0.0", "batch", "2", []byte{2})
		e3, _ := chainscript.NewEvidence("1.0.0", "bcbatch", "3", []byte{3})
		e4, _ := chainscript.NewEvidence("1.0.0", "generic", "4", []byte{4})
		testEvidences := []*chainscript.Evidence{e1, e2, e3, e4}

		for _, e := range testEvidences {
			err := a.AddEvidence(ctx, linkHash4, e)
			assert.NoError(t, err, "a.AddEvidence()")
		}

		got, err := a.FindSegments(ctx, &store.SegmentFilter{
			Pagination: store.Pagination{
				Limit: segmentsTotalCount,
			},
			LinkHashes: []chainscript.LinkHash{linkHash4},
		})
		verifyResultsCount(t, err, got, 1)
		assert.True(t, len(got.Segments[0].Meta.Evidences) >= 4)
	})

	t.Run("Resists to SQL injections in process filter", func(t *testing.T) {
		ctx := context.Background()
		slice, err := a.FindSegments(ctx, &store.SegmentFilter{
			Pagination: store.Pagination{
				Limit: segmentsTotalCount,
			},
			Process: "Foo' or 'bar' = 'bar'--",
		})
		verifyResultsCount(t, err, slice, 0)
	})

	t.Run("Resists to SQL injections in mapIDs filter", func(t *testing.T) {
		ctx := context.Background()
		slice, err := a.FindSegments(ctx, &store.SegmentFilter{
			Pagination: store.Pagination{
				Limit: segmentsTotalCount,
			},
			MapIDs: []string{"Foo') or 'bar' = 'bar'--", "plap"},
		})
		verifyResultsCount(t, err, slice, 0)
	})

}

// BenchmarkFindSegments benchmarks finding segments.
func (f Factory) BenchmarkFindSegments(b *testing.B, numLinks int, createLinkFunc CreateLinkFunc, filterFunc FilterFunc) {
	a := f.initAdapterB(b)
	defer f.freeAdapter(a)

	for i := 0; i < numLinks; i++ {
		_, err := a.CreateLink(context.Background(), createLinkFunc(b, numLinks, i))
		if err != nil {
			b.Fatal(err)
		}
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
		} else if s.Segments == nil {
			b.Error("s = nil want types.SegmentSlice")
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
		_, err := a.CreateLink(context.Background(), createLinkFunc(b, numLinks, i))
		if err != nil {
			b.Fatal(err)
		}
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
			} else if s.Segments == nil {
				b.Error("s = nil want types.SegmentSlice")
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
