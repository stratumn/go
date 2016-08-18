// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"log"
	"math/rand"

	"github.com/pkg/profile"

	"github.com/stratumn/goprivate/merkle"
)

const size = 10000
const paths = 10000

func main() {
	leaves := make([]merkle.Hash, size)
	for i := 0; i < size; i++ {
		leaves[i] = randomHash()
	}

	tree, err := merkle.NewStaticTree(leaves)
	if err != nil {
		log.Fatal(err)
	}

	defer profile.Start().Stop()

	for i := 0; i < paths; i++ {
		tree.Path(i % size)
	}
}

var letters = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randomHash() (hash merkle.Hash) {
	for i := range hash {
		hash[i] = letters[rand.Intn(len(letters))]
	}
	return
}
