package merkle_test

import (
	"crypto/sha256"
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"reflect"
	"testing"

	"github.com/stratumn/goprivate/merkle"
	"github.com/stratumn/goprivate/merkle/merkletesting"
)

func TestHashTripletValidateOK(t *testing.T) {
	var (
		left  = merkletesting.RandomHash()
		right = merkletesting.RandomHash()
		h     = merkle.HashTriplet{Left: left, Right: right}
		hash  = sha256.New()
	)

	if _, err := hash.Write(left[:]); err != nil {
		t.Fatal(err)
	}
	if _, err := hash.Write(right[:]); err != nil {
		t.Fatal(err)
	}

	copy(h.Parent[:], hash.Sum(nil))

	if err := h.Validate(); err != nil {
		t.Log(err)
		t.Fatal("expected error to be nil")
	}
}

func TestHashTripletValidateNotOK(t *testing.T) {
	h := merkle.HashTriplet{
		Left:   merkletesting.RandomHash(),
		Right:  merkletesting.RandomHash(),
		Parent: merkletesting.RandomHash(),
	}
	if err := h.Validate(); err == nil {
		t.Fatal("expected error not to be nil")
	}
}

func TestPathValidateOK(t *testing.T) {
	var (
		pathABCDE0 merkle.Path
		pathABCDE4 merkle.Path
	)
	if err := loadPath("testdata/path-abcde-0.json", &pathABCDE0); err != nil {
		t.Fatal(err)
	}
	if err := loadPath("testdata/path-abcde-4.json", &pathABCDE4); err != nil {
		t.Fatal(err)
	}

	if err := pathABCDE0.Validate(); err != nil {
		t.Log(err)
		t.Error("expected error to be nil")
	}
	if err := pathABCDE4.Validate(); err != nil {
		t.Log(err)
		t.Error("expected error to be nil")
	}
}

func TestPathValidateNotOK(t *testing.T) {
	var (
		pathInvalid0 merkle.Path
		pathInvalid1 merkle.Path
	)
	if err := loadPath("testdata/path-invalid-0.json", &pathInvalid0); err != nil {
		t.Fatal(err)
	}
	if err := loadPath("testdata/path-invalid-1.json", &pathInvalid1); err != nil {
		t.Fatal(err)
	}

	if err := pathInvalid0.Validate(); err == nil {
		t.Fatal("expected error not to be nil")
	}
	if err := pathInvalid1.Validate(); err == nil {
		t.Fatal("expected error not to be nil")
	}
}

func TestTreeConsistency(t *testing.T) {
	for i := 0; i < 10; i++ {
		leaves := make([]merkle.Hash, 1+rand.Intn(1000))
		for j := range leaves {
			leaves[j] = merkletesting.RandomHash()
		}

		static, err := merkle.NewStaticTree(leaves)
		if err != nil {
			t.Fatal(err)
		}
		if static == nil {
			t.Fatal("expected tree not to be nil")
		}

		dyn := merkle.NewDynTree(len(leaves) * 2)
		if dyn == nil {
			t.Fatal("expected tree not to be nil")
		}
		for _, leaf := range leaves {
			if err := dyn.Add(leaf); err != nil {
				t.Fatal(err)
			}
		}

		if static.Root() != dyn.Root() {
			t.Fatal("expected roots to be the same")
		}

		for j := range leaves {
			p1 := static.Path(j)
			p2 := dyn.Path(j)

			if !reflect.DeepEqual(p1, p2) {
				json1, _ := json.MarshalIndent(p1, "", "  ")
				json2, _ := json.MarshalIndent(p2, "", "  ")
				t.Logf("static: %s; dyn: %s\n", json1, json2)
				t.Error("expected paths to be the same")
			}
		}
	}
}

func loadPath(filename string, path *merkle.Path) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(data, path); err != nil {
		return err
	}
	return nil
}
