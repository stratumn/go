// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package treetestcases contains test cases to test Merkle tree implementation.
package treetestcases

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/stratumn/goprivate/merkle"
	"github.com/stratumn/goprivate/merkle/merkletesting"
)

var (
	pathA0     merkle.Path
	pathAB0    merkle.Path
	pathAB1    merkle.Path
	pathABC0   merkle.Path
	pathABC1   merkle.Path
	pathABC2   merkle.Path
	pathABCD0  merkle.Path
	pathABCD1  merkle.Path
	pathABCD2  merkle.Path
	pathABCD3  merkle.Path
	pathABCDE0 merkle.Path
	pathABCDE1 merkle.Path
	pathABCDE2 merkle.Path
	pathABCDE3 merkle.Path
	pathABCDE4 merkle.Path
)

func loadPath(filename string, path *merkle.Path) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	if err = json.Unmarshal(data, path); err != nil {
		panic(err)
	}
}

// LoadFixtures loads test fixtures and should be called before running the tests.
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

// Factory contains functions to allocate and free a Merkle tree.
type Factory struct {
	// New create a Merkle tree from leaves.
	New func(leaves []merkle.Hash) (merkle.Tree, error)

	// Free is an optional function to free a Merkle tree.
	Free func(tree merkle.Tree)
}

// RunTests runs all the tests.
func (f Factory) RunTests(t *testing.T) {
	t.Run("NumLeaves", f.TestNumLeaves)
	t.Run("Root", f.TestRoot)
	t.Run("Leaf", f.TestLeaf)
	t.Run("Path", f.TestPath)
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

// TestNumLeaves tests that the implementation returns the correct number of leaves.
func (f Factory) TestNumLeaves(t *testing.T) {
	tree, err := f.New([]merkle.Hash{merkletesting.RandomHash(), merkletesting.RandomHash(), merkletesting.RandomHash()})
	if err != nil {
		t.Fatal(err)
	}
	if tree == nil {
		t.Fatal("expected tree not to be nil")
	}
	defer f.free(tree)

	var (
		a = tree.NumLeaves()
		e = 3
	)
	if a != e {
		t.Logf("actual: %d; expected: %d\n", a, e)
		t.Error("unexpected number of leaves")
	}
}

// TestRoot tests that the implementation computes the root correctly.
func (f Factory) TestRoot(t *testing.T) {
	grid := [...]struct {
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

	for _, row := range grid {
		var leaves = make([]merkle.Hash, len(row.leaves), len(row.leaves))
		for i, s := range row.leaves {
			leaves[i] = sha256.Sum256([]byte(s))
		}

		tree, err := f.New(leaves)
		if err != nil {
			t.Fatal(err)
		}
		if tree == nil {
			t.Fatal("expected tree not to be nil")
		}
		defer f.free(tree)

		a := hex.EncodeToString(atos(tree.Root()))
		e := row.expected

		if a != e {
			t.Logf("actual: %s; expected: %s\n", a, e)
			t.Error("unexpected root")
		}
	}
}

// TestLeaf tests that the implementation correctly returns leaves.
func (f Factory) TestLeaf(t *testing.T) {
	for i := 1; i < 128; i++ {
		var leaves []merkle.Hash
		for j := 0; j < i; j++ {
			leaves = append(leaves, merkletesting.RandomHash())
		}

		tree, err := f.New(leaves)
		if err != nil {
			t.Fatal(err)
		}
		if tree == nil {
			t.Fatal("expected tree not to be nil")
		}
		defer f.free(tree)

		for j := 0; j < i; j++ {
			a := tree.Leaf(j)
			e := leaves[j]
			if !reflect.DeepEqual(a, e) {
				t.Logf("actual: %s; expected: %s\n", a, e)
				t.Error("unexpected leaf")
			}
		}
	}
}

// TestPath tests that the implementation correctly computes paths.
func (f Factory) TestPath(t *testing.T) {
	grid := [...]struct {
		leaves   []string
		expected []merkle.Path
	}{
		{
			[]string{"a"},
			[]merkle.Path{pathA0},
		},
		{
			[]string{"a", "b"},
			[]merkle.Path{pathAB0, pathAB1},
		},
		{
			[]string{"a", "b", "c"},
			[]merkle.Path{pathABC0, pathABC1, pathABC2},
		},
		{
			[]string{"a", "b", "c", "d"},
			[]merkle.Path{pathABCD0, pathABCD1, pathABCD2, pathABCD3},
		},
		{
			[]string{"a", "b", "c", "d", "e"},
			[]merkle.Path{pathABCDE0, pathABCDE1, pathABCDE2, pathABCDE3, pathABCDE4},
		},
	}

	for _, row := range grid {
		var leaves = make([]merkle.Hash, len(row.leaves), len(row.leaves))
		for i, s := range row.leaves {
			leaves[i] = sha256.Sum256([]byte(s))
		}

		tree, err := f.New(leaves)
		if err != nil {
			t.Fatal(err)
		}
		if tree == nil {
			t.Fatal("expected tree not to be nil")
		}
		defer f.free(tree)

		for i := range row.leaves {
			var (
				a = tree.Path(i)
				e = row.expected[i]
			)

			if !reflect.DeepEqual(e, a) {
				t.Logf("actual: %v; expected: %v\n", a, e)
				t.Error("unexpected root")
			}
		}
	}
}

// BenchmarkCreateWithSize benchmarks creating trees of given size.
func (f Factory) BenchmarkCreateWithSize(b *testing.B, size int) {
	leaves := make([]merkle.Hash, size)
	for i := 0; i < size; i++ {
		leaves[i] = merkletesting.RandomHash()
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tree, err := f.New(leaves)
		if err != nil {
			b.Fatal(err)
		}
		if tree == nil {
			b.Fatal("expected tree not to be nil")
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
}

// BenchmarkPathWithSize benchmarks computing paths for trees of given size.
func (f Factory) BenchmarkPathWithSize(b *testing.B, size int) {
	leaves := make([]merkle.Hash, size)
	for i := 0; i < size; i++ {
		leaves[i] = merkletesting.RandomHash()
	}

	tree, err := f.New(leaves)
	if err != nil {
		b.Fatal(err)
	}
	if tree == nil {
		b.Fatal("expected tree not to be nil")
	}
	defer f.free(tree)

	b.ResetTimer()

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
}

func atos(a merkle.Hash) []byte {
	return a[:]
}
