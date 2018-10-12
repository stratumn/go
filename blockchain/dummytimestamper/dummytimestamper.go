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

// Package dummytimestamper implements a fake blockchain timestamper which can be used for testing.
package dummytimestamper

import (
	"crypto/sha256"
	"encoding/json"

	"github.com/stratumn/go-core/blockchain"
	"github.com/stratumn/go-core/monitoring/errorcode"
	"github.com/stratumn/go-core/types"
)

const networkString = "dummytimestamper"

// Description describes this Timestamper
const Description = "Dummy Timestamper"

// Component name for monitoring.
const Component = "dummytimestamper"

// Network is the identifier of the dummy network.
type Network struct{}

// Timestamper is the type that implements fmt.Stringer.
func (Network) String() string {
	return networkString
}

// Timestamper is the type that implements github.com/stratumn/go-core/blockchain.Timestamper.
type Timestamper struct{}

// Network implements fmt.Stringer.
func (Timestamper) Network() blockchain.Network {
	return Network{}
}

// Timestamp implements github.com/stratumn/go-core/blockchain.Timestamper.
func (Timestamper) Timestamp(data interface{}) (types.TransactionID, error) {
	js, err := json.Marshal(data)
	if err != nil {
		return nil, types.WrapError(err, errorcode.InvalidArgument, Component, "json.Marshal")
	}
	sum := sha256.Sum256(js)
	return sum[:], nil
}

// TimestampHash implements github.com/stratumn/go-core/blockchain.HashTimestamper.
func (Timestamper) TimestampHash(hash *types.Bytes32) (types.TransactionID, error) {
	sum := sha256.Sum256(hash[:])
	return sum[:], nil
}

// GetInfo implements github.com/stratumn/go-core/blockchain.HashTimestamper.
func (t Timestamper) GetInfo() *blockchain.Info {
	return &blockchain.Info{
		Network:     t.Network(),
		Description: Description,
	}
}
