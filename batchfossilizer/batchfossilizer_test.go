// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package batchfossilizer

import (
	"crypto/sha256"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestGetInfo(t *testing.T) {
	a, err := New(&Config{})
	if err != nil {
		t.Fatalf("New(): err: %s", err)
	}
	got, err := a.GetInfo()
	if err != nil {
		t.Fatalf("a.GetInfo(): err: %s", err)
	}
	if _, ok := got.(*Info); !ok {
		t.Errorf("a.GetInfo(): info = %#v want *Info", got)
	}
}

func TestFossilize(t *testing.T) {
	a, err := New(&Config{Interval: interval})
	if err != nil {
		t.Fatalf("New(): err: %s", err)
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

func TestFossilize_MaxLeaves(t *testing.T) {
	a, err := New(&Config{Interval: interval, MaxLeaves: 4})
	if err != nil {
		t.Fatalf("New(): err: %s", err)
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

func TestFossilize_Interval(t *testing.T) {
	a, err := New(&Config{Interval: interval})
	if err != nil {
		t.Fatalf("New(): err: %s", err)
	}
	tests := []fossilizeTest{
		{atos(sha256.Sum256([]byte("a"))), []byte("test a 1"), pathABC0, 0, false},
		{atos(sha256.Sum256([]byte("b"))), []byte("test b 1"), pathABC1, 0, false},
		{atos(sha256.Sum256([]byte("c"))), []byte("test c 1"), pathABC2, interval * 10, false},
		{atos(sha256.Sum256([]byte("a"))), []byte("test a 2"), pathABCD0, 0, false},
		{atos(sha256.Sum256([]byte("b"))), []byte("test b 2"), pathABCD1, 0, false},
		{atos(sha256.Sum256([]byte("c"))), []byte("test c 2"), pathABCD2, 0, false},
		{atos(sha256.Sum256([]byte("d"))), []byte("test d 2"), pathABCD3, interval * 10, false},
		{atos(sha256.Sum256([]byte("a"))), []byte("test a 3"), pathABC0, 0, false},
		{atos(sha256.Sum256([]byte("b"))), []byte("test b 3"), pathABC1, 0, false},
		{atos(sha256.Sum256([]byte("c"))), []byte("test c 3"), pathABC2, 0, false},
	}
	testFossilizeMultiple(t, a, tests, true, true)
}

func TestStop_StopBatch(t *testing.T) {
	path, err := ioutil.TempDir("", "batchfossilizer")
	if err != nil {
		t.Fatalf("ioutil.TempDir(): err: %s", err)
	}
	defer os.RemoveAll(path)

	a, err := New(&Config{Interval: interval, Path: path, StopBatch: true})
	if err != nil {
		t.Fatalf("New(): err: %s", err)
	}

	go func() {
		if err := a.Start(); err != nil {
			t.Fatalf("a.Start(): err: %s", err)
		}
	}()

	if err := a.Fossilize(atos(sha256.Sum256([]byte("a"))), []byte("test a")); err != nil {
		t.Errorf("a.Fossilize(): err: %s", err)
	}
	if err := a.Fossilize(atos(sha256.Sum256([]byte("b"))), []byte("test b")); err != nil {
		t.Errorf("a.Fossilize(): err: %s", err)
	}
	if err := a.Fossilize(atos(sha256.Sum256([]byte("c"))), []byte("test c")); err != nil {
		t.Errorf("a.Fossilize(): err: %s", err)
	}
	if err := a.Fossilize(atos(sha256.Sum256([]byte("d"))), []byte("test d")); err != nil {
		t.Errorf("a.Fossilize(): err: %s", err)
	}
	if err := a.Fossilize(atos(sha256.Sum256([]byte("e"))), []byte("test e")); err != nil {
		t.Errorf("a.Fossilize(): err: %s", err)
	}

	go a.Stop()

	tests := []fossilizeTest{
		{atos(sha256.Sum256([]byte("a"))), []byte("test a"), pathABCDE0, 0, false},
		{atos(sha256.Sum256([]byte("b"))), []byte("test b"), pathABCDE1, 0, false},
		{atos(sha256.Sum256([]byte("c"))), []byte("test c"), pathABCDE2, 0, false},
		{atos(sha256.Sum256([]byte("d"))), []byte("test d"), pathABCDE3, 0, false},
		{atos(sha256.Sum256([]byte("e"))), []byte("test e"), pathABCDE4, 0, false},
	}
	testFossilizeMultiple(t, a, tests, false, false)
}

func TestFossilize_Archive(t *testing.T) {
	path, err := ioutil.TempDir("", "batchfossilizer")
	if err != nil {
		t.Fatalf("ioutil.TempDir(): err: %s", err)
	}
	defer os.RemoveAll(path)

	a, err := New(&Config{Path: path, Archive: true, MaxLeaves: 5})
	if err != nil {
		t.Fatalf("New(): err: %s", err)
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
		t.Errorf("os.Stat(): err: %s", err)
	}
}

func TestNew_recover(t *testing.T) {
	path, err := ioutil.TempDir("", "batchfossilizer")
	if err != nil {
		t.Fatalf("ioutil.TempDir(): err: %s", err)
	}
	defer os.RemoveAll(path)

	a, err := New(&Config{Path: path, StopBatch: false})
	if err != nil {
		t.Fatalf("New(): err: %s", err)
	}

	go func() {
		if err := a.Start(); err != nil {
			t.Fatalf("a.Start(): err: %s", err)
		}
	}()

	if err := a.Fossilize(atos(sha256.Sum256([]byte("a"))), []byte("test a")); err != nil {
		t.Errorf("a.Fossilize(): err: %s", err)
	}
	if err := a.Fossilize(atos(sha256.Sum256([]byte("b"))), []byte("test b")); err != nil {
		t.Errorf("a.Fossilize(): err: %s", err)
	}
	if err := a.Fossilize(atos(sha256.Sum256([]byte("c"))), []byte("test c")); err != nil {
		t.Errorf("a.Fossilize(): err: %s", err)
	}
	if err := a.Fossilize(atos(sha256.Sum256([]byte("d"))), []byte("test d")); err != nil {
		t.Errorf("a.Fossilize(): err: %s", err)
	}
	if err := a.Fossilize(atos(sha256.Sum256([]byte("e"))), []byte("test e")); err != nil {
		t.Errorf("a.Fossilize(): err: %s", err)
	}

	<-a.Started()
	a.Stop()

	a, err = New(&Config{Interval: interval, Path: path})
	if err != nil {
		t.Fatalf("New(): err: %s", err)
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