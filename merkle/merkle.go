// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package merkle contains types and functions to create and work with Merkle trees.
package merkle

import (
	"crypto/sha256"
)

const (
	// HashByteLen is the length of a hash or leaf measured in bytes.
	HashByteLen = sha256.Size
)

// Hash is a binary encoded 32-bit hash.
type Hash [HashByteLen]byte

// HashPair is a pair of hashes.
type HashPair [2]Hash

// Path contains the necessary hashes to go from a leaf to a Merkle root.
type Path []HashPair

// Tree must be implemented by Merkle tree implementations.
type Tree interface {
	// NumNodes returns the total number of nodes, including the root and leaves.
	NumNodes() int

	// NumLeaves returns the number of leaves.
	NumLeaves() int

	// Depth returns the depth of the tree.
	Depth() int

	// Leaf returns the Merkle root.
	Root() Hash

	// Leaf returns the leaf at the specified index.
	Leaf(index int) Hash

	// Path returns the path of a leaf to the Merkle root.
	Path(index int) Path
}
