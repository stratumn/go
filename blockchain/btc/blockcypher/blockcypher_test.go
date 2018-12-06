// Copyright 2016-2018 Stratumn SAS. All rights reserved.
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
	"time"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/stratumn/go-core/blockchain/btc"
	"github.com/stratumn/go-core/testutil"
	"github.com/stratumn/go-core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindUnspent(t *testing.T) {
	ctx := context.Background()
	testAddr := "n4XCm5oQmo98uGhAJDxQ8wGsqA2YoGrKNX"
	bcy := New(&Config{Network: btc.NetworkTest3})

	addr, err := btcutil.DecodeAddress(testAddr, &chaincfg.TestNet3Params)
	require.NoError(t, err)

	var addr20 types.ReversedBytes20
	copy(addr20[:], addr.ScriptAddress())

	t.Run("with enough coins", func(t *testing.T) {
		res, err := bcy.FindUnspent(ctx, &addr20, 1000000)
		require.NoError(t, err)
		assert.Truef(t, res.Sum >= 1000000, "invalid result sum %d", res.Sum)
		assert.Truef(t, res.Total >= res.Sum, "invalid total amount %d", res.Total)
		assert.True(t, len(res.Outputs) > 0, "missing res outputs")

		for _, output := range res.Outputs {
			// Avoid being throttled by blockcypher
			<-time.After(time.Second)

			tx, err := bcy.api.GetTX(output.TXHash.String(), nil)
			require.NoError(t, err)
			assert.True(t, testutil.ContainsString(tx.Addresses, testAddr), "can't find address in output addresses")
		}
	})

	t.Run("without enough coins", func(t *testing.T) {
		_, err = bcy.FindUnspent(ctx, &addr20, 1000000000000)
		require.Error(t, err)
	})
}
