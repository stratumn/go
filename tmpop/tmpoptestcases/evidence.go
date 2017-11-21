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
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/tmpop"
)

// TestEvidence tests if the evidence is correctly inserted and updated on segments
func (f Factory) TestEvidence(t *testing.T) {
	h := f.initTMPop(t, nil)
	defer f.free()
	s1 := commitMockTx(t, h)

	got := &cs.Segment{}
	err := makeQuery(h, tmpop.GetSegment, s1.GetLinkHash(), got)
	if err != nil {
		t.Fatal(err)
	}

	evidence := got.Meta.GetEvidence(h.GetHeader().GetChainId())

	proof := evidence.Proof.(*tmpop.TendermintFullProof)

	if proof == nil {
		t.Fatalf("h.Commit(): expected original proof not to be nil")
	}
	if proof.Original.BlockHeight != height {
		t.Errorf("h.Commit(): Expected originalEvidence.BlockHeight to contain %v, got %v", height, proof.Original.BlockHeight)
	}

	gotState, wantState := evidence.State, cs.PendingEvidence
	if strings.Compare(gotState, wantState) != 0 {
		t.Errorf("h.Commit(): Expected state to be %s since the next block has not been commited, got %s", wantState, gotState)
	}

	// Create a new Block that confirms the AppHash
	commitMockTx(t, h)

	err = makeQuery(h, tmpop.GetSegment, s1.GetLinkHash(), got)
	if err != nil {
		t.Fatal(err)
	}

	evidence = got.Meta.GetEvidence(h.GetHeader().GetChainId())

	gotState, wantState = evidence.State, cs.CompleteEvidence
	if strings.Compare(gotState, wantState) != 0 {
		t.Errorf("h.Commit(): Expected state to be %s since the next block has been commited, got %s", wantState, gotState)

	}
	if !evidence.Proof.Verify(s1.GetLinkHash()) {
		t.Errorf("TendermintProof.Verify(): Expected proof %v to be valid", evidence.Proof.FullProof())

	}
}

// TestTendermintProof tests the format and the validity of a tendermint proof
func (f Factory) TestTendermintProof(t *testing.T) {
	h := f.initTMPop(t, nil)
	defer f.free()

	t.Run("TestTime()", func(t *testing.T) {
		s := commitMockTx(t, h)

		queried := &cs.Segment{}
		err := makeQuery(h, tmpop.GetSegment, s.GetLinkHash(), queried)
		if err != nil {
			t.Fatal(err)
		}

		e := queried.Meta.GetEvidence(h.GetHeader().GetChainId())
		got := e.Proof.Time()
		if got != 0 {
			t.Errorf("TendermintProof.Time(): Expected timestamp to be %d, got %d", 0, got)
		}

		commitMockTx(t, h)
		err = makeQuery(h, tmpop.GetSegment, s.GetLinkHash(), queried)
		if err != nil {
			t.Fatal(err)
		}

		e = queried.Meta.GetEvidence(h.GetHeader().GetChainId())
		want := h.GetHeader().GetTime()
		got = e.Proof.Time()
		if got != want {
			t.Errorf("TendermintProof.Time(): Expected timestamp to be %d, got %d", want, got)
		}

	})

	t.Run("TestFullProof()", func(t *testing.T) {
		s := commitMockTx(t, h)

		queried := &cs.Segment{}
		err := makeQuery(h, tmpop.GetSegment, s.GetLinkHash(), queried)
		if err != nil {
			t.Fatal(err)
		}

		e := queried.Meta.GetEvidence(h.GetHeader().GetChainId())
		got := e.Proof.FullProof()
		if got == nil {
			t.Errorf("TendermintProof.FullProof(): Expected proof to be a json-formatted bytes array, got %v", got)
		}

		commitMockTx(t, h)
		err = makeQuery(h, tmpop.GetSegment, s.GetLinkHash(), queried)
		if err != nil {
			t.Fatal(err)
		}

		e = queried.Meta.GetEvidence(h.GetHeader().GetChainId())
		wantDifferent := got
		got = e.Proof.FullProof()
		if got == nil {
			t.Errorf("TendermintProof.FullProof(): Expected proof to be a json-formatted bytes array, got %v", got)
		}
		if bytes.Compare(got, wantDifferent) == 0 {
			t.Errorf("TendermintProof.FullProof(): Expected proof after appHash validation to be complete, got %s", string(got))
		}
		if err := json.Unmarshal(got, &tmpop.TendermintProof{}); err != nil {
			t.Errorf("TendermintProof.FullProof(): Could not unmarshal bytes proof, err = %+v", err)
		}

	})

	t.Run("TestVerify()", func(t *testing.T) {
		s := commitMockTx(t, h)

		queried := &cs.Segment{}
		err := makeQuery(h, tmpop.GetSegment, s.GetLinkHash(), queried)
		if err != nil {
			t.Fatal(err)
		}

		e := queried.Meta.GetEvidence(h.GetHeader().GetChainId())
		got := e.Proof.Verify(s.GetLinkHash())
		if got == true {
			t.Errorf("TendermintProof.Verify(): Expected incomplete original proof to be false, got %v", got)
		}

		commitMockTx(t, h)
		if err = makeQuery(h, tmpop.GetSegment, s.GetLinkHash(), queried); err != nil {
			t.Fatal(err)
		}

		e = queried.Meta.GetEvidence(h.GetHeader().GetChainId())

		if got = e.Proof.Verify(s.GetLinkHash()); got != true {
			t.Errorf("TendermintProof.Verify(): Expected original proof to be true, got %v", got)
		}

	})
}
