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

// Package blockchain defines primitives to work with blockchains.
package blockchain

import (
	"fmt"

	"github.com/stratumn/sdk/types"
)

// Info is the info returned by GetInfo.
type Info struct {
	Network     Network
	Description string
}

// Network represents a blockchain network.
type Network interface {
	fmt.Stringer
}

// Timestamper must be able to timestamp data.
type Timestamper interface {
	// Timestamp timestamps data on a blockchain.
	Timestamp(date interface{}) (types.TransactionID, error)
}

// HashTimestamper must be able to timestamp a hash.
type HashTimestamper interface {
	// GetInfo returns information on the Timestamper
	GetInfo() *Info

	// TimestampHash timestamps a hash on a blockchain.
	TimestampHash(hash *types.Bytes32) (types.TransactionID, error)
}
