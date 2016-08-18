// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package merkle

import (
	"crypto/sha256"
	"errors"
	"math"
)

// StaticTree is designed for Merkle trees with leaves that do not change.
// It is ideal when computing a tree from a batch of hashes.
type StaticTree struct {
	numLeaves int

	// We use a single buffer to store all the hashes top down.
	// For instance, given the tree:
	//
	// 0        I
	//         / \
	// 1      H   \
	//       /  \  \
	// 2    F   G   \
	//     / \ / \   \
	// 3   A B C D   E
	//
	// The buffer will contain {I,H,F,G,A,B,C,D,E}.
	buffer []byte

	// These slices map levels to the buffer.
	// For instance, given the tree above:
	// levels[0] = {I}
	// levels[1] = {H}
	// levels[2] = {F,G}
	// levels[3] = {A,B,C,D,E}
	levels [][]byte
}

// NewStaticTree creates a static Merkle tree from a slice of leaves.
func NewStaticTree(leaves []Hash) (*StaticTree, error) {
	numLeaves := len(leaves)
	if numLeaves < 1 {
		return nil, errors.New("tree should have at least one leaf")
	}

	tree := alloc(numLeaves)
	tree.copyLeaves(leaves)
	tree.compute()

	return tree, nil
}

// Root returns the Merkle root of the tree.
func (t *StaticTree) Root() (hash Hash) {
	copy(hash[:], t.buffer[:])
	return
}

// Leaf implements Tree.Leaf.
func (t *StaticTree) Leaf(index int) (hash Hash) {
	offset := index * HashByteLen
	copy(hash[:], t.levels[len(t.levels)-1][offset:])
	return
}

// Path implements Tree.Path.
func (t *StaticTree) Path(index int) Path {
	// 0        I
	//         / \
	// 1      H   \
	//       /  \  \
	// 2    F   G   \
	//     / \ / \   \
	// 3   A B C D   E
	//
	// Comments refer to this Merkle tree.

	var (
		depth  = len(t.levels)
		path   = make(Path, 0, depth-1)
		left   Hash
		right  Hash
		orphan *Hash
	)

	l := len(t.levels) - 1
	if l == 0 {
		return Path{}
	}

	// Note we don't care about the root level (Merkle root), because it can be computed
	// from the last pair of the path.
	for ; l > 0; l-- {
		level := t.levels[l]
		levelLen := len(level) / HashByteLen

		if orphan != nil {
			if levelLen%2 == 1 {
				// This case happens for instance if we at level 1, up from E.
				// E was previously marked as orphan, and H is available.
				// We can append HashPair{H,E}.
				copy(left[:], level[HashByteLen*(levelLen-1):])
				path = append(path, HashPair{left, *orphan})
				orphan = nil
			}
			// Otherwise we keep going up the tree.
		} else if levelLen == 1 {
			// This case happens for instance if we are at node H.
			// We must go down until we find an orphan, E in the example.
			// We can append HashPair{H,E}.
			copy(left[:], level[:])
			for rl := l + 1; rl < depth; rl++ {
				rlevel := t.levels[rl]
				rlevelLen := len(rlevel) / HashByteLen
				if rlevelLen%2 > 0 {
					copy(right[:], rlevel[HashByteLen*(rlevelLen-1):])
					break
				}
			}
			path = append(path, HashPair{left, right})
			orphan = nil
		} else if index%2 == 0 {
			if index+1 < levelLen {
				// This case happens for instance if we are at node F.
				// We can simply use node G as the right node.
				// We can append HashPair{F,G}.
				copy(left[:], level[HashByteLen*(index):])
				copy(right[:], level[HashByteLen*(index+1):])
				path = append(path, HashPair{left, right})
			} else {
				// This case happens for instance if we are at node E.
				// We don't have a neighbor available at this level, so we tag E
				// as orphan, which will eventually be used as a
				// right node further up.
				copy(right[:], level[HashByteLen*(index):])
				orphan = &right
			}
		} else {
			// This case happens for instance if we are at node G.
			// We can simply use node F as the left node.
			// We can append HashPair{F,G}.
			copy(left[:], level[HashByteLen*(index-1):])
			copy(right[:], level[HashByteLen*(index):])
			path = append(path, HashPair{left, right})
		}

		index /= 2
	}

	return path
}

// Allocates memory for the buffer and creates the level slices that map to the buffer.
func alloc(numLeaves int) *StaticTree {
	var (
		bufl      = staticTreeBufferLen(numLeaves)
		buf       = make([]byte, bufl)
		levelLens = staticTreeLevelsLen(numLeaves)
		depth     = len(levelLens)
		tree      = &StaticTree{numLeaves, buf, make([][]byte, depth, depth)}
		start     = 0
		end       = 0
	)

	for i, l := range levelLens {
		end = start + l*HashByteLen
		tree.levels[i] = tree.buffer[start:end]
		start = end
	}

	return tree
}

// Copies the leaves at the end of the buffer.
func (t *StaticTree) copyLeaves(leaves []Hash) {
	level := t.levels[len(t.levels)-1]
	for i, v := range leaves {
		copy(level[i*HashByteLen:], v[:])
	}
}

// Computes all the hashes. Assumes that the leaves have been copied to the buffer.
func (t *StaticTree) compute() {
	// 0        I
	//         / \
	// 1      H   \
	//       /  \  \
	// 2    F   G   \
	//     / \ / \   \
	// 3   A B C D   E
	//
	// Comments refer to this Merkle tree.

	var (
		depth  = len(t.levels)
		orphan []byte
	)

	for l := depth - 1; l > 0; l-- {
		var (
			parent      = t.levels[l-1]
			level       = t.levels[l]
			levelLength = len(level)
		)

		// Iterate nodes of the current level two by two.
		for start := 0; start < levelLength; start += 2 * HashByteLen {
			var (
				end  = start + HashByteLen
				left = level[start:end]
			)

			if end < levelLength || orphan != nil {
				var right []byte

				if end >= levelLength && orphan != nil {
					// This case happens for instance if we are computing I. H doesn't have a
					// right neighbor, but E was previously marked as orphan.
					// We can compute Hash(I,E).
					right = orphan
					orphan = nil
				} else {
					// This case happens for instance if we are computing H.
					// We can simply compute Hash(F,G).
					right = level[end : end+HashByteLen]
				}

				hash := sha256.New()
				if _, err := hash.Write(left); err != nil {
					panic(err)
				}
				if _, err := hash.Write(right); err != nil {
					panic(err)
				}

				copy(parent[start/2:], hash.Sum(nil))
			} else {
				// This case happens for instance if we are trying to compute the parent of E.
				// We don't have a parent at the next level, so we mark E as orphan,
				// which will eventually be used further up.
				orphan = left
			}
		}
	}
}

// Returns the number of tree nodes needed for the given number of leaves.
func numStaticTreeNodes(numLeaves int) int {
	return numLeaves*2 - 1
}

// Returns the length of the buffer needed for the given number of leaves.
func staticTreeBufferLen(numLeaves int) int {
	return HashByteLen * numStaticTreeNodes(numLeaves)
}

// Returns the length of each tree level needed for the given number of leaves.
func staticTreeLevelsLen(numLeaves int) []int {
	var (
		depth  = int(math.Ceil(math.Log2(float64(numLeaves)))) + 1
		lens   = make([]int, depth)
		curr   = numLeaves
		orphan = false
	)

	for level := depth - 1; level >= 0; level-- {
		lens[level] = curr

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

	return lens
}
