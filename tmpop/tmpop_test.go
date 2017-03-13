// Copyright 2017 Stratumn SAS. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package tmpop

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"testing"

	"github.com/stratumn/sdk/cs/cstesting"
	"github.com/stratumn/sdk/filestore"
	"github.com/stratumn/sdk/store"
	"github.com/stratumn/sdk/store/storetesting"
	"github.com/stratumn/sdk/types"
	tmtypes "github.com/tendermint/abci/types"

	"strings"

	"reflect"

	"fmt"

	"github.com/stratumn/sdk/cs"
)

func createDefaultStore() store.Adapter {
	return &storetesting.MockAdapter{}
}

func (t *TMPop) readAppHash() []byte {
	return t.LoadLastBlock().AppHash
}

func (t *TMPop) readHeight() uint64 {
	return t.LoadLastBlock().Height
}

func createDefaultTMPop(a store.Adapter) *TMPop {
	if a == nil {
		a = createDefaultStore()
	}
	dir, err := ioutil.TempDir("", "db")
	if err != nil {
		log.Fatal("cannot create temp directory")
	}

	return New(a, &Config{DbDir: dir})
}

func makeMockTx(t *testing.T) (*cs.Segment, []byte) {
	s := cstesting.RandomSegment()

	res, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	return s, res
}

func TestInfo(t *testing.T) {
	h := createDefaultTMPop(nil)

	got := h.Info()

	if !strings.Contains(got.Data, "TMPop") {
		t.Errorf("a.Info(): expected to contain TMPop got %d", got)
	}
}

func TestCommit_SavesLastBlockInfo(t *testing.T) {
	h := createDefaultTMPop(nil)

	height := uint64(12)

	_, tx := makeMockTx(t)
	h.BeginBlock(nil, &tmtypes.Header{
		Height: height,
	})
	h.DeliverTx(tx)

	commitResult := h.Commit()
	if commitResult.IsErr() {
		t.Errorf("a.Commit(): failed: %v", commitResult.Log)
	}

	got := h.readHeight()
	if got != height {
		t.Errorf("a.Commit(): expected commit to save the last block height, got %v, expected %v",
			got, height)
	}

	hashGot := h.readAppHash()
	if len(hashGot) == 0 {
		t.Errorf("a.Commit(): expected commit to save the last app hash, got %v", hashGot)
	}
}

func TestCommit_AppendsEvidence(t *testing.T) {
	a := &storetesting.MockAdapter{}
	h := createDefaultTMPop(a)
	chainID := "MyChain"

	height := uint64(12)

	_, tx := makeMockTx(t)
	h.BeginBlock(nil, &tmtypes.Header{
		Height:  height,
		ChainId: chainID,
	})
	h.DeliverTx(tx)

	a.MockSaveSegment.Fn = func(s *cs.Segment) error {
		evidence := fmt.Sprint(s.Meta["evidence"])
		if !strings.Contains(evidence, fmt.Sprint(height)) {
			return fmt.Errorf("Expected evidence to contain %v, got %v", height, evidence)
		}
		if !strings.Contains(evidence, chainID) {
			return fmt.Errorf("Expected evidence to contain %v, got %v", chainID, evidence)
		}
		return nil
	}

	commitResult := h.Commit()
	if commitResult.IsErr() {
		t.Errorf("a.Commit(): failed: %v", commitResult.Log)
	}
}

func TestQuery(t *testing.T) {
	a := &storetesting.MockAdapter{}
	h := createDefaultTMPop(a)

	segment, _ := makeMockTx(t)

	t.Run("GetInfo()", func(t *testing.T) {
		fakeName := "Fake Name"
		a.MockGetInfo.Fn = func() (interface{}, error) { return &filestore.Info{Name: fakeName}, nil }

		info := &Info{}
		err := h.makeQuery(GetInfo, nil, info)
		if err != nil {
			t.Fatal(err)
		}

		if info.AdapterInfo.(map[string]interface{})["name"] != fakeName {
			t.Errorf("h.Query(): expected GetInfo to return name %v, got %v", fakeName,
				info.AdapterInfo.(map[string]interface{}))
		}
		if info.Name != Name {
			t.Errorf("h.Query(): expected GetInfo to return name %v, got %v", info.Name, Name)
		}
	})

	t.Run("GetSegment()", func(t *testing.T) {
		a.MockGetSegment.Fn = func(linkHash *types.Bytes32) (*cs.Segment, error) {
			if linkHash.String() == segment.GetLinkHash().String() {
				return segment, nil
			}
			t.Errorf("Unexpected link Hash, wanted: %v, got %v", segment.GetLinkHash(), linkHash)
			return nil, nil
		}

		got := &cs.Segment{}
		err := h.makeQuery(GetSegment, segment.GetLinkHash(), got)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(segment, got) {
			t.Errorf("h.Query(): expected GetSegment to return %v, got %v", segment, got)
		}
	})

	t.Run("FindSegments()", func(t *testing.T) {
		a.MockFindSegments.Fn = func(filter *store.Filter) (cs.SegmentSlice, error) {
			res := make(cs.SegmentSlice, 1, 1)
			if filter.MapID == segment.Link.GetMapID() {
				res[0] = segment
			} else {
				t.Errorf("Unexpected Map ID, wanted: %v, got %v", segment.Link.GetMapID(), filter.MapID)
			}
			return res, nil
		}

		args := &store.Filter{
			Pagination:   store.Pagination{},
			MapID:        segment.Link.GetMapID(),
			PrevLinkHash: segment.Link.GetPrevLinkHash(),
			Tags:         segment.Link.GetTags(),
		}
		got := make(cs.SegmentSlice, 0)
		err := h.makeQuery(FindSegments, args, &got)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(segment, got[0]) {
			t.Errorf("h.Query(): expected FindSegments to return %v, got %v", segment, got[0])
		}

	})

	t.Run("GetMapIDs()", func(t *testing.T) {
		mapID := segment.Link.GetMapID()
		limit := 1
		a.MockGetMapIDs.Fn = func(pagination *store.Pagination) ([]string, error) {
			if pagination.Limit != limit {
				t.Errorf("Expected limit %v, got %v", limit, pagination.Limit)
			}

			res := []string{mapID}

			return res, nil
		}
		args := &store.Pagination{
			Limit: limit,
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
}

func (t *TMPop) makeQuery(name string, args interface{}, res interface{}) error {
	bytes, err := BuildQueryBinary(name, args)
	if err != nil {
		return err
	}

	q := t.Query(bytes)
	if q.IsErr() {
		return errors.New(q.Error())
	}
	return json.Unmarshal(q.Data, &res)
}
