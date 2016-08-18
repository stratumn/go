// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package merkle

import (
	"crypto/sha256"
	"encoding/hex"
	"io/ioutil"
	"math/rand"
	"reflect"
	"testing"
)

func TestNewStaticTree(t *testing.T) {
	tree, err := NewStaticTree([]Hash{randomHash()})
	if err != nil {
		t.Fatal(err)
	}
	if tree == nil {
		t.Fatal("expected tree not to be nil")
	}

	_ = Tree(tree)
}

func TestNewStaticTreeNoLeaves(t *testing.T) {
	_, err := NewStaticTree(nil)
	if err == nil {
		t.Fatal("expected error not to be nil")
	}
	if err.Error() != "tree should have at least one leaf" {
		t.Log(err)
		t.Fatal("unexpected error message")
	}
}

func TestStaticTreeNumLeaves(t *testing.T) {
	tree, err := NewStaticTree([]Hash{randomHash(), randomHash(), randomHash()})
	if err != nil {
		t.Fatal(err)
	}
	if tree == nil {
		t.Fatal("expected tree not to be nil")
	}

	var (
		a = tree.NumLeaves()
		e = 3
	)
	if a != e {
		t.Logf("actual: %d; expected: %d\n", a, e)
		t.Error("unexpected number of leaves")
	}
}

func TestStaticTreeRoot(t *testing.T) {
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
		var leaves = make([]Hash, len(row.leaves), len(row.leaves))
		for i, s := range row.leaves {
			leaves[i] = sha256.Sum256([]byte(s))
		}

		tree, err := NewStaticTree(leaves)
		if err != nil {
			t.Fatal(err)
		}
		if tree == nil {
			t.Fatal("expected tree not to be nil")
		}

		a := hex.EncodeToString(atos(tree.Root()))
		e := row.expected

		if a != e {
			t.Logf("actual: %s; expected: %s\n", a, e)
			t.Error("unexpected root")
		}
	}
}

func TestStaticTreeRootVulnerability1(t *testing.T) {
	tree1, err := NewStaticTree([]Hash{
		sha256.Sum256([]byte("a")),
		sha256.Sum256([]byte("b")),
		sha256.Sum256([]byte("c")),
	})
	if err != nil {
		t.Fatal(err)
	}
	if tree1 == nil {
		t.Fatal("expected tree not to be nil")
	}

	tree2, err := NewStaticTree([]Hash{
		sha256.Sum256([]byte("a")),
		sha256.Sum256([]byte("b")),
		sha256.Sum256([]byte("c")),
		sha256.Sum256([]byte("c")),
	})
	if err != nil {
		t.Fatal(err)
	}
	if tree2 == nil {
		t.Fatal("expected tree not to be nil")
	}

	root1 := hex.EncodeToString(atos(tree1.Root()))
	root2 := hex.EncodeToString(atos(tree2.Root()))

	if root1 == root2 {
		t.Log(root1)
		t.Error("expected root to be different")
	}
}

func TestStaticTreeRootVulnerability2(t *testing.T) {
	tree1, err := NewStaticTree([]Hash{
		sha256.Sum256([]byte("a")),
		sha256.Sum256([]byte("b")),
		sha256.Sum256([]byte("c")),
	})
	if err != nil {
		t.Fatal(err)
	}
	if tree1 == nil {
		t.Fatal("expected tree not to be nil")
	}

	tree2, err := NewStaticTree([]Hash{
		sha256.Sum256([]byte("a")),
		sha256.Sum256([]byte("b")),
		sha256.Sum256([]byte("c")),
		sha256.Sum256([]byte("")),
	})
	if err != nil {
		t.Fatal(err)
	}
	if tree2 == nil {
		t.Fatal("expected tree not to be nil")
	}

	root1 := hex.EncodeToString(atos(tree1.Root()))
	root2 := hex.EncodeToString(atos(tree2.Root()))

	if root1 == root2 {
		t.Log(root1)
		t.Error("expected root to be different")
	}
}

