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

	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/cs/cstesting"
)

// TestCheckTx tests what happens when the ABCI method CheckTx() is called
func (f Factory) TestCheckTx(t *testing.T) {
	h := f.initTMPop(t, nil)
	defer f.free()

	_, tx := makeSaveSegmentTx(t)

	res := h.CheckTx(tx)

	if !res.IsOK() {
		t.Errorf("Expected CheckTx to return an OK result, got %v", res)
	}
}

// TestTx tests each transaction type processed by doTx()
func (f Factory) TestTx(t *testing.T) {
	h := f.initTMPop(t, nil)
	defer f.free()

	t.Run("WriteSaveSegment()", func(t *testing.T) {
		want := commitMockTx(t, h)

		got, err := f.adapter.GetSegment(want.GetLinkHash())
		if err != nil {
			t.Fatal(err)
		}

		ev := got.Meta.GetEvidence(h.GetHeader().GetChainId())
		got.Meta.Evidences = nil
		if !reflect.DeepEqual(want, got) {
			gotJS, _ := json.MarshalIndent(got, "", "  ")
			wantJS, _ := json.MarshalIndent(want, "", "  ")
			t.Errorf("h.Commit(): expected to return %s, got %s", wantJS, gotJS)
		}
		got.Meta.AddEvidence(*ev)
	})

	t.Run("WriteDoubleSaveSegment()", func(t *testing.T) {
		s := cstesting.RandomSegment()
		tx := makeSaveSegmentTxFromSegment(t, s)
		h.BeginBlock(requestBeginBlock)

		h.DeliverTx(tx)
		res := h.Commit()

		got, err := f.adapter.GetSegment(s.GetLinkHash())

		if err != nil {
			t.Fatal(err)
		}
		ev := got.Meta.GetEvidence(h.GetHeader().GetChainId())
		if ev.State != cs.PendingEvidence {
			t.Errorf("h.DeliverTx(): wrong evidence state after saving segment, got %s, want %s", ev.State, cs.PendingEvidence)
		}

		// We try to save a segment with the same link (and linkHash)
		// but with new evidences
		newRequest := requestBeginBlock
		newRequest.Header.AppHash = res.Data.Bytes()

		h.BeginBlock(newRequest)
		s.Meta.Evidences = nil
		s.Meta.AddEvidence(cs.Evidence{
			Provider: "test1",
			Backend:  "TMPop",
		})
		s.Meta.AddEvidence(cs.Evidence{
			Provider: "test2",
			Backend:  "TMPop",
			Proof:    nil,
		})
		tx = makeSaveSegmentTxFromSegment(t, s)
		h.DeliverTx(tx)
		h.Commit()

		got, err = f.adapter.GetSegment(s.GetLinkHash())

		if err != nil {
			t.Fatal(err)
		}
		ev = got.Meta.GetEvidence(h.GetHeader().GetChainId())
		if ev.State != cs.CompleteEvidence {
			t.Errorf("h.DeliverTx(): wrong evidence state after saving segment, got %s, want %s", ev.State, cs.CompleteEvidence)
		}
		if len(got.Meta.Evidences) != 3 {
			t.Errorf("h.DeliverTx(): wrong length of segment.Meta.Evidences, got %d want %d", len(got.Meta.Evidences), 3)
		}
	})

	t.Run("WriteDeleteSegment", func(t *testing.T) {
		segment := commitMockTx(t, h)

		tx := makeDeleteSegmentTx(t, segment)
		commitTx(t, h, tx)

		got, err := f.adapter.GetSegment(segment.GetLinkHash())

		if err != nil {
			t.Fatal(err)
		}

		if got != nil {
			t.Errorf("h.GetSegment(): expected to return nil, got %s", got)
		}
	})
}
