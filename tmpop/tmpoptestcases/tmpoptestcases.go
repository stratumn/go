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
	height = uint64(1)

	chainID = "testChain"
)

var requestBeginBlock = abci.RequestBeginBlock{
	Hash: []byte{},
	Header: &abci.Header{
		Height:  height,
		ChainId: chainID,
		AppHash: []byte{},
	},
}

// Factory wraps functions to allocate and free an adapter, and is used to run
// the tests on tmpop using this adapter.
type Factory struct {
	// New creates an adapter.
	New func() (store.Adapter, error)

	// Free is an optional function to free an adapter.
	Free func(adapter store.Adapter)

	adapter store.Adapter
}

// RunTests runs all the tests.
func (f Factory) RunTests(t *testing.T) {
	t.Run("New", f.TestNew)
	t.Run("TestBeginBlockSavesLastBlockInfo", f.TestBeginBlockSavesLastBlockInfo)
	t.Run("TestCheckTx", f.TestCheckTx)
	t.Run("TestTx", f.TestTx)
	t.Run("TestEvidence", f.TestEvidence)
	t.Run("TestQuery", f.TestQuery)
	t.Run("TestValidation", f.TestValidation)

}

func (f Factory) free() {
	if f.Free != nil {
		f.Free(f.adapter)
	}
}

func (f *Factory) initTMPop(t *testing.T, config *tmpop.Config) *tmpop.TMPop {
	if f.adapter == nil {
		var err error
		if f.adapter, err = f.New(); err != nil {
			t.Fatalf("f.New(): err: %s", err)
		}
		if f.adapter == nil {
			t.Fatal("a = nil want store.Adapter")
		}
	}
	if config == nil {
		config = &tmpop.Config{}
	}
	h, err := tmpop.New(f.adapter, config)
	if err != nil {
		t.Fatalf("tmpop.New(): err: %s", err)
	}

	// reset header
	requestBeginBlock.Header = &abci.Header{
		Height:  height,
		ChainId: chainID,
		AppHash: []byte{},
	}

	return h
}

// TestNew tests what happens when tmpop is stopped and restarted with the same adapter
func (f Factory) TestNew(t *testing.T) {
	h1 := f.initTMPop(t, nil)

	want := commitMockTx(t, h1)
	commitMockTx(t, h1)

	h2 := f.initTMPop(t, nil)

	got := &cs.Segment{}
	err := makeQuery(h2, tmpop.GetSegment, want.GetLinkHash(), got)
	if err != nil {
		t.Fatal(err)
	}

	got.Meta.Evidences = nil
	if !reflect.DeepEqual(want, got) {
		gotJS, _ := json.MarshalIndent(got, "", "  ")
		wantJS, _ := json.MarshalIndent(want, "", "  ")
		t.Errorf("New(): expected new TMPop to have access to existing segment %s, got:\n %s", wantJS, gotJS)
	}

	gotLastBlock, _ := tmpop.ReadLastBlock(f.adapter)
	if gotLastBlock.Height != 1 {
		t.Errorf("a.New(): expected new TMPop to start on the last block height, got %v, expected %v",
			gotLastBlock.Height, 1)
	}

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

func makeSaveSegmentTx(t *testing.T) (*cs.Segment, []byte) {
	s := cstesting.RandomSegment()
	return s, makeSaveSegmentTxFromSegment(t, s)
}

func makeSaveSegmentTxFromSegment(t *testing.T, s *cs.Segment) []byte {
	tx := tmpop.Tx{
		TxType:  tmpop.SaveSegment,
		Segment: s,
	}
	res, err := json.Marshal(tx)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func makeDeleteSegmentTx(t *testing.T, s *cs.Segment) []byte {
	tx := tmpop.Tx{
		TxType:   tmpop.DeleteSegment,
		LinkHash: s.GetLinkHash(),
	}
	res, err := json.Marshal(tx)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func commitMockTx(t *testing.T, h *tmpop.TMPop) *cs.Segment {
	s, tx := makeSaveSegmentTx(t)

	commitTx(t, h, tx)
	return s
}

func commitTx(t *testing.T, h *tmpop.TMPop, tx []byte) {
	h.BeginBlock(requestBeginBlock)

	h.DeliverTx(tx)

	commitResult := h.Commit()
	if commitResult.IsErr() {
		t.Errorf("a.Commit(): failed: %v", commitResult.Log)
	}
	requestBeginBlock.Header.AppHash = commitResult.Data
	requestBeginBlock.Header.Height++
}
