// Copyright 2017-2018 Stratumn SAS. All rights reserved.
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
	json "github.com/gibson042/canonicaljson-go"
	"github.com/pkg/errors"
	"github.com/stratumn/go-chainscript"
	batchevidences "github.com/stratumn/go-core/batchfossilizer/evidences"
	"github.com/stratumn/go-core/types"
)

const (
	// BcBatchFossilizerName is the name used as the BcBatchProof backend.
	BcBatchFossilizerName = "bcbatch"

	// Version1_0_0 uses canonical-JSON to serialize a timestamped merkle path
	// and a merkle root along with a transaction ID.
	Version1_0_0 = "1.0.0"

	// Version used for new bcbatch proofs.
	Version = Version1_0_0
)

// Errors used by the bcbatch evidence.
var (
	ErrInvalidBackend = errors.New("backend is not batch")
	ErrUnknownVersion = errors.New("unknown evidence version")
)

// BcBatchProof implements the chainscript.Proof interface.
type BcBatchProof struct {
	Batch         batchevidences.BatchProof `json:"batch"`
	TransactionID types.TransactionID       `json:"txid"`
}

// Time returns the timestamp from the block header.
func (p *BcBatchProof) Time() uint64 {
	return uint64(p.Batch.Timestamp)
}

// Verify returns true if the proof of a given linkHash is correct.
func (p *BcBatchProof) Verify(linkHash interface{}) bool {
	return p.Batch.Verify(linkHash)
}

// Evidence wraps the proof in a versioned evidence.
func (p *BcBatchProof) Evidence(provider string) (*chainscript.Evidence, error) {
	proof, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}

	return chainscript.NewEvidence(Version, BcBatchFossilizerName, provider, proof)
}

// UnmarshalProof unmarshals the bcbatch proof contained in an evidence.
func UnmarshalProof(e *chainscript.Evidence) (*BcBatchProof, error) {
	if e.Backend != BcBatchFossilizerName {
		return nil, ErrInvalidBackend
	}

	if len(e.Provider) == 0 {
		return nil, chainscript.ErrMissingProvider
	}

	switch e.Version {
	case Version1_0_0:
		var proof BcBatchProof
		err := json.Unmarshal(e.Proof, &proof)
		if err != nil {
			return nil, err
		}

		return &proof, nil
	default:
		return nil, ErrUnknownVersion
	}
}
