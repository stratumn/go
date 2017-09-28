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

package types

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// MerkleNodeHashes contains a left, right, and parent hash.
type MerkleNodeHashes struct {
	Left   Bytes32 `json:"left"`
	Right  Bytes32 `json:"right"`
	Parent Bytes32 `json:"parent"`
}

// Path contains the necessary hashes to go from a leaf to a Merkle root.
type Path []MerkleNodeHashes

// Validate validates the integrity of a hash triplet.
func (h MerkleNodeHashes) Validate() error {
	hash := sha256.New()

	if _, err := hash.Write(h.Left[:]); err != nil {
		return err
	}
	if _, err := hash.Write(h.Right[:]); err != nil {
		return err
	}

	var expected Bytes32
	copy(expected[:], hash.Sum(nil))

	if h.Parent != expected {
		var (
			got  = h.Parent.String()
			want = hex.EncodeToString(expected[:])
		)
		return fmt.Errorf("unexpected parent hash got %q want %q", got, want)
	}

	return nil
}

// Validate validates the integrity of a Merkle path.
func (p Path) Validate() error {
	for i, h := range p {
		if err := h.Validate(); err != nil {
			return err
		}

		if i < len(p)-1 {
			up := p[i+1]

			if h.Parent != up.Left && h.Parent != up.Right {
				var (
					e  = hex.EncodeToString(h.Parent[:])
					a1 = hex.EncodeToString(up.Left[:])
					a2 = hex.EncodeToString(up.Right[:])
				)
				return fmt.Errorf("could not find parent hash %q, got %q and %q", e, a1, a2)
			}
		}
	}

	return nil
}

// TransactionID is a blockchain transaction ID.
type TransactionID []byte

// String returns a hex encoded string.
func (txid TransactionID) String() string {
	return hex.EncodeToString(txid)
}

// MarshalJSON implements encoding/json.Marshaler.MarshalJSON.
func (txid TransactionID) MarshalJSON() ([]byte, error) {
	return json.Marshal(txid.String())
}

// UnmarshalJSON implements encoding/json.Unmarshaler.UnmarshalJSON.
func (txid *TransactionID) UnmarshalJSON(data []byte) (err error) {
	var s string
	if err = json.Unmarshal(data, &s); err != nil {
		return
	}
	*txid, err = hex.DecodeString(s)
	return
}
