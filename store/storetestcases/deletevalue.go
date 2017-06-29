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
		t.Fatal("s2 = nil want []byte")
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
