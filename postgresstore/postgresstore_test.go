// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by the license that can be found in the
// LICENSE file.

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
