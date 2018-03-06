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

package cs_test

import (
	"crypto/sha256"
	"encoding/json"
	"math/rand"
	"testing"
	"time"

	"github.com/stratumn/go-indigocore/cs/evidences"
	"github.com/stratumn/go-indigocore/merkle"
	"github.com/stratumn/go-indigocore/testutil"
	"github.com/stratumn/go-indigocore/types"
	"github.com/stretchr/testify/assert"
	crypto "github.com/tendermint/go-crypto"
	tmtypes "github.com/tendermint/tendermint/types"
)

func TestTendermintProof(t *testing.T) {
	for _, tt := range []struct {
		name string
		test func(*testing.T)
	}{{
		"time",
		func(t *testing.T) {
			e := &evidences.TendermintProof{Header: &tmtypes.Header{Time: time.Unix(42, 0)}}
			assert.Equal(t, uint64(42), e.Time(), "Invalid proof time")
		},
	}, {
		"full-proof",
		func(t *testing.T) {
			_, e := CreateTendermintProof(t, 2)
			fullProof := e.FullProof()
			assert.NoError(t, json.Unmarshal(fullProof, &evidences.TendermintProof{}),
				"Could not unmarshal bytes proof")
		},
	}, {
		"single-link",
		func(t *testing.T) {
			linkHash, e := CreateTendermintProof(t, 1)
			assert.True(t, e.Verify(linkHash), "Proof should be valid")
		},
	}, {
		"validations-hash",
		func(t *testing.T) {
			linkHash, e := CreateTendermintProof(t, 5)
			assert.True(t, e.Verify(linkHash), "Proof should be valid before modification")

			e.ValidationsHash = testutil.RandomHash()
			assert.False(t, e.Verify(linkHash), "Proof should not be correct if validations hash changed")
		},
	}, {
		"merkle-root",
		func(t *testing.T) {
			linkHash, e := CreateTendermintProof(t, 3)
			assert.True(t, e.Verify(linkHash), "Proof should be valid before modification")

			e.Root = linkHash
			assert.False(t, e.Verify(linkHash), "Proof should not be correct if merkle root changed")
		},
	}, {
		"previous-app-hash",
		func(t *testing.T) {
			linkHash, e := CreateTendermintProof(t, 4)
			assert.True(t, e.Verify(linkHash), "Proof should be valid before modification")

			e.Header.AppHash = linkHash[:]
			assert.False(t, e.Verify(linkHash), "Proof should not be correct if previous app hash changed")
		},
	}, {
		"missing-votes",
		func(t *testing.T) {
			linkHash, e := CreateTendermintProof(t, 4)
			assert.True(t, e.Verify(linkHash), "Proof should be valid before modification")

			e.HeaderVotes = nil
			assert.False(t, e.Verify(linkHash), "Proof should not be correct if votes are missing")
		},
	}, {
		"missing-public-key",
		func(t *testing.T) {
			linkHash, e := CreateTendermintProof(t, 5)
			assert.True(t, e.Verify(linkHash), "Proof should be valid before modification")

			e.HeaderVotes[0].PubKey = &crypto.PubKey{}
			assert.False(t, e.Verify(linkHash), "Proof should not be correct if public key is missing")
		},
	}, {
		"public-key-mismatch",
		func(t *testing.T) {
			linkHash, e := CreateTendermintProof(t, 2)
			assert.True(t, e.Verify(linkHash), "Proof should be valid before modification")

			e.HeaderVotes[0].PubKey = e.NextHeaderVotes[0].PubKey
			assert.False(t, e.Verify(linkHash), "Proof should not be correct if public key doesn't match")
		},
	}, {
		"invalid-signature",
		func(t *testing.T) {
			linkHash, e := CreateTendermintProof(t, 3)
			assert.True(t, e.Verify(linkHash), "Proof should be valid before modification")

			e.HeaderVotes[0].Vote.Signature = e.NextHeaderVotes[0].Vote.Signature
			assert.False(t, e.Verify(linkHash), "Proof should not be correct if signature is invalid")
		},
	}, {
		"invalid-next-signature",
		func(t *testing.T) {
			linkHash, e := CreateTendermintProof(t, 3)
			assert.True(t, e.Verify(linkHash), "Proof should be valid before modification")

			e.NextHeaderVotes[0].Vote.Signature = e.HeaderVotes[0].Vote.Signature
			assert.False(t, e.Verify(linkHash), "Proof should not be correct if next signature is invalid")
		},
	}, {
		"header-mismatch",
		func(t *testing.T) {
			linkHash, e := CreateTendermintProof(t, 4)
			assert.True(t, e.Verify(linkHash), "Proof should be valid before modification")

			e.Header.Height += 42
			assert.False(t, e.Verify(linkHash), "Proof should not be correct if header has been modified")
		},
	}, {
		"next-header-mismatch",
		func(t *testing.T) {
			linkHash, e := CreateTendermintProof(t, 4)
			assert.True(t, e.Verify(linkHash), "Proof should be valid before modification")

			e.NextHeader.Height--
			assert.False(t, e.Verify(linkHash), "Proof should not be correct if next header has been modified")
		},
	}, {
		"invalid-multiple-votes",
		func(t *testing.T) {
			linkHash, e := CreateTendermintProof(t, 3)
			assert.True(t, e.Verify(linkHash), "Proof should be valid before modification")

			moreInvalidVotes := vote(e.Header)
			moreInvalidVotes[0].Vote.Height = 0
			moreValidVotes := vote(e.Header)

			e.HeaderVotes = append(e.HeaderVotes, moreInvalidVotes...)
			e.HeaderVotes = append(e.HeaderVotes, moreValidVotes...)

			assert.False(t, e.Verify(linkHash), "Proof should not be correct if next header has been modified")
		},
	}} {
		t.Run(tt.name, tt.test)
	}
}

