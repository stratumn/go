// Copyright 2017 Stratumn SAS. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package merkle_test

import (
	"testing"

	"github.com/stratumn/sdk/merkle"
	"github.com/stratumn/sdk/merkle/treetestcases"
	"github.com/stratumn/sdk/testutil"
	"github.com/stratumn/sdk/types"
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
