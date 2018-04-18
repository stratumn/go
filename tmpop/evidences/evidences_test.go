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

package evidences_test

import (
	"crypto/sha256"
	"encoding/json"
	"math/rand"
	"testing"
	"time"

	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/testutil"
	"github.com/stratumn/go-indigocore/tmpop/evidences"
	"github.com/stratumn/go-indigocore/types"
	"github.com/stratumn/merkle"
	mktypes "github.com/stratumn/merkle/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	crypto "github.com/tendermint/go-crypto"
	tmtypes "github.com/tendermint/tendermint/types"
)

var validators []*tmtypes.Validator
var validatorsPrivKeys map[*tmtypes.Validator]crypto.PrivKeyEd25519

func init() {
	validators = make([]*tmtypes.Validator, 3)
	validatorsPrivKeys = make(map[*tmtypes.Validator]crypto.PrivKeyEd25519)

	sk1 := crypto.GenPrivKeyEd25519()
	pk1 := sk1.PubKey()
	v1 := &tmtypes.Validator{
		Address:     pk1.Address(),
		PubKey:      pk1,
		VotingPower: 30,
	}

	sk2 := crypto.GenPrivKeyEd25519()
	pk2 := sk2.PubKey()
	v2 := &tmtypes.Validator{
		Address:     pk2.Address(),
		PubKey:      pk2,
		VotingPower: 10,
	}

	sk3 := crypto.GenPrivKeyEd25519()
	pk3 := sk3.PubKey()
	v3 := &tmtypes.Validator{
		Address:     pk3.Address(),
		PubKey:      pk3,
		VotingPower: 20,
	}

	validators[0] = v1
	validatorsPrivKeys[v1] = sk1

	validators[1] = v2
	validatorsPrivKeys[v2] = sk2

	validators[2] = v3
	validatorsPrivKeys[v3] = sk3
}