func TestStaticTreeLeaf(t *testing.T) {
	for i := 1; i < 128; i++ {
		var leaves []Hash
		for j := 0; j < i; j++ {
			leaves = append(leaves, randomHash())
		}

		tree, err := NewStaticTree(leaves)
		if err != nil {
			t.Fatal(err)
		}
		if tree == nil {
			t.Fatal("expected tree not to be nil")
		}

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

func TestStaticTreePath(t *testing.T) {
	grid := [...]struct {
		leaves   []string
		expected [][]string
	}{
		{
			[]string{"a"},
			[][]string{
				[]string{},
			},
		},
		{
			[]string{"a", "b"},
			[][]string{
				[]string{
					"ca978112ca1bbdcafac231b39a23dc4da786eff8147c4e72b9807785afee48bb",
					"3e23e8160039594a33894f6564e1b1348bbd7a0088d42c4acb73eeaed59c009d",
					"e5a01fee14e0ed5c48714f22180f25ad8365b53f9779f79dc4a3d7e93963f94a",
				},
				[]string{
					"ca978112ca1bbdcafac231b39a23dc4da786eff8147c4e72b9807785afee48bb",
					"3e23e8160039594a33894f6564e1b1348bbd7a0088d42c4acb73eeaed59c009d",
					"e5a01fee14e0ed5c48714f22180f25ad8365b53f9779f79dc4a3d7e93963f94a",
				},
			},
		},
		{
			[]string{"a", "b", "c"},
			[][]string{
				[]string{
					"ca978112ca1bbdcafac231b39a23dc4da786eff8147c4e72b9807785afee48bb",
					"3e23e8160039594a33894f6564e1b1348bbd7a0088d42c4acb73eeaed59c009d",
					"e5a01fee14e0ed5c48714f22180f25ad8365b53f9779f79dc4a3d7e93963f94a",

					"e5a01fee14e0ed5c48714f22180f25ad8365b53f9779f79dc4a3d7e93963f94a",
					"2e7d2c03a9507ae265ecf5b5356885a53393a2029d241394997265a1a25aefc6",
					"7075152d03a5cd92104887b476862778ec0c87be5c2fa1c0a90f87c49fad6eff",
				},
				[]string{
					"ca978112ca1bbdcafac231b39a23dc4da786eff8147c4e72b9807785afee48bb",
					"3e23e8160039594a33894f6564e1b1348bbd7a0088d42c4acb73eeaed59c009d",
					"e5a01fee14e0ed5c48714f22180f25ad8365b53f9779f79dc4a3d7e93963f94a",

					"e5a01fee14e0ed5c48714f22180f25ad8365b53f9779f79dc4a3d7e93963f94a",
					"2e7d2c03a9507ae265ecf5b5356885a53393a2029d241394997265a1a25aefc6",
					"7075152d03a5cd92104887b476862778ec0c87be5c2fa1c0a90f87c49fad6eff",
				},
				[]string{
					"e5a01fee14e0ed5c48714f22180f25ad8365b53f9779f79dc4a3d7e93963f94a",
					"2e7d2c03a9507ae265ecf5b5356885a53393a2029d241394997265a1a25aefc6",
					"7075152d03a5cd92104887b476862778ec0c87be5c2fa1c0a90f87c49fad6eff",
				},
			},
		},
		{
			[]string{"a", "b", "c", "d"},
			[][]string{
				[]string{
					"ca978112ca1bbdcafac231b39a23dc4da786eff8147c4e72b9807785afee48bb",
					"3e23e8160039594a33894f6564e1b1348bbd7a0088d42c4acb73eeaed59c009d",
					"e5a01fee14e0ed5c48714f22180f25ad8365b53f9779f79dc4a3d7e93963f94a",

					"e5a01fee14e0ed5c48714f22180f25ad8365b53f9779f79dc4a3d7e93963f94a",
					"bffe0b34dba16bc6fac17c08bac55d676cded5a4ade41fe2c9924a5dde8f3e5b",
					"14ede5e8e97ad9372327728f5099b95604a39593cac3bd38a343ad76205213e7",
				},
				[]string{
					"ca978112ca1bbdcafac231b39a23dc4da786eff8147c4e72b9807785afee48bb",
					"3e23e8160039594a33894f6564e1b1348bbd7a0088d42c4acb73eeaed59c009d",
					"e5a01fee14e0ed5c48714f22180f25ad8365b53f9779f79dc4a3d7e93963f94a",

					"e5a01fee14e0ed5c48714f22180f25ad8365b53f9779f79dc4a3d7e93963f94a",
					"bffe0b34dba16bc6fac17c08bac55d676cded5a4ade41fe2c9924a5dde8f3e5b",
					"14ede5e8e97ad9372327728f5099b95604a39593cac3bd38a343ad76205213e7",
				},
				[]string{
					"2e7d2c03a9507ae265ecf5b5356885a53393a2029d241394997265a1a25aefc6",
					"18ac3e7343f016890c510e93f935261169d9e3f565436429830faf0934f4f8e4",
					"bffe0b34dba16bc6fac17c08bac55d676cded5a4ade41fe2c9924a5dde8f3e5b",

					"e5a01fee14e0ed5c48714f22180f25ad8365b53f9779f79dc4a3d7e93963f94a",
					"bffe0b34dba16bc6fac17c08bac55d676cded5a4ade41fe2c9924a5dde8f3e5b",
					"14ede5e8e97ad9372327728f5099b95604a39593cac3bd38a343ad76205213e7",
				},
				[]string{
					"2e7d2c03a9507ae265ecf5b5356885a53393a2029d241394997265a1a25aefc6",
					"18ac3e7343f016890c510e93f935261169d9e3f565436429830faf0934f4f8e4",
					"bffe0b34dba16bc6fac17c08bac55d676cded5a4ade41fe2c9924a5dde8f3e5b",

					"e5a01fee14e0ed5c48714f22180f25ad8365b53f9779f79dc4a3d7e93963f94a",
					"bffe0b34dba16bc6fac17c08bac55d676cded5a4ade41fe2c9924a5dde8f3e5b",
					"14ede5e8e97ad9372327728f5099b95604a39593cac3bd38a343ad76205213e7",
				},
			},
		},
		{
			[]string{"a", "b", "c", "d", "e"},
			[][]string{
				[]string{
					"ca978112ca1bbdcafac231b39a23dc4da786eff8147c4e72b9807785afee48bb",
					"3e23e8160039594a33894f6564e1b1348bbd7a0088d42c4acb73eeaed59c009d",
					"e5a01fee14e0ed5c48714f22180f25ad8365b53f9779f79dc4a3d7e93963f94a",

					"e5a01fee14e0ed5c48714f22180f25ad8365b53f9779f79dc4a3d7e93963f94a",
					"bffe0b34dba16bc6fac17c08bac55d676cded5a4ade41fe2c9924a5dde8f3e5b",
					"14ede5e8e97ad9372327728f5099b95604a39593cac3bd38a343ad76205213e7",

					"14ede5e8e97ad9372327728f5099b95604a39593cac3bd38a343ad76205213e7",
					"3f79bb7b435b05321651daefd374cdc681dc06faa65e374e38337b88ca046dea",
					"d71f8983ad4ee170f8129f1ebcdd7440be7798d8e1c80420bf11f1eced610dba",
				},
				[]string{
					"ca978112ca1bbdcafac231b39a23dc4da786eff8147c4e72b9807785afee48bb",
					"3e23e8160039594a33894f6564e1b1348bbd7a0088d42c4acb73eeaed59c009d",
					"e5a01fee14e0ed5c48714f22180f25ad8365b53f9779f79dc4a3d7e93963f94a",

					"e5a01fee14e0ed5c48714f22180f25ad8365b53f9779f79dc4a3d7e93963f94a",
					"bffe0b34dba16bc6fac17c08bac55d676cded5a4ade41fe2c9924a5dde8f3e5b",
					"14ede5e8e97ad9372327728f5099b95604a39593cac3bd38a343ad76205213e7",

					"14ede5e8e97ad9372327728f5099b95604a39593cac3bd38a343ad76205213e7",
					"3f79bb7b435b05321651daefd374cdc681dc06faa65e374e38337b88ca046dea",
					"d71f8983ad4ee170f8129f1ebcdd7440be7798d8e1c80420bf11f1eced610dba",
				},
				[]string{
					"2e7d2c03a9507ae265ecf5b5356885a53393a2029d241394997265a1a25aefc6",
					"18ac3e7343f016890c510e93f935261169d9e3f565436429830faf0934f4f8e4",
					"bffe0b34dba16bc6fac17c08bac55d676cded5a4ade41fe2c9924a5dde8f3e5b",

					"e5a01fee14e0ed5c48714f22180f25ad8365b53f9779f79dc4a3d7e93963f94a",
					"bffe0b34dba16bc6fac17c08bac55d676cded5a4ade41fe2c9924a5dde8f3e5b",
					"14ede5e8e97ad9372327728f5099b95604a39593cac3bd38a343ad76205213e7",

					"14ede5e8e97ad9372327728f5099b95604a39593cac3bd38a343ad76205213e7",
					"3f79bb7b435b05321651daefd374cdc681dc06faa65e374e38337b88ca046dea",
					"d71f8983ad4ee170f8129f1ebcdd7440be7798d8e1c80420bf11f1eced610dba",
				},
				[]string{
					"2e7d2c03a9507ae265ecf5b5356885a53393a2029d241394997265a1a25aefc6",
					"18ac3e7343f016890c510e93f935261169d9e3f565436429830faf0934f4f8e4",
					"bffe0b34dba16bc6fac17c08bac55d676cded5a4ade41fe2c9924a5dde8f3e5b",

					"e5a01fee14e0ed5c48714f22180f25ad8365b53f9779f79dc4a3d7e93963f94a",
					"bffe0b34dba16bc6fac17c08bac55d676cded5a4ade41fe2c9924a5dde8f3e5b",
					"14ede5e8e97ad9372327728f5099b95604a39593cac3bd38a343ad76205213e7",

					"14ede5e8e97ad9372327728f5099b95604a39593cac3bd38a343ad76205213e7",
					"3f79bb7b435b05321651daefd374cdc681dc06faa65e374e38337b88ca046dea",
					"d71f8983ad4ee170f8129f1ebcdd7440be7798d8e1c80420bf11f1eced610dba",
				},
				[]string{
					"14ede5e8e97ad9372327728f5099b95604a39593cac3bd38a343ad76205213e7",
					"3f79bb7b435b05321651daefd374cdc681dc06faa65e374e38337b88ca046dea",
					"d71f8983ad4ee170f8129f1ebcdd7440be7798d8e1c80420bf11f1eced610dba",
				},
			},
		},
	}

	for _, row := range grid {
		var leaves = make([]Hash, len(row.leaves), len(row.leaves))
		for i, s := range row.leaves {
			leaves[i] = sha256.Sum256([]byte(s))
		}

		tree, err := NewStaticTree(leaves)
		if err != nil {
			t.Fatal(err)
		}
		if tree == nil {
			t.Fatal("expected tree not to be nil")
		}

		for i := range row.leaves {
			var (
				path = tree.Path(i)
				e    = row.expected[i]
				a    []string
			)

			for _, p := range path {
				a = append(a, hex.EncodeToString(p.Left[:]), hex.EncodeToString(p.Right[:]), hex.EncodeToString(p.Parent[:]))
			}

			if !(len(e) == 0 && len(a) == 0) && !reflect.DeepEqual(e, a) {
				t.Logf("actual: %v; expected: %v\n", a, e)
				t.Error("unexpected root")
			}
		}
	}
}

func TestStaticTreeReader(t *testing.T) {
	tree, err := NewStaticTree([]Hash{
		sha256.Sum256([]byte("a")),
		sha256.Sum256([]byte("b")),
	})
	if err != nil {
		t.Fatal(err)
	}
	if tree == nil {
		t.Fatal("expected tree not to be nil")
	}

	b, err := ioutil.ReadAll(tree)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(b[:HashByteLen], atos(tree.Root())) {
		t.Error("unexpected data")
	}

	for i := 0; i < 2; i++ {
		offset := (i + 1) * HashByteLen
		if !reflect.DeepEqual(b[offset:offset+HashByteLen], atos(tree.Leaf(i))) {
			t.Error("unexpected data")
		}
	}
}

func TestNumStaticTreeNodes(t *testing.T) {
	grid := [...]int{
		1, 1,
		2, 3,
		3, 5,
		4, 7,
		5, 9,
		6, 11,
		7, 13,
		8, 15,
	}

	for i := 0; i < len(grid); i += 2 {
		a := numStaticTreeNodes(grid[i])
		e := grid[i+1]
		if a != e {
			t.Logf("actual: %d; expected: %d\n", a, e)
			t.Error("unexpected buffer length")
		}
	}
}

func TestStaticTreeLevelsLen(t *testing.T) {
	grid := [...]struct {
		numLeaves int
		expected  []int
	}{
		{1, []int{1}},
		{2, []int{1, 2}},
		{3, []int{1, 1, 3}},
		{4, []int{1, 2, 4}},
		{5, []int{1, 1, 2, 5}},
		{6, []int{1, 1, 3, 6}},
		{7, []int{1, 2, 3, 7}},
		{8, []int{1, 2, 4, 8}},
		{9, []int{1, 1, 2, 4, 9}},
		{10, []int{1, 1, 2, 5, 10}},
		{11, []int{1, 1, 3, 5, 11}},
		{12, []int{1, 1, 3, 6, 12}},
		{13, []int{1, 2, 3, 6, 13}},
		{14, []int{1, 2, 3, 7, 14}},
		{15, []int{1, 2, 4, 7, 15}},
		{16, []int{1, 2, 4, 8, 16}},
	}

	for _, row := range grid {
		a := staticTreeLevelsLen(row.numLeaves)
		e := row.expected
		if !reflect.DeepEqual(a, e) {
			t.Logf("actual: %v; expected: %v\n", a, e)
			t.Error("unexpected level lengths")
		}
	}
}

func benchmarkNewStaticTree(size int, b *testing.B) {
	leaves := make([]Hash, size)
	for i := 0; i < size; i++ {
		leaves[i] = randomHash()
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := NewStaticTree(leaves); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkNewStaticTree10(b *testing.B) {
	benchmarkNewStaticTree(10, b)
}

func BenchmarkNewStaticTree100(b *testing.B) {
	benchmarkNewStaticTree(100, b)
}

func BenchmarkNewStaticTree1000(b *testing.B) {
	benchmarkNewStaticTree(1000, b)
}

func BenchmarkNewStaticTree10000(b *testing.B) {
	benchmarkNewStaticTree(10000, b)
}

func BenchmarkNewStaticTree100000(b *testing.B) {
	benchmarkNewStaticTree(100000, b)
}

func BenchmarkNewStaticTree1000000(b *testing.B) {
	benchmarkNewStaticTree(1000000, b)
}

func BenchmarkSha256_1000000(b *testing.B) {
	size := 1000000
	leaves := make([]Hash, size)
	for i := 0; i < size; i++ {
		leaves[i] = randomHash()
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for j := 0; j < size; j++ {
			sha256.Sum256(leaves[j][:])
		}
	}
}

func benchmarkStaticTreePath(size int, b *testing.B) {
	leaves := make([]Hash, size)
	for i := 0; i < size; i++ {
		leaves[i] = randomHash()
	}

	tree, err := NewStaticTree(leaves)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tree.Path(i % size)
	}
}

func BenchmarkStaticTreePath10(b *testing.B) {
	benchmarkStaticTreePath(10, b)
}

func BenchmarkStaticTreePath100(b *testing.B) {
	benchmarkStaticTreePath(100, b)
}

func BenchmarkStaticTreePath1000(b *testing.B) {
	benchmarkStaticTreePath(1000, b)
}

func BenchmarkStaticTreePath10000(b *testing.B) {
	benchmarkStaticTreePath(10000, b)
}

func BenchmarkStaticTreePath100000(b *testing.B) {
	benchmarkStaticTreePath(100000, b)
}

func BenchmarkStaticTreePath1000000(b *testing.B) {
	benchmarkStaticTreePath(1000000, b)
}

var letters = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randomHash() (hash Hash) {
	for i := range hash {
		hash[i] = letters[rand.Intn(len(letters))]
	}
	return
}

func atos(a Hash) []byte {
	return a[:]
}
