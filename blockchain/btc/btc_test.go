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

package btc_test

import (
	"testing"

	"github.com/btcsuite/btcutil"
	"github.com/pkg/errors"
	"github.com/stratumn/go-indigocore/blockchain/btc"
	"github.com/stretchr/testify/assert"
)

func TestGetNetworkFromWIF(t *testing.T) {
	type testCase struct {
		name            string
		wif             string
		err             string
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
			err:  errors.Wrap(btcutil.ErrMalformedPrivateKey, btc.ErrBadWIF.Error()).Error(),
		},
		{
			name: "unknown bitcoin network",
			wif:  "5KrPNVvAhnRBNMYRJUq58YMfyUMyVMQrQhhfFtcbT9rK67poC3F",
			err:  btc.ErrUnknownBitcoinNetwork.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			net, err := btc.GetNetworkFromWIF(tt.wif)
			if tt.err == "" {
				assert.NoError(t, err)
				assert.Equal(t, net, tt.expectedNetwork)
			} else {
				assert.EqualError(t, err, tt.err)
			}
		})
	}
}

func TestNetworkString(t *testing.T) {
	if got, want := btc.NetworkTest3.String(), "bitcoin:test3"; got != want {
		t.Errorf("NetworkTest3.String() = %s want %s", got, want)
	}
}

func TestNetworkID(t *testing.T) {
	if got, want := btc.NetworkTest3.ID(), byte(0x6F); got != want {
		t.Errorf(`NetworkTest3.String() = "%x" want "%x"`, got, want)
	}
	if got, want := btc.NetworkMain.ID(), byte(0x00); got != want {
		t.Errorf(`NetworkTest3.String() = "%x" want "%x"`, got, want)
	}
}
