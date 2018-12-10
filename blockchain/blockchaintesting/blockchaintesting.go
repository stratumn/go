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

// Package blockchaintesting contains test utilies for packages that depend
// on the blockchain packages.
package blockchaintesting

import (
	"context"

	"github.com/stratumn/go-core/blockchain"
	"github.com/stratumn/go-core/blockchain/btc"
	"github.com/stratumn/go-core/types"
)

// MockTimestamper mocks the Timestamper and HashTimestamper interfaces.
type MockTimestamper struct {
	txid types.TransactionID
	err  error
}

// NewTimestamper creates a new mock timestamper.
func NewTimestamper() *MockTimestamper {
	return &MockTimestamper{}
}

// WithError configures the mock to return an error.
func (t *MockTimestamper) WithError(err error) *MockTimestamper {
	t.err = err
	return t
}

// WithTransactionID configures the mock to return the given transaction ID.
func (t *MockTimestamper) WithTransactionID(txid types.TransactionID) *MockTimestamper {
	t.txid = txid
	return t
}

// GetInfo returns dummy information.
func (t *MockTimestamper) GetInfo() *blockchain.Info {
	return &blockchain.Info{
		Network:     btc.Network("mock"),
		Description: "mock timestamper",
	}
}

// Timestamp returns a dummy transaction ID.
func (t *MockTimestamper) Timestamp(ctx context.Context, _ interface{}) (types.TransactionID, error) {
	return t.TimestampHash(ctx, nil)
}

// TimestampHash returns a dummy transaction ID.
func (t *MockTimestamper) TimestampHash(_ context.Context, _ []byte) (types.TransactionID, error) {
	if t.err != nil {
		return nil, t.err
	}

	return t.txid, nil
}
