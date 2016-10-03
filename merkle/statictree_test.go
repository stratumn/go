// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by the license that can be found in the
// LICENSE file.

package merkle_test

import (
	"testing"

	"github.com/stratumn/go/types"
	"github.com/stratumn/goprivate/merkle"
	"github.com/stratumn/goprivate/merkle/treetestcases"
)

func TestNewStaticTree_noLeaves(t *testing.T) {
	_, err := merkle.NewStaticTree(nil)
	if err == nil {
		t.Error("NewStaticTree(): err = nil want Error")
	}
	if got, want := err.Error(), "tree should have at least one leaf"; got != want {
		t.Errorf("NewStaticTree(): err.Error() = %q want %q", got, want)
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
