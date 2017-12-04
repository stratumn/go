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
	storetestcases.Factory{
		New:  createAdapter,
		Free: freeAdapter,
	}.RunTests(t)
}

func TestStoreV2(t *testing.T) {
	storetestcases.Factory{
		NewV2:  createAdapterV2,
		FreeV2: freeAdapterV2,
	}.RunTestsV2(t)
}

func createAdapter() (store.Adapter, error) {
	a, err := New(&Config{URL: fmt.Sprintf("%s:%s", domain, port), DB: dbName})
	if err != nil {
		return nil, err
	}
	if err := a.Create(); err != nil {
		return nil, err
	}
	return a, err
}

func freeAdapter(a store.Adapter) {
	if err := a.(*Store).Clean(); err != nil {
		panic(err)
	}
}

func createAdapterV2() (store.AdapterV2, error) {
	a, err := New(&Config{URL: fmt.Sprintf("%s:%s", domain, port), DB: dbName})
	if err != nil {
		return nil, err
	}
	if err := a.Create(); err != nil {
		return nil, err
	}
	return a, err
}

func freeAdapterV2(a store.AdapterV2) {
	if err := a.(*Store).Clean(); err != nil {
		panic(err)
	}
}
