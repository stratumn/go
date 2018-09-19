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

package btctesting

import (
	"errors"
	"testing"

	"github.com/stratumn/go-chainscript/chainscripttest"
	"github.com/stratumn/go-core/blockchain/btc"
	"github.com/stratumn/go-core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockFindUnspent(t *testing.T) {
	a := &Mock{}

	var addr1 types.ReversedBytes20
	copy(addr1[:], chainscripttest.RandomHash())
	_, err := a.FindUnspent(&addr1, 1000)
	require.NoError(t, err, "a.FindUnspent()")

	a.MockFindUnspent.Fn = func(*types.ReversedBytes20, int64) (btc.UnspentResult, error) {
		return btc.UnspentResult{Sum: 10000}, nil
	}

	var addr2 types.ReversedBytes20
	copy(addr2[:], chainscripttest.RandomHash())
	_, err = a.FindUnspent(&addr2, 2000)
	assert.NoError(t, err)

	assert.Equal(t, 2, a.MockFindUnspent.CalledCount)
	assert.Equal(t, []*types.ReversedBytes20{&addr1, &addr2}, a.MockFindUnspent.CalledWithAddress)
	assert.Equal(t, addr2.String(), a.MockFindUnspent.LastCalledWithAddress.String())
	assert.Equal(t, []int64{1000, 2000}, a.MockFindUnspent.CalledWithAmount)
	assert.Equal(t, int64(2000), a.MockFindUnspent.LastCalledWithAmount)
}

func TestMockBroadcast(t *testing.T) {
	a := &Mock{}

	tx1 := []byte(chainscripttest.RandomHash())
	err := a.Broadcast(tx1)
	require.NoError(t, err, "a.Broadcast()")

	a.MockBroadcast.Fn = func(raw []byte) error { return errors.New("error") }

	tx2 := []byte(chainscripttest.RandomHash())
	err = a.Broadcast(tx2)
	require.Error(t, err, "a.Broadcast()")

	assert.Equal(t, 2, a.MockBroadcast.CalledCount)
	assert.Equal(t, [][]byte{tx1, tx2}, a.MockBroadcast.CalledWith)
	assert.Equal(t, tx2, a.MockBroadcast.LastCalledWith)
}
