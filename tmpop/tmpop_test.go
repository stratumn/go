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

package tmpop

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/stratumn/sdk/cs/cstesting"
	"github.com/stratumn/sdk/dummystore"
	"github.com/stratumn/sdk/store"
	"github.com/stratumn/sdk/store/storetesting"
	"github.com/stratumn/sdk/testutil"
	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/merkleeyes/iavl"

	"strings"

	"github.com/stratumn/sdk/cs"
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

func TestNew(t *testing.T) {
	a := dummystore.New(&dummystore.Config{})
	h1 := createDefaultTMPop(a, t)

	want := commitMockTx(t, h1)
	commitMockTx(t, h1)

	h2 := createDefaultTMPop(a, t)

	got := &cs.Segment{}
	err := h2.makeQuery(GetSegment, want.GetLinkHash(), got)
	if err != nil {
		t.Fatal(err)
	}

	got.Meta.Evidences = nil
	if !reflect.DeepEqual(want, got) {
		gotJS, _ := json.MarshalIndent(got, "", "  ")
		wantJS, _ := json.MarshalIndent(want, "", "  ")
		t.Errorf("New(): expected new TMPop to have access to existing segment %s, got:\n %s", wantJS, gotJS)
	}

	gotHeight := h2.readHeight(t)
	if gotHeight != 1 {
		t.Errorf("a.New(): expected new TMPop to start on the last block height, got %v, expected %v",
			gotHeight, 1)
	}

}

func TestInfo(t *testing.T) {
	h := createDefaultTMPop(nil, t)

	got := h.Info(abci.RequestInfo{
		Version: "UT",
	})

	if !strings.Contains(got.Data, Name) {
		t.Errorf("a.Info(): expected to contain %s got %v", Name, got)
	}
}

func TestBeginBlock_SavesLastBlockInfo(t *testing.T) {
	h := createDefaultTMPop(dummystore.New(&dummystore.Config{}), t)

	height := uint64(2)

	req := requestBeginBlock
	req.Header.Height = height
	hash := req.GetHeader().GetAppHash()

	h.BeginBlock(req)

	got := h.readHeight(t)
	if got != (height - 1) {
		t.Errorf("a.Commit(): expected BeginBlock to save the last block height, got %v, expected %v",
			got, height-1)
	}

	hashGot := h.readAppHash(t)
	if bytes.Compare(hashGot, hash) != 0 {
		t.Errorf("a.Commit(): expected BeginBlock to save the last app hash, expected %v, got %v", hash, hashGot)
	}
}

func TestCheckTx(t *testing.T) {
	h := createDefaultTMPop(nil, t)

	_, tx := makeSaveSegmentTx(t)

	res := h.CheckTx(tx)

	if !res.IsOK() {
		t.Errorf("Expected CheckTx to return an OK result, got %v", res)
	}
}

