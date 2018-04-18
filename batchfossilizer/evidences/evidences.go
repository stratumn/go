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

// Package evidences defines batchfossilizer evidence types.
package evidences

import (
	"encoding/json"

	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/types"
	mktypes "github.com/stratumn/merkle/types"
)

var (
	//BatchFossilizerName is the name used as the BatchProof backend
	BatchFossilizerName = "batch"
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

func init() {
	cs.DeserializeMethods[BatchFossilizerName] = func(rawProof json.RawMessage) (cs.Proof, error) {
		p := BatchProof{}
		if err := json.Unmarshal(rawProof, &p); err != nil {
			return nil, err
		}
		return &p, nil
	}
}
