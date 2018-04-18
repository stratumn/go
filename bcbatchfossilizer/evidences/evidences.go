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

// Package evidences defines bcbatchfossilizer evidence types.
package evidences

import (
	"encoding/json"

	batchevidences "github.com/stratumn/go-indigocore/batchfossilizer/evidences"
	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/types"
)

var (
	//BcBatchFossilizerName is the name used as the BcBatchProof backend
	BcBatchFossilizerName = "bcbatch"
)

// BcBatchProof implements the Proof interface
type BcBatchProof struct {
	Batch         batchevidences.BatchProof `json:"batch"`
	TransactionID types.TransactionID       `json:"txid"`
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

func init() {
	cs.DeserializeMethods[BcBatchFossilizerName] = func(rawProof json.RawMessage) (cs.Proof, error) {
		p := BcBatchProof{}
		if err := json.Unmarshal(rawProof, &p); err != nil {
			return nil, err
		}
		return &p, nil
	}
}
