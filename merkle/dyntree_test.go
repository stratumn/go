// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package merkle_test

import (
	"testing"

	"github.com/stratumn/goprivate/merkle"
	"github.com/stratumn/goprivate/merkle/treetestcases"
	"github.com/stratumn/goprivate/testutil"
	"github.com/stratumn/goprivate/types"
)

func TestNewDynTree(t *testing.T) {
	tree := merkle.NewDynTree(16)
	if tree == nil {
		t.Fatal("expected tree not to be nil")
	}
}

func TestDynTree(t *testing.T) {
	treetestcases.Factory{
		New: func(leaves []types.Bytes32) (merkle.Tree, error) {
			tree := merkle.NewDynTree(len(leaves) * 2)
			for _, leaf := range leaves {
				tree.Add(&leaf)
			}
			return tree, nil
		},
	}.RunTests(t)
}

func TestDynTreeUpdate(t *testing.T) {
	tree := merkle.NewDynTree(16)

	for i := 0; i < 10; i++ {
		tree.Add(testutil.RandomHash())
	}

	r0 := tree.Root()
	l2 := tree.Leaf(2)
	l5 := tree.Leaf(5)

	tree.Update(2, testutil.RandomHash())

	r1 := tree.Root()

	if r1 == r0 {
		t.Fatal("expected root to change")
	}

	tree.Update(5, testutil.RandomHash())

	if tree.Root() == r1 {
		t.Fatal("expected root to change")
	}

	tree.Update(5, &l5)

	if tree.Root() != r1 {
		t.Fatal("unexpected root")
	}

	tree.Update(2, &l2)

	if tree.Root() != r0 {
		t.Fatal("unexpected root")
	}
}

func BenchmarkDynTree(b *testing.B) {
	treetestcases.Factory{
		New: func(leaves []types.Bytes32) (merkle.Tree, error) {
			tree := merkle.NewDynTree(len(leaves) * 2)
			for _, leaf := range leaves {
				tree.Add(&leaf)
			}
			return tree, nil
		},
	}.RunBenchmarks(b)
}