func TestQuery(t *testing.T) {
	h := createDefaultTMPop(dummystore.New(&dummystore.Config{}), t)

	t.Run("GetInfo()", func(t *testing.T) {
		info := &Info{}
		err := h.makeQuery(GetInfo, nil, info)
		if err != nil {
			t.Fatal(err)
		}

		if info.Name != Name {
			t.Errorf("h.Query(): expected GetInfo to return name %v, got %v", info.Name, Name)
		}
	})

	t.Run("GetSegment()", func(t *testing.T) {
		want := commitMockTx(t, h)

		got := &cs.Segment{}
		err := h.makeQuery(GetSegment, want.GetLinkHash(), got)
		if err != nil {
			t.Fatal(err)
		}

		got.Meta.Evidences = nil
		if !reflect.DeepEqual(want, got) {
			gotJS, _ := json.MarshalIndent(got, "", "  ")
			wantJS, _ := json.MarshalIndent(want, "", "  ")
			t.Errorf("h.Query(): expected GetSegment to return %s, got:\n %s", wantJS, gotJS)
		}
	})

	t.Run("GetValue()", func(t *testing.T) {
		k, want, tx := makeSaveValueTx(t)
		commitTx(t, h, tx)

		b, err := BuildQueryBinary(k)
		if err != nil {
			t.Fatal(err)
		}

		q := h.Query(abci.RequestQuery{
			Path: GetValue,
			Data: b,
		})

		if got := q.Value; bytes.Compare(got, want) != 0 {
			t.Errorf("h.Query(): expected GetValue to return %s, got:\n %s", want, got)
		}
	})

	t.Run("FindSegments()", func(t *testing.T) {
		want := commitMockTx(t, h)

		wantedPrevLinkHashStr := want.Link.GetPrevLinkHashString()
		args := &store.SegmentFilter{
			Pagination: store.Pagination{
				Limit: store.DefaultLimit,
			},
			MapIDs:       []string{want.Link.GetMapID()},
			PrevLinkHash: &wantedPrevLinkHashStr,
			Tags:         want.Link.GetTags(),
		}
		gots := cs.SegmentSlice{}
		err := h.makeQuery(FindSegments, args, &gots)
		if err != nil {
			t.Fatal(err)
		}
		if len(gots) != 1 {
			t.Fatalf("h.Query(): unexpected size for FindSegments result, got %v", gots)
		}
		got := gots[0]

		got.Meta.Evidences = nil
		if !reflect.DeepEqual(want, got) {
			gotJS, _ := json.MarshalIndent(got, "", "  ")
			wantJS, _ := json.MarshalIndent(want, "", "  ")
			t.Errorf("h.Query(): expected FindSegments to return %s, got %s", wantJS, gotJS)
		}
	})

	t.Run("GetMapIDs()", func(t *testing.T) {
		a := &storetesting.MockAdapter{}
		h := createDefaultTMPop(a, t)
		segment, _ := makeSaveSegmentTx(t)
		mapID := segment.Link.GetMapID()
		limit := 1
		a.MockGetMapIDs.Fn = func(filter *store.MapFilter) ([]string, error) {
			if filter.Pagination.Limit != limit {
				t.Errorf("Expected limit %v, got %v", limit, filter.Pagination.Limit)
			}

			res := []string{mapID}

			return res, nil
		}
		args := &store.MapFilter{
			Pagination: store.Pagination{
				Limit: limit,
			},
		}

		var got []string
		err := h.makeQuery(GetMapIDs, args, &got)
		if err != nil {
			t.Fatal(err)
		}
		if mapID != got[0] {
			t.Errorf("h.Query(): expected GetMapIDs to return %v, got %v", mapID, got[0])
		}
	})

	t.Run("Unsupported Query", func(t *testing.T) {
		q := h.Query(abci.RequestQuery{
			Path: "Unsupported",
		})
		if got, want := q.GetCode(), abci.CodeType_UnknownRequest; got != want {
			t.Errorf("h.Query(): expected unsupported query to return %v, got %v", want, got)
		}

		q = h.Query(abci.RequestQuery{
			Path:   FindSegments,
			Height: 12,
		})
		if got, want := q.GetCode(), abci.CodeType_InternalError; got != want {
			t.Errorf("h.Query(): expected unsupported query to return %v, got %v", want, got)
		}
	})
}

