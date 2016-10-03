// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by the license that can be found in the
// LICENSE file.

package merkle

import (
	"crypto/sha256"
	"errors"
	"math"

	"github.com/stratumn/go/types"
)

// StaticTree is designed for Merkle trees with leaves that do not change.
// It is ideal when computing a tree from a batch of hashes.
type StaticTree struct {
	// We use a single buffer to store all the hashes top down.
	// For instance, given the tree:
	//
	// 0        I
	//         / \
	// 1      H   \
	//       / \   \
	// 2    F   G   \
	//     / \ / \   \
	// 3   A B C D   E
	//
	// The buffer will contain {I,H,F,G,A,B,C,D,E}.
	buffer []types.Bytes32

	// These slices map rows to the buffer.
	// For instance, given the tree above:
	// rows[0] = {I}
	// rows[1] = {H}
	// rows[2] = {F,G}
	// rows[3] = {A,B,C,D,E}
	rows [][]types.Bytes32
}

// NewStaticTree creates a static Merkle tree from a slice of leaves.
func NewStaticTree(leaves []types.Bytes32) (*StaticTree, error) {
	numLeaves := len(leaves)
	if numLeaves < 1 {
		return nil, errors.New("tree should have at least one leaf")
	}

	tree := alloc(numLeaves)
	tree.copyLeaves(leaves)
	tree.compute()

	return tree, nil
}

// LeavesLen implements Tree.LeavesLen.
func (t *StaticTree) LeavesLen() int {
	return len(t.rows[len(t.rows)-1])
}

// Root implements Tree.Root.
func (t *StaticTree) Root() *types.Bytes32 {
	r := *&t.buffer[0]
	return &r
}

// Leaf implements Tree.Leaf.
func (t *StaticTree) Leaf(index int) *types.Bytes32 {
	l := *&t.rows[len(t.rows)-1][index]
	return &l
}

// Path implements Tree.Path.
func (t *StaticTree) Path(index int) Path {
	row := len(t.rows) - 1
	if row < 0 {
		return Path{}
	}

	var (
		col   = index
		depth = 0
		path  = make(Path, row)
	)

	for row > 0 {
		t.triplet(&path[depth], row, col)
		row, col = t.parent(row, col)
		depth++
	}

	return path[:depth]
}

// Allocates memory for the buffer and creates the row slices that map to the
// buffer.
func alloc(numLeaves int) *StaticTree {
	var (
		bufl    = staticTreeBufferLen(numLeaves)
		buf     = make([]types.Bytes32, bufl)
		rowsLen = staticTreeRowsLen(numLeaves)
		depth   = len(rowsLen)
		tree    = &StaticTree{buf, make([][]types.Bytes32, depth)}
		start   = 0
		end     = 0
	)

	for i, l := range rowsLen {
		end = start + l
		tree.rows[i] = tree.buffer[start:end]
		start = end
	}

	return tree
}

// Copies the leaves at the end of the buffer.
func (t *StaticTree) copyLeaves(leaves []types.Bytes32) {
	copy(t.rows[len(t.rows)-1], leaves)
}

// Computes all the hashes. Assumes that the leaves have been copied to the
// buffer.
func (t *StaticTree) compute() {
	hash := sha256.New()
	for row := len(t.rows) - 2; row >= 0; row-- {
		rowLen := len(t.rows[row])
		for col := 0; col < rowLen; col++ {
			r, c := t.dleft(row, col)
			// Write never returns an error.
			hash.Write(t.rows[r][c][:])
			r, c = t.dright(row, col)
			hash.Write(t.rows[r][c][:])
			t.write(hash.Sum(nil), row, col)
			hash.Reset()
		}
	}
}

// Computes the values of a hash triplet for given row and column.
func (t *StaticTree) triplet(triplet *HashTriplet, row, col int) {
	r, c := t.left(row, col)
	if r >= 0 {
		t.read(triplet.Left[:], r, c)
		t.read(triplet.Right[:], row, col)
	} else {
		t.read(triplet.Left[:], row, col)
		r, c = t.right(row, col)
		t.read(triplet.Right[:], r, c)
	}

	if row > 0 {
		r, c = t.parent(row, col)
		t.read(triplet.Parent[:], r, c)
	}
}

// Reads the hash for given row and column.
func (t *StaticTree) read(dst []byte, row, col int) {
	copy(dst, t.rows[row][col][:])
}

// Writes the hash for given row and column.
func (t *StaticTree) write(src []byte, row, col int) {
	copy(t.rows[row][col][:], src)
}

// Returns the position of the node to the left of given row and column.
func (t *StaticTree) left(row, col int) (int, int) {
	if row < 1 {
		return -1, -1
	}
	r, c := t.parent(row, col)
	r, c = t.dleft(r, c)
	if r == row && c == col {
		return -1, -1
	}
	return r, c
}

// Returns the position of the node to the right of given row and column.
func (t *StaticTree) right(row, col int) (int, int) {
	if row < 1 {
		return -1, -1
	}
	r, c := t.parent(row, col)
	r, c = t.dright(r, c)
	if r == row && c == col {
		return -1, -1
	}
	return r, c
}

// Returns the position of the parent node of given row and column.
func (t *StaticTree) parent(row, col int) (int, int) {
	r, c := row-1, col/2
	for r >= 0 {
		if c < len(t.rows[r]) {
			return r, c
		}
		r, c = r-1, c/2
	}
	return -1, -1
}

// Returns the position of the left child node of given row and column.
func (t *StaticTree) dleft(row, col int) (int, int) {
	if row >= len(t.rows) {
		return -1, -1
	}
	return row + 1, col * 2
}

// Returns the position of the right child node of given row and column.
func (t *StaticTree) dright(row, col int) (int, int) {
	r, c := row+1, col*2+1
	for r < len(t.rows) {
		if c < len(t.rows[r]) {
			return r, c
		}
		r, c = r+1, c*2 // Note no plus one (orphan)!
	}
	return -1, -1
}

// Returns the number of tree nodes needed for the given number of leaves.
func numStaticTreeNodes(numLeaves int) int {
	return numLeaves*2 - 1
}

// Returns the length of the buffer needed for the given number of leaves.
func staticTreeBufferLen(numLeaves int) int {
	return numStaticTreeNodes(numLeaves)
}

// Returns the length of each tree row needed for the given number of leaves.
func staticTreeRowsLen(numLeaves int) []int {
	var (
		depth   = int(math.Ceil(math.Log2(float64(numLeaves)))) + 1
		lengths = make([]int, depth)
		curr    = numLeaves
		orphan  = false
	)

	for row := depth - 1; row >= 0; row-- {
		lengths[row] = curr

		if curr%2 > 0 {
			if orphan {
				curr++
				orphan = false
			} else {
				orphan = true
			}
		}

		curr /= 2
	}

	return lengths
}
