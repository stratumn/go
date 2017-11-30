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

// Package treetestcases contains test cases to test Merkle tree implementation.
package treetestcases

import (
	"crypto/sha256"
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"reflect"
	"testing"

	"github.com/stratumn/sdk/merkle"
	"github.com/stratumn/sdk/testutil"
	"github.com/stratumn/sdk/types"
)

var (
	pathA0     types.Path
	pathAB0    types.Path
	pathAB1    types.Path
	pathABC0   types.Path
	pathABC1   types.Path
	pathABC2   types.Path
	pathABCD0  types.Path
	pathABCD1  types.Path
	pathABCD2  types.Path
	pathABCD3  types.Path
	pathABCDE0 types.Path
	pathABCDE1 types.Path
	pathABCDE2 types.Path
	pathABCDE3 types.Path
	pathABCDE4 types.Path
)

func loadPath(filename string, path *types.Path) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	if err = json.Unmarshal(data, path); err != nil {
		panic(err)
	}
}

// LoadFixtures loads test fixtures and should be called before running the
// tests.
func LoadFixtures(testdatapath string) {
	// Load fixtures.
	loadPath(testdatapath+"/path-a-0.json", &pathA0)
	loadPath(testdatapath+"/path-ab-0.json", &pathAB0)
	loadPath(testdatapath+"/path-ab-1.json", &pathAB1)
	loadPath(testdatapath+"/path-abc-0.json", &pathABC0)
	loadPath(testdatapath+"/path-abc-1.json", &pathABC1)
	loadPath(testdatapath+"/path-abc-2.json", &pathABC2)
	loadPath(testdatapath+"/path-abcd-0.json", &pathABCD0)
	loadPath(testdatapath+"/path-abcd-1.json", &pathABCD1)
	loadPath(testdatapath+"/path-abcd-2.json", &pathABCD2)
	loadPath(testdatapath+"/path-abcd-3.json", &pathABCD3)
	loadPath(testdatapath+"/path-abcde-0.json", &pathABCDE0)
	loadPath(testdatapath+"/path-abcde-1.json", &pathABCDE1)
	loadPath(testdatapath+"/path-abcde-2.json", &pathABCDE2)
	loadPath(testdatapath+"/path-abcde-3.json", &pathABCDE3)
	loadPath(testdatapath+"/path-abcde-4.json", &pathABCDE4)
}

// Factory wraps functions to allocate and free a Merkle tree, and is used to
// run the tests on a Merkle tree implementation.
type Factory struct {
	// New create a Merkle tree from leaves.
	New func(leaves []types.Bytes32) (merkle.Tree, error)

	// Free is an optional function to free a Merkle tree.
	Free func(tree merkle.Tree)
}

// RunTests runs all the tests.
func (f Factory) RunTests(t *testing.T) {
	t.Run("NumLeaves", f.TestNumLeaves)
	t.Run("Root", f.TestRoot)
	t.Run("Leaf", f.TestLeaf)
	t.Run("Path", f.TestPath)
	t.Run("PathRandom", f.TestPathRandom)
}

// RunBenchmarks runs all the benchmarks.
func (f Factory) RunBenchmarks(b *testing.B) {
	b.Run("Create", f.BenchmarkCreate)
	b.Run("Path", f.BenchmarkPath)
}

func (f Factory) free(tree merkle.Tree) {
	if f.Free != nil {
		f.Free(tree)
	}
}

// TestNumLeaves tests that the implementation returns the correct number of
// leaves.
func (f Factory) TestNumLeaves(t *testing.T) {
	tree, err := f.New([]types.Bytes32{*testutil.RandomHash(), *testutil.RandomHash(), *testutil.RandomHash()})
	if err != nil {
		t.Fatalf("f.New(): err: %s", err)
	}
	defer f.free(tree)

	if got, want := tree.LeavesLen(), 3; got != want {
		t.Errorf("tree.LeavesLen() = %d want %d", got, want)
	}
}

