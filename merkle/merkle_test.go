package merkle_test

import (
	"crypto/sha256"
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/stratumn/goprivate/merkle"
	"github.com/stratumn/goprivate/merkle/treetestcases"
)

func TestHashTripletValidateOK(t *testing.T) {
	var (
		left  = treetestcases.RandomHash()
		right = treetestcases.RandomHash()
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
		treetestcases.RandomHash(),
		treetestcases.RandomHash(),
		treetestcases.RandomHash(),
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
