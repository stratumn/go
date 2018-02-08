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

// Package btctesting defines helpers to test Bitcoin.
package btctesting

import (
	"github.com/stratumn/go-indigocore/blockchain/btc"
	"github.com/stratumn/go-indigocore/types"
)

// Mock is used to mock a UnspentFinder and Broadcaster.
//
// It implements github.com/stratumn/go-indigocore/fossilizer.Adapter.
type Mock struct {
	// The mock for the FindUnspent function.
	MockFindUnspent MockFindUnspent

	// The mock for the Broadcast function.
	MockBroadcast MockBroadcast
}

// MockFindUnspent mocks the FindUnspent function.
type MockFindUnspent struct {
	// The number of times the function was called.
	CalledCount int

	// The address that was passed to each call.
	CalledWithAddress []*types.ReversedBytes20

	// The amount that was passed to each call.
	CalledWithAmount []int64

	// The last address that was passed.
	LastCalledWithAddress *types.ReversedBytes20

	// The last amount that was passed.
	LastCalledWithAmount int64

	// An optional implementation of the function.
	Fn func(*types.ReversedBytes20, int64) ([]btc.Output, int64, error)
}

// MockBroadcast mocks the Broadcast function.
type MockBroadcast struct {
	// The number of times the function was called.
	CalledCount int

	// The transaction that was passed to each call.
	CalledWith [][]byte

	// The last transaction that was passed.
	LastCalledWith []byte

	// An optional implementation of the function.
	Fn func([]byte) error
}

// FindUnspent implements
// github.com/stratumn/go-indigocore/blockchain/btc.UnspentFinder.FindUnspent.
func (a *Mock) FindUnspent(address *types.ReversedBytes20, amount int64) ([]btc.Output, int64, error) {
	a.MockFindUnspent.CalledCount++
	a.MockFindUnspent.CalledWithAddress = append(a.MockFindUnspent.CalledWithAddress, address)
	a.MockFindUnspent.LastCalledWithAddress = address
	a.MockFindUnspent.CalledWithAmount = append(a.MockFindUnspent.CalledWithAmount, amount)
	a.MockFindUnspent.LastCalledWithAmount = amount

	if a.MockFindUnspent.Fn != nil {
		return a.MockFindUnspent.Fn(address, amount)
	}

	return nil, 0, nil
}

// Broadcast implements
// github.com/stratumn/go-indigocore/blockchain/btc.Broadcaster.Broadcast.
func (a *Mock) Broadcast(raw []byte) error {
	a.MockBroadcast.CalledCount++
	a.MockBroadcast.CalledWith = append(a.MockBroadcast.CalledWith, raw)
	a.MockBroadcast.LastCalledWith = raw

	if a.MockBroadcast.Fn != nil {
		return a.MockBroadcast.Fn(raw)
	}

	return nil
}
