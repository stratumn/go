// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// that can be found in the LICENSE file.

// Package btc defines primitives to work with Bitcoin.
package btc

import "github.com/stratumn/goprivate/types"
import "github.com/btcsuite/btcd/chaincfg"

// Network represents a Bitcoin network.
type Network string

const (
	// NetworkTest3 is an identified for the test Bitcoin network.
	NetworkTest3 Network = "bitcoin:test3"

	// NetworkMain is an identified for the main Bitcoin network.
	NetworkMain Network = "bitcoin:main"
)

// String implements github.com/stratumn/goprivate/blockchain.Network.
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
	TX     []byte
	TXHash types.Bytes32
	Index  int
}

// UnspentFinder is find unspent outputs.
type UnspentFinder interface {
	// FindUnspent find unspent outputs for the given address and the required amount.
	// It returns the outputs and the total amount of the outputs.
	FindUnspent(address160 *types.Bytes20, amount int64) (outputs []Output, total int64, err error)
}

// Broadcaster is able to broadcast raw Bitcoin transactions.
type Broadcaster interface {
	// Broadcast broadcasts a raw transaction.
	Broadcast(raw []byte) error
}
