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

package tmpoptestcases

import (
	"testing"

	abci "github.com/tendermint/abci/types"

	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/cs/cstesting"
	"github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/tmpop"
	"github.com/stratumn/go-indigocore/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestQuery tests each query request type implemented by TMPop
func (f Factory) TestQuery(t *testing.T) {
	h, req := f.newTMPop(t, nil)
	defer f.free()

	link1, req := commitRandomLink(t, h, req)

	invalidLink := cstesting.InvalidLinkWithProcess(link1.GetProcess())
	invalidLinkHash, _ := invalidLink.Hash()
	req = commitLink(t, h, invalidLink, req)

	link2 := cstesting.RandomLinkWithProcess(link1.GetProcess())
	linkHash2, _ := link2.Hash()
	commitLink(t, h, link2, req)

	t.Run("Info() returns correct last seen height and app hash", func(t *testing.T) {
		abciInfo := h.Info(abci.RequestInfo{})
		assert.Equal(t, int64(3), abciInfo.LastBlockHeight)
	})

	t.Run("GetInfo() correctly returns name", func(t *testing.T) {
		info := &tmpop.Info{}
		err := makeQuery(h, tmpop.GetInfo, nil, info)
		assert.NoError(t, err)
		assert.EqualValues(t, tmpop.Name, info.Name)
	})

	t.Run("AddEvidence() adds an external evidence", func(t *testing.T) {
		evidence := &cs.Evidence{Backend: "dummy", Provider: "1"}
		evidenceRequest := &struct {
			LinkHash *types.Bytes32
			Evidence *cs.Evidence
		}{
			linkHash2,
			evidence,
		}
		err := makeQuery(h, tmpop.AddEvidence, evidenceRequest, nil)
		assert.NoError(t, err)

		got := &cs.Segment{}
		err = makeQuery(h, tmpop.GetSegment, linkHash2, got)
		assert.NoError(t, err)
		require.Len(t, got.Meta.Evidences, 1, "Segment should have an evidence added")

		storedEvidence := got.Meta.GetEvidence("1")
		assert.True(t, storedEvidence.Backend == evidence.Backend && storedEvidence.Provider == evidence.Provider)
	})

	t.Run("GetSegment()", func(t *testing.T) {
		verifyLinkStored(t, h, link2)
	})

	t.Run("FindSegments()", func(t *testing.T) {
		wantedPrevLinkHashStr := link2.GetPrevLinkHashString()
		args := &store.SegmentFilter{
			Pagination: store.Pagination{
				Limit: store.DefaultLimit,
			},
			MapIDs:       []string{link2.GetMapID()},
			PrevLinkHash: &wantedPrevLinkHashStr,
			Tags:         link2.GetTags(),
		}
		gots := cs.SegmentSlice{}
		err := makeQuery(h, tmpop.FindSegments, args, &gots)
		assert.NoError(t, err)
		require.Len(t, gots, 1, "Unexpected number of segments")

		got := gots[0]
		assert.EqualValues(t, link2, &got.Link)
	})

	t.Run("FindSegments() skips invalid links", func(t *testing.T) {
		args := &store.SegmentFilter{
			Pagination: store.Pagination{
				Limit: store.DefaultLimit,
			},
			Process: link1.GetProcess(),
		}
		gots := cs.SegmentSlice{}
		err := makeQuery(h, tmpop.FindSegments, args, &gots)
		assert.NoError(t, err)
		assert.Len(t, gots, 2, "Unexpected number of segments")

		for _, segment := range gots {
			assert.NotEqual(t, *invalidLinkHash, *segment.GetLinkHash(),
				"Invalid segment found in FindSegments")
		}
	})

	t.Run("GetMapIDs()", func(t *testing.T) {
		args := &store.MapFilter{
			Pagination: store.Pagination{
				Limit: store.DefaultLimit,
			},
		}

		var got []string
		err := makeQuery(h, tmpop.GetMapIDs, args, &got)
		assert.NoError(t, err)
		assert.Len(t, got, 2, "Unexpected number of maps")

		mapIdsFound := make(map[string]bool)
		for _, mapID := range got {
			mapIdsFound[mapID] = true
		}

		for _, mapID := range []string{link1.GetMapID(), link2.GetMapID()} {
			_, found := mapIdsFound[mapID]
			assert.True(t, found, "Couldn't find map id %s", mapID)
		}
	})

	t.Run("Pending events are delivered only once", func(t *testing.T) {
		var events []*store.Event
		err := makeQuery(h, tmpop.PendingEvents, nil, &events)
		assert.NoError(t, err)
		assert.Len(t, events, 2, "We should have two saved links events (no evidence since Tendermint Core is not connected)")

		err = makeQuery(h, tmpop.PendingEvents, nil, &events)
		assert.NoError(t, err)
		assert.Len(t, events, 0, "Events should not be delivered twice")
	})

	t.Run("Unsupported Query", func(t *testing.T) {
		q := h.Query(abci.RequestQuery{
			Path: "Unsupported",
		})
		assert.EqualValues(t, tmpop.CodeTypeNotImplemented, q.GetCode())

		q = h.Query(abci.RequestQuery{
			Path:   tmpop.FindSegments,
			Height: 12,
		})
		assert.EqualValues(t, tmpop.CodeTypeInternalError, q.GetCode())
	})
}
