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
	"encoding/json"
	"reflect"
	"testing"

	abci "github.com/tendermint/abci/types"

	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/cs/cstesting"
	"github.com/stratumn/sdk/store"
	"github.com/stratumn/sdk/tmpop"
	"github.com/stratumn/sdk/types"
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
	req = commitLink(t, h, link2, req)

	t.Run("Info() returns correct last seen height and app hash", func(t *testing.T) {
		abciInfo := h.Info(abci.RequestInfo{})
		if abciInfo.LastBlockHeight != 3 {
			t.Errorf("Invalid LastBlockHeight: expected %d, got %d",
				3, abciInfo.LastBlockHeight)
		}
	})

	t.Run("GetInfo() correctly returns name", func(t *testing.T) {
		info := &tmpop.Info{}
		err := makeQuery(h, tmpop.GetInfo, nil, info)
		if err != nil {
			t.Fatal(err)
		}

		if info.Name != tmpop.Name {
			t.Errorf("h.Query(): expected GetInfo to return name %v, got %v", info.Name, tmpop.Name)
		}
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
		if err := makeQuery(h, tmpop.AddEvidence, evidenceRequest, nil); err != nil {
			t.Fatal(err)
		}

		got := &cs.Segment{}
		if err := makeQuery(h, tmpop.GetSegment, linkHash2, got); err != nil {
			t.Fatal(err)
		}

		if len(got.Meta.Evidences) != 1 {
			t.Fatalf("Segment should have an evidence added")
		}

		storedEvidence := got.Meta.GetEvidence("1")
		if storedEvidence.Backend != evidence.Backend || storedEvidence.Provider != evidence.Provider {
			t.Errorf("Unexpected evidence stored: got %v, want %v",
				storedEvidence, evidence)
		}
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
		if err != nil {
			t.Fatal(err)
		}
		if len(gots) != 1 {
			t.Fatalf("h.Query(): unexpected size for FindSegments result, got %v", gots)
		}

		got := gots[0]
		if want, got := *link2, got.Link; !reflect.DeepEqual(want, got) {
			gotJS, _ := json.MarshalIndent(got, "", "  ")
			wantJS, _ := json.MarshalIndent(want, "", "  ")
			t.Errorf("h.Query(): expected FindSegments to return %s, got %s", wantJS, gotJS)
		}
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
		if err != nil {
			t.Fatal(err)
		}
		if len(gots) != 2 {
			t.Fatalf("h.Query(): unexpected size for FindSegments result, got %v", gots)
		}

		for _, segment := range gots {
			if segment.GetLinkHash().Equals(invalidLinkHash) {
				t.Errorf("Invalid segment found in FindSegments")
			}
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
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 2 {
			t.Errorf("Invalid number of maps found: expected %d, got %d",
				2, len(got))
		}

		mapIdsFound := make(map[string]bool)
		for _, mapID := range got {
			mapIdsFound[mapID] = true
		}

		for _, mapID := range []string{link1.GetMapID(), link2.GetMapID()} {
			if _, found := mapIdsFound[mapID]; found == false {
				t.Errorf("Couldn't find map id %s", mapID)
			}
		}
	})

	t.Run("Unsupported Query", func(t *testing.T) {
		q := h.Query(abci.RequestQuery{
			Path: "Unsupported",
		})
		if got, want := q.GetCode(), abci.CodeType_UnknownRequest; got != want {
			t.Errorf("h.Query(): expected unsupported query to return %v, got %v", want, got)
		}

		q = h.Query(abci.RequestQuery{
			Path:   tmpop.FindSegments,
			Height: 12,
		})
		if got, want := q.GetCode(), abci.CodeType_InternalError; got != want {
			t.Errorf("h.Query(): expected unsupported query to return %v, got %v", want, got)
		}
	})
}
