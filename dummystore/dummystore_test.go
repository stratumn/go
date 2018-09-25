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

package dummystore

import (
	"testing"

	"github.com/stratumn/go-core/store"
	"github.com/stratumn/go-core/store/storetestcases"
	"github.com/stratumn/go-core/tmpop/tmpoptestcases"
)

func TestDummyStore(t *testing.T) {
	factory := storetestcases.Factory{
		New: func() (store.Adapter, error) {
			return New(&Config{}), nil
		},
		NewKeyValueStore: func() (store.KeyValueStore, error) {
			s := New(&Config{})
			return s, nil
		},
	}

	factory.RunStoreTests(t)
	factory.RunKeyValueStoreTests(t)
}

func TestDummyTMPop(t *testing.T) {
	tmpoptestcases.Factory{
		New: func() (store.Adapter, store.KeyValueStore, error) {
			dummyStore := New(&Config{})
			return dummyStore, dummyStore, nil
		},
	}.RunTests(t)
}
