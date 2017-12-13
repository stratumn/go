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

	"github.com/stratumn/sdk/cs"
	// This package imports every package defining its own implementation of the cs.Proof interface
	// The init() function of each package gets called hence providing a way for cs.Evidence.UnmarshalJSON to deserialize any kind of proof
	_ "github.com/stratumn/sdk/dummyfossilizer"
	"github.com/stratumn/sdk/types"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/go-crypto"
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
	Path      types.Path     `json:"merklePath"`
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

// TendermintSignature is a signature by one of the Tendermint nodes
type TendermintSignature struct {
	PubKey    crypto.PubKey    `json:"pub_key"`
	Signature crypto.Signature `json:"signature"`
}

// TendermintProof implements the Proof interface
type TendermintProof struct {
	BlockHeight int64 `json:"blockHeight"`

	Root            *types.Bytes32 `json:"merkleRoot"`
	Path            types.Path     `json:"merklePath"`
	ValidationsHash *types.Bytes32 `json:"validationsHash"`

	// The header and its signatures are needed to validate
	// the previous app hash and metadata such as the height and time
	Header     abci.Header            `json:"header"`
	Signatures []*TendermintSignature `json:"signatures"`

	// The next header and its signatures are needed to validate
	// the app hash representing the validations and merkle path
	NextHeader     abci.Header            `json:"nextHeader"`
	NextSignatures []*TendermintSignature `json:"nextSignatures"`
}

// Time returns the timestamp from the block header
func (p *TendermintProof) Time() uint64 {
	return uint64(p.Header.GetTime())
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
		// If the tree contains a single element, it's valid if it's the root
		if lh.Equals(p.Root) {
			return true
		}

		return false
	}

	if err := p.Path.Validate(); err != nil {
		return false
	}

	// And that the merkle path starts at the given link hash

	if !lh.Equals(&p.Path[0].Left) && !lh.Equals(&p.Path[0].Right) {
		return false
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
