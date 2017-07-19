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

	"bytes"
)

// TestGetValue tests what happens when you get an existing segment.
func (f Factory) TestGetValue(t *testing.T) {
	a := f.initAdapter(t)
	defer f.free(a)

	k := testutil.RandomKey()
	v1 := testutil.RandomValue()

	a.SaveValue(k, v1)

	v2, err := a.GetValue(k)
	if err != nil {
		t.Fatalf("a.GetValue(): err: %s", err)
	}

	if got := v2; got == nil {
		t.Fatal("s2 = nil want []byte")
	}

	if got, want := v2, v1; bytes.Compare(got, want) != 0 {
		t.Errorf("s2 = %s\n want%s", got, want)
	}
}

// TestGetValueNotFound tests what happens when you get a nonexistent segment.
func (f Factory) TestGetValueNotFound(t *testing.T) {
	a := f.initAdapter(t)
	defer f.free(a)

	v, err := a.GetValue(testutil.RandomKey())
	if err != nil {
		t.Fatalf("a.GetValue(): err: %s", err)
	}

	if got := v; got != nil {
		t.Errorf("s = %s\n want nil", got)
	}
}

// BenchmarkGetValue benchmarks getting existing segments.
func (f Factory) BenchmarkGetValue(b *testing.B) {
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
		if v, err := a.GetValue(values[i]); err != nil {
			b.Fatal(err)
		} else if v == nil {
			b.Error("s = nil want []byte")
		}
	}
}

// BenchmarkGetValueParallel benchmarks getting existing segments in parallel.
func (f Factory) BenchmarkGetValueParallel(b *testing.B) {
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
			if v, err := a.GetValue(values[i]); err != nil {
				b.Error(err)
			} else if v == nil {
				b.Error("s = nil want *cs.Segment")
			}
		}
	})
}
