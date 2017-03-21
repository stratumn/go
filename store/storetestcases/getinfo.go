// Copyright 2017 Stratumn SAS. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package storetestcases

import (
	"testing"
)

// TestGetInfo tests what happens when you get information about the adapter.
func (f Factory) TestGetInfo(t *testing.T) {
	a := f.initAdapter(t)
	defer f.free(a)

	info, err := a.GetInfo()
	if err != nil {
		t.Fatalf("a.GetInfo(): err: %s", err)
	}
	if info == nil {
		t.Fatal("info = nil want interface{}")
	}
}
