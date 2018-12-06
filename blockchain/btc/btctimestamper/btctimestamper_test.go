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

package btctimestamper

import (
	"context"
	"encoding/hex"
	"testing"

	"github.com/stratumn/go-chainscript/chainscripttest"
	"github.com/stratumn/go-core/blockchain/btc"
	"github.com/stratumn/go-core/blockchain/btc/btctesting"
	"github.com/stratumn/go-core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNetwork_NetworkTest3(t *testing.T) {
	ts, err := New(&Config{
		WIF: "924v2d7ryXJjnbwB6M9GsZDEjAkfE9aHeQAG1j8muA4UEjozeAJ",
		Fee: int64(10000),
	})
	require.NoError(t, err)
	assert.Equal(t, btc.NetworkTest3, ts.Network())
}

func TestNetwork_NetworkMain(t *testing.T) {
	ts, err := New(&Config{
		WIF: "L3Wbnfn57Fc547FLSkm6iCzAaHmLArNUBCYx6q8LdxWoEMoFZmLH",
		Fee: int64(10000),
	})
	require.NoError(t, err)
	assert.Equal(t, btc.NetworkMain, ts.Network())
}

func TestTimestamperTimestampHash(t *testing.T) {
	ctx := context.Background()
	mock := &btctesting.Mock{}
	mock.MockFindUnspent.Fn = func(context.Context, *types.ReversedBytes20, int64) (btc.UnspentResult, error) {
		PKScriptHex := "76a914fc56f7f9f80cfba26f300c77b893c39ed89351ff88ac"
		PKScript, _ := hex.DecodeString(PKScriptHex)
		output := btc.Output{Index: 0, PKScript: PKScript}
		if err := output.TXHash.Unstring("c805dd0fbf728e6b7e6c4e5d4ddfaba0089291145453aafb762bcff7a8afe2f5"); err != nil {
			return btc.UnspentResult{}, err
		}

		return btc.UnspentResult{
			Outputs: []btc.Output{output},
			Sum:     6241000,
		}, nil
	}

	ts, err := New(&Config{
		WIF:           "924v2d7ryXJjnbwB6M9GsZDEjAkfE9aHeQAG1j8muA4UEjozeAJ",
		UnspentFinder: mock,
		Broadcaster:   mock,
		Fee:           int64(10000),
	})
	require.NoError(t, err)

	_, err = ts.TimestampHash(ctx, chainscripttest.RandomHash())
	assert.NoError(t, err)
	assert.Equal(t, 1, mock.MockBroadcast.CalledCount)
}
