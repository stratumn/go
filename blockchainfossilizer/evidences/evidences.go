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

// Package evidences defines blockchainfossilizer evidence types.
package evidences

import (
	"bytes"
	"time"

	json "github.com/gibson042/canonicaljson-go"
	"github.com/pkg/errors"
	"github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-core/monitoring/errorcode"
	"github.com/stratumn/go-core/types"
)

const (
	// BlockchainFossilizerName is the name used as the blockchain proof backend.
	BlockchainFossilizerName = "blockchainfossilizer"

	// Version1_0_0 uses canonical-JSON to serialize a timestamped hash along
	// with a transaction ID.
	Version1_0_0 = "1.0.0"

	// Version used for new blockchain proofs.
	Version = Version1_0_0
)

// Errors used by the blockchain evidence.
var (
	ErrInvalidBackend = errors.New("backend is not blockchain")
	ErrUnknownVersion = errors.New("unknown evidence version")
)

// BlockchainProof implements the chainscript.Proof interface.
// It contains the data that was stored on the blockchain and the ID of the
// transaction that stored that data.
// It also includes a server timestamp (not trusted).
type BlockchainProof struct {
	Data          []byte              `json:"data"`
	Timestamp     int64               `json:"timestamp"`
	TransactionID types.TransactionID `json:"txid"`
}

// New creates a new blockchain proofs.
func New(data []byte, txid []byte) *BlockchainProof {
	return &BlockchainProof{
		Data:          data,
		Timestamp:     time.Now().Unix(),
		TransactionID: txid,
	}
}

// Time returns the server timestamp.
func (p *BlockchainProof) Time() uint64 {
	return uint64(p.Timestamp)
}

// Verify returns true if the proof is correct for the given data.
func (p *BlockchainProof) Verify(data interface{}) bool {
	if len(p.TransactionID) == 0 {
		return false
	}

	dataBytes, ok := data.([]byte)
	if !ok {
		return false
	}

	return bytes.Equal(dataBytes, p.Data)
}

// Evidence wraps the proof in a versioned evidence.
func (p *BlockchainProof) Evidence(provider string) (*chainscript.Evidence, error) {
	proof, err := json.Marshal(p)
	if err != nil {
		return nil, types.WrapError(err, errorcode.InvalidArgument, BlockchainFossilizerName, "json.Marshal")
	}

	e, err := chainscript.NewEvidence(Version, BlockchainFossilizerName, provider, proof)
	if err != nil {
		return nil, types.WrapError(err, errorcode.InvalidArgument, BlockchainFossilizerName, "failed to create evidence")
	}

	return e, nil
}

// UnmarshalProof unmarshals the blockchain proof contained in an evidence.
func UnmarshalProof(e *chainscript.Evidence) (*BlockchainProof, error) {
	if e.Backend != BlockchainFossilizerName {
		return nil, types.WrapError(ErrInvalidBackend, errorcode.InvalidArgument, BlockchainFossilizerName, "failed to unmarshal proof")
	}

	if len(e.Provider) == 0 {
		return nil, types.WrapError(chainscript.ErrMissingProvider, errorcode.InvalidArgument, BlockchainFossilizerName, "failed to unmarshal proof")
	}

	switch e.Version {
	case Version1_0_0:
		var proof BlockchainProof
		err := json.Unmarshal(e.Proof, &proof)
		if err != nil {
			return nil, types.WrapError(err, errorcode.InvalidArgument, BlockchainFossilizerName, "json.Unmarshal")
		}

		return &proof, nil
	default:
		return nil, types.WrapError(ErrUnknownVersion, errorcode.InvalidArgument, BlockchainFossilizerName, "failed to unmarshal proof")
	}
}
