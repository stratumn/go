// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package batchfossilizer

import (
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stratumn/goprivate/merkle"
)

const interval = 10 * time.Millisecond

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

func TestGetInfo(t *testing.T) {
	a, err := New(&Config{})
	if err != nil {
		t.Fatal(err)
	}
	info, err := a.GetInfo()
	if err != nil {
		t.Fatal(err)
	}
	if info == nil {
		t.Fatal("info is nil")
	}
}

func TestFossilize(t *testing.T) {
	a, err := New(&Config{Interval: interval})
	if err != nil {
		t.Fatal(err)
	}
	tests := []fossilizeTest{
		{atos(sha256.Sum256([]byte("a"))), []byte("test a"), pathABCDE0, 0, false},
		{atos(sha256.Sum256([]byte("b"))), []byte("test b"), pathABCDE1, 0, false},
		{atos(sha256.Sum256([]byte("c"))), []byte("test c"), pathABCDE2, 0, false},
		{atos(sha256.Sum256([]byte("d"))), []byte("test d"), pathABCDE3, 0, false},
		{atos(sha256.Sum256([]byte("e"))), []byte("test e"), pathABCDE4, 0, false},
	}
	testFossilizeMultiple(t, a, tests, true, true)
}

func TestFossilizeMaxLeaves(t *testing.T) {
	a, err := New(&Config{Interval: interval, MaxLeaves: 4})
	if err != nil {
		t.Fatal(err)
	}
	tests := []fossilizeTest{
		{atos(sha256.Sum256([]byte("a"))), []byte("test a 1"), pathABCD0, 0, false},
		{atos(sha256.Sum256([]byte("b"))), []byte("test b 1"), pathABCD1, 0, false},
		{atos(sha256.Sum256([]byte("c"))), []byte("test c 1"), pathABCD2, 0, false},
		{atos(sha256.Sum256([]byte("d"))), []byte("test d 1"), pathABCD3, 0, false},
		{atos(sha256.Sum256([]byte("a"))), []byte("test a 2"), pathABC0, 0, false},
		{atos(sha256.Sum256([]byte("b"))), []byte("test b 2"), pathABC1, 0, false},
		{atos(sha256.Sum256([]byte("c"))), []byte("test c 2"), pathABC2, 0, false},
	}
	testFossilizeMultiple(t, a, tests, true, true)
}

func TestFossilizeInterval(t *testing.T) {
	a, err := New(&Config{Interval: interval})
	if err != nil {
		t.Fatal(err)
	}
	tests := []fossilizeTest{
		{atos(sha256.Sum256([]byte("a"))), []byte("test a 1"), pathABC0, 0, false},
		{atos(sha256.Sum256([]byte("b"))), []byte("test b 1"), pathABC1, 0, false},
		{atos(sha256.Sum256([]byte("c"))), []byte("test c 1"), pathABC2, interval * 2, false},
		{atos(sha256.Sum256([]byte("a"))), []byte("test a 2"), pathABCD0, 0, false},
		{atos(sha256.Sum256([]byte("b"))), []byte("test b 2"), pathABCD1, 0, false},
		{atos(sha256.Sum256([]byte("c"))), []byte("test c 2"), pathABCD2, 0, false},
		{atos(sha256.Sum256([]byte("d"))), []byte("test d 2"), pathABCD3, interval * 2, false},
		{atos(sha256.Sum256([]byte("a"))), []byte("test a 3"), pathABC0, 0, false},
		{atos(sha256.Sum256([]byte("b"))), []byte("test b 3"), pathABC1, 0, false},
		{atos(sha256.Sum256([]byte("c"))), []byte("test c 3"), pathABC2, 0, false},
	}
	testFossilizeMultiple(t, a, tests, true, true)
}

func TestFossilizeStopped(t *testing.T) {
	a, err := New(&Config{Interval: interval})
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		if err := a.Start(); err != nil {
			t.Fatal(err)
		}
	}()

	if err := a.Stop(); err != nil {
		t.Fatal(err)
	}

	if err := a.Fossilize(atos(sha256.Sum256([]byte("test"))), []byte("test meta")); err == nil {
		t.Fatal("expected error not to be nil")
	}
}

