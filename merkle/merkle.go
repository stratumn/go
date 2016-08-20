// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package merkle contains types and functions to create and work with Merkle trees.
package merkle

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

const (
	// HashByteLen is the length of a hash or leaf measured in bytes.
	HashByteLen = sha256.Size
)

// Hash is a binary encoded 32-byte hash.
type Hash [HashByteLen]byte

// MarshalJSON implements encoding/json.Marshaler.MarshalJSON.
func (h Hash) MarshalJSON() ([]byte, error) {
	return json.Marshal(hex.EncodeToString(h[:]))
}

// UnmarshalJSON implements encoding/json.Unmarshaler.UnmarshalJSON.
func (h *Hash) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	if _, err := hex.Decode(h[:], []byte(s)); err != nil {
		return err
	}

	return nil
}

// HashTriplet contains a left, right, and parent hash.
type HashTriplet struct {
	Left   Hash `json:"left"`
	Right  Hash `json:"right"`
	Parent Hash `json:"parent"`
}

// Validate validates the integrity of a hash triplet.
func (h HashTriplet) Validate() error {
	hash := sha256.New()

	if _, err := hash.Write(h.Left[:]); err != nil {
		return err
	}
	if _, err := hash.Write(h.Right[:]); err != nil {
		return err
	}

	var expected Hash
	copy(expected[:], hash.Sum(nil))

	if h.Parent != expected {
		var (
			a = hex.EncodeToString(h.Parent[:])
			e = hex.EncodeToString(expected[:])
		)
		return fmt.Errorf("unexpected parent hash, got %s expected %s\n", a, e)
	}

	return nil
}

// Path contains the necessary hashes to go from a leaf to a Merkle root.
type Path []HashTriplet

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
				return fmt.Errorf("could not find parent hash %s, got %s and %s\n", e, a1, a2)
			}
		}
	}

	return nil
}

// Tree must be implemented by Merkle tree implementations.
type Tree interface {
	// NumLeaves returns the number of leaves.
	NumLeaves() int

	// Leaf returns the Merkle root.
	Root() Hash

	// Leaf returns the leaf at the specified index.
	Leaf(index int) Hash

	// Path returns the path of a leaf to the Merkle root.
	Path(index int) Path
}
