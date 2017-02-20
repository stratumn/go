// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by the license that can be found in the
// LICENSE file.

// Package merkle contains types and functions to create and work with Merkle
// trees.
package merkle

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/stratumn/sdk/types"
)

const (
	// HashByteSize is the length of a hash or leaf measured in bytes.
	HashByteSize = types.Bytes32Size
)

// HashTriplet contains a left, right, and parent hash.
type HashTriplet struct {
	Left   types.Bytes32 `json:"left"`
	Right  types.Bytes32 `json:"right"`
	Parent types.Bytes32 `json:"parent"`
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

	var expected types.Bytes32
	copy(expected[:], hash.Sum(nil))

	if h.Parent != expected {
		var (
			got  = h.Parent.String()
			want = hex.EncodeToString(expected[:])
		)
		return fmt.Errorf("unexpected parent hash got %q want %q\n", got, want)
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
				return fmt.Errorf("could not find parent hash %q, got %q and %q\n", e, a1, a2)
			}
		}
	}

	return nil
}

// Tree must be implemented by Merkle tree implementations.
type Tree interface {
	// NumLeaves returns the number of leaves.
	LeavesLen() int

	// Leaf returns the Merkle root.
	Root() *types.Bytes32

	// Leaf returns the leaf at the specified index.
	Leaf(index int) *types.Bytes32

	// Path returns the path of a leaf to the Merkle root.
	Path(index int) Path
}