func TestTx(t *testing.T) {
	store := dummystore.New(&dummystore.Config{})
	h := createDefaultTMPop(store, t)

	t.Run("WriteSaveSegment()", func(t *testing.T) {
		want := commitMockTx(t, h)

		got, err := store.GetSegment(want.GetLinkHash())
		if err != nil {
			t.Fatal(err)
		}

		ev := got.Meta.GetEvidence(h.header.ChainId)
		got.Meta.Evidences = nil
		if !reflect.DeepEqual(want, got) {
			gotJS, _ := json.MarshalIndent(got, "", "  ")
			wantJS, _ := json.MarshalIndent(want, "", "  ")
			t.Errorf("h.Commit(): expected to return %s, got %s", wantJS, gotJS)
		}
		got.Meta.AddEvidence(*ev)
	})

	t.Run("WriteDoubleSaveSegment()", func(t *testing.T) {
		s := cstesting.RandomSegment()
		s.Meta.Data["test"] = "test"
		tx := makeSaveSegmentTxFromSegment(t, s)
		h.BeginBlock(requestBeginBlock)

		h.DeliverTx(tx)
		res := h.Commit()

		got, err := store.GetSegment(s.GetLinkHash())
		if err != nil {
			t.Fatal(err)
		}
		ev := got.Meta.GetEvidence(h.header.ChainId)
		if ev.State != cs.PendingEvidence {
			t.Errorf("h.DeliverTx(): wrong evidence state after saving segment, got %s, want %s", ev.State, cs.PendingEvidence)
		}
		if got.Meta.Data["test"] != "test" {
			t.Errorf("h.DeliverTx(): wrong segment.Meta.Data after saving, got %s, want %s", got.Meta.Data, s.Meta.Data)
		}
		// We try to save a segment with the same link (and linkHash)
		// but with new evidences and meta data
		newRequest := requestBeginBlock
		newRequest.Header.AppHash = res.Data.Bytes()

		h.BeginBlock(newRequest)
		s.Meta.Evidences = nil
		s.Meta.AddEvidence(cs.Evidence{
			Provider: "test1",
			Backend:  "TMPop",
		})
		s.Meta.AddEvidence(cs.Evidence{
			Provider: "test2",
			Backend:  "TMPop",
			Proof:    nil,
		})
		s.Meta.Data["random"] = nil
		s.Meta.Data["test"] = "random"
		tx = makeSaveSegmentTxFromSegment(t, s)
		h.DeliverTx(tx)
		h.Commit()

		got, err = store.GetSegment(s.GetLinkHash())
		if err != nil {
			t.Fatal(err)
		}
		ev = got.Meta.GetEvidence(h.header.ChainId)
		if ev.State != cs.CompleteEvidence {
			t.Errorf("h.DeliverTx(): wrong evidence state after saving segment, got %s, want %s", ev.State, cs.CompleteEvidence)
		}
		if len(got.Meta.Evidences) != 3 {
			t.Errorf("h.DeliverTx(): wrong length of segment.Meta.Evidences, got %d want %d", len(got.Meta.Evidences), 3)
		}
		if got.Meta.Data["test"] != "random" || got.Meta.Data["random"] != nil {
			t.Errorf("h.DeliverTx(): wrong segment.Meta.Data after saving, got %s, want %s", got.Meta.Data, s.Meta.Data)
		}
	})

	t.Run("WriteDeleteSegment", func(t *testing.T) {
		segment := commitMockTx(t, h)

		tx := makeDeleteSegmentTx(t, segment)
		commitTx(t, h, tx)

		got, err := store.GetSegment(segment.GetLinkHash())
		if err != nil {
			t.Fatal(err)
		}

		if got != nil {
			t.Errorf("h.Commit(): expected to return nil, got %s", got)
		}
	})

	t.Run("WriteSaveValue()", func(t *testing.T) {
		k, want, tx := makeSaveValueTx(t)

		commitTx(t, h, tx)

		got, err := store.GetValue(k)
		if err != nil {
			t.Fatal(err)
		}

		if bytes.Compare(want, got) != 0 {
			t.Errorf("h.Commit(): expected to return %s, got %s", want, got)
		}
	})

	t.Run("WriteDeleteValue", func(t *testing.T) {
		k, _, txSave := makeSaveValueTx(t)
		commitTx(t, h, txSave)

		tx := makeDeleteValueTx(t, k)
		commitTx(t, h, tx)

		got, err := store.GetValue(k)
		if err != nil {
			t.Fatal(err)
		}
		if got != nil {
			t.Errorf("h.Commit(): expected to return nil, got %s", got)
		}
	})
}

func TestEvidence(t *testing.T) {
	h := createDefaultTMPop(dummystore.New(&dummystore.Config{}), t)
	s1 := commitMockTx(t, h)

	got := &cs.Segment{}
	err := h.makeQuery(GetSegment, s1.GetLinkHash(), got)
	if err != nil {
		t.Fatal(err)
	}

	evidence := got.Meta.GetEvidence(h.header.ChainId)

	proof := evidence.Proof.(*TendermintFullProof)

	if proof == nil {
		t.Fatalf("h.Commit(): expected original proof not to be nil")
	}
	if proof.Original.BlockHeight != height {
		t.Errorf("h.Commit(): Expected originalEvidence.BlockHeight to contain %v, got %v", height, proof.Original.BlockHeight)
	}

	gotState, wantState := evidence.State, cs.PendingEvidence
	if strings.Compare(gotState, wantState) != 0 {
		t.Errorf("h.Commit(): Expected state to be %s since the next block has not been commited, got %s", wantState, gotState)
	}

	// Create a new Block that confirms the AppHash
	commitMockTx(t, h)

	err = h.makeQuery(GetSegment, s1.GetLinkHash(), got)
	if err != nil {
		t.Fatal(err)
	}

	evidence = got.Meta.GetEvidence(h.header.ChainId)

	gotState, wantState = evidence.State, cs.CompleteEvidence
	if strings.Compare(gotState, wantState) != 0 {
		t.Errorf("h.Commit(): Expected state to be %s since the next block has been commited, got %s", wantState, gotState)

	}
	if !evidence.Proof.Verify(s1.GetLinkHash()) {
		t.Errorf("TendermintProof.Verify(): Expected proof %v to be valid", evidence.Proof.FullProof())

	}
}

