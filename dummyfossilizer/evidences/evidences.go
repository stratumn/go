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

package evidences

import (
	"encoding/json"

	"github.com/stratumn/go-indigocore/cs"
)

const (
	// Name is the name set in the fossilizer's information.
	Name = "dummy"
)

// DummyProof implements the cs.Proof interface
type DummyProof struct {
	Timestamp uint64 `json:"timestamp"`
}

// Time returns the timestamp from the block header
func (p *DummyProof) Time() uint64 {
	return p.Timestamp
}

// FullProof returns a JSON formatted proof
func (p *DummyProof) FullProof() []byte {
	bytes, err := json.MarshalIndent(p, "", "   ")
	if err != nil {
		return nil
	}
	return bytes
}

// Verify returns true if the proof of a given linkHash is correct
func (p *DummyProof) Verify(interface{}) bool {
	return true
}

// init needs to define a way to deserialize a DummyProof
func init() {
	cs.DeserializeMethods[Name] = func(rawProof json.RawMessage) (cs.Proof, error) {
		p := DummyProof{}
		if err := json.Unmarshal(rawProof, &p); err != nil {
			return nil, err
		}
		return &p, nil
	}
}
