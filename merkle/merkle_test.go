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
	"encoding/json"
	"math/rand"
	"reflect"
	"testing"

	"github.com/stratumn/sdk/merkle"
	"github.com/stratumn/sdk/testutil"
	"github.com/stratumn/sdk/types"
)

func TestTreeConsistency(t *testing.T) {
	for i := 0; i < 10; i++ {
		tests := make([]types.Bytes32, 1+rand.Intn(1000))
		for j := range tests {
			tests[j] = *testutil.RandomHash()
		}

		static, err := merkle.NewStaticTree(tests)
		if err != nil {
			t.Fatalf("merkle.NewStaticTree(): err: %s", err)
		}

		dyn := merkle.NewDynTree(len(tests) * 2)
		for _, leaf := range tests {
			dyn.Add(&leaf)
		}

		if got, want := static.Root().String(), dyn.Root().String(); got != want {
			t.Errorf("static.Root() = %q want %q", got, want)
		}

		for j := range tests {
			p1 := static.Path(j)
			p2 := dyn.Path(j)

			if !reflect.DeepEqual(p1, p2) {
				got, _ := json.MarshalIndent(p1, "", "  ")
				want, _ := json.MarshalIndent(p2, "", "  ")
				t.Errorf("test#%d: static.Path() = %s\nwant %s\n", j, got, want)
			}
		}
	}
}
