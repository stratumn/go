// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// LICENSE file.

// Package dummytimestamper implements a fake blockchain timestamper which can be used for testing.
package dummytimestamper

import (
	"crypto/sha256"
	"encoding/json"

	"github.com/stratumn/go/types"
	"github.com/stratumn/goprivate/blockchain"
)

const networkString = "dummy"

// Network is the identifier of the dummy network.
type Network struct{}

// Timestamper is the type that implements fmt.Stringer.
func (Network) String() string {
	return networkString
}

// Timestamper is the type that implements github.com/stratumn/goprivate/blockchain.Timestamper.
type Timestamper struct{}

// Network implements fmt.Stringer.
func (Timestamper) Network() blockchain.Network {
	return Network{}
}

// Timestamp implements github.com/stratumn/goprivate/blockchain.Timestamper.
func (Timestamper) Timestamp(data interface{}) (blockchain.TransactionID, error) {
	js, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(js)
	return sum[:], nil
}

// TimestampHash implements github.com/stratumn/goprivate/blockchain.HashTimestamper.
func (Timestamper) TimestampHash(hash *types.Bytes32) (blockchain.TransactionID, error) {
	sum := sha256.Sum256(hash[:])
	return sum[:], nil
}