func TestTendermintProof(t *testing.T) {
	for _, tt := range []struct {
		name string
		test func(*testing.T)
	}{{
		"deserialize",
		func(t *testing.T) {
			jsonTestEvidences := `[{
				"backend": "TMPop",
				"provider": "testTendermintChain",
				"proof": {
					"block_height": 42,
					"merkle_root":
					  "59655543574b735862477952414f6d4254764b534a666a7a614c62745a73794d",
					"merkle_path": [],
					"validations_hash":
					  "4e7377594e734752757373566d616f7a465a4273624f4a694651475a736e7754",
					"header": {
					  "chain_id": "testchain",
					  "height": 42,
					  "time": "1970-01-01T01:00:42+01:00",
					  "num_txs": 1,
					  "last_block_id": {
					    "hash":
					      "4765754474527A514D44516959434F6867484F7667536579634A504A48594E75",
					    "parts": { "total": 0, "hash": "" }
					  },
					  "total_txs": 1,
					  "last_commit_hash": "",
					  "data_hash": "",
					  "validators_hash": "15F8466100A26952601BF93E18D738AC2278DB5F",
					  "consensus_hash": "",
					  "app_hash":
					    "4B536D566F69474C4F7062554F7045644B7570644F4D6552566A61527A4C4E54",
					  "last_results_hash": "",
					  "evidence_hash": ""
					},
					"header_votes": [],
					"header_validator_set": {
					  "validators": [],
					  "proposer": null
					},
					"next_header": {
					  "chain_id": "testchain",
					  "height": 43,
					  "time": "1970-01-01T01:00:43+01:00",
					  "num_txs": 0,
					  "last_block_id": {
					    "hash": "C74CD8C8FF218EBB95EBF35F03A8F21E1BBFDD30",
					    "parts": { "total": 0, "hash": "" }
					  },
					  "total_txs": 0,
					  "last_commit_hash": "",
					  "data_hash": "",
					  "validators_hash": "15F8466100A26952601BF93E18D738AC2278DB5F",
					  "consensus_hash": "",
					  "app_hash":
					    "0394E1FDEF366159793C5544577EF7207C6FFFEE7F157FB91E2A9807FB9D3F1D",
					  "last_results_hash": "",
					  "evidence_hash": ""
					},
					"next_header_votes": [],
					"next_header_validator_set": {
					  "validators": [],
					  "proposer": null
					}
				}
			}]`
			proofs := cs.Evidences{}
			require.NoError(t, json.Unmarshal([]byte(jsonTestEvidences), &proofs))
			require.Len(t, proofs, 1)
			assert.Equal(t, "testTendermintChain", proofs[0].Provider)
			assert.NotNil(t, proofs[0].Proof)
			assert.Equal(t, "testchain", proofs[0].Proof.(*evidences.TendermintProof).Header.ChainID)
		},
	}, {
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

			e.HeaderVotes[0].PubKey = e.HeaderVotes[1].PubKey
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
	}, {
		"missing-validator-set",
		func(t *testing.T) {
			linkHash, e := CreateTendermintProof(t, 3)
			e.NextHeaderValidatorSet = nil

			assert.False(t, e.Verify(linkHash), "Proof should not be correct if validator set is missing")
		},
	}, {
		"invalid-validator-set",
		func(t *testing.T) {
			linkHash, e := CreateTendermintProof(t, 5)
			e.HeaderValidatorSet = &tmtypes.ValidatorSet{
				Validators: []*tmtypes.Validator{
					validators[1],
					validators[2],
				},
			}

			assert.False(t, e.Verify(linkHash), "Proof should not be correct if validator set doesn't match header's ValidatorHash")
		},
	}, {
		"validator-minority",
		func(t *testing.T) {
			linkHash, e := CreateTendermintProof(t, 1)
			// If we remove the vote from validator 3, there's less than 2/3 of the voting power.
			e.HeaderVotes = e.HeaderVotes[:2]

			assert.False(t, e.Verify(linkHash), "Proof should not be correct if voting power is less than 2/3")
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

	validatorSet := &tmtypes.ValidatorSet{Validators: validators}
	validatorsHash := validatorSet.Hash()

	header := &tmtypes.Header{
		AppHash:        appHash[:],
		ChainID:        "testchain",
		Height:         42,
		LastBlockID:    tmtypes.BlockID{Hash: testutil.RandomHash()[:]},
		NumTxs:         int64(linksCount),
		Time:           time.Unix(42, 0),
		TotalTxs:       int64(linksCount),
		ValidatorsHash: validatorsHash,
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
		ValidatorsHash: validatorsHash,
	}

	e := &evidences.TendermintProof{
		BlockHeight:            42,
		Root:                   types.NewBytes32FromBytes(tree.Root()),
		Path:                   merklePath,
		ValidationsHash:        validationsHash,
		Header:                 header,
		HeaderVotes:            vote(header),
		HeaderValidatorSet:     validatorSet,
		NextHeader:             nextHeader,
		NextHeaderVotes:        vote(nextHeader),
		NextHeaderValidatorSet: validatorSet,
	}

	return linkHash, e
}

// createMerkleTree creates linksCount random links and builds
// a merkle tree from it. It also returns the merkle path for
// the chosen link.
func createMerkleTree(linksCount int) (*types.Bytes32, *merkle.StaticTree, mktypes.Path) {
	position := rand.Intn(linksCount)
	linkHash := testutil.RandomHash()

	treeLeaves := make([][]byte, linksCount)
	for i := 0; i < linksCount; i++ {
		if i == position {
			treeLeaves[i] = linkHash[:]
		} else {
			treeLeaves[i] = testutil.RandomHash()[:]
		}
	}

	tree, _ := merkle.NewStaticTree(treeLeaves)

	return linkHash, tree, tree.Path(position)
}

// vote creates a valid vote for a given header.
// It simulates nodes signing a header and is crucial for the proof.
func vote(header *tmtypes.Header) []*evidences.TendermintVote {
	votes := make([]*evidences.TendermintVote, len(validators))
	for i := 0; i < len(validators); i++ {
		validator := validators[i]
		privKey := validatorsPrivKeys[validator]

		v := &evidences.TendermintVote{
			PubKey: &(validator.PubKey),
			Vote: &tmtypes.Vote{
				BlockID:          tmtypes.BlockID{Hash: header.Hash()},
				Height:           header.Height,
				ValidatorAddress: validator.Address,
				ValidatorIndex:   i,
			},
		}

		sig := privKey.Sign(v.Vote.SignBytes(header.ChainID))
		v.Vote.Signature = sig

		votes[i] = v
	}

	return votes
}