func TestTendermintProof(t *testing.T) {
	h := createDefaultTMPop(dummystore.New(&dummystore.Config{}), t)

	t.Run("TestTime()", func(t *testing.T) {
		s := commitMockTx(t, h)

		queried := &cs.Segment{}
		err := h.makeQuery(GetSegment, s.GetLinkHash(), queried)
		if err != nil {
			t.Fatal(err)
		}

		e := queried.Meta.GetEvidence(h.header.ChainId)
		got := e.Proof.Time()
		if got != 0 {
			t.Errorf("TendermintProof.Time(): Expected timestamp to be %d, got %d", 0, got)
		}

		commitMockTx(t, h)
		err = h.makeQuery(GetSegment, s.GetLinkHash(), queried)
		if err != nil {
			t.Fatal(err)
		}

		e = queried.Meta.GetEvidence(h.header.ChainId)
		want := h.header.GetTime()
		got = e.Proof.Time()
		if got != want {
			t.Errorf("TendermintProof.Time(): Expected timestamp to be %d, got %d", want, got)
		}

	})

	t.Run("TestFullProof()", func(t *testing.T) {
		s := commitMockTx(t, h)

		queried := &cs.Segment{}
		err := h.makeQuery(GetSegment, s.GetLinkHash(), queried)
		if err != nil {
			t.Fatal(err)
		}

		e := queried.Meta.GetEvidence(h.header.ChainId)
		got := e.Proof.FullProof()
		if got == nil {
			t.Errorf("TendermintProof.FullProof(): Expected proof to be a json-formatted bytes array, got %v", got)
		}

		commitMockTx(t, h)
		err = h.makeQuery(GetSegment, s.GetLinkHash(), queried)
		if err != nil {
			t.Fatal(err)
		}

		e = queried.Meta.GetEvidence(h.header.ChainId)
		wantDifferent := got
		got = e.Proof.FullProof()
		if got == nil {
			t.Errorf("TendermintProof.FullProof(): Expected proof to be a json-formatted bytes array, got %v", got)
		}
		if bytes.Compare(got, wantDifferent) == 0 {
			t.Errorf("TendermintProof.FullProof(): Expected proof after appHash validation to be complete, got %s", string(got))
		}
		if err := json.Unmarshal(got, &TendermintProof{}); err != nil {
			t.Errorf("TendermintProof.FullProof(): Could not unmarshal bytes proof, err = %+v", err)
		}

	})

	t.Run("TestVerify()", func(t *testing.T) {
		s := commitMockTx(t, h)

		queried := &cs.Segment{}
		err := h.makeQuery(GetSegment, s.GetLinkHash(), queried)
		if err != nil {
			t.Fatal(err)
		}

		e := queried.Meta.GetEvidence(h.header.ChainId)
		got := e.Proof.Verify(s.GetLinkHash())
		if got == true {
			t.Errorf("TendermintProof.Verify(): Expected incomplete original proof to be false, got %v", got)
		}

		commitMockTx(t, h)
		if err = h.makeQuery(GetSegment, s.GetLinkHash(), queried); err != nil {
			t.Fatal(err)
		}

		e = queried.Meta.GetEvidence(h.header.ChainId)

		if got = e.Proof.Verify(s.GetLinkHash()); got != true {
			t.Errorf("TendermintProof.Verify(): Expected original proof to be true, got %v", got)
		}

	})
}

func TestValidation(t *testing.T) {
	h := createDefaultTMPop(dummystore.New(&dummystore.Config{}), t)
	h.config.ValidatorFilename = filepath.Join("testdata", "rules.json")

	s := cstesting.RandomSegment()
	s.Link.Meta["process"] = "testProcess"
	s.Link.Meta["action"] = "init"
	s.Link.State["string"] = "test"
	tx := makeSaveSegmentTxFromSegment(t, s)

	h.BeginBlock(requestBeginBlock)
	res := h.DeliverTx(tx)

	if res.IsErr() {
		t.Errorf("a.Commit(): failed: %v", res.Log)
	}

	s = cstesting.RandomSegment()
	s.Link.Meta["process"] = "testProcess"
	s.Link.Meta["action"] = "init"
	s.Link.State["string"] = 42
	tx = makeSaveSegmentTxFromSegment(t, s)

	h.BeginBlock(requestBeginBlock)
	res = h.DeliverTx(tx)

	if !res.IsErr() {
		t.Error("a.DeliverTx(): want error")
	}

	if res.Code != CodeTypeValidation {
		t.Errorf("res.Code = got %d want %d", res.Code, CodeTypeValidation)
	}
}

