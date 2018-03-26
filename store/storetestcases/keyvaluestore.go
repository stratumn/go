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
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stratumn/go-indigocore/testutil"
	"github.com/stretchr/testify/assert"
)

// TestKeyValueStore runs all tests for the store.KeyValueStore interface
func (f Factory) TestKeyValueStore(t *testing.T) {
	a := f.initKeyValueStore(t)
	defer f.freeKeyValueStore(a)

	t.Run("SetValue", func(t *testing.T) {
		ctx := context.Background()
		key := testutil.RandomKey()
		value := testutil.RandomValue()

		err := a.SetValue(ctx, key, value)
		assert.NoError(t, err, "a.SetValue()")

		updatedValue := testutil.RandomValue()
		err = a.SetValue(ctx, key, updatedValue)
		assert.NoError(t, err, "a.SetValue()")

		storedValue, err := a.GetValue(ctx, key)
		assert.NoError(t, err, "a.GetValue()")
		assert.EqualValues(t, updatedValue, storedValue, "a.GetValue()")
	})

	t.Run("GetValue", func(t *testing.T) {
		ctx := context.Background()
		k := testutil.RandomKey()
		v1 := testutil.RandomValue()

		a.SetValue(ctx, k, v1)
		v2, err := a.GetValue(ctx, k)
		assert.NoError(t, err, "a.GetValue()")
		assert.EqualValues(t, v1, v2, "a.GetValue()")
	})

	t.Run("GetValue not found", func(t *testing.T) {
		ctx := context.Background()
		v, err := a.GetValue(ctx, testutil.RandomKey())
		assert.NoError(t, err, "a.GetValue()")
		assert.Nil(t, v, "Not found value")
	})

	t.Run("DeleteValue", func(t *testing.T) {
		ctx := context.Background()
		key := testutil.RandomKey()
		value1 := testutil.RandomValue()
		a.SetValue(ctx, key, value1)

		value2, err := a.DeleteValue(ctx, key)
		assert.NoError(t, err, "a.DeleteValue()")
		assert.EqualValues(t, value1, value2, "a.DeleteValue() should return the deleted value")

		value2, err = a.GetValue(ctx, key)
		assert.NoError(t, err, "a.GetValue()")
		assert.Nil(t, value2, "Deleted value should not be found")
	})

	t.Run("DeleteValue not found", func(t *testing.T) {
		ctx := context.Background()
		v, err := a.DeleteValue(ctx, testutil.RandomKey())
		assert.NoError(t, err, "a.DeleteValue()")
		assert.Nil(t, v, "Not found value should be nil")
	})
}

// BenchmarkGetValue benchmarks getting existing values.
func (f Factory) BenchmarkGetValue(b *testing.B) {
	a := f.initKeyValueStoreB(b)
	defer f.freeKeyValueStore(a)

	values := make([][]byte, b.N)
	for i := 0; i < b.N; i++ {
		v := testutil.RandomKey()
		a.SetValue(context.Background(), v, v)
		values[i] = v
	}

	b.ResetTimer()
	log.SetOutput(ioutil.Discard)

	for i := 0; i < b.N; i++ {
		if v, err := a.GetValue(context.Background(), values[i]); err != nil {
			b.Fatal(err)
		} else if v == nil {
			b.Error("s = nil want []byte")
		}
	}
}

// BenchmarkGetValueParallel benchmarks getting existing values in parallel.
func (f Factory) BenchmarkGetValueParallel(b *testing.B) {
	a := f.initKeyValueStoreB(b)
	defer f.freeKeyValueStore(a)

	values := make([][]byte, b.N)
	for i := 0; i < b.N; i++ {
		v := testutil.RandomKey()
		a.SetValue(context.Background(), v, v)
		values[i] = v
	}

	var counter uint64

	b.ResetTimer()
	log.SetOutput(ioutil.Discard)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			i := atomic.AddUint64(&counter, 1) - 1
			if v, err := a.GetValue(context.Background(), values[i]); err != nil {
				b.Error(err)
			} else if v == nil {
				b.Error("s = nil want *cs.Segment")
			}
		}
	})
}

// BenchmarkSetValue benchmarks saving new values.
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
		if err := a.SetValue(context.Background(), slice[i], slice[i]); err != nil {
			b.Error(err)
		}
	}
}

// BenchmarkSetValueParallel benchmarks saving new values in parallel.
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
			if err := a.SetValue(context.Background(), slice[i], slice[i]); err != nil {
				b.Error(err)
			}
		}
	})
}

func searchNewKey(values map[string][]byte) ([]byte, string) {
	for {
		k := testutil.RandomKey()
		strkey := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(k)), ""), "[]")
		if _, present := values[strkey]; !present {
			return k, strkey
		}
	}
}

// BenchmarkDeleteValue benchmarks deleting existing segments.
func (f Factory) BenchmarkDeleteValue(b *testing.B) {
	a := f.initKeyValueStoreB(b)
	defer f.freeKeyValueStore(a)

	values := make(map[string][]byte, b.N)
	for i := 0; i < b.N; i++ {
		k, strkey := searchNewKey(values)
		v := testutil.RandomValue()
		a.SetValue(context.Background(), k, v)
		values[strkey] = k
	}

	b.ResetTimer()
	log.SetOutput(ioutil.Discard)

	for _, k := range values {
		if s, err := a.DeleteValue(context.Background(), k); err != nil {
			b.Error(err)
		} else if s == nil {
			b.Error("s = nil want []byte")
		}
	}
}

// BenchmarkDeleteValueParallel benchmarks deleting existing segments in
// parallel.
func (f Factory) BenchmarkDeleteValueParallel(b *testing.B) {
	a := f.initKeyValueStoreB(b)
	defer f.freeKeyValueStore(a)

	mapvalues := make(map[string][]byte, b.N)
	for i := 0; i < b.N; i++ {
		k, strkey := searchNewKey(mapvalues)
		v := testutil.RandomValue()
		a.SetValue(context.Background(), k, v)
		mapvalues[strkey] = k
	}
	values := make([][]byte, 0, b.N)
	for _, v := range mapvalues {
		values = append(values, v)
	}

	var counter uint64

	b.ResetTimer()
	log.SetOutput(ioutil.Discard)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			i := atomic.AddUint64(&counter, 1) - 1
			if s, err := a.DeleteValue(context.Background(), values[i]); err != nil {
				b.Error(err)
			} else if s == nil {
				b.Error("s = nil want *cs.Segment")
			}
		}
	})
}
