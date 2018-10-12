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

package evidences

import (
	json "github.com/gibson042/canonicaljson-go"
	"github.com/pkg/errors"
	"github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-core/monitoring/errorcode"
	"github.com/stratumn/go-core/types"
)

const (
	// Name is the name set in the fossilizer's information.
	Name = "dummyfossilizer"

	// Version1_0_0 uses canonical-JSON to serialize a timestamp.
	Version1_0_0 = "1.0.0"

	// Version used for new batch proofs.
	Version = Version1_0_0
)

// Errors used by the batch evidence.
var (
	ErrInvalidBackend = errors.New("backend is not dummyfossilizer")
	ErrUnknownVersion = errors.New("unknown evidence version")
)

// DummyProof implements the chainscript.Proof interface.
type DummyProof struct {
	Timestamp uint64 `json:"timestamp"`
}

// Time returns the timestamp.
func (p *DummyProof) Time() uint64 {
	return p.Timestamp
}

// Verify returns true if the proof of a given linkHash is correct.
func (p *DummyProof) Verify(interface{}) bool {
	return true
}

// Evidence wraps the proof in a versioned evidence.
func (p *DummyProof) Evidence(provider string) (*chainscript.Evidence, error) {
	proof, err := json.Marshal(p)
	if err != nil {
		return nil, types.WrapError(err, errorcode.InvalidArgument, Name, "json.Marshal")
	}

	e, err := chainscript.NewEvidence(Version, Name, provider, proof)
	if err != nil {
		return nil, types.WrapError(err, errorcode.InvalidArgument, Name, "failed to create evidence")
	}

	return e, nil
}

// UnmarshalProof unmarshals the dummy proof contained in an evidence.
func UnmarshalProof(e *chainscript.Evidence) (*DummyProof, error) {
	if e.Backend != Name {
		return nil, types.WrapError(ErrInvalidBackend, errorcode.InvalidArgument, Name, "failed to unmarshal proof")
	}

	if len(e.Provider) == 0 {
		return nil, types.WrapError(chainscript.ErrMissingProvider, errorcode.InvalidArgument, Name, "failed to unmarshal proof")
	}

	switch e.Version {
	case Version1_0_0:
		var proof DummyProof
		err := json.Unmarshal(e.Proof, &proof)
		if err != nil {
			return nil, types.WrapError(err, errorcode.InvalidArgument, Name, "json.Unmarshal")
		}

		return &proof, nil
	default:
		return nil, types.WrapError(ErrUnknownVersion, errorcode.InvalidArgument, Name, "failed to unmarshal proof")
	}
}
