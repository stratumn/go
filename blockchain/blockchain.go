// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// LICENSE file.

// Package blockchain defines primitives to work with blockchains.
package blockchain

import (
	"encoding/hex"
	"encoding/json"
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

// TransactionID is a blockchain transaction ID.
type TransactionID []byte

// String returns a hex encoded string.
func (txid TransactionID) String() string {
	return hex.EncodeToString(txid)
}

// MarshalJSON implements encoding/json.Marshaler.MarshalJSON.
func (txid TransactionID) MarshalJSON() ([]byte, error) {
	return json.Marshal(txid.String())
}

// UnmarshalJSON implements encoding/json.Unmarshaler.UnmarshalJSON.
func (txid *TransactionID) UnmarshalJSON(data []byte) (err error) {
	var s string
	if err = json.Unmarshal(data, &s); err != nil {
		return
	}
	*txid, err = hex.DecodeString(s)
	return
}

// Timestamper must be able to timestamp data.
type Timestamper interface {
	Networker

	// Timestamp timestamps data on a blockchain.
	Timestamp(date interface{}) (TransactionID, error)
}

// HashTimestamper must be able to timestamp a hash.
type HashTimestamper interface {
	Networker

	// TimestampHash timestamps a hash on a blockchain.
	TimestampHash(hash *types.Bytes32) (TransactionID, error)
}
