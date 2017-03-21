// Copyright 2017 Stratumn SAS. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package storetestcases

import (
	"io/ioutil"
	"log"
	"sync/atomic"
	"testing"

	"github.com/stratumn/sdk/testutil"

	"bytes"
)

// TestDeleteValue tests what happens when you delete an existing value.
func (f Factory) TestDeleteValue(t *testing.T) {
	a := f.initAdapter(t)
	defer f.free(a)

	key := testutil.RandomKey()
	value1 := testutil.RandomValue()
	a.SaveValue(key, value1)

	value2, err := a.DeleteValue(key)
	if err != nil {
		t.Fatalf("a.DeleteValue(): err: %s", err)
	}

	if got := value2; got == nil {
		t.Error("s2 = nil want []byte")
	}

	if got, want := value2, value1; bytes.Compare(got, want) != 0 {
		t.Errorf("s2 = %s\n want%s", got, want)
	}

	value2, err = a.GetValue(key)
	if err != nil {
		t.Fatalf("a.GetValue(): err: %s", err)
	}
	if got := value2; got != nil {
		t.Errorf("s2 = %s\n want nil", got)
	}
}

// TestDeleteValueNotFound tests what happens when you delete a nonexistent
// value.
func (f Factory) TestDeleteValueNotFound(t *testing.T) {
	a := f.initAdapter(t)
	defer f.free(a)

	v, err := a.DeleteValue(testutil.RandomKey())
	if err != nil {
		t.Fatalf("a.DeleteValue(): err: %s", err)
	}

	if got := v; got != nil {
		t.Errorf("s = %s\n want nil", got)
	}
}

// BenchmarkDeleteValue benchmarks deleting existing segments.
func (f Factory) BenchmarkDeleteValue(b *testing.B) {
	a := f.initAdapterB(b)
	defer f.free(a)

	values := make([][]byte, b.N)
	for i := 0; i < b.N; i++ {
		v := testutil.RandomKey()
		a.SaveValue(v, v)
		values[i] = v
	}

	b.ResetTimer()
	log.SetOutput(ioutil.Discard)

	for i := 0; i < b.N; i++ {
		if s, err := a.DeleteValue(values[i]); err != nil {
			b.Error(err)
		} else if s == nil {
			b.Error("s = nil want []byte")
		}
	}
}

// BenchmarkDeleteValueParallel benchmarks deleting existing segments in
// parallel.
func (f Factory) BenchmarkDeleteValueParallel(b *testing.B) {
	a := f.initAdapterB(b)
	defer f.free(a)

	values := make([][]byte, b.N)
	for i := 0; i < b.N; i++ {
		v := testutil.RandomKey()
		a.SaveValue(v, v)
		values[i] = v
	}

	var counter uint64

	b.ResetTimer()
	log.SetOutput(ioutil.Discard)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			i := atomic.AddUint64(&counter, 1) - 1
			if s, err := a.DeleteValue(values[i]); err != nil {
				b.Error(err)
			} else if s == nil {
				b.Error("s = nil want *cs.Segment")
			}
		}
	})
}
