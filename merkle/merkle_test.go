// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package merkle_test

import (
	"crypto/sha256"
	"encoding/json"
	"math/rand"
	"reflect"
	"testing"

	"github.com/stratumn/go/testutil"
	"github.com/stratumn/go/types"
	"github.com/stratumn/goprivate/merkle"
)

func TestHashTripletValidate_OK(t *testing.T) {
	var (
		left  = *testutil.RandomHash()
		right = *testutil.RandomHash()
		h     = merkle.HashTriplet{Left: left, Right: right}
		hash  = sha256.New()
	)

	if _, err := hash.Write(left[:]); err != nil {
		t.Fatalf("hash.Write(): err: %s", err)
	}
	if _, err := hash.Write(right[:]); err != nil {
		t.Fatalf("hash.Write(): err: %s", err)
	}

	copy(h.Parent[:], hash.Sum(nil))

	if err := h.Validate(); err != nil {
		t.Errorf("h.Validate(): err: %s", err)
	}
}

func TestHashTripletValidate_Error(t *testing.T) {
	h := merkle.HashTriplet{
		Left:   *testutil.RandomHash(),
		Right:  *testutil.RandomHash(),
		Parent: *testutil.RandomHash(),
	}
	if err := h.Validate(); err == nil {
		t.Error("h.Validate(): err = nil want Error")
	}
}

func TestPathValidate_OK(t *testing.T) {
	var (
		pathABCDE0 merkle.Path
		pathABCDE4 merkle.Path
	)
	if err := loadPath("testdata/path-abcde-0.json", &pathABCDE0); err != nil {
		t.Fatalf("loadPath(): err: %s", err)
	}
	if err := loadPath("testdata/path-abcde-4.json", &pathABCDE4); err != nil {
		t.Fatalf("loadPath(): err: %s", err)
	}

	if err := pathABCDE0.Validate(); err != nil {
		t.Errorf("pathABCDE0.Validate(): err: %s", err)
	}
	if err := pathABCDE4.Validate(); err != nil {
		t.Errorf("pathABCDE4.Validate(): err: %s", err)
	}
}

func TestPathValidate_Error(t *testing.T) {
	var (
		pathInvalid0 merkle.Path
		pathInvalid1 merkle.Path
	)
	if err := loadPath("testdata/path-invalid-0.json", &pathInvalid0); err != nil {
		t.Fatalf("loadPath(): err: %s", err)
	}
	if err := loadPath("testdata/path-invalid-1.json", &pathInvalid1); err != nil {
		t.Fatalf("loadPath(): err: %s", err)
	}

	if err := pathInvalid0.Validate(); err == nil {
		t.Error("pathInvalid0.Validate(): err = nil want Error")
	}
	if err := pathInvalid1.Validate(); err == nil {
		t.Error("pathInvalid1.Validate(): err = nil want Error")
	}
}

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
			if err := dyn.Add(&leaf); err != nil {
				t.Errorf("dyn.Add(): err: %s", err)
			}
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
				t.Errorf("test#%d: static.Path() = %s\nwant %s\n", got, want)
			}
		}
	}
}
