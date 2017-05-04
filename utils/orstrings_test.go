// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by the license that can be found in the
// LICENSE file.

package utils

import "testing"

func TestOrStrings(t *testing.T) {
	want := "test"
	got := OrStrings("", want)

	if got != want {
		t.Errorf("Expected %s to equal %s", got, want)
	}
}