func TestNewRecover(t *testing.T) {
	path, err := ioutil.TempDir("", "batchfossilizer")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(path)

	a, err := New(&Config{Path: path, StopBatch: false})
	if err != nil {
		t.Fatal(err)
	}
	if err := a.Fossilize(atos(sha256.Sum256([]byte("a"))), []byte("test a")); err != nil {
		t.Fatal(err)
	}
	if err := a.Fossilize(atos(sha256.Sum256([]byte("b"))), []byte("test b")); err != nil {
		t.Fatal(err)
	}
	if err := a.Fossilize(atos(sha256.Sum256([]byte("c"))), []byte("test c")); err != nil {
		t.Fatal(err)
	}
	if err := a.Fossilize(atos(sha256.Sum256([]byte("d"))), []byte("test d")); err != nil {
		t.Fatal(err)
	}
	if err := a.Fossilize(atos(sha256.Sum256([]byte("e"))), []byte("test e")); err != nil {
		t.Fatal(err)
	}
	if err := a.Stop(); err != nil {
		t.Fatal(err)
	}

	a, err = New(&Config{Interval: interval, Path: path})
	if err != nil {
		t.Fatal(err)
	}
	tests := []fossilizeTest{
		{atos(sha256.Sum256([]byte("a"))), []byte("test a"), pathABCDE0, 0, false},
		{atos(sha256.Sum256([]byte("b"))), []byte("test b"), pathABCDE1, 0, false},
		{atos(sha256.Sum256([]byte("c"))), []byte("test c"), pathABCDE2, 0, false},
		{atos(sha256.Sum256([]byte("d"))), []byte("test d"), pathABCDE3, 0, false},
		{atos(sha256.Sum256([]byte("e"))), []byte("test e"), pathABCDE4, 0, false},
	}
	testFossilizeMultiple(t, a, tests, true, false)
}

func TestStopBatch(t *testing.T) {
	path, err := ioutil.TempDir("", "batchfossilizer")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(path)

	a, err := New(&Config{Interval: interval, Path: path, StopBatch: true})
	go func() {
		if err := a.Start(); err != nil {
			t.Fatal(err)
		}
	}()

	if err != nil {
		t.Fatal(err)
	}
	if err := a.Fossilize(atos(sha256.Sum256([]byte("a"))), []byte("test a")); err != nil {
		t.Fatal(err)
	}
	if err := a.Fossilize(atos(sha256.Sum256([]byte("b"))), []byte("test b")); err != nil {
		t.Fatal(err)
	}
	if err := a.Fossilize(atos(sha256.Sum256([]byte("c"))), []byte("test c")); err != nil {
		t.Fatal(err)
	}
	if err := a.Fossilize(atos(sha256.Sum256([]byte("d"))), []byte("test d")); err != nil {
		t.Fatal(err)
	}
	if err := a.Fossilize(atos(sha256.Sum256([]byte("e"))), []byte("test e")); err != nil {
		t.Fatal(err)
	}

	go func() {
		if err := a.Stop(); err != nil {
			t.Fatal(err)
		}
	}()

	tests := []fossilizeTest{
		{atos(sha256.Sum256([]byte("a"))), []byte("test a"), pathABCDE0, 0, false},
		{atos(sha256.Sum256([]byte("b"))), []byte("test b"), pathABCDE1, 0, false},
		{atos(sha256.Sum256([]byte("c"))), []byte("test c"), pathABCDE2, 0, false},
		{atos(sha256.Sum256([]byte("d"))), []byte("test d"), pathABCDE3, 0, false},
		{atos(sha256.Sum256([]byte("e"))), []byte("test e"), pathABCDE4, 0, false},
	}
	testFossilizeMultiple(t, a, tests, false, false)
}

func TestArchive(t *testing.T) {
	path, err := ioutil.TempDir("", "batchfossilizer")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(path)

	a, err := New(&Config{Path: path, Archive: true, MaxLeaves: 5})
	if err != nil {
		t.Fatal(err)
	}

	tests := []fossilizeTest{
		{atos(sha256.Sum256([]byte("a"))), []byte("test a"), pathABCDE0, 0, false},
		{atos(sha256.Sum256([]byte("b"))), []byte("test b"), pathABCDE1, 0, false},
		{atos(sha256.Sum256([]byte("c"))), []byte("test c"), pathABCDE2, 0, false},
		{atos(sha256.Sum256([]byte("d"))), []byte("test d"), pathABCDE3, 0, false},
		{atos(sha256.Sum256([]byte("e"))), []byte("test e"), pathABCDE4, 0, false},
	}
	testFossilizeMultiple(t, a, tests, true, true)

	archive := filepath.Join(path, "d71f8983ad4ee170f8129f1ebcdd7440be7798d8e1c80420bf11f1eced610dba")

	if _, err := os.Stat(archive); err != nil {
		t.Fatal(err)
	}
}
