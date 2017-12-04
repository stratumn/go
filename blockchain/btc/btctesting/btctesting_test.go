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
	"reflect"
	"testing"

	"github.com/stratumn/sdk/testutil"
	"github.com/stratumn/sdk/types"
	"github.com/stratumn/sdk/blockchain/btc"
)

func TestMockFindUnspent(t *testing.T) {
	a := &Mock{}

	var addr1 types.ReversedBytes20
	copy(addr1[:], testutil.RandomHash()[:])
	if _, _, err := a.FindUnspent(&addr1, 1000); err != nil {
		t.Fatalf("a.FindUnspent(): err: %s", err)
	}

	a.MockFindUnspent.Fn = func(*types.ReversedBytes20, int64) ([]btc.Output, int64, error) { return nil, 10000, nil }

	var addr2 types.ReversedBytes20
	copy(addr2[:], testutil.RandomHash()[:])
	if _, _, err := a.FindUnspent(&addr2, 2000); err != nil {
		t.Errorf("a.FindUnspent(): err: %s", err)
	}

	if got, want := a.MockFindUnspent.CalledCount, 2; got != want {
		t.Errorf(`a.MockFindUnspent.CalledCount = %d want %d`, got, want)
	}
	got, want := a.MockFindUnspent.CalledWithAddress, []*types.ReversedBytes20{&addr1, &addr2}
	if !reflect.DeepEqual(got, want) {
		t.Errorf(`a.MockFindUnspent.CalledWithAddress = %q want %q`, got, want)
	}
	if got, want := a.MockFindUnspent.LastCalledWithAddress.String(), addr2.String(); got != want {
		t.Errorf(`a.MockFindUnspent.LastCalledWithAddress = %q want %q`, got, want)
	}
	if got, want := a.MockFindUnspent.CalledWithAmount, []int64{1000, 2000}; !reflect.DeepEqual(got, want) {
		t.Errorf(`a.MockFindUnspent.CalledWithAmount = %q want %q`, got, want)
	}
	if got, want := a.MockFindUnspent.LastCalledWithAmount, int64(2000); got != want {
		t.Errorf(`a.MockFindUnspent.LastCalledWithAmount = %d want %d`, got, want)
	}
}

func TestMockBroadcast(t *testing.T) {
	a := &Mock{}

	tx1 := testutil.RandomHash()[:]
	if err := a.Broadcast(tx1); err != nil {
		t.Errorf("a.Broadcast(): err: %s", err)
	}

	a.MockBroadcast.Fn = func(raw []byte) error { return errors.New("error") }

	tx2 := testutil.RandomHash()[:]
	if err := a.Broadcast(tx2); err == nil {
		t.Error("a.Broadcast(): err = nil want Error")
	}

	if got, want := a.MockBroadcast.CalledCount, 2; got != want {
		t.Errorf(`a.MockBroadcast.CalledCount = %d want %d`, got, want)
	}
	got, want := a.MockBroadcast.CalledWith, [][]byte{tx1, tx2}
	if !reflect.DeepEqual(got, want) {
		t.Errorf(`a.MockBroadcast.CalledWith = %q want %q`, got, want)
	}
	if got, want := a.MockBroadcast.LastCalledWith, tx2; !reflect.DeepEqual(got, want) {
		t.Errorf(`a.MockBroadcast.LastCalledWith = %q want %q`, got, want)
	}
}
