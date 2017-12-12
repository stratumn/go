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
	"github.com/stratumn/sdk/merkle"
	"github.com/stratumn/sdk/store"
	"github.com/stratumn/sdk/tmpop"
	"github.com/stratumn/sdk/tmpop/tmpoptestcases/mocks"
	"github.com/stratumn/sdk/types"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/abci/types"
)

// TestTendermintEvidence tests that evidence is correctly added.
func (f Factory) TestTendermintEvidence(t *testing.T) {
	h, req := f.newTMPop(t, nil)
	defer f.free()

	tmClientMock := new(tmpoptestcasesmocks.MockedTendermintClient)
	h.ConnectTendermint(tmClientMock)

	// First block contains an invalid link
	invalidLink := cstesting.RandomLink()
	invalidLink.Meta["mapId"] = nil
	invalidLinkHash, _ := invalidLink.Hash()
	req = commitLink(t, h, invalidLink, req)
	previousAppHash := req.Header.AppHash
	tmClientMock.On("Block", 1).Return(&tmpop.Block{
		Header: &abci.Header{Height: uint64(1)},
	})

	// Second block contains two valid links
	link1 := cstesting.RandomLink()
	linkHash1, _ := link1.Hash()

	link2 := cstesting.RandomLink()
	linkHash2, _ := link2.Hash()

	req = commitTxs(t, h, req, [][]byte{makeCreateLinkTx(t, link1), makeCreateLinkTx(t, link2)})
	appHash := req.Header.AppHash

	expectedTx1 := &tmpop.Tx{TxType: tmpop.CreateLink, Link: link1}
	expectedTx2 := &tmpop.Tx{TxType: tmpop.CreateLink, Link: link2}
	expectedBlock := &tmpop.Block{
		Header: &abci.Header{
			Height:  uint64(2),
			AppHash: previousAppHash,
		},
		Txs: []*tmpop.Tx{expectedTx1, expectedTx2},
	}
	tmClientMock.On("Block", 2).Return(expectedBlock)

	// Third block contains one valid link
	link3, req := commitRandomLink(t, h, req)
	linkHash3, _ := link3.Hash()

	t.Run("Adds evidence when block is signed", func(t *testing.T) {
		got := &cs.Segment{}
		err := makeQuery(h, tmpop.GetSegment, linkHash1, got)
		assert.NoError(t, err)

		evidence := got.Meta.GetEvidence(h.GetCurrentHeader().GetChainId())
		assert.NotNil(t, evidence, "Evidence is missing")

		proof := evidence.Proof.(*evidences.TendermintProof)
		assert.NotNil(t, proof, "h.Commit(): expected proof not to be nil")
		assert.Equal(t, uint64(2), proof.BlockHeight, "Invalid block height in proof")

		tree, _ := merkle.NewStaticTree([]types.Bytes32{*linkHash1, *linkHash2})
		assert.EqualValues(t, tree.Root(), proof.Root, "Invalid proof merkle root")
		assert.EqualValues(t, tree.Path(0), proof.Path, "Invalid proof merkle path")

		expectedAppHash, _ := tmpop.ComputeAppHash(
			types.NewBytes32FromBytes(previousAppHash),
			types.NewBytes32FromBytes(nil),
			tree.Root())
		assert.EqualValues(t, expectedAppHash[:], appHash, "Invalid app hash generated")
	})

	t.Run("Creates evidence events when block is signed", func(t *testing.T) {
		var events []*store.Event
		err := makeQuery(h, tmpop.PendingEvents, nil, &events)
		assert.NoError(t, err)

		var evidenceEvents []*store.Event
		for _, event := range events {
			if event.EventType == store.SavedEvidences {
				evidenceEvents = append(evidenceEvents, event)
			}
		}

		assert.Equal(t, 1, len(evidenceEvents), "Invalid number of events")
		savedEvidences := evidenceEvents[0].Data.(map[string]*cs.Evidence)
		assert.Equal(t, 2, len(savedEvidences), "Invalid number of evidence produced")
		assert.NotNil(t, savedEvidences[linkHash1.String()], "Evidence missing for %x", *linkHash1)
		assert.NotNil(t, savedEvidences[linkHash2.String()], "Evidence missing for %x", *linkHash2)
	})

	t.Run("Does not add evidence right after commit", func(t *testing.T) {
		got := &cs.Segment{}
		err := makeQuery(h, tmpop.GetSegment, linkHash3, got)
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
		assert.Zero(t, got.Link, "Link should not be found")
		assert.Zero(t, len(got.Meta.Evidences), "Evidence should not be added to invalid link")
	})
}
