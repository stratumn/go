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

// TestSaveValue tests what happens when you save a new value.
func (f Factory) TestSaveValue(t *testing.T) {
	a := f.initKeyValueStore(t)
	defer f.freeKeyValueStore(a)

	key := testutil.RandomKey()
	value := testutil.RandomValue()

	err := a.SetValue(key, value)
	assert.NoError(t, err, "a.SetValue()")

	updatedValue := testutil.RandomValue()
	err = a.SetValue(key, updatedValue)
	assert.NoError(t, err, "a.SetValue()")

	storedValue, err := a.GetValue(key)
	assert.NoError(t, err, "a.GetValue()")
	assert.EqualValues(t, updatedValue, storedValue, "a.GetValue()")
}

// BenchmarkSetValue benchmarks saving new segments.
func (f Factory) BenchmarkSetValue(b *testing.B) {
	a := f.initKeyValueStoreB(b)
	defer f.freeKeyValueStore(a)

	slice := make([][]byte, b.N)
	for i := 0; i < b.N; i++ {
		slice[i] = testutil.RandomKey()
	}

	b.ResetTimer()
	log.SetOutput(ioutil.Discard)

	for i := 0; i < b.N; i++ {
		if err := a.SetValue(slice[i], slice[i]); err != nil {
			b.Error(err)
		}
	}
}

// BenchmarkSetValueParallel benchmarks saving new segments in parallel.
func (f Factory) BenchmarkSetValueParallel(b *testing.B) {
	a := f.initKeyValueStoreB(b)
	defer f.freeKeyValueStore(a)

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
			if err := a.SetValue(slice[i], slice[i]); err != nil {
				b.Error(err)
			}
		}
	})
}
