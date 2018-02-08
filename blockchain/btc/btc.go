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

import "github.com/stratumn/go-indigocore/types"
import "github.com/btcsuite/btcd/chaincfg"

// Network represents a Bitcoin network.
type Network string

const (
	// NetworkTest3 is an identified for the test Bitcoin network.
	NetworkTest3 Network = "bitcoin:test3"

	// NetworkMain is an identified for the main Bitcoin network.
	NetworkMain Network = "bitcoin:main"
)

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
