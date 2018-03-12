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

// Package evidences defines any type of proof that can be used in a chainscript segment
// It is needed by a store to know how to deserialize a segment containing any type of proof
package evidences

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"

	"github.com/stratumn/go-indigocore/cs"
	// This package imports every package defining its own implementation of the cs.Proof interface
	// The init() function of each package gets called hence providing a way for cs.Evidence.UnmarshalJSON to deserialize any kind of proof
	_ "github.com/stratumn/go-indigocore/dummyfossilizer"
	"github.com/stratumn/go-indigocore/types"
	mktypes "github.com/stratumn/merkle/types"
	"github.com/tendermint/go-crypto"
	tmtypes "github.com/tendermint/tendermint/types"
)

var (
	//BatchFossilizerName is the name used as the BatchProof backend
	BatchFossilizerName = "batch"
	//BcBatchFossilizerName is the name used as the BcBatchProof backend
	BcBatchFossilizerName = "bcbatch"
	// TMPopName is the name used as the Tendermint PoP backend
	TMPopName = "TMPop"
)

// BatchProof implements the Proof interface
type BatchProof struct {
	Timestamp int64          `json:"timestamp"`
	Root      *types.Bytes32 `json:"merkleRoot"`
	Path      mktypes.Path   `json:"merklePath"`
}

// Time returns the timestamp from the block header
func (p *BatchProof) Time() uint64 {
	return uint64(p.Timestamp)
}

// FullProof returns a JSON formatted proof
func (p *BatchProof) FullProof() []byte {
	bytes, err := json.MarshalIndent(p, "", "   ")
	if err != nil {
		return nil
	}
	return bytes
}

// Verify returns true if the proof of a given linkHash is correct
func (p *BatchProof) Verify(linkHash interface{}) bool {
	err := p.Path.Validate()
	if err != nil {
		return false
	}
	return true
}

// BcBatchProof implements the Proof interface
type BcBatchProof struct {
	Batch         BatchProof          `json:"batch"`
	TransactionID types.TransactionID `json:"txid"`
}

// Time returns the timestamp from the block header
func (p *BcBatchProof) Time() uint64 {
	return uint64(p.Batch.Timestamp)
}

// FullProof returns a JSON formatted proof
func (p *BcBatchProof) FullProof() []byte {
	bytes, err := json.MarshalIndent(p, "", "   ")
	if err != nil {
		return nil
	}
	return bytes
}

// Verify returns true if the proof of a given linkHash is correct
func (p *BcBatchProof) Verify(linkHash interface{}) bool {
	err := p.Batch.Path.Validate()
	if err != nil {
		return false
	}
	return true
}

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
	Header      *tmtypes.Header   `json:"header"`
	HeaderVotes []*TendermintVote `json:"header_votes"`

	// The next header and its votes are needed to validate
	// the app hash representing the validations and merkle path.
	NextHeader      *tmtypes.Header   `json:"next_header"`
	NextHeaderVotes []*TendermintVote `json:"next_header_votes"`
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

	// We validate that nodes signed the header.
	if !p.validateVotes(p.Header, p.HeaderVotes) {
		return false
	}

	// We validate that nodes signed the next header.
	if !p.validateVotes(p.NextHeader, p.NextHeaderVotes) {
		return false
	}

	return true
}

// validateVotes verifies that votes are correctly signed
// and refer to the given header.
func (p *TendermintProof) validateVotes(header *tmtypes.Header, votes []*TendermintVote) bool {
	if len(votes) == 0 {
		return false
	}

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
	}

	return true
}

func init() {
	cs.DeserializeMethods[BatchFossilizerName] = func(rawProof json.RawMessage) (cs.Proof, error) {
		p := BatchProof{}
		if err := json.Unmarshal(rawProof, &p); err != nil {
			return nil, err
		}
		return &p, nil
	}
	cs.DeserializeMethods[BcBatchFossilizerName] = func(rawProof json.RawMessage) (cs.Proof, error) {
		p := BcBatchProof{}
		if err := json.Unmarshal(rawProof, &p); err != nil {
			return nil, err
		}
		return &p, nil
	}
	cs.DeserializeMethods[TMPopName] = func(rawProof json.RawMessage) (cs.Proof, error) {
		p := TendermintProof{}
		if err := json.Unmarshal(rawProof, &p); err != nil {
			return nil, err
		}
		return &p, nil
	}
}
