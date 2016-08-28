// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// that can be found in the LICENSE file.

package blockcypher

import (
	"encoding/hex"
	"testing"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/stratumn/go/testutil"
	"github.com/stratumn/goprivate/blockchain/btc"
	"github.com/stratumn/goprivate/types"
)

func TestFindUnspentOK(t *testing.T) {
	bcy := New(btc.NetworkTest3, "")

	addr, err := btcutil.DecodeAddress("n4XCm5oQmo98uGhAJDxQ8wGsqA2YoGrKNX", &chaincfg.TestNet3Params)
	if err != nil {
		t.Fatal(err)
	}
	var addr20 types.Bytes20
	copy(addr20[:], addr.ScriptAddress())

	outputs, total, err := bcy.FindUnspent(&addr20, 1000000)
	if err != nil {
		t.Fatal(err)
	}
	if total < 1000000 {
		t.Logf("actual: %d; expected: %d", total, 1000000)
		t.Fatal("unexpected total")
	}
	if len(outputs) < 1 {
		t.Fatal("expected outputs")
	}

	for _, output := range outputs {
		var TXHash types.Bytes32
		// Invert bytes!
		for i, b := range output.TXHash {
			TXHash[types.Bytes32Size-i-1] = b
		}
		TXHashString := hex.EncodeToString(TXHash[:])
		tx, err := bcy.api.GetTX(TXHashString, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !testutil.ContainsString(tx.Addresses, "n4XCm5oQmo98uGhAJDxQ8wGsqA2YoGrKNX") {
			t.Log(tx.Addresses)
			t.Fatal("unexpected output")
		}
	}
}

func TestFindUnspentNotEnough(t *testing.T) {
	api := New(btc.NetworkTest3, "")

	addr, err := btcutil.DecodeAddress("n4XCm5oQmo98uGhAJDxQ8wGsqA2YoGrKNX", &chaincfg.TestNet3Params)
	if err != nil {
		t.Fatal(err)
	}
	var addr20 types.Bytes20
	copy(addr20[:], addr.ScriptAddress())

	_, _, err = api.FindUnspent(&addr20, 1000000000000)
	if err == nil {
		t.Fatal("expected error not to be nil")
	}
}
