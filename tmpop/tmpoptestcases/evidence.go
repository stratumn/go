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

	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/cs/cstesting"
	"github.com/stratumn/sdk/tmpop"
)

// TestTendermintEvidence tests that evidence is correctly added.
func (f Factory) TestTendermintEvidence(t *testing.T) {
	h, req := f.newTMPop(t, nil)
	defer f.free()

	invalidLink := cstesting.RandomLink()
	invalidLink.Meta["mapId"] = nil
	invalidLinkHash, _ := invalidLink.Hash()
	req = commitLink(t, h, invalidLink, req)

	link1, req := commitRandomLink(t, h, req)
	linkHash1, _ := link1.Hash()
	link2, req := commitRandomLink(t, h, req)
	linkHash2, _ := link2.Hash()

	t.Run("Adds evidence when block is signed", func(t *testing.T) {
		got := &cs.Segment{}
		err := makeQuery(h, tmpop.GetSegment, linkHash1, got)
		if err != nil {
			t.Fatal(err)
		}

		evidence := got.Meta.GetEvidence(h.GetCurrentHeader().GetChainId())
		if evidence == nil {
			t.Fatalf("Evidence is missing")
		}

		proof := evidence.Proof.(*tmpop.TendermintFullProof)
		if proof == nil {
			t.Fatalf("h.Commit(): expected original proof not to be nil")
		}
		if !proof.Verify(linkHash1) {
			t.Errorf("TendermintProof.Verify(): Expected proof %v to be valid", proof.FullProof())
		}
	})

	t.Run("Does not add evidence right after commit", func(t *testing.T) {
		got := &cs.Segment{}
		err := makeQuery(h, tmpop.GetSegment, linkHash2, got)
		if err != nil {
			t.Fatal(err)
		}

		if len(got.Meta.Evidences) != 0 {
			t.Errorf("Link should not have evidence before the next block signs the committed state")
		}
	})

	// Test that if an invalid link was added to a block (which can happen
	// if validations change between the checkTx and deliverTx messages),
	// we don't generate evidence for it.
	t.Run("Does not add evidence to invalid links", func(t *testing.T) {
		got := &cs.Segment{}
		err := makeQuery(h, tmpop.GetSegment, invalidLinkHash, got)
		if err != nil {
			t.Fatal(err)
		}

		if len(got.Meta.Evidences) != 0 {
			t.Errorf("Evidence should not be added to invalid link")
		}
	})
}

// TestTendermintProof tests the format and the validity of a tendermint proof.
func (f Factory) TestTendermintProof(t *testing.T) {
	// TODO: when the format of tendermint proof is updated, update those tests
	t.Fail()

	// h := f.initTMPop(t, nil)
	// defer f.free()

	// t.Run("TestTime()", func(t *testing.T) {
	// 	s := commitMockTx(t, h)

	// 	queried := &cs.Segment{}
	// 	err := makeQuery(h, tmpop.GetSegment, s.GetLinkHash(), queried)
	// 	if err != nil {
	// 		t.Fatal(err)
	// 	}

	// 	e := queried.Meta.GetEvidence(h.GetHeader().GetChainId())
	// 	got := e.Proof.Time()
	// 	if got != 0 {
	// 		t.Errorf("TendermintProof.Time(): Expected timestamp to be %d, got %d", 0, got)
	// 	}

	// 	commitMockTx(t, h)
	// 	err = makeQuery(h, tmpop.GetSegment, s.GetLinkHash(), queried)
	// 	if err != nil {
	// 		t.Fatal(err)
	// 	}

	// 	e = queried.Meta.GetEvidence(h.GetHeader().GetChainId())
	// 	want := h.GetHeader().GetTime()
	// 	got = e.Proof.Time()
	// 	if got != want {
	// 		t.Errorf("TendermintProof.Time(): Expected timestamp to be %d, got %d", want, got)
	// 	}

	// })

	// t.Run("TestFullProof()", func(t *testing.T) {
	// 	s := commitMockTx(t, h)

	// 	queried := &cs.Segment{}
	// 	err := makeQuery(h, tmpop.GetSegment, s.GetLinkHash(), queried)
	// 	if err != nil {
	// 		t.Fatal(err)
	// 	}

	// 	e := queried.Meta.GetEvidence(h.GetHeader().GetChainId())
	// 	got := e.Proof.FullProof()
	// 	if got == nil {
	// 		t.Errorf("TendermintProof.FullProof(): Expected proof to be a json-formatted bytes array, got %v", got)
	// 	}

	// 	commitMockTx(t, h)
	// 	err = makeQuery(h, tmpop.GetSegment, s.GetLinkHash(), queried)
	// 	if err != nil {
	// 		t.Fatal(err)
	// 	}

	// 	e = queried.Meta.GetEvidence(h.GetHeader().GetChainId())
	// 	wantDifferent := got
	// 	got = e.Proof.FullProof()
	// 	if got == nil {
	// 		t.Errorf("TendermintProof.FullProof(): Expected proof to be a json-formatted bytes array, got %v", got)
	// 	}
	// 	if bytes.Compare(got, wantDifferent) == 0 {
	// 		t.Errorf("TendermintProof.FullProof(): Expected proof after appHash validation to be complete, got %s", string(got))
	// 	}
	// 	if err := json.Unmarshal(got, &tmpop.TendermintProof{}); err != nil {
	// 		t.Errorf("TendermintProof.FullProof(): Could not unmarshal bytes proof, err = %+v", err)
	// 	}

	// })

	// t.Run("TestVerify()", func(t *testing.T) {
	// 	s := commitMockTx(t, h)

	// 	queried := &cs.Segment{}
	// 	err := makeQuery(h, tmpop.GetSegment, s.GetLinkHash(), queried)
	// 	if err != nil {
	// 		t.Fatal(err)
	// 	}

	// 	e := queried.Meta.GetEvidence(h.GetHeader().GetChainId())
	// 	got := e.Proof.Verify(s.GetLinkHash())
	// 	if got == true {
	// 		t.Errorf("TendermintProof.Verify(): Expected incomplete original proof to be false, got %v", got)
	// 	}

	// 	commitMockTx(t, h)
	// 	if err = makeQuery(h, tmpop.GetSegment, s.GetLinkHash(), queried); err != nil {
	// 		t.Fatal(err)
	// 	}

	// 	e = queried.Meta.GetEvidence(h.GetHeader().GetChainId())

	// 	if got = e.Proof.Verify(s.GetLinkHash()); got != true {
	// 		t.Errorf("TendermintProof.Verify(): Expected original proof to be true, got %v", got)
	// 	}

	// })
}