// TestRoot tests that the implementation computes the root correctly.
func (f Factory) TestRoot(t *testing.T) {
	tests := [...]struct {
		leaves   []string
		expected string
	}{
		{[]string{"a"}, "ca978112ca1bbdcafac231b39a23dc4da786eff8147c4e72b9807785afee48bb"},
		{[]string{"a", "b"}, "e5a01fee14e0ed5c48714f22180f25ad8365b53f9779f79dc4a3d7e93963f94a"},
		{[]string{"a", "b", "c"}, "7075152d03a5cd92104887b476862778ec0c87be5c2fa1c0a90f87c49fad6eff"},
		{[]string{"a", "b", "c", "d"}, "14ede5e8e97ad9372327728f5099b95604a39593cac3bd38a343ad76205213e7"},
		{[]string{"a", "b", "c", "d", "e"}, "d71f8983ad4ee170f8129f1ebcdd7440be7798d8e1c80420bf11f1eced610dba"},
		{[]string{"a", "b", "c", "d", "e", "f"}, "1f7379539707bcaea00564168d1d4d626b09b73f8a2a365234c62d763f854da2"},
		{[]string{"a", "b", "c", "d", "e", "f", "g"}, "e2a80e0e872a6c6eaed37b4c1f220e1935004805585b5f99617e48e9c8fe4034"},
		{[]string{"a", "b", "c", "d", "e", "f", "g", "h"}, "bd7c8a900be9b67ba7df5c78a652a8474aedd78adb5083e80e49d9479138a23f"},
		{[]string{"a", "b", "c", "d", "e", "f", "g", "h", "i"}, "386ced54bdc7456fecfc9b43018bbda2fe0a105f4cf7cad6bbb429c18fe852cc"},
	}

	for i, test := range tests {
		var leaves = make([]types.Bytes32, len(test.leaves), len(test.leaves))
		for i, s := range test.leaves {
			leaves[i] = sha256.Sum256([]byte(s))
		}

		tree, err := f.New(leaves)
		if err != nil {
			t.Fatalf("f.New(): err: %s", err)
		}
		defer f.free(tree)

		if got, want := tree.Root().String(), test.expected; got != want {
			t.Errorf("test#%d: tree.Root() = %q want %q", i, got, want)
		}
	}
}

// TestLeaf tests that the implementation correctly returns leaves.
func (f Factory) TestLeaf(t *testing.T) {
	for i := 1; i < 128; i++ {
		var leaves []types.Bytes32
		for j := 0; j < i; j++ {
			leaves = append(leaves, *testutil.RandomHash())
		}

		tree, err := f.New(leaves)
		if err != nil {
			t.Fatalf("f.New(): err: %s", err)
		}
		defer f.free(tree)

		for j := 0; j < i; j++ {
			if got, want := tree.Leaf(j).String(), leaves[j].String(); got != want {
				t.Errorf("test#%d: tree.Leaf(%d) = %q want %q", i, j, got, want)
			}
		}
	}
}

// TestPath tests that the implementation correctly computes paths.
func (f Factory) TestPath(t *testing.T) {
	tests := [...]struct {
		leaves   []string
		expected []types.Path
	}{
		{
			[]string{"a"},
			[]types.Path{pathA0},
		},
		{
			[]string{"a", "b"},
			[]types.Path{pathAB0, pathAB1},
		},
		{
			[]string{"a", "b", "c"},
			[]types.Path{pathABC0, pathABC1, pathABC2},
		},
		{
			[]string{"a", "b", "c", "d"},
			[]types.Path{pathABCD0, pathABCD1, pathABCD2, pathABCD3},
		},
		{
			[]string{"a", "b", "c", "d", "e"},
			[]types.Path{pathABCDE0, pathABCDE1, pathABCDE2, pathABCDE3, pathABCDE4},
		},
	}

	for i, test := range tests {
		var leaves = make([]types.Bytes32, len(test.leaves), len(test.leaves))
		for i, s := range test.leaves {
			leaves[i] = sha256.Sum256([]byte(s))
		}

		tree, err := f.New(leaves)
		if err != nil {
			t.Fatalf("f.New(): err: %s", err)
		}
		defer f.free(tree)

		for j := range test.leaves {
			var (
				got  = tree.Path(j)
				want = test.expected[j]
			)
			if !reflect.DeepEqual(got, want) {
				g, _ := json.MarshalIndent(got, "", "  ")
				w, _ := json.MarshalIndent(want, "", "  ")
				t.Errorf("test#%d: tree.Path(%d) = %s\nwant %s", i, j, g, w)
			}
		}
	}
}

