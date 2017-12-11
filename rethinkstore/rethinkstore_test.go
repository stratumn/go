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
	"fmt"
	"testing"

	"github.com/stratumn/sdk/store"
	"github.com/stratumn/sdk/store/storetestcases"
)

func TestExists(t *testing.T) {
	a, err := New(&Config{URL: fmt.Sprintf("%s:%s", domain, port), DB: dbName})
	if err != nil {
		t.Errorf("err: New(): %s", err)
	}
	got, err := a.Exists()
	if err != nil {
		t.Errorf("err: a.Exists(): %s", err)
	}
	if got {
		t.Errorf("err: a.Exists(): exists = true want false")
	}
	if err := a.Create(); err != nil {
		t.Errorf("err: a.Create(): %s", err)
	}
	defer a.Drop()
	got, err = a.Exists()
	if err != nil {
		t.Errorf("err: a.Exists(): %s", err)
	}
	if !got {
		t.Errorf("err: a.Exists(): exists = false want true")
	}
}

func TestStore(t *testing.T) {
	factory := storetestcases.Factory{
		New:               createAdapter,
		NewKeyValueStore:  createKeyValueStore,
		Free:              freeAdapter,
		FreeKeyValueStore: freeKeyValueStore,
	}

	factory.RunStoreTests(t)
	factory.RunKeyValueStoreTests(t)
}

func createStore() (*Store, error) {
	a, err := New(&Config{URL: fmt.Sprintf("%s:%s", domain, port), DB: dbName})
	if err != nil {
		return nil, err
	}
	if err := a.Create(); err != nil {
		return nil, err
	}
	return a, err
}

func createAdapter() (store.Adapter, error) {
	return createStore()
}

func createKeyValueStore() (store.KeyValueStore, error) {
	return createStore()
}

func freeStore(a *Store) {
	if err := a.Clean(); err != nil {
		panic(err)
	}
}

func freeAdapter(a store.Adapter) {
	freeStore(a.(*Store))
}

func freeKeyValueStore(a store.KeyValueStore) {
	freeStore(a.(*Store))
}
