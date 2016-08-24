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
		TXHex := "0100000001746e5a4154a3f772099c9dbc1837c267d72aeb396fc665b5713568" +
			"633a276051000000008a47304402201c366ec94d73671fd0380b7fb27010301c" +
			"05b15eed299ec4a14344b663a5e91802205d755d7f8e1c0bacf732d83b30bd9e" +
			"fca66ca52c2f6c8a3216e7ee7169807816014104d33318fd473f461a54f3fe2a" +
			"eeeb2ba00be07d578598f36e6612b740720d56b630b5d2bd7b9a57f74ae6439a" +
			"6b921c79b6f05cb939a298c2d9833ff8c6b441feffffffff02e83a5f00000000" +
			"001976a914fc56f7f9f80cfba26f300c77b893c39ed89351ff88ac0000000000" +
			"000000226a2058566c427a67626169434d52416a577768544863746375417868" +
			"784b5146446100000000"
		TX, _ := hex.DecodeString(TXHex)
		TXHashHex := "c805dd0fbf728e6b7e6c4e5d4ddfaba0089291145453aafb762bcff7a8afe2f5"
		TXHash, _ := hex.DecodeString(TXHashHex)
		output := btc.Output{Index: 0, TX: TX}
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
