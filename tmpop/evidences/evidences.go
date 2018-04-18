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

// Package evidences defines Tendermint proofs.
// It is needed by a store to know how to deserialize a segment containing
// a Tendermint proof.
package evidences

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"

	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/types"
	mktypes "github.com/stratumn/merkle/types"
	"github.com/tendermint/go-crypto"
	tmtypes "github.com/tendermint/tendermint/types"
)

var (
	// TMPopName is the name used as the Tendermint PoP backend
	TMPopName = "TMPop"
)

// TendermintVote is a signed vote by one of the Tendermint validator nodes.
type TendermintVote struct {
	PubKey *crypto.PubKey `json:"pub_key"`
	Vote   *tmtypes.Vote  `json:"vote"`
}

// TendermintProof implements the Proof interface.
type TendermintProof struct {
	BlockHeight int64 `json:"block_height"`

	Root            *types.Bytes32 `json:"merkle_root"`
	Path            mktypes.Path   `json:"merkle_path"`
	ValidationsHash *types.Bytes32 `json:"validations_hash"`

	// The header and its votes are needed to validate
	// the previous app hash and metadata such as the height and time.
	Header             *tmtypes.Header       `json:"header"`
	HeaderVotes        []*TendermintVote     `json:"header_votes"`
	HeaderValidatorSet *tmtypes.ValidatorSet `json:"header_validator_set"`

	// The next header and its votes are needed to validate
	// the app hash representing the validations and merkle path.
	NextHeader             *tmtypes.Header       `json:"next_header"`
	NextHeaderVotes        []*TendermintVote     `json:"next_header_votes"`
	NextHeaderValidatorSet *tmtypes.ValidatorSet `json:"next_header_validator_set"`
}

// Time returns the timestamp from the block header
func (p *TendermintProof) Time() uint64 {
	return uint64(p.Header.Time.Unix())
}

// FullProof returns a JSON formatted proof
func (p *TendermintProof) FullProof() []byte {
	bytes, err := json.MarshalIndent(p, "", "   ")
	if err != nil {
		return nil
	}
	return bytes
}

// Verify returns true if the proof of a given linkHash is correct
func (p *TendermintProof) Verify(linkHash interface{}) bool {
	lh, ok := linkHash.(*types.Bytes32)
	if ok != true {
		return false
	}

	// We first verify that the app hash is correct

	hash := sha256.New()
	if _, err := hash.Write(types.NewBytes32FromBytes(p.Header.AppHash)[:]); err != nil {
		return false
	}

	validationsHash := p.ValidationsHash
	if validationsHash == nil {
		validationsHash = &types.Bytes32{}
	}

	if _, err := hash.Write(validationsHash[:]); err != nil {
		return false
	}
	if _, err := hash.Write(p.Root[:]); err != nil {
		return false
	}

	expectedAppHash := hash.Sum(nil)
	if bytes.Compare(expectedAppHash, p.NextHeader.AppHash) != 0 {
		return false
	}

	// Then we validate the merkle path

	if len(p.Path) == 0 {
		// If the tree contains a single element,
		// it's valid only if it's the root.
		if !lh.Equals(p.Root) {
			return false
		}
	} else {
		// Otherwise the path needs to be valid.
		if err := p.Path.Validate(); err != nil {
			return false
		}

		// And it should start at the given link hash.
		if !lh.EqualsBytes(p.Path[0].Left) && !lh.EqualsBytes(p.Path[0].Right) {
			return false
		}
	}

	// If validator set doesn't match the header's validatorHash,
	// someone tampered with the validator set.
	if !p.validateValidatorSet() {
		return false
	}

	// We validate that nodes signed the header.
	if !p.validateVotes(p.Header, p.HeaderVotes, p.HeaderValidatorSet) {
		return false
	}

	// We validate that nodes signed the next header.
	if !p.validateVotes(p.NextHeader, p.NextHeaderVotes, p.NextHeaderValidatorSet) {
		return false
	}

	return true
}

// validateValidatorSet verifies that the signed headers
// align with the given validator set.
func (p *TendermintProof) validateValidatorSet() bool {
	if p.HeaderValidatorSet == nil || p.NextHeaderValidatorSet == nil {
		return false
	}

	if p.Header.ValidatorsHash == nil || p.NextHeader.ValidatorsHash == nil {
		return false
	}

	if !bytes.Equal(p.HeaderValidatorSet.Hash(), p.Header.ValidatorsHash.Bytes()) {
		return false
	}

	if !bytes.Equal(p.NextHeaderValidatorSet.Hash(), p.NextHeader.ValidatorsHash.Bytes()) {
		return false
	}

	return true
}

// validateVotes verifies that votes are correctly signed
// and refer to the given header.
func (p *TendermintProof) validateVotes(header *tmtypes.Header, votes []*TendermintVote, validatorSet *tmtypes.ValidatorSet) bool {
	if len(votes) == 0 {
		return false
	}

	votesPower := int64(0)

	for _, v := range votes {
		if v == nil || v.PubKey == nil || v.PubKey.Empty() || v.Vote == nil || v.Vote.BlockID.IsZero() {
			return false
		}

		// If the vote isn't for the the given header,
		// no need to verify the signatures.
		if bytes.Compare(v.Vote.BlockID.Hash.Bytes(), header.Hash().Bytes()) != 0 {
			return false
		}

		if err := v.Vote.Verify(header.ChainID, *v.PubKey); err != nil {
			return false
		}

		_, validator := validatorSet.GetByIndex(v.Vote.ValidatorIndex)
		if validator == nil {
			return false
		}

		votesPower += validator.VotingPower
	}

	// We need more than 2/3 of the votes for the proof to be accepted.
	if 3*votesPower <= 2*validatorSet.TotalVotingPower() {
		return false
	}

	return true
}

func init() {
	cs.DeserializeMethods[TMPopName] = func(rawProof json.RawMessage) (cs.Proof, error) {
		p := TendermintProof{}
		if err := json.Unmarshal(rawProof, &p); err != nil {
			return nil, err
		}
		return &p, nil
	}
}
