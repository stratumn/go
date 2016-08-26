// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// that can be found in the LICENSE file.

package btctimestamper

import (
	"encoding/hex"
	"testing"

	"github.com/stratumn/goprivate/blockchain/btc"
	"github.com/stratumn/goprivate/blockchain/btc/btctesting"
	"github.com/stratumn/goprivate/testutil"
	"github.com/stratumn/goprivate/types"
)

func TestTimestamperNetworkTest3(t *testing.T) {
	ts, err := New(&Config{
		WIF: "924v2d7ryXJjnbwB6M9GsZDEjAkfE9aHeQAG1j8muA4UEjozeAJ",
		Fee: int64(10000),
	})
	if err != nil {
		t.Fatal(err)
	}
	if ts == nil {
		t.Fatal("expected timestamper not to be nil")
	}

	if ts.Network() != btc.NetworkTest3 {
		t.Logf("actual: %s, expected: %s", ts.Network(), btc.NetworkTest3)
		t.Fatal("unexpected network")
	}
}

func TestTimestamperNetworkMain(t *testing.T) {
	ts, err := New(&Config{
		WIF: "L3Wbnfn57Fc547FLSkm6iCzAaHmLArNUBCYx6q8LdxWoEMoFZmLH",
		Fee: int64(10000),
	})
	if err != nil {
		t.Fatal(err)
	}
	if ts == nil {
		t.Fatal("expected timestamper not to be nil")
	}

	if ts.Network() != btc.NetworkMain {
		t.Logf("actual: %s, expected: %s", ts.Network(), btc.NetworkMain)
		t.Fatal("unexpected network")
	}
}

func TestTimestamperTimestampHash(t *testing.T) {
	mock := &btctesting.Mock{}
	mock.MockFindUnspent.Fn = func(address160 *types.Bytes20, amount int64) ([]btc.Output, int64, error) {
		PKScriptHex := "76a914fc56f7f9f80cfba26f300c77b893c39ed89351ff88ac"
		PKScript, _ := hex.DecodeString(PKScriptHex)
		output := btc.Output{Index: 0, PKScript: PKScript}
		TXHashHex := "c805dd0fbf728e6b7e6c4e5d4ddfaba0089291145453aafb762bcff7a8afe2f5"
		TXHash, _ := hex.DecodeString(TXHashHex)
		copy(output.TXHash[:], TXHash)

		return []btc.Output{output}, 6241000, nil
	}

	ts, err := New(&Config{
		WIF:           "924v2d7ryXJjnbwB6M9GsZDEjAkfE9aHeQAG1j8muA4UEjozeAJ",
		UnspentFinder: mock,
		Broadcaster:   mock,
		Fee:           int64(10000),
	})
	if err != nil {
		t.Fatal(err)
	}
	if ts == nil {
		t.Fatal("expected timestamper not to be nil")
	}

	if _, err := ts.TimestampHash(testutil.RandomHash()); err != nil {
		t.Fatal(err)
	}

	if mock.MockBroadcast.CalledCount != 1 {
		t.Logf("actual: %d, expected: %d", mock.MockBroadcast.CalledCount, 1)
		t.Fatal("expected Broadcast to be called once")
	}
}
