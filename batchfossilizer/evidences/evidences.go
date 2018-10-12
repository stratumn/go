// Copyright 2016-2018 Stratumn SAS. All rights reserved.
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
	json "github.com/gibson042/canonicaljson-go"
	"github.com/pkg/errors"
	"github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-core/monitoring/errorcode"
	"github.com/stratumn/go-core/types"
	mktypes "github.com/stratumn/merkle/types"
)

const (
	// BatchFossilizerName is the name used as the BatchProof backend.
	BatchFossilizerName = "batchfossilizer"

	// Version1_0_0 uses canonical-JSON to serialize a timestamped merkle path
	// and a merkle root.
	Version1_0_0 = "1.0.0"

	// Version used for new batch proofs.
	Version = Version1_0_0
)

// Errors used by the batch evidence.
var (
	ErrInvalidBackend = errors.New("backend is not batchfossilizer")
	ErrUnknownVersion = errors.New("unknown evidence version")
)

// BatchProof implements the chainscript.Proof interface.
type BatchProof struct {
	Timestamp int64          `json:"timestamp"`
	Root      *types.Bytes32 `json:"merkleRoot"`
	Path      mktypes.Path   `json:"merklePath"`
}

// Time returns the timestamp from the block header.
func (p *BatchProof) Time() uint64 {
	return uint64(p.Timestamp)
}

// Verify returns true if the proof of a given linkHash is correct.
func (p *BatchProof) Verify(_ interface{}) bool {
	err := p.Path.Validate()
	return err == nil
}

// Evidence wraps the proof in a versioned evidence.
func (p *BatchProof) Evidence(provider string) (*chainscript.Evidence, error) {
	proof, err := json.Marshal(p)
	if err != nil {
		return nil, types.WrapError(err, errorcode.InvalidArgument, BatchFossilizerName, "json.Marshal")
	}

	e, err := chainscript.NewEvidence(Version, BatchFossilizerName, provider, proof)
	if err != nil {
		return nil, types.WrapError(err, errorcode.InvalidArgument, BatchFossilizerName, "failed to create evidence")
	}

	return e, nil
}

// UnmarshalProof unmarshals the batch proof contained in an evidence.
func UnmarshalProof(e *chainscript.Evidence) (*BatchProof, error) {
	if e.Backend != BatchFossilizerName {
		return nil, types.WrapError(ErrInvalidBackend, errorcode.InvalidArgument, BatchFossilizerName, "failed to unmarshal proof")
	}

	if len(e.Provider) == 0 {
		return nil, types.WrapError(chainscript.ErrMissingProvider, errorcode.InvalidArgument, BatchFossilizerName, "failed to unmarshal proof")
	}

	switch e.Version {
	case Version1_0_0:
		var proof BatchProof
		err := json.Unmarshal(e.Proof, &proof)
		if err != nil {
			return nil, types.WrapError(err, errorcode.InvalidArgument, BatchFossilizerName, "json.Unmarshal")
		}

		return &proof, nil
	default:
		return nil, types.WrapError(ErrUnknownVersion, errorcode.InvalidArgument, BatchFossilizerName, "failed to unmarshal proof")
	}
}
