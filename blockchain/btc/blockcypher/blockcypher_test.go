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

package blockcypher

import (
	"context"
	"testing"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/stratumn/go-core/blockchain/btc"
	"github.com/stratumn/go-core/testutil"
	"github.com/stratumn/go-core/types"
)

func TestFindUnspent(t *testing.T) {
	bcy := New(&Config{Network: btc.NetworkTest3})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go bcy.Start(ctx)

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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go bcy.Start(ctx)

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
