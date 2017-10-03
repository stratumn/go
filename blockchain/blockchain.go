// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// LICENSE file.

// Package blockchain defines primitives to work with blockchains.
package blockchain

import (
	"fmt"

	"github.com/stratumn/sdk/types"
)

// Network represents a blockchain network.
type Network interface {
	fmt.Stringer
}

// Networker must be able to return a network identifier.
type Networker interface {
	// Network returns the network identifier of the blockchain.
	Network() Network
}

// Timestamper must be able to timestamp data.
type Timestamper interface {
	Networker

	// Timestamp timestamps data on a blockchain.
	Timestamp(date interface{}) (types.TransactionID, error)
}

// HashTimestamper must be able to timestamp a hash.
type HashTimestamper interface {
	Networker

	// TimestampHash timestamps a hash on a blockchain.
	TimestampHash(hash *types.Bytes32) (types.TransactionID, error)
}
