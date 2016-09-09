// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package merkle_test

import (
	"testing"

	"github.com/stratumn/go/testutil"
	"github.com/stratumn/go/types"
	"github.com/stratumn/goprivate/merkle"
	"github.com/stratumn/goprivate/merkle/treetestcases"
)

func TestDynTree(t *testing.T) {
	treetestcases.Factory{
		New: func(leaves []types.Bytes32) (merkle.Tree, error) {
			tree := merkle.NewDynTree(len(leaves))
			for _, leaf := range leaves {
				tree.Add(&leaf)
			}
			return tree, nil
		},
	}.RunTests(t)
}

func TestDynTreePause(t *testing.T) {
	treetestcases.Factory{
		New: func(leaves []types.Bytes32) (merkle.Tree, error) {
			tree := merkle.NewDynTree(len(leaves))
			tree.Pause()
			for _, leaf := range leaves {
				tree.Add(&leaf)
			}
			tree.Resume()
			return tree, nil
		},
	}.RunTests(t)
}

func TestDynTreeUpdate(t *testing.T) {
	tree := merkle.NewDynTree(16)

	for i := 0; i < 10; i++ {
		tree.Add(testutil.RandomHash())
	}

	var (
		r0 = tree.Root()
		l2 = tree.Leaf(2)
		l5 = tree.Leaf(5)
	)

	tree.Update(2, testutil.RandomHash())
	r1 := tree.Root()
	if got, notWant := r1.String(), r0.String(); got == notWant {
		t.Errorf("tree.Root() = %q want not %q", got, notWant)
	}

	tree.Update(5, testutil.RandomHash())
	if got, notWant := tree.Root().String(), r1.String(); got == notWant {
		t.Errorf("tree.Root() = %q want not %q", got, notWant)
	}

	tree.Update(5, l5)
	if got, want := tree.Root().String(), r1.String(); got != want {
		t.Errorf("tree.Root() = %q want %q", got, want)
	}

	tree.Update(2, l2)
	if got, want := tree.Root().String(), r0.String(); got != want {
		t.Errorf("tree.Root() = %q want %q", got, want)
	}
}

func BenchmarkDynTree(b *testing.B) {
	treetestcases.Factory{
		New: func(leaves []types.Bytes32) (merkle.Tree, error) {
			tree := merkle.NewDynTree(len(leaves))
			for _, leaf := range leaves {
				tree.Add(&leaf)
			}
			return tree, nil
		},
	}.RunBenchmarks(b)
}

func BenchmarkDynTreePause(b *testing.B) {
	treetestcases.Factory{
		New: func(leaves []types.Bytes32) (merkle.Tree, error) {
			tree := merkle.NewDynTree(len(leaves))
			tree.Pause()
			for _, leaf := range leaves {
				tree.Add(&leaf)
			}
			tree.Resume()
			return tree, nil
		},
	}.RunBenchmarks(b)
}
