// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// that can be found in the LICENSE file.

package dummytimestamper

import (
	"testing"

	"github.com/stratumn/go/testutil"
	"github.com/stratumn/go/types"
)

func TestNetworkString(t *testing.T) {
	n := Network{}
	if got := n.String(); got != networkString {
		t.Errorf("n.String() = %q want %q", got, networkString)
	}
}

func TestTimestamperNetwork(t *testing.T) {
	ts := Timestamper{}
	if n, ok := ts.Network().(Network); !ok {
		t.Errorf("ts.Network = %#v want Network", n)
	}
}

func TestTimestamperTimestamp(t *testing.T) {
	ts := Timestamper{}
	if _, err := ts.Timestamp(map[string]types.Bytes32{"hash": *testutil.RandomHash()}); err != nil {
		t.Errorf("ts.Timestamp(): err: %s", err)
	}
}

func TestTimestamperTimestampHash(t *testing.T) {
	ts := Timestamper{}
	if _, err := ts.TimestampHash(testutil.RandomHash()); err != nil {
		t.Errorf("ts.TimestampHash(): err: %s", err)
	}
}
