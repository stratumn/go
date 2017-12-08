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
	"github.com/stretchr/testify/assert"
)

// TestGetValue tests what happens when you get an existing segment.
func (f Factory) TestGetValue(t *testing.T) {
	a := f.initKeyValueStore(t)
	defer f.freeKeyValueStore(a)

	k := testutil.RandomKey()
	v1 := testutil.RandomValue()

	a.SetValue(k, v1)
	v2, err := a.GetValue(k)
	assert.NoError(t, err, "a.GetValue()")
	assert.EqualValues(t, v1, v2, "a.GetValue()")
}

// TestGetValueNotFound tests what happens when you get a nonexistent segment.
func (f Factory) TestGetValueNotFound(t *testing.T) {
	a := f.initKeyValueStore(t)
	defer f.freeKeyValueStore(a)

	v, err := a.GetValue(testutil.RandomKey())
	assert.NoError(t, err, "a.GetValue()")
	assert.Nil(t, v, "Not found value")
}

// BenchmarkGetValue benchmarks getting existing segments.
func (f Factory) BenchmarkGetValue(b *testing.B) {
	a := f.initKeyValueStoreB(b)
	defer f.freeKeyValueStore(a)

	values := make([][]byte, b.N)
	for i := 0; i < b.N; i++ {
		v := testutil.RandomKey()
		a.SetValue(v, v)
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
	a := f.initKeyValueStoreB(b)
	defer f.freeKeyValueStore(a)

	values := make([][]byte, b.N)
	for i := 0; i < b.N; i++ {
		v := testutil.RandomKey()
		a.SetValue(v, v)
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
