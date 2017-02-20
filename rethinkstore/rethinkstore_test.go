// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by the license that can be found in the
// LICENSE file.

package rethinkstore

import (
	"testing"

	"github.com/stratumn/sdk/store"
	"github.com/stratumn/sdk/store/storetestcases"
)

func TestStore(t *testing.T) {
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
	}.RunTests(t)
}

func TestExists(t *testing.T) {
	a, err := New(&Config{URL: "localhost:28015", DB: "test"})
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
