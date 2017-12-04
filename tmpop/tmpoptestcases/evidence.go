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
	"github.com/stratumn/sdk/store"
	"github.com/stratumn/sdk/tmpop"
	"github.com/stratumn/sdk/tmpop/tmpoptestcases/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestTendermintEvidence tests that evidence is correctly added.
func (f Factory) TestTendermintEvidence(t *testing.T) {
	h, req := f.newTMPop(t, nil)
	defer f.free()

	tmClientMock := new(tmpoptestcasesmocks.MockedTendermintClient)
	tmClientMock.On("FireEvent", mock.Anything, mock.Anything)

	h.ConnectTendermint(tmClientMock)

	invalidLink := cstesting.RandomLink()
	invalidLink.Meta["mapId"] = nil
	invalidLinkHash, _ := invalidLink.Hash()
	req = commitLink(t, h, invalidLink, req)
	tmClientMock.On("Block", 1).Return(&tmpop.Block{})

	link1 := cstesting.RandomLink()
	req = commitLink(t, h, link1, req)
	linkHash1, _ := link1.Hash()
	expectedTx := &tmpop.Tx{TxType: tmpop.CreateLink, Link: link1}
	expectedBlock := &tmpop.Block{Txs: []*tmpop.Tx{expectedTx}}
	tmClientMock.On("Block", 2).Return(expectedBlock)

	link2, req := commitRandomLink(t, h, req)
	linkHash2, _ := link2.Hash()

	t.Run("Adds evidence when block is signed", func(t *testing.T) {
		got := &cs.Segment{}
		err := makeQuery(h, tmpop.GetSegment, linkHash1, got)
		assert.NoError(t, err)

		evidence := got.Meta.GetEvidence(h.GetCurrentHeader().GetChainId())
		assert.NotNil(t, evidence, "Evidence is missing")

		proof := evidence.Proof.(*evidences.TendermintProof)
		assert.NotNil(t, proof, "h.Commit(): expected proof not to be nil")
		assert.Equal(t, uint64(2), proof.BlockHeight, "Invalid block height in proof")

		evidenceEventFired := false
		for _, tmClientCall := range tmClientMock.Calls {
			if tmClientCall.Method != "FireEvent" {
				continue
			}

			storeEvent := tmClientCall.Arguments.Get(1).(tmpop.StoreEventsData).StoreEvents[0]
			if storeEvent.EventType == store.SavedEvidence {
				evidenceEventFired = true
			}
		}
		assert.True(t, evidenceEventFired, "Missing evidence event")
	})

	t.Run("Does not add evidence right after commit", func(t *testing.T) {
		got := &cs.Segment{}
		err := makeQuery(h, tmpop.GetSegment, linkHash2, got)
		assert.NoError(t, err)
		assert.Zero(t, len(got.Meta.Evidences),
			"Link should not have evidence before the next block signs the committed state")
	})

	// Test that if an invalid link was added to a block (which can happen
	// if validations change between the checkTx and deliverTx messages),
	// we don't generate evidence for it.
	t.Run("Does not add evidence to invalid links", func(t *testing.T) {
		got := &cs.Segment{}
		err := makeQuery(h, tmpop.GetSegment, invalidLinkHash, got)
		assert.NoError(t, err)
		assert.Zero(t, len(got.Meta.Evidences), "Evidence should not be added to invalid link")
	})
}
