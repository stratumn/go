// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by the license that can be found in the
// LICENSE file.

package bcbatchfossilizer

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/stratumn/goprivate/merkle"
)

const interval = 100 * time.Millisecond

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

func TestMain(m *testing.M) {
	seed := int64(time.Now().Nanosecond())
	fmt.Printf("using seed %d\n", seed)
	rand.Seed(seed)

	loadPath("testdata/path-a-0.json", &pathA0)
	loadPath("testdata/path-ab-0.json", &pathAB0)
	loadPath("testdata/path-ab-1.json", &pathAB1)
	loadPath("testdata/path-abc-0.json", &pathABC0)
	loadPath("testdata/path-abc-1.json", &pathABC1)
	loadPath("testdata/path-abc-2.json", &pathABC2)
	loadPath("testdata/path-abcd-0.json", &pathABCD0)
	loadPath("testdata/path-abcd-1.json", &pathABCD1)
	loadPath("testdata/path-abcd-2.json", &pathABCD2)
	loadPath("testdata/path-abcd-3.json", &pathABCD3)
	loadPath("testdata/path-abcde-0.json", &pathABCDE0)
	loadPath("testdata/path-abcde-1.json", &pathABCDE1)
	loadPath("testdata/path-abcde-2.json", &pathABCDE2)
	loadPath("testdata/path-abcde-3.json", &pathABCDE3)
	loadPath("testdata/path-abcde-4.json", &pathABCDE4)

	flag.Parse()
	os.Exit(m.Run())
}
