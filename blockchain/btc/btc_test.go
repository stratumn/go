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

package btc_test

import (
	"testing"

	"github.com/btcsuite/btcutil"
	"github.com/stratumn/go-core/blockchain/btc"
	"github.com/stratumn/go-core/monitoring/errorcode"
	"github.com/stratumn/go-core/types"
	"github.com/stretchr/testify/assert"
)

func TestGetNetworkFromWIF(t *testing.T) {
	type testCase struct {
		name            string
		wif             string
		err             error
		expectedNetwork btc.Network
	}

	tests := []testCase{
		{
			name:            "test network WIF",
			wif:             "924v2d7ryXJjnbwB6M9GsZDEjAkfE9aHeQAG1j8muA4UEjozeAJ",
			expectedNetwork: btc.NetworkTest3,
		},
		{
			name:            "main network WIF",
			wif:             "5HueCGU8rMjxEXxiPuD5BDku4MkFqeZyd4dZ1jvhTVqvbTLvyTJ",
			expectedNetwork: btc.NetworkMain,
		},
		{
			name: "invalid WIF",
			wif:  "fakeWIF",
			err:  types.WrapError(btcutil.ErrMalformedPrivateKey, errorcode.InvalidArgument, "btc", btc.ErrBadWIF.Error()),
		},
		{
			name: "unknown bitcoin network",
			wif:  "5KrPNVvAhnRBNMYRJUq58YMfyUMyVMQrQhhfFtcbT9rK67poC3F",
			err:  types.WrapError(btc.ErrUnknownBitcoinNetwork, errorcode.InvalidArgument, "btc", "invalid network"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			net, err := btc.GetNetworkFromWIF(tt.wif)
			if tt.err == nil {
				assert.NoError(t, err)
				assert.Equal(t, net, tt.expectedNetwork)
			} else {
				assert.EqualError(t, err, tt.err.Error())
			}
		})
	}
}

func TestNetworkString(t *testing.T) {
	assert.Equal(t, "bitcoin:test3", btc.NetworkTest3.String())
}

func TestNetworkID(t *testing.T) {
	assert.Equal(t, byte(0x6F), btc.NetworkTest3.ID())
	assert.Equal(t, byte(0x00), btc.NetworkMain.ID())
}
