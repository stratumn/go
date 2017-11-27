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

package tmpop

import (
	"encoding/json"

	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/types"
	abci "github.com/tendermint/abci/types"
	merkle "github.com/tendermint/merkleeyes/iavl"
	tmtypes "github.com/tendermint/tendermint/types"
)

// init needs to define a way to deserialize a TendermintProof
func init() {
	cs.DeserializeMethods[Name] = func(rawProof json.RawMessage) (cs.Proof, error) {
		p := TendermintFullProof{}
		if err := json.Unmarshal(rawProof, &p); err != nil {
			return nil, err
		}
		return &p, nil
	}
}

// TendermintProof implements the Proof interface
type TendermintProof struct {
	BlockHeight uint64           `json:"blockHeight"`
	Header      abci.Header      `json:"header"`
	MerkleProof merkle.IAVLProof `json:"merkleProof"`
	Signatures  []tmtypes.Vote   `json:"signatures"`
}

// Time returns the timestamp from the block header
func (p *TendermintProof) Time() uint64 {
	return p.Header.GetTime()
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
	checkedLinkHash, exists := linkHash.(*types.Bytes32)
	if exists != true {
		return false
	}
	return p.MerkleProof.Verify(checkedLinkHash[:], nil, p.Header.AppHash)
}

// TendermintFullProof implements the Proof interface
type TendermintFullProof struct {
	Original TendermintProof `json:"original"`
	Current  TendermintProof `json:"current"`
}

// Time returns the timestamp from the block header
func (p *TendermintFullProof) Time() uint64 {
	return p.Original.Time()
}

// FullProof returns a JSON formatted proof
func (p *TendermintFullProof) FullProof() []byte {
	bytes, err := json.MarshalIndent(p, "", "   ")
	if err != nil {
		return nil
	}
	return bytes
}

// Verify returns true if the proof of a given linkHash is correct
func (p *TendermintFullProof) Verify(linkHash interface{}) bool {
	return p.Original.Verify(linkHash) && p.Current.Verify(linkHash)
}
