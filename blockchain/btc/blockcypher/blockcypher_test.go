// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// LICENSE file.

package blockcypher

import (
	"testing"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/stratumn/go/testutil"
	"github.com/stratumn/go/types"
	"github.com/stratumn/goprivate/blockchain/btc"
)

func TestFindUnspent(t *testing.T) {
	bcy := New(&Config{Network: btc.NetworkTest3})
	go bcy.Start()
	defer bcy.Stop()

	addr, err := btcutil.DecodeAddress("n4XCm5oQmo98uGhAJDxQ8wGsqA2YoGrKNX", &chaincfg.TestNet3Params)
	if err != nil {
		t.Fatalf("btcutil.DecodeAddress(): err: %s", err)
	}
	var addr20 types.ReversedBytes20
	copy(addr20[:], addr.ScriptAddress())

	outputs, total, err := bcy.FindUnspent(&addr20, 1000000)

	if err != nil {
		t.Errorf("bcy.FindUnspent(): err: %s", err)
	}
	if total < 1000000 {
		t.Errorf("bcy.FindUnspent(): total = %d want %d", total, 1000000)
	}
	if l := len(outputs); l < 1 {
		t.Errorf("bcy.FindUnspent(): len(outputs) = %d want > 0", l)
	}

	for _, output := range outputs {
		tx, err := bcy.api.GetTX(output.TXHash.String(), nil)
		if err != nil {
			t.Errorf("bcy.api.GetTX(): err: %s", err)
		}
		if !testutil.ContainsString(tx.Addresses, "n4XCm5oQmo98uGhAJDxQ8wGsqA2YoGrKNX") {
			t.Errorf("bcy.FindUnspent(): can't find address in output addresses %s", tx.Addresses)
		}
	}
}

func TestFindUnspent_notEnough(t *testing.T) {
	bcy := New(&Config{Network: btc.NetworkTest3})
	go bcy.Start()
	defer bcy.Stop()

	addr, err := btcutil.DecodeAddress("n4XCm5oQmo98uGhAJDxQ8wGsqA2YoGrKNX", &chaincfg.TestNet3Params)
	if err != nil {
		t.Fatalf("btcutil.DecodeAddress(): err: %s", err)
	}
	var addr20 types.ReversedBytes20
	copy(addr20[:], addr.ScriptAddress())

	_, _, err = bcy.FindUnspent(&addr20, 1000000000000)
	if err == nil {
		t.Errorf("bcy.FindUnspent(): err = nil want Error")
	}
}
