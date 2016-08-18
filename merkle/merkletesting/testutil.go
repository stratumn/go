// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package merkletesting contains helpers to test Merkle trees.
package merkletesting

import (
	"math/rand"

	"github.com/stratumn/goprivate/merkle"
)

var letters = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// RandomHash creates a random hash.
func RandomHash() (hash merkle.Hash) {
	for i := range hash {
		hash[i] = letters[rand.Intn(len(letters))]
	}
	return
}
