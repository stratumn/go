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
