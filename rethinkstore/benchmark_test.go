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

package rethinkstore

import (
	"testing"

	"github.com/stratumn/sdk/store"
	"github.com/stratumn/sdk/store/storetestcases"
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
