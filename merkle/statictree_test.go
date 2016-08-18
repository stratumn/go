// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package merkle_test

import (
	"flag"
	"os"
	"testing"

	"github.com/stratumn/goprivate/merkle"
	"github.com/stratumn/goprivate/merkle/treetestcases"
)

func TestMain(m *testing.M) {
	treetestcases.LoadFixtures("treetestcases/testdata")
	flag.Parse()
	os.Exit(m.Run())
}

func TestNewStaticTree(t *testing.T) {
	tree, err := merkle.NewStaticTree([]merkle.Hash{treetestcases.RandomHash()})
	if err != nil {
		t.Fatal(err)
	}
	if tree == nil {
		t.Fatal("expected tree not to be nil")
	}

	// Compiling will fail if interface is not implemented.
	_ = merkle.Tree(tree)
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
		New: func(leaves []merkle.Hash) (merkle.Tree, error) {
			return merkle.NewStaticTree(leaves)
		},
	}.RunTests(t)
}

func BenchmarkStaticTree(b *testing.B) {
	treetestcases.Factory{
		New: func(leaves []merkle.Hash) (merkle.Tree, error) {
			return merkle.NewStaticTree(leaves)
		},
	}.RunBenchmarks(b)
}