func (tmpop *TMPop) makeQuery(name string, args interface{}, res interface{}) error {
	bytes, err := BuildQueryBinary(args)
	if err != nil {
		return err
	}

	q := tmpop.Query(abci.RequestQuery{
		Data: bytes,
		Path: name,
	})
	return json.Unmarshal(q.Value, &res)
}

func readIAVLProof(raw map[string]interface{}) (*iavl.IAVLProof, error) {
	var nodes []iavl.IAVLProofInnerNode

	nodesI := raw["InnerNodes"].([]interface{})

	for _, nodeI := range nodesI {

		node := nodeI.(map[string]interface{})

		leftHash, err := base64.StdEncoding.DecodeString(node["Left"].(string))
		if err != nil {
			return nil, err
		}

		rightHash, err := base64.StdEncoding.DecodeString(node["Right"].(string))
		if err != nil {
			return nil, err
		}

		nodes = append(nodes, iavl.IAVLProofInnerNode{
			Height: int8(node["Height"].(float64)),
			Size:   int(node["Size"].(float64)),
			Left:   leftHash,
			Right:  rightHash,
		})
	}

	leafHash, err := base64.StdEncoding.DecodeString(raw["LeafHash"].(string))
	if err != nil {
		return nil, err
	}

	rootHash, err := base64.StdEncoding.DecodeString(raw["RootHash"].(string))
	if err != nil {
		return nil, err
	}

	return &iavl.IAVLProof{
		LeafHash:   leafHash,
		InnerNodes: nodes,
		RootHash:   rootHash,
	}, nil
}

func readHeader(raw map[string]interface{}) (*abci.Header, error) {
	appHash, err := base64.StdEncoding.DecodeString(raw["app_hash"].(string))
	if err != nil {
		return nil, err
	}
	return &abci.Header{
		AppHash: appHash,
	}, nil
}

func createDefaultStore() store.Adapter {
	return &storetesting.MockAdapter{}
}

func (tmpop *TMPop) readAppHash(t *testing.T) []byte {
	res, err := readLastBlock(tmpop.adapter)
	if err != nil {
		t.Fatal(err)
	}
	return res.AppHash
}

func (tmpop *TMPop) readHeight(t *testing.T) uint64 {
	res, err := readLastBlock(tmpop.adapter)
	if err != nil {
		t.Fatal(err)
	}
	return res.Height
}

func createDefaultTMPop(a store.Adapter, t *testing.T) *TMPop {
	if a == nil {
		a = createDefaultStore()
	}

	tmpop, err := New(a, &Config{})
	if err != nil {
		t.Fatal(err)
	}

	// reset header
	requestBeginBlock.Header = &abci.Header{
		Height:  height,
		ChainId: chainID,
		AppHash: []byte{},
	}
	return tmpop
}

func makeSaveSegmentTx(t *testing.T) (*cs.Segment, []byte) {
	s := cstesting.RandomSegment()
	return s, makeSaveSegmentTxFromSegment(t, s)
}

func makeSaveSegmentTxFromSegment(t *testing.T, s *cs.Segment) []byte {
	tx := Tx{
		TxType:  SaveSegment,
		Segment: s,
	}
	res, err := json.Marshal(tx)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func makeDeleteSegmentTx(t *testing.T, s *cs.Segment) []byte {
	tx := Tx{
		TxType:   DeleteSegment,
		LinkHash: s.GetLinkHash(),
	}
	res, err := json.Marshal(tx)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func makeSaveValueTx(t *testing.T) (key, value, txBytes []byte) {
	k := testutil.RandomKey()
	v := testutil.RandomValue()

	tx := Tx{
		TxType: SaveValue,
		Key:    k,
		Value:  v,
	}
	txBytes, err := json.Marshal(tx)
	if err != nil {
		t.Fatal(err)
	}
	return key, value, txBytes
}

func makeDeleteValueTx(t *testing.T, key []byte) []byte {
	tx := Tx{
		TxType: DeleteValue,
		Key:    key,
	}
	res, err := json.Marshal(tx)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func commitMockTx(t *testing.T, h *TMPop) *cs.Segment {
	s, tx := makeSaveSegmentTx(t)

	commitTx(t, h, tx)
	return s
}

func commitTx(t *testing.T, h *TMPop, tx []byte) {
	h.BeginBlock(requestBeginBlock)

	h.DeliverTx(tx)

	commitResult := h.Commit()
	if commitResult.IsErr() {
		t.Errorf("a.Commit(): failed: %v", commitResult.Log)
	}
	requestBeginBlock.Header.AppHash = commitResult.Data
	requestBeginBlock.Header.Height++
}
