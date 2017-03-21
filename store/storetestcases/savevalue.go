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
)

// TestSaveValue tests what happens when you save a new value.
func (f Factory) TestSaveValue(t *testing.T) {
	a := f.initAdapter(t)
	defer f.free(a)

	key := testutil.RandomKey()
	value := testutil.RandomValue()

	if err := a.SaveValue(key, value); err != nil {
		t.Fatalf("a.SaveValue(): err: %s", err)
	}
}

// BenchmarkSaveValue benchmarks saving new segments.
func (f Factory) BenchmarkSaveValue(b *testing.B) {
	a := f.initAdapterB(b)
	defer f.free(a)

	slice := make([][]byte, b.N)
	for i := 0; i < b.N; i++ {
		slice[i] = testutil.RandomKey()
	}

	b.ResetTimer()
	log.SetOutput(ioutil.Discard)

	for i := 0; i < b.N; i++ {
		if err := a.SaveValue(slice[i], slice[i]); err != nil {
			b.Error(err)
		}
	}
}

// BenchmarkSaveValueParallel benchmarks saving new segments in parallel.
func (f Factory) BenchmarkSaveValueParallel(b *testing.B) {
	a := f.initAdapterB(b)
	defer f.free(a)

	slice := make([][]byte, b.N)
	for i := 0; i < b.N; i++ {
		slice[i] = testutil.RandomKey()
	}

	var counter uint64

	b.ResetTimer()
	log.SetOutput(ioutil.Discard)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			i := atomic.AddUint64(&counter, 1) - 1
			if err := a.SaveValue(slice[i], slice[i]); err != nil {
				b.Error(err)
			}
		}
	})
}
