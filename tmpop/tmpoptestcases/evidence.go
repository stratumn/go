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
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/cs/cstesting"
	"github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/testutil"
	"github.com/stratumn/go-indigocore/tmpop"
	"github.com/stratumn/go-indigocore/tmpop/evidences"
	"github.com/stratumn/go-indigocore/tmpop/tmpoptestcases/mocks"
	"github.com/stratumn/go-indigocore/types"
	"github.com/stratumn/merkle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	crypto "github.com/tendermint/go-crypto"
	tmtypes "github.com/tendermint/tendermint/types"
)

var tmvalidator *tmtypes.Validator
var validatorPrivKey crypto.PrivKeyEd25519

func init() {
	validatorPrivKey = crypto.GenPrivKeyEd25519()
	tmvalidator = &tmtypes.Validator{
		Address:     validatorPrivKey.PubKey().Address(),
		PubKey:      validatorPrivKey.PubKey(),
		VotingPower: 42,
	}
}

// TestTendermintEvidence tests that evidence is correctly added.
func (f Factory) TestTendermintEvidence(t *testing.T) {
	h, req := f.newTMPop(t, nil)
	defer f.free()

	ctrl := gomock.NewController(t)
	tmClientMock := tmpoptestcasesmocks.NewMockTendermintClient(ctrl)
	h.ConnectTendermint(tmClientMock)

	// We setup a blockchain containing the following blocks:
	//  * Block1 contains one invalid link -> no evidence should be generated
	//  * Block2 contains two valid links -> evidence should be generated at block 5
	//  * Block3 contains one valid link -> evidence should be generated at block 6
	//  * Block4 contains one invalid link -> no evidence should be generated
	//  * Block5 contains one valid link -> votes will be missing for block6 so evidence should not be generated
	//  * Block6 contains one valid link -> votes will be missing for this block so evidence should not be generated
	//  * Block7 contains one valid link
	//  * Block8 contains one valid link -> Tendermint Core error so no evidence should be generated
	//  * Block9 contains one valid link -> but an invalid app hash so no evidence should be generated
	//  * Block10 contains one valid link -> no evidence should be generated because we stop before block 13
	//  * Block11 contains one valid link -> no evidence should be generated because we stop before block 13
	//  * Block12 contains one valid link -> no evidence should be generated because we stop before block 13

	validatorSet := &tmtypes.ValidatorSet{Validators: []*tmtypes.Validator{tmvalidator}}
	validatorsHash := validatorSet.Hash()

	appHashes := make([][]byte, 13)
	appHashes[0] = req.Header.AppHash

	createHeader := func(height int64) *tmtypes.Header {
		var appHash []byte
		if height > 0 {
			appHash = appHashes[height-1]
		}
		return &tmtypes.Header{
			AppHash:        appHash,
			ChainID:        req.Header.ChainID,
			Height:         height,
			Time:           time.Unix(42, 42),
			ValidatorsHash: validatorsHash,
		}
	}

	blocks := make([]*tmpop.Block, 13)
	blocks[0] = &tmpop.Block{Header: createHeader(0)}

	// Block1
	invalidLink1 := cstesting.NewLinkBuilder().Invalid().Build()
	invalidLinkHash1, _ := invalidLink1.Hash()
	req = commitLink(t, h, invalidLink1, req)

	appHashes[1] = req.Header.AppHash
	blocks[1] = &tmpop.Block{
		Header:     createHeader(1),
		Txs:        []*tmpop.Tx{&tmpop.Tx{TxType: tmpop.CreateLink, Link: invalidLink1}},
		Validators: validatorSet,
	}
	tmClientMock.EXPECT().Block(gomock.Any(), int64(1)).Return(blocks[1], nil).AnyTimes()

	// Block2
	link1 := cstesting.RandomLink()
	linkHash1, _ := link1.Hash()
	link2 := cstesting.RandomLink()
	linkHash2, _ := link2.Hash()
	req = commitTxs(t, h, req, [][]byte{makeCreateLinkTx(t, link1), makeCreateLinkTx(t, link2)})

	appHashes[2] = req.Header.AppHash
	blocks[2] = &tmpop.Block{
		Header: createHeader(2),
		Txs: []*tmpop.Tx{
			&tmpop.Tx{TxType: tmpop.CreateLink, Link: link1},
			&tmpop.Tx{TxType: tmpop.CreateLink, Link: link2},
		},
		Validators: validatorSet,
		Votes:      vote(blocks[1].Header),
	}
	tmClientMock.EXPECT().Block(gomock.Any(), int64(2)).Return(blocks[2], nil).AnyTimes()

	// Block3
	link3, req := commitRandomLink(t, h, req)
	linkHash3, _ := link3.Hash()

	appHashes[3] = req.Header.AppHash
	blocks[3] = &tmpop.Block{
		Header:     createHeader(3),
		Txs:        []*tmpop.Tx{&tmpop.Tx{TxType: tmpop.CreateLink, Link: link3}},
		Validators: validatorSet,
		Votes:      vote(blocks[2].Header),
	}
	tmClientMock.EXPECT().Block(gomock.Any(), int64(3)).Return(blocks[3], nil).AnyTimes()

	// Block4
	invalidLink2 := cstesting.NewLinkBuilder().Invalid().Build()
	req = commitLink(t, h, invalidLink2, req)

	appHashes[4] = req.Header.AppHash
	blocks[4] = &tmpop.Block{
		Header:     createHeader(4),
		Txs:        []*tmpop.Tx{&tmpop.Tx{TxType: tmpop.CreateLink, Link: invalidLink2}},
		Validators: validatorSet,
		Votes:      vote(blocks[3].Header),
	}
	tmClientMock.EXPECT().Block(gomock.Any(), int64(4)).Return(blocks[4], nil).AnyTimes()

	// Block5
	link4, req := commitRandomLink(t, h, req)
	linkHash4, _ := link4.Hash()

	appHashes[5] = req.Header.AppHash
	blocks[5] = &tmpop.Block{
		Header:     createHeader(5),
		Txs:        []*tmpop.Tx{&tmpop.Tx{TxType: tmpop.CreateLink, Link: link4}},
		Validators: validatorSet,
		Votes:      vote(blocks[4].Header),
	}
	tmClientMock.EXPECT().Block(gomock.Any(), int64(5)).Return(blocks[5], nil).AnyTimes()

	// Block6
	link5, req := commitRandomLink(t, h, req)

	appHashes[6] = req.Header.AppHash
	blocks[6] = &tmpop.Block{
		Header:     createHeader(6),
		Txs:        []*tmpop.Tx{&tmpop.Tx{TxType: tmpop.CreateLink, Link: link5}},
		Validators: validatorSet,
		Votes:      vote(blocks[5].Header),
	}
	tmClientMock.EXPECT().Block(gomock.Any(), int64(6)).Return(blocks[6], nil).AnyTimes()

	// Block7: missing votes for the previous block
	link6, req := commitRandomLink(t, h, req)

	appHashes[7] = req.Header.AppHash
	blocks[7] = &tmpop.Block{
		Header:     createHeader(7),
		Txs:        []*tmpop.Tx{&tmpop.Tx{TxType: tmpop.CreateLink, Link: link6}},
		Validators: validatorSet,
		// No votes here: should not generate evidence
	}
	tmClientMock.EXPECT().Block(gomock.Any(), int64(7)).Return(blocks[7], nil).AnyTimes()

	// Block8: error from Tendermint Core
	link7, req := commitRandomLink(t, h, req)
	linkHash7, _ := link7.Hash()

	// Invalid app hash to prevent next block from producing valid proofs.
	appHashes[8] = testutil.RandomHash()[:]
	blocks[8] = &tmpop.Block{
		Header:     createHeader(8),
		Txs:        []*tmpop.Tx{&tmpop.Tx{TxType: tmpop.CreateLink, Link: link7}},
		Validators: validatorSet,
		Votes:      vote(blocks[7].Header),
	}
	tmClientMock.EXPECT().Block(gomock.Any(), int64(8)).Return(nil, errors.New("internal error")).AnyTimes()

	// Block9
	link8, req := commitRandomLink(t, h, req)
	linkHash8, _ := link8.Hash()

	appHashes[9] = req.Header.AppHash
	blocks[9] = &tmpop.Block{
		Header:     createHeader(9),
		Txs:        []*tmpop.Tx{&tmpop.Tx{TxType: tmpop.CreateLink, Link: link8}},
		Validators: validatorSet,
		Votes:      vote(blocks[8].Header),
	}
	tmClientMock.EXPECT().Block(gomock.Any(), int64(9)).Return(blocks[9], nil).AnyTimes()

	// Block10
	link9, req := commitRandomLink(t, h, req)
	linkHash9, _ := link9.Hash()

	appHashes[10] = req.Header.AppHash
	blocks[10] = &tmpop.Block{
		Header:     createHeader(10),
		Txs:        []*tmpop.Tx{&tmpop.Tx{TxType: tmpop.CreateLink, Link: link9}},
		Validators: validatorSet,
		Votes:      vote(blocks[9].Header),
	}
	tmClientMock.EXPECT().Block(gomock.Any(), int64(10)).Return(blocks[10], nil).AnyTimes()

	// Block11
	link10, req := commitRandomLink(t, h, req)

	appHashes[11] = req.Header.AppHash
	blocks[11] = &tmpop.Block{
		Header:     createHeader(11),
		Txs:        []*tmpop.Tx{&tmpop.Tx{TxType: tmpop.CreateLink, Link: link10}},
		Validators: validatorSet,
		Votes:      vote(blocks[10].Header),
	}
	tmClientMock.EXPECT().Block(gomock.Any(), int64(11)).Return(blocks[11], nil).AnyTimes()

	// Block12
	link11, req := commitRandomLink(t, h, req)

	appHashes[12] = req.Header.AppHash
	blocks[12] = &tmpop.Block{
		Header:     createHeader(12),
		Txs:        []*tmpop.Tx{&tmpop.Tx{TxType: tmpop.CreateLink, Link: link11}},
		Validators: validatorSet,
		Votes:      vote(blocks[11].Header),
	}
	tmClientMock.EXPECT().Block(gomock.Any(), int64(12)).Return(blocks[12], nil).AnyTimes()

	t.Run("Adds evidence when block is properly signed", func(t *testing.T) {
		got := &cs.Segment{}
		err := makeQuery(h, tmpop.GetSegment, linkHash2, got)
		assert.NoError(t, err)

		evidence := got.Meta.GetEvidence(h.GetCurrentHeader().GetChainID())
		require.NotNil(t, evidence, "Evidence is missing")

		proof := evidence.Proof.(*evidences.TendermintProof)
		assert.NotNil(t, proof, "h.Commit(): expected proof not to be nil")
		assert.Equal(t, int64(2), proof.BlockHeight, "Invalid block height in proof")

		tree, _ := merkle.NewStaticTree([][]byte{linkHash1[:], linkHash2[:]})
		assert.EqualValues(t, tree.Root(), proof.Root[:], "Invalid proof merkle root")
		assert.EqualValues(t, tree.Path(0), proof.Path[:], "Invalid proof merkle path")

		expectedAppHash, _ := tmpop.ComputeAppHash(
			types.NewBytes32FromBytes(appHashes[1]),
			types.NewBytes32FromBytes(nil),
			types.NewBytes32FromBytes(tree.Root()))
		assert.EqualValues(t, expectedAppHash[:], appHashes[2], "Invalid app hash generated")

		assert.True(t, proof.Verify(linkHash2), "Proof should verify")
	})

	t.Run("Creates evidence events to notify store", func(t *testing.T) {
		var events []*store.Event
		err := makeQuery(h, tmpop.PendingEvents, nil, &events)
		assert.NoError(t, err)

		var evidenceEvents []*store.Event
		for _, event := range events {
			if event.EventType == store.SavedEvidences {
				evidenceEvents = append(evidenceEvents, event)
			}
		}

		require.Len(t, evidenceEvents, 2, "Invalid number of events")

		savedEvidences := evidenceEvents[0].Data.(map[string]*cs.Evidence)
		assert.Len(t, savedEvidences, 2, "Invalid number of evidence produced")
		assert.NotNil(t, savedEvidences[linkHash1.String()], "Evidence missing for %x", *linkHash1)
		assert.NotNil(t, savedEvidences[linkHash2.String()], "Evidence missing for %x", *linkHash2)

		savedEvidences = evidenceEvents[1].Data.(map[string]*cs.Evidence)
		assert.Len(t, savedEvidences, 1, "Invalid number of evidence produced")
		assert.NotNil(t, savedEvidences[linkHash3.String()], "Evidence missing for %x", *linkHash3)
	})

	t.Run("Does not add evidence right after commit", func(t *testing.T) {
		got := &cs.Segment{}
		err := makeQuery(h, tmpop.GetSegment, linkHash9, got)
		assert.NoError(t, err)
		assert.Empty(t, got.Meta.Evidences, "Link should not have evidence before the next block is signed")
	})

	// It is possible to add invalid links to a block.
	// It can happen if validation rules change between
	// the checkTx and deliverTx messages.
	// It's ok to have such links in the blockchain, but
	// we should not generate evidence for them.
	t.Run("Does not add evidence to invalid links", func(t *testing.T) {
		got := &cs.Segment{}
		err := makeQuery(h, tmpop.GetSegment, invalidLinkHash1, got)
		assert.NoError(t, err)
		assert.Zero(t, got.Link, "Link should not be found")
		assert.Empty(t, got.Meta.Evidences, "Evidence should not be added to invalid link")
	})

	t.Run("Does not add evidence if signatures are missing", func(t *testing.T) {
		got := &cs.Segment{}
		err := makeQuery(h, tmpop.GetSegment, linkHash4, got)
		assert.NoError(t, err)
		assert.Empty(
			t,
			got.Meta.Evidences,
			"Link should not have evidence if signatures are missing",
		)
	})

	t.Run("Does not add evidence in case of Tendermint Core error", func(t *testing.T) {
		got := &cs.Segment{}
		err := makeQuery(h, tmpop.GetSegment, linkHash7, got)
		assert.NoError(t, err)
		assert.Empty(
			t,
			got.Meta.Evidences,
			"Link should not have evidence in case of Tendermint Core error",
		)
	})

	t.Run("Does not add evidence if app hash doesn't match", func(t *testing.T) {
		got := &cs.Segment{}
		err := makeQuery(h, tmpop.GetSegment, linkHash8, got)
		assert.NoError(t, err)
		assert.Empty(
			t,
			got.Meta.Evidences,
			"Link should not have evidence in case of app hash mismatch",
		)
	})
}

// vote creates a valid vote for a given header.
// It simulates nodes signing a header and is crucial for the proof.
func vote(header *tmtypes.Header) []*evidences.TendermintVote {
	v := &evidences.TendermintVote{
		PubKey: &tmvalidator.PubKey,
		Vote: &tmtypes.Vote{
			BlockID:          tmtypes.BlockID{Hash: header.Hash()},
			Height:           header.Height,
			ValidatorAddress: tmvalidator.PubKey.Address(),
			ValidatorIndex:   0,
		},
	}

	sig := validatorPrivKey.Sign(v.Vote.SignBytes(header.ChainID))
	v.Vote.Signature = sig

	return []*evidences.TendermintVote{v}
}
