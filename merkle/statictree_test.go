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

func TestNewStaticTree(t *testing.T) {
	tree, err := merkle.NewStaticTree([]types.Bytes32{testutil.RandomHash()})
	if err != nil {
		t.Fatal(err)
	}
	if tree == nil {
		t.Fatal("expected tree not to be nil")
	}
}

func TestNewStaticTreeNoLeaves(t *testing.T) {
	_, err := merkle.NewStaticTree(nil)
	if err == nil {
		t.Fatal("expected error not to be nil")
	}
	if err.Error() != "tree should have at least one leaf" {
		t.Log(err)
		t.Fatal("unexpected error message")
	}
}

func TestStaticTree(t *testing.T) {
	treetestcases.Factory{
		New: func(leaves []types.Bytes32) (merkle.Tree, error) {
			return merkle.NewStaticTree(leaves)
		},
	}.RunTests(t)
}

func BenchmarkStaticTree(b *testing.B) {
	treetestcases.Factory{
		New: func(leaves []types.Bytes32) (merkle.Tree, error) {
			return merkle.NewStaticTree(leaves)
		},
	}.RunBenchmarks(b)
}
