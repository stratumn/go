// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by the license that can be found in the
// LICENSE file.

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
