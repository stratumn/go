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

package batchfossilizer

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stratumn/merkle/types"
)

const interval = 100 * time.Millisecond

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

func TestMain(m *testing.M) {
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