// CreateTendermintProof creates a valid Tendermint proof.
// It creates linksCount random links to include in a block,
// generates a valid block and its proof, and returns the link
// and the evidence.
func CreateTendermintProof(t *testing.T, linksCount int) (*types.Bytes32, *evidences.TendermintProof) {
	validationsHash := testutil.RandomHash()
	appHash := testutil.RandomHash()
	linkHash, tree, merklePath := createMerkleTree(linksCount)

	header := &tmtypes.Header{
		AppHash:        appHash[:],
		ChainID:        "testchain",
		Height:         42,
		LastBlockID:    tmtypes.BlockID{Hash: testutil.RandomHash()[:]},
		NumTxs:         int64(linksCount),
		Time:           time.Unix(42, 0),
		TotalTxs:       int64(linksCount),
		ValidatorsHash: testutil.RandomHash()[:],
	}

	hash := sha256.New()
	hash.Write(appHash[:])
	hash.Write(validationsHash[:])
	hash.Write(tree.Root()[:])
	nextAppHash := hash.Sum(nil)

	nextHeader := &tmtypes.Header{
		AppHash:        nextAppHash,
		ChainID:        "testchain",
		Height:         43,
		LastBlockID:    tmtypes.BlockID{Hash: header.Hash()},
		Time:           time.Unix(43, 0),
		ValidatorsHash: testutil.RandomHash()[:],
	}

	e := &evidences.TendermintProof{
		BlockHeight:     42,
		Root:            tree.Root(),
		Path:            merklePath,
		ValidationsHash: validationsHash,
		Header:          header,
		HeaderVotes:     vote(header),
		NextHeader:      nextHeader,
		NextHeaderVotes: vote(nextHeader),
	}

	return linkHash, e
}

// createMerkleTree creates linksCount random links and builds
// a merkle tree from it. It also returns the merkle path for
// the chosen link.
func createMerkleTree(linksCount int) (*types.Bytes32, *merkle.StaticTree, types.Path) {
	position := rand.Intn(linksCount)
	linkHash := testutil.RandomHash()

	treeLeaves := make([]types.Bytes32, linksCount)
	for i := 0; i < linksCount; i++ {
		if i == position {
			treeLeaves[i] = *linkHash
		} else {
			treeLeaves[i] = *testutil.RandomHash()
		}
	}

	tree, _ := merkle.NewStaticTree(treeLeaves)

	return linkHash, tree, tree.Path(position)
}

// vote creates a valid vote for a given header.
// It simulates nodes signing a header and is crucial for the proof.
func vote(header *tmtypes.Header) []*evidences.TendermintVote {
	privKey := crypto.GenPrivKeyEd25519()
	pubKey := privKey.PubKey()

	v := &evidences.TendermintVote{
		PubKey: &pubKey,
		Vote: &tmtypes.Vote{
			BlockID:          tmtypes.BlockID{Hash: header.Hash()},
			Height:           header.Height,
			ValidatorAddress: pubKey.Address(),
		},
	}

	sig := privKey.Sign(tmtypes.SignBytes(header.ChainID, v.Vote))
	v.Vote.Signature = sig

	return []*evidences.TendermintVote{v}
}
