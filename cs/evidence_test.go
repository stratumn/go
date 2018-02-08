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

	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/cs/evidences"
	"github.com/stratumn/go-indigocore/merkle"
	"github.com/stratumn/go-indigocore/testutil"
	"github.com/stratumn/go-indigocore/types"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/abci/types"
)

const (
	height      = int64(1)
	TestChainId = "testChain"
)

var (
	header = &abci.Header{
		Height:  height,
		ChainId: TestChainId,
	}

	TestEvidence = cs.Evidence{
		Provider: TestChainId,
		Backend:  "generic",
		Proof: &cs.GenericProof{
			Timestamp: 1507187163,
			Data:      "data",
			Pubkey:    []byte("random"),
			Signature: []byte("hash"),
		},
	}

	JSONTestWrongEvidences = `
		[{
			"backend": "random",
			"provider": "testChain",
			"proof": {}
		}]`

	JSONTestEvidences = `
		[
		{
			"backend": "generic",
			"provider": "testChain0",
			"proof": {}
		},
		{
			"backend": "dummy",
			"provider": "testChain1",
			"proof": {}
		},
		{
			"backend": "batch",
			"provider": "testChain2",
			"proof": {
				"original": {
					"Timestamp": 1,
					"Root": null,
					"Path": []
				},
				"current": {}
			}

		},
		{
			"backend": "bcbatch",
			"provider": "testChain3",
			"proof": {}
		},
		{
			"backend": "TMPop",
			"provider": "testChain4",
			"proof": {
				"original": {
					"blockHeight": 2,
					"header": {
						"app_hash": "eQg/Pe/PaO3amW1Jkn+vpTH0ldU=",
						"chain_id": "test-chain-D9z7EJ",
						"height": 2744,
						"last_block_id": {
							"hash": "g6/ewNOerE0++Lg3f2xEnSxhNW0=",
							"parts": {
								"hash": "fH/X/N6l+2q3XHPqMsXOM0W4d1I=",
								"total": 1
							}
						}
					},
					"merkleProof": {
						"InnerNodes": [{
							"Height": 1,
							"Left": "",
							"Right": "DWDp+PJUmkoXzXthUq4cIsEyPvw=",
							"Size": 2
						}],
						"LeafHash": "wisCGdiVRAAMNDosjHdlEvPmCo0=",
						"RootHash": "eQg/Pe/PaO3amW1Jkn+vpTH0ldU="
					}

				}
			}
		}
		]
	`
)

func TestSerializeEvidence(t *testing.T) {
	evidences := cs.Evidences{}

	if err := json.Unmarshal([]byte(JSONTestEvidences), &evidences); err != nil {
		t.Error(err)
	}

	if len(evidences) != 5 {
		t.Errorf("Could not parse all the evidences")
	}

	for _, e := range evidences {
		if e.Provider == "" || e.Proof == nil {
			t.Errorf("Could not parse evidence with backend %s", e.Backend)
		}
	}
}

func TestSerializeWrongEvidence(t *testing.T) {
	evidences := cs.Evidences{}

	if err := json.Unmarshal([]byte(JSONTestWrongEvidences), &evidences); err == nil {
		t.Errorf("Should have failed because of unknown type of evidence")
	}

}

func TestGenericProof(t *testing.T) {
	p := TestEvidence.Proof
	t.Run("Time()", func(t *testing.T) {
		got, want := p.Time(), uint64(1507187163)
		if got != want {
			t.Errorf(`Evidence.originalProof.Time() = %d, want %d`, got, want)
		}
	})

	t.Run("FullProof()", func(t *testing.T) {
		got := p.FullProof()
		if err := json.Unmarshal(got, &cs.GenericProof{}); err != nil {
			t.Errorf("Could not unmarshal bytes proof, err = %+v", err)
		}
	})

	t.Run("Verify()", func(t *testing.T) {
		if got, want := p.Verify(""), true; got != want {
			t.Errorf(`Evidence.originalProof.Verify() = %v, want %v`, got, want)
		}
	})
}

func TestTendermintProof(t *testing.T) {
	createValidProof := func(t *testing.T, linksCount int) (*types.Bytes32, *evidences.TendermintProof) {
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

		validationsHash := testutil.RandomHash()
		previousAppHash := testutil.RandomHash()
		tree, _ := merkle.NewStaticTree(treeLeaves)

		hash := sha256.New()
		hash.Write(previousAppHash[:])
		hash.Write(validationsHash[:])
		hash.Write(tree.Root()[:])
		appHash := hash.Sum(nil)

		e := &evidences.TendermintProof{
			BlockHeight:     42,
			Root:            tree.Root(),
			Path:            tree.Path(position),
			ValidationsHash: validationsHash,
			Header:          abci.Header{AppHash: previousAppHash[:]},
			NextHeader:      abci.Header{AppHash: appHash},
		}

		assert.True(t, e.Verify(linkHash), "Proof should be verified")

		return linkHash, e
	}

	t.Run("Time()", func(t *testing.T) {
		e := &evidences.TendermintProof{
			Header: abci.Header{Time: int64(42)},
		}

		assert.Equal(t, uint64(42), e.Time(), "Invalid proof time")
	})

	t.Run("FullProof()", func(t *testing.T) {
		e := &evidences.TendermintProof{
			BlockHeight:     int64(42),
			Header:          abci.Header{Time: int64(42)},
			ValidationsHash: testutil.RandomHash(),
		}

		fullProof := e.FullProof()
		assert.NoError(t, json.Unmarshal(fullProof, &evidences.TendermintProof{}),
			"Could not unmarshal bytes proof")
	})

	t.Run("Verify() succeeds for single element merkle tree", func(t *testing.T) {
		createValidProof(t, 1)
	})

	t.Run("Verify() fails if validations hash changed", func(t *testing.T) {
		linkHash, e := createValidProof(t, 5)
		e.ValidationsHash = testutil.RandomHash()
		assert.False(t, e.Verify(linkHash), "Proof should not be correct if validations hash changed")
	})

	t.Run("Verify() fails if merkle tree is invalid", func(t *testing.T) {
		linkHash, e := createValidProof(t, 3)
		e.Root = linkHash
		assert.False(t, e.Verify(linkHash), "Proof should not be correct if merkle root changed")
	})

	t.Run("Verify() fails if previous app hash has changed", func(t *testing.T) {
		linkHash, e := createValidProof(t, 4)
		e.Header.AppHash = linkHash[:]
		assert.False(t, e.Verify(linkHash), "Proof should not be correct if previous app hash changed")
	})
}