// TestPathRandom tests that the implementation correctly computes paths given
// random trees.
func (f Factory) TestPathRandom(t *testing.T) {
	for i := 0; i < 10; i++ {
		tests := make([]types.Bytes32, 2+rand.Intn(10000))
		for j := range tests {
			tests[j] = *testutil.RandomHash()
		}

		tree, err := f.New(tests)
		if err != nil {
			t.Fatalf("f.New(): err: %s", err)
		}
		defer f.free(tree)

		for j := range tests {
			path := tree.Path(j)
			if err := path.Validate(); err != nil {
				t.Errorf("path.Validate(): err: %s", err)
			}

			if got, want := path[len(path)-1].Parent.String(), tree.Root().String(); got != want {
				t.Errorf("test#%d: tree.Path(%d) last parent = %q want %q", i, j, got, want)
			}
		}
	}
}

// BenchmarkCreateWithSize benchmarks creating trees of given size.
func (f Factory) BenchmarkCreateWithSize(b *testing.B, size int) {
	leaves := make([]types.Bytes32, size)
	for i := 0; i < size; i++ {
		leaves[i] = *testutil.RandomHash()
	}

	b.ResetTimer()
	log.SetOutput(ioutil.Discard)

	for i := 0; i < b.N; i++ {
		tree, err := f.New(leaves)
		if err != nil {
			b.Errorf("f.New(): err: %s", err)
		}
		defer f.free(tree)
	}
}

// BenchmarkCreate benchmarks creating trees of different sizes.
func (f Factory) BenchmarkCreate(b *testing.B) {
	b.Run("10-leaves", func(b *testing.B) { f.BenchmarkCreateWithSize(b, 10) })
	b.Run("100-leaves", func(b *testing.B) { f.BenchmarkCreateWithSize(b, 100) })
	b.Run("1000-leaves", func(b *testing.B) { f.BenchmarkCreateWithSize(b, 1000) })
	b.Run("10000-leaves", func(b *testing.B) { f.BenchmarkCreateWithSize(b, 10000) })
	b.Run("100000-leaves", func(b *testing.B) { f.BenchmarkCreateWithSize(b, 100000) })
}

// BenchmarkPathWithSize benchmarks computing paths for trees of given size.
func (f Factory) BenchmarkPathWithSize(b *testing.B, size int) {
	leaves := make([]types.Bytes32, size)
	for i := 0; i < size; i++ {
		leaves[i] = *testutil.RandomHash()
	}

	tree, err := f.New(leaves)
	if err != nil {
		b.Fatalf("f.New(): err: %s", err)
	}
	defer f.free(tree)

	b.ResetTimer()
	log.SetOutput(ioutil.Discard)

	for i := 0; i < b.N; i++ {
		tree.Path(i % size)
	}
}

// BenchmarkPath benchmarks computing paths for different sizes.
func (f Factory) BenchmarkPath(b *testing.B) {
	b.Run("10-leaves", func(b *testing.B) { f.BenchmarkPathWithSize(b, 10) })
	b.Run("100-leaves", func(b *testing.B) { f.BenchmarkPathWithSize(b, 100) })
	b.Run("1000-leaves", func(b *testing.B) { f.BenchmarkPathWithSize(b, 1000) })
	b.Run("10000-leaves", func(b *testing.B) { f.BenchmarkPathWithSize(b, 10000) })
	b.Run("100000-leaves", func(b *testing.B) { f.BenchmarkPathWithSize(b, 100000) })
}

func atos(a types.Bytes32) []byte {
	return a[:]
}
