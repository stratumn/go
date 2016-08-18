// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package merkle_test

import (
	"testing"

	"github.com/stratumn/goprivate/merkle"
	"github.com/stratumn/goprivate/merkle/treetestcases"
)

func TestNewDynTree(t *testing.T) {
	tree := merkle.NewDynTree(16)
	if tree == nil {
		t.Fatal("expected tree not to be nil")
	}
}

func TestDynTree(t *testing.T) {
	treetestcases.Factory{
		New: func(leaves []merkle.Hash) (merkle.Tree, error) {
			tree := merkle.NewDynTree(len(leaves) * 2)
			for _, leaf := range leaves {
				tree.Add(leaf)
			}
			return tree, nil
		},
	}.RunTests(t)
}

func BenchmarkDynTree(b *testing.B) {
	treetestcases.Factory{
		New: func(leaves []merkle.Hash) (merkle.Tree, error) {
			tree := merkle.NewDynTree(len(leaves) * 2)
			for _, leaf := range leaves {
				tree.Add(leaf)
			}
			return tree, nil
		},
	}.RunBenchmarks(b)
}
