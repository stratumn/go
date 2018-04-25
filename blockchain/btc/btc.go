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

// Package btc defines primitives to work with Bitcoin.
package btc

import (
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/pkg/errors"
	"github.com/stratumn/go-indigocore/types"
)

// Network represents a Bitcoin network.
type Network string

const (
	// NetworkTest3 is an identified for the test Bitcoin network.
	NetworkTest3 Network = "bitcoin:test3"

	// NetworkMain is an identified for the main Bitcoin network.
	NetworkMain Network = "bitcoin:main"
)

var (
	// ErrUnknownBitcoinNetwork is returned when the network ID associated to the WIF is unknown.
	ErrUnknownBitcoinNetwork = errors.New("WIF encoded private key uses unknown Bitcoin network")

	// ErrBadWIF is returned when the WIF encoded private key could not be decoded
	ErrBadWIF = errors.New("Failed to decode WIF encoded private key")
)

// GetNetworkFromWIF returns the network ID associated to a bitcoin wallet.
func GetNetworkFromWIF(key string) (Network, error) {
	WIF, err := btcutil.DecodeWIF(key)
	if err != nil {
		return "", errors.Wrap(err, ErrBadWIF.Error())
	}

	var network Network
	if WIF.IsForNet(&chaincfg.TestNet3Params) {
		network = NetworkTest3
	} else if WIF.IsForNet(&chaincfg.MainNetParams) {
		network = NetworkMain
	} else {
		return "", ErrUnknownBitcoinNetwork
	}
	return network, nil
}

// String implements fmt.Stringer.
func (n Network) String() string {
	return string(n)
}

// ID returns the byte ID of the network.
func (n Network) ID() byte {
	switch n {
	case NetworkTest3:
		return chaincfg.TestNet3Params.PubKeyHashAddrID
	case NetworkMain:
		return chaincfg.MainNetParams.PubKeyHashAddrID
	}

	return 0
}

// Output represents a transaction output.
type Output struct {
	TXHash   types.ReversedBytes32
	PKScript []byte
	Index    int
}

// UnspentFinder is find unspent outputs.
type UnspentFinder interface {
	// FindUnspent find unspent outputs for the given address and the
	// required amount. It returns the outputs and the total amount of the
	// outputs.
	FindUnspent(address *types.ReversedBytes20, amount int64) (outputs []Output, total int64, err error)
}

// Broadcaster is able to broadcast raw Bitcoin transactions.
type Broadcaster interface {
	// Broadcast broadcasts a raw transaction.
	Broadcast(raw []byte) error
}
