// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by the license that can be found in the
// LICENSE file.

package rethinkstore

import (
	"testing"

	"github.com/stratumn/go/store"
	"github.com/stratumn/go/store/storetestcases"
)

func BenchmarkStoreSoft(b *testing.B) {
	storetestcases.Factory{
		New: func() (store.Adapter, error) {
			a, err := New(&Config{URL: "localhost:28015", DB: "test"})
			if err := a.Create(); err != nil {
				return nil, err
			}
			return a, err
		},
		Free: func(a store.Adapter) {
			if err := a.(*Store).Drop(); err != nil {
				panic(err)
			}
		},
	}.RunBenchmarks(b)
}

func BenchmarkStoreHard(b *testing.B) {
	storetestcases.Factory{
		New: func() (store.Adapter, error) {
			a, err := New(&Config{
				URL:  "localhost:28015",
				DB:   "test",
				Hard: true,
			})
			if err := a.Create(); err != nil {
				return nil, err
			}
			return a, err
		},
		Free: func(a store.Adapter) {
			if err := a.(*Store).Drop(); err != nil {
				panic(err)
			}
		},
	}.RunBenchmarks(b)
}
