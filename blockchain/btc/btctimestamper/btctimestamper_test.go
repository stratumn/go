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

package btctimestamper

import (
	"encoding/hex"
	"testing"

	"github.com/stratumn/go-indigocore/blockchain/btc"
	"github.com/stratumn/go-indigocore/blockchain/btc/btctesting"
	"github.com/stratumn/go-indigocore/testutil"
	"github.com/stratumn/go-indigocore/types"
)

func TestNetwork_NetworkTest3(t *testing.T) {
	ts, err := New(&Config{
		WIF: "924v2d7ryXJjnbwB6M9GsZDEjAkfE9aHeQAG1j8muA4UEjozeAJ",
		Fee: int64(10000),
	})
	if err != nil {
		t.Fatalf("New(): err: %s", err)
	}

	if got := ts.Network(); got != btc.NetworkTest3 {
		t.Errorf("ts.Network() = %q want %q", got, btc.NetworkTest3)
	}
}

func TestNetwork_NetworkMain(t *testing.T) {
	ts, err := New(&Config{
		WIF: "L3Wbnfn57Fc547FLSkm6iCzAaHmLArNUBCYx6q8LdxWoEMoFZmLH",
		Fee: int64(10000),
	})
	if err != nil {
		t.Fatalf("New(): err: %s", err)
	}

	if got := ts.Network(); got != btc.NetworkMain {
		t.Errorf("ts.Network() = %q want %q", got, btc.NetworkMain)
	}
}

func TestTimestamperTimestampHash(t *testing.T) {
	mock := &btctesting.Mock{}
	mock.MockFindUnspent.Fn = func(*types.ReversedBytes20, int64) ([]btc.Output, int64, error) {
		PKScriptHex := "76a914fc56f7f9f80cfba26f300c77b893c39ed89351ff88ac"
		PKScript, _ := hex.DecodeString(PKScriptHex)
		output := btc.Output{Index: 0, PKScript: PKScript}
		if err := output.TXHash.Unstring("c805dd0fbf728e6b7e6c4e5d4ddfaba0089291145453aafb762bcff7a8afe2f5"); err != nil {
			return nil, 0, err
		}
		return []btc.Output{output}, 6241000, nil
	}

	ts, err := New(&Config{
		WIF:           "924v2d7ryXJjnbwB6M9GsZDEjAkfE9aHeQAG1j8muA4UEjozeAJ",
		UnspentFinder: mock,
		Broadcaster:   mock,
		Fee:           int64(10000),
	})
	if err != nil {
		t.Fatalf("New(): err: %s", err)
	}

	if _, err := ts.TimestampHash(testutil.RandomHash()); err != nil {
		t.Fatalf("ts.TimestampHash(): err: %s", err)
	}

	if got := mock.MockBroadcast.CalledCount; got != 1 {
		t.Errorf("ts.TimestampHash(): Broadcast() called %d time(s) want 1 time", mock.MockBroadcast.CalledCount)
	}
}
