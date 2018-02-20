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

package agenttestcases

import (
	"testing"

	"github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/testutil"
	"github.com/stretchr/testify/assert"
)

// TestFindSegmentsOK tests the client's ability to handle a FindSegment request.
func (f Factory) TestFindSegmentsOK(t *testing.T) {
	process := "test"
	expected := 20
	for i := 0; i < expected; i++ {
		f.Client.CreateMap(process, nil, "test")
	}

	filter := store.SegmentFilter{
		Process: process,
		Pagination: store.Pagination{
			Limit: expected,
		},
	}
	sgmts, err := f.Client.FindSegments(&filter)
	assert.NoError(t, err)
	assert.NotNil(t, sgmts)
	assert.Equal(t, expected, len(sgmts))
}

// TestFindSegmentsTags tests the client's ability to handle a FindSegment request
// when tags are set in the filter.
func (f Factory) TestFindSegmentsTags(t *testing.T) {
	process, tag := "test", "tag"
	f.Client.CreateMap(process, nil, tag)

	filter := store.SegmentFilter{
		Process: process,
		Tags:    []string{tag},
		Pagination: store.Pagination{
			Limit: 20,
		},
	}
	found, err := f.Client.FindSegments(&filter)
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.True(t, len(found) > 0)
	for _, s := range found {
		assert.Equal(t, []string{tag}, s.Link.Meta.Tags)
	}
}

// TestFindSegmentsLinkHashes tests the client's ability to handle a FindSegment request
// when LinkHashes are set in the filter.
func (f Factory) TestFindSegmentsLinkHashes(t *testing.T) {
	process := "test"
	parent, _ := f.Client.CreateMap(process, nil, "test")

	filter := store.SegmentFilter{
		Process:    process,
		LinkHashes: []string{parent.Meta.GetLinkHashString()},
		Pagination: store.Pagination{
			Limit: 20,
		},
	}
	found, err := f.Client.FindSegments(&filter)
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, 1, len(found))
	assert.Equal(t, parent.Meta.LinkHash, found[0].Meta.LinkHash)
}

// TestFindSegmentsMapIDs tests the client's ability to handle a FindSegment request
// when a map ID is set in the filter.
func (f Factory) TestFindSegmentsMapIDs(t *testing.T) {
	process := "test"
	parent, _ := f.Client.CreateMap(process, nil, "test")

	filter := store.SegmentFilter{
		Process: process,
		MapIDs:  []string{parent.Link.Meta.MapID},
		Pagination: store.Pagination{
			Limit: 20,
		},
	}
	found, err := f.Client.FindSegments(&filter)
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, 1, len(found))
	assert.Equal(t, parent.Meta.LinkHash, found[0].Meta.LinkHash)
}

// TestFindSegmentsLimit tests the client's ability to handle a FindSegment request
// when a limit is set in the filter, and when this liit is set to -1
func (f Factory) TestFindSegmentsLimit(t *testing.T) {
	process := "test"
	created := 30
	for i := 0; i < created; i++ {
		f.Client.CreateMap(process, nil, "test")
	}

	t.Run("With a limit", func(t *testing.T) {
		limit := 5
		filter := store.SegmentFilter{
			Process: process,
			Pagination: store.Pagination{
				Limit: limit,
			},
		}
		sgmts, err := f.Client.FindSegments(&filter)
		assert.NoError(t, err)
		assert.NotNil(t, sgmts)
		assert.Equal(t, limit, len(sgmts))
	})

	t.Run("Without a limit", func(t *testing.T) {
		limit := -1
		filter := store.SegmentFilter{
			Process: process,
			Pagination: store.Pagination{
				Limit: limit,
			},
		}
		sgmts, err := f.Client.FindSegments(&filter)
		assert.NoError(t, err)
		assert.NotNil(t, sgmts)
		assert.True(t, len(sgmts) > created)
	})

}

// TestFindSegmentsNoMatch tests the client's ability to handle a FindSegment request
// when no segment is found.
func (f Factory) TestFindSegmentsNoMatch(t *testing.T) {
	process := "wrong"
	filter := store.SegmentFilter{
		Process: process,
	}
	sgmts, err := f.Client.FindSegments(&filter)
	assert.EqualError(t, err, "process 'wrong' does not exist")
	assert.Nil(t, sgmts)
}

// TestFindSegmentsNotFound tests the client's ability to handle a FindSegment request
// when no segment is found.
func (f Factory) TestFindSegmentsNotFound(t *testing.T) {
	process, prevLinkHash := "test", testutil.RandomHash().String()
	filter := store.SegmentFilter{
		Process:      process,
		PrevLinkHash: &prevLinkHash,
	}
	sgmts, err := f.Client.FindSegments(&filter)
	assert.NoError(t, err)
	assert.Len(t, sgmts, 0)
}
