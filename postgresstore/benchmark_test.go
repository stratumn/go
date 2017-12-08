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

package postgresstore

import (
	"testing"

	"github.com/stratumn/sdk/store"
	"github.com/stratumn/sdk/store/storetestcases"
)

func BenchmarkStore(b *testing.B) {
	factory := storetestcases.Factory{
		New:               createAdapterB,
		NewKeyValueStore:  createKeyValueStoreB,
		Free:              freeAdapterB,
		FreeKeyValueStore: freeKeyValueStoreB,
	}

	factory.RunStoreBenchmarks(b)
	factory.RunKeyValueStoreBenchmarks(b)
}

func createStoreB() (*Store, error) {
	a, err := New(&Config{URL: "postgres://postgres@localhost/postgres?sslmode=disable"})
	if err := a.Create(); err != nil {
		return nil, err
	}
	if err := a.Prepare(); err != nil {
		return nil, err
	}
	return a, err
}

func createAdapterB() (store.Adapter, error) {
	return createStore()
}

func createKeyValueStoreB() (store.KeyValueStore, error) {
	return createStore()
}

func freeStoreB(s *Store) {
	if err := s.Drop(); err != nil {
		panic(err)
	}
}

func freeAdapterB(s store.Adapter) {
	freeStore(s.(*Store))
}

func freeKeyValueStoreB(s store.KeyValueStore) {
	freeStore(s.(*Store))
}
