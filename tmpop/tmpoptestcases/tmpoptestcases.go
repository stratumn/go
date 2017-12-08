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

package tmpoptestcases

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/cs/cstesting"
	"github.com/stratumn/sdk/store"
	"github.com/stratumn/sdk/tmpop"

	abci "github.com/tendermint/abci/types"
)

const (
	chainID = "testChain"
)

// Factory wraps functions to allocate and free an adapter, and is used to run
// the tests on tmpop using this adapter.
type Factory struct {
	// New creates an adapter and a key-value store.
	New func() (store.Adapter, store.KeyValueStore, error)

	// Free is an optional function to free the adapter and key-value store.
	Free func(adapter store.Adapter, kv store.KeyValueStore)

	adapter store.Adapter
	kv      store.KeyValueStore
}

// RunTests runs all the tests.
func (f Factory) RunTests(t *testing.T) {
	t.Run("TestLastBlock", f.TestLastBlock)
	t.Run("TestTendermintEvidence", f.TestTendermintEvidence)
	t.Run("TestQuery", f.TestQuery)
	t.Run("TestCheckTx", f.TestCheckTx)
	t.Run("TestDeliverTx", f.TestDeliverTx)
	t.Run("TestCommitTx", f.TestCommitTx)
	t.Run("TestValidation", f.TestValidation)
}

func (f Factory) free() {
	if f.Free != nil {
		f.Free(f.adapter, f.kv)
	}
}

// newTMPop creates a new TMPoP from scratch (no previous history)
func (f *Factory) newTMPop(t *testing.T, config *tmpop.Config) (*tmpop.TMPop, abci.RequestBeginBlock) {
	var err error
	if f.adapter, f.kv, err = f.New(); err != nil {
		t.Fatalf("f.New(): err: %s", err)
	}
	if f.adapter == nil {
		t.Fatal("a = nil want store.Adapter")
	}
	if f.kv == nil {
		t.Fatalf("kv = nil want store.KeyValueStore")
	}

	if config == nil {
		config = &tmpop.Config{}
	}
	h, err := tmpop.New(f.adapter, f.kv, config)
	if err != nil {
		t.Fatalf("tmpop.New(): err: %s", err)
	}

	return h, makeBeginBlock([]byte{}, uint64(1))
}

func makeQuery(h *tmpop.TMPop, name string, args interface{}, res interface{}) error {
	bytes, err := tmpop.BuildQueryBinary(args)
	if err != nil {
		return err
	}

	q := h.Query(abci.RequestQuery{
		Data: bytes,
		Path: name,
	})

	return json.Unmarshal(q.Value, &res)
}

func makeCreateRandomLinkTx(t *testing.T) (*cs.Link, []byte) {
	l := cstesting.RandomLink()
	return l, makeCreateLinkTx(t, l)
}

func makeCreateLinkTx(t *testing.T, l *cs.Link) []byte {
	tx := tmpop.Tx{
		TxType: tmpop.CreateLink,
		Link:   l,
	}
	res, err := json.Marshal(tx)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func makeBeginBlock(appHash []byte, height uint64) abci.RequestBeginBlock {
	return abci.RequestBeginBlock{
		Hash: []byte{},
		Header: &abci.Header{
			Height:  height,
			ChainId: chainID,
			AppHash: appHash,
		},
	}
}

func commitLink(t *testing.T, h *tmpop.TMPop, l *cs.Link, requestBeginBlock abci.RequestBeginBlock) abci.RequestBeginBlock {
	tx := makeCreateLinkTx(t, l)
	nextBeginBlock := commitTx(t, h, requestBeginBlock, tx)
	return nextBeginBlock
}

func commitRandomLink(t *testing.T, h *tmpop.TMPop, requestBeginBlock abci.RequestBeginBlock) (*cs.Link, abci.RequestBeginBlock) {
	l, tx := makeCreateRandomLinkTx(t)
	nextBeginBlock := commitTx(t, h, requestBeginBlock, tx)
	return l, nextBeginBlock
}

func commitTx(t *testing.T, h *tmpop.TMPop, requestBeginBlock abci.RequestBeginBlock, tx []byte) abci.RequestBeginBlock {
	return commitTxs(t, h, requestBeginBlock, [][]byte{tx})
}

func commitTxs(t *testing.T, h *tmpop.TMPop, requestBeginBlock abci.RequestBeginBlock, txs [][]byte) abci.RequestBeginBlock {
	h.BeginBlock(requestBeginBlock)

	for _, tx := range txs {
		h.DeliverTx(tx)
	}

	commitResult := h.Commit()
	if commitResult.IsErr() {
		t.Errorf("a.Commit(): failed: %v", commitResult.Log)
	}

	return makeBeginBlock(commitResult.Data, requestBeginBlock.Header.Height+1)
}

func verifyLinkStored(t *testing.T, h *tmpop.TMPop, link *cs.Link) {
	linkHash, _ := link.Hash()

	got := &cs.Segment{}
	err := makeQuery(h, tmpop.GetSegment, linkHash, got)
	if err != nil {
		t.Fatal(err)
	}

	if want, got := *link, got.Link; !reflect.DeepEqual(want, got) {
		gotJS, _ := json.MarshalIndent(got, "", "  ")
		wantJS, _ := json.MarshalIndent(want, "", "  ")
		t.Errorf("h.Commit(): expected to return %s, got %s", wantJS, gotJS)
	}
}
