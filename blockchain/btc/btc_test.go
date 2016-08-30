// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// that can be found in the LICENSE file.

package btc

import "testing"

func TestNetworkString(t *testing.T) {
	if got, want := NetworkTest3.String(), "bitcoin:test3"; got != want {
		t.Errorf("NetworkTest3.String() = %s want %s", got, want)
	}
}

func TestNetworkID(t *testing.T) {
	if got, want := NetworkTest3.ID(), byte(0x6F); got != want {
		t.Errorf(`NetworkTest3.String() = "%x" want "%x"`, got, want)
	}
	if got, want := NetworkMain.ID(), byte(0x00); got != want {
		t.Errorf(`NetworkTest3.String() = "%x" want "%x"`, got, want)
	}
}
