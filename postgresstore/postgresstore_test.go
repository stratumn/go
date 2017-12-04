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

func TestStore(t *testing.T) {
	storetestcases.Factory{
		New: func() (store.Adapter, error) {
			a, err := New(&Config{URL: "postgres://postgres@localhost/sdk_test?sslmode=disable"})
			if err := a.Create(); err != nil {
				return nil, err
			}
			if err := a.Prepare(); err != nil {
				return nil, err
			}
			return a, err
		},
		Free: func(a store.Adapter) {
			if err := a.(*Store).Drop(); err != nil {
				panic(err)
			}
			if err := a.(*Store).Close(); err != nil {
				panic(err)
			}
		},
	}.RunTests(t)
}

func TestStoreV2(t *testing.T) {
	storetestcases.Factory{
		NewV2: func() (store.AdapterV2, error) {
			a, err := New(&Config{URL: "postgres://postgres@localhost/sdk_test?sslmode=disable"})
			if err := a.Create(); err != nil {
				return nil, err
			}
			if err := a.Prepare(); err != nil {
				return nil, err
			}
			return a, err
		},
		FreeV2: func(a store.AdapterV2) {
			if err := a.(*Store).Drop(); err != nil {
				panic(err)
			}
			if err := a.(*Store).Close(); err != nil {
				panic(err)
			}
		},
	}.RunTestsV2(t)
}
