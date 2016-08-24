// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// that can be found in the LICENSE file.

package dummytimestamper

import (
	"testing"

	"github.com/stratumn/goprivate/testutil"
	"github.com/stratumn/goprivate/types"
)

func TestNetwork(t *testing.T) {
	n := Network{}
	if s := n.String(); s != networkString {
		t.Logf("actual: %s, expected %s\n", s, networkString)
		t.Fatal("unexpected network string")
	}
}

func TestTimestamperNetwork(t *testing.T) {
	ts := Timestamper{}

	if n, ok := ts.Network().(Network); !ok {
		t.Fatalf("expected network to be a Network, got %v\n", n)
	}
}

func TestTimestamperTimestamp(t *testing.T) {
	ts := Timestamper{}

	if _, err := ts.Timestamp(map[string]types.Bytes32{"hash": *testutil.RandomHash()}); err != nil {
		t.Fatal(err)
	}
}

func TestTimestamperTimestampHash(t *testing.T) {
	ts := Timestamper{}

	if _, err := ts.TimestampHash(testutil.RandomHash()); err != nil {
		t.Fatal(err)
	}
}