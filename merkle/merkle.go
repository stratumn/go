// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by the license that can be found in the
// LICENSE file.

// Package merkle contains types and functions to create and work with Merkle
// trees.
package merkle

import (
	"github.com/stratumn/sdk/types"
)

const (
	// HashByteSize is the length of a hash or leaf measured in bytes.
	HashByteSize = types.Bytes32Size
)

// Tree must be implemented by Merkle tree implementations.
type Tree interface {
	// NumLeaves returns the number of leaves.
	LeavesLen() int

	// Leaf returns the Merkle root.
	Root() *types.Bytes32

	// Leaf returns the leaf at the specified index.
	Leaf(index int) *types.Bytes32

	// Path returns the path of a leaf to the Merkle root.
	Path(index int) types.Path
}
