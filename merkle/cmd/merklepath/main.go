// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/stratumn/goprivate/merkle"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatal("unexpected number of arguments")
	}

	var (
		a = os.Args[1:]
		l = a[:len(a)-1]
		i = a[len(a)-1]
		s = len(l)
	)

	leaves := make([]merkle.Hash, s)
	for j, v := range l {
		leaves[j] = sha256.Sum256([]byte(v))
	}

	index, err := strconv.Atoi(i)
	if err != nil {
		log.Fatal(err)
	}

	tree, err := merkle.NewStaticTree(leaves)
	if err != nil {
		log.Fatal(err)
	}

	b, err := json.MarshalIndent(tree.Path(index), "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(b))
}
