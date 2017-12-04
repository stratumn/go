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

package dummytimestamper

import (
	"testing"

	"github.com/stratumn/sdk/testutil"
	"github.com/stratumn/sdk/types"
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
