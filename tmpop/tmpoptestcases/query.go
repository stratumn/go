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
	"github.com/stratumn/sdk/store"
	"github.com/stratumn/sdk/store/storetesting"
	"github.com/stratumn/sdk/tmpop"
)

// TestQuery tests each query request type implemented by TMPop
func (f Factory) TestQuery(t *testing.T) {
	h := f.initTMPop(t, nil)
	defer f.free()

	t.Run("GetInfo()", func(t *testing.T) {
		info := &tmpop.Info{}
		err := makeQuery(h, tmpop.GetInfo, nil, info)
		if err != nil {
			t.Fatal(err)
		}

		if info.Name != tmpop.Name {
			t.Errorf("h.Query(): expected GetInfo to return name %v, got %v", info.Name, tmpop.Name)
		}
	})

	t.Run("GetSegment()", func(t *testing.T) {
		want := commitMockTx(t, h)

		got := &cs.Segment{}
		err := makeQuery(h, tmpop.GetSegment, want.GetLinkHash(), got)
		if err != nil {
			t.Fatal(err)
		}

		got.Meta.Evidences = nil
		if !reflect.DeepEqual(want, got) {
			gotJS, _ := json.MarshalIndent(got, "", "  ")
			wantJS, _ := json.MarshalIndent(want, "", "  ")
			t.Errorf("h.Query(): expected GetSegment to return %s, got:\n %s", wantJS, gotJS)
		}
	})

	t.Run("FindSegments()", func(t *testing.T) {
		want := commitMockTx(t, h)

		wantedPrevLinkHashStr := want.Link.GetPrevLinkHashString()
		args := &store.SegmentFilter{
			Pagination: store.Pagination{
				Limit: store.DefaultLimit,
			},
			MapIDs:       []string{want.Link.GetMapID()},
			PrevLinkHash: &wantedPrevLinkHashStr,
			Tags:         want.Link.GetTags(),
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

		got.Meta.Evidences = nil
		if !reflect.DeepEqual(want, got) {
			gotJS, _ := json.MarshalIndent(got, "", "  ")
			wantJS, _ := json.MarshalIndent(want, "", "  ")
			t.Errorf("h.Query(): expected FindSegments to return %s, got %s", wantJS, gotJS)
		}
	})

	t.Run("GetMapIDs()", func(t *testing.T) {
		a := &storetesting.MockAdapter{}
		f.adapter = nil
		f.New = func() (store.Adapter, error) {
			return a, nil
		}
		h := f.initTMPop(t, nil)
		segment, _ := makeSaveSegmentTx(t)
		mapID := segment.Link.GetMapID()
		limit := 1
		a.MockGetMapIDs.Fn = func(filter *store.MapFilter) ([]string, error) {
			if filter.Pagination.Limit != limit {
				t.Errorf("Expected limit %v, got %v", limit, filter.Pagination.Limit)
			}

			res := []string{mapID}

			return res, nil
		}
		args := &store.MapFilter{
			Pagination: store.Pagination{
				Limit: limit,
			},
		}

		var got []string
		err := makeQuery(h, tmpop.GetMapIDs, args, &got)
		if err != nil {
			t.Fatal(err)
		}
		if mapID != got[0] {
			t.Errorf("h.Query(): expected GetMapIDs to return %v, got %v", mapID, got[0])
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
