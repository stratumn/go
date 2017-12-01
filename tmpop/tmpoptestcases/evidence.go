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
	"github.com/stratumn/sdk/cs/evidences"
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

		proof := evidence.Proof.(*evidences.TendermintProof)
		if proof == nil {
			t.Fatalf("h.Commit(): expected proof not to be nil")
		}
		if proof.BlockHeight != 1 {
			t.Fatalf("Invalid block height in proof: want %d, got %d",
				1, proof.BlockHeight)
		}
		if got, want := proof.Time(), proof.Header.GetTime(); got != want {
			t.Fatalf("Invalid time in proof: want %d, got %d", want, got)
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
