// Copyright 2017-2018 Stratumn SAS. All rights reserved.
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
	"context"
	"crypto/sha256"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-core/batchfossilizer/evidences"
	"github.com/stratumn/go-core/fossilizer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetInfo(t *testing.T) {
	t.Parallel()
	a, err := New(&Config{})
	if err != nil {
		t.Fatalf("New(): err: %s", err)
	}
	got, err := a.GetInfo(context.Background())
	if err != nil {
		t.Fatalf("a.GetInfo(): err: %s", err)
	}
	if _, ok := got.(*Info); !ok {
		t.Errorf("a.GetInfo(): info = %#v want *Info", got)
	}
}

func TestFossilize(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	ctx := context.Background()

	t.Parallel()
	path, err := ioutil.TempDir("", "batchfossilizer")
	if err != nil {
		t.Fatalf("ioutil.TempDir(): err: %s", err)
	}
	defer os.RemoveAll(path)

	a, err := New(&Config{Interval: interval, Path: path, StopBatch: true})
	if err != nil {
		t.Fatalf("New(): err: %s", err)
	}
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		if err := a.Start(ctx); err != nil && errors.Cause(err) != context.Canceled {
			t.Fatalf("a.Start(): err: %s", err)
		}
	}()

	if err := a.Fossilize(ctx, atos(sha256.Sum256([]byte("a"))), []byte("test a")); err != nil {
		t.Errorf("a.Fossilize(): err: %s", err)
	}
	if err := a.Fossilize(ctx, atos(sha256.Sum256([]byte("b"))), []byte("test b")); err != nil {
		t.Errorf("a.Fossilize(): err: %s", err)
	}
	if err := a.Fossilize(ctx, atos(sha256.Sum256([]byte("c"))), []byte("test c")); err != nil {
		t.Errorf("a.Fossilize(): err: %s", err)
	}
	if err := a.Fossilize(ctx, atos(sha256.Sum256([]byte("d"))), []byte("test d")); err != nil {
		t.Errorf("a.Fossilize(): err: %s", err)
	}
	if err := a.Fossilize(ctx, atos(sha256.Sum256([]byte("e"))), []byte("test e")); err != nil {
		t.Errorf("a.Fossilize(): err: %s", err)
	}

	cancel()

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
	t.Parallel()
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
	ctx := context.Background()

	t.Parallel()
	path, err := ioutil.TempDir("", "batchfossilizer")
	if err != nil {
		t.Fatalf("ioutil.TempDir(): err: %s", err)
	}
	defer os.RemoveAll(path)

	a, err := New(&Config{Path: path, StopBatch: false})
	if err != nil {
		t.Fatalf("New(): err: %s", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		if err := a.Start(ctx); err != nil && errors.Cause(err) != context.Canceled {
			t.Fatalf("a.Start(): err: %s", err)
		}
	}()

	if err := a.Fossilize(ctx, atos(sha256.Sum256([]byte("a"))), []byte("test a")); err != nil {
		t.Errorf("a.Fossilize(): err: %s", err)
	}
	if err := a.Fossilize(ctx, atos(sha256.Sum256([]byte("b"))), []byte("test b")); err != nil {
		t.Errorf("a.Fossilize(): err: %s", err)
	}
	if err := a.Fossilize(ctx, atos(sha256.Sum256([]byte("c"))), []byte("test c")); err != nil {
		t.Errorf("a.Fossilize(): err: %s", err)
	}
	if err := a.Fossilize(ctx, atos(sha256.Sum256([]byte("d"))), []byte("test d")); err != nil {
		t.Errorf("a.Fossilize(): err: %s", err)
	}
	if err := a.Fossilize(ctx, atos(sha256.Sum256([]byte("e"))), []byte("test e")); err != nil {
		t.Errorf("a.Fossilize(): err: %s", err)
	}

	<-a.Started()
	cancel()

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

func TestSetTransformer(t *testing.T) {
	t.Parallel()
	a, err := New(&Config{Interval: interval})
	if err != nil {
		t.Fatalf("New(): err: %s", err)
	}
	transformerCalled := false
	transformer := func(evidence *chainscript.Evidence, data, meta []byte) (*fossilizer.Result, error) {
		transformerCalled = true
		return &fossilizer.Result{
			Evidence: *evidence,
			Data:     data,
			Meta:     meta,
		}, nil
	}
	a.SetTransformer(transformer)

	tests := []fossilizeTest{
		{atos(sha256.Sum256([]byte("a"))), []byte("test a"), pathABCDE0, 0, false},
		{atos(sha256.Sum256([]byte("b"))), []byte("test b"), pathABCDE1, 0, false},
		{atos(sha256.Sum256([]byte("c"))), []byte("test c"), pathABCDE2, 0, false},
		{atos(sha256.Sum256([]byte("d"))), []byte("test d"), pathABCDE3, 0, false},
		{atos(sha256.Sum256([]byte("e"))), []byte("test e"), pathABCDE4, 0, false},
	}
	testFossilizeMultiple(t, a, tests, true, true)

	if !transformerCalled {
		t.Errorf("a.transform() was not called")
	}
}

func TestBatchProof(t *testing.T) {
	t.Parallel()

	a, err := New(&Config{
		Interval: interval,
	})
	require.NoError(t, err)

	tests := []fossilizeTest{
		{atos(sha256.Sum256([]byte("a"))), []byte("test a"), pathABCDE0, 0, false},
		{atos(sha256.Sum256([]byte("b"))), []byte("test b"), pathABCDE1, 0, false},
		{atos(sha256.Sum256([]byte("c"))), []byte("test c"), pathABCDE2, 0, false},
		{atos(sha256.Sum256([]byte("d"))), []byte("test d"), pathABCDE3, 0, false},
		{atos(sha256.Sum256([]byte("e"))), []byte("test e"), pathABCDE4, 0, false},
	}
	results := testFossilizeMultiple(t, a, tests, true, true)

	t.Run("TestTime()", func(t *testing.T) {
		for _, r := range results {
			e, err := evidences.UnmarshalProof(&r.Evidence)
			require.NoError(t, err)

			assert.Equal(t, uint64(time.Now().Unix()), e.Time())
		}
	})

	t.Run("TestVerify()", func(t *testing.T) {
		for _, r := range results {
			e, err := evidences.UnmarshalProof(&r.Evidence)
			require.NoError(t, err)

			assert.True(t, e.Verify(nil))
		}
	})
}
