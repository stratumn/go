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

package storetesting

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/cs/cstesting"
	"github.com/stratumn/sdk/store"
	"github.com/stratumn/sdk/testutil"
	"github.com/stratumn/sdk/types"
)

func TestMockAdapter_GetInfo(t *testing.T) {
	a := &MockAdapter{}

	if _, err := a.GetInfo(); err != nil {
		t.Fatalf("a.GetInfo(): err: %s", err)
	}

	a.MockGetInfo.Fn = func() (interface{}, error) { return map[string]string{"name": "test"}, nil }
	info, err := a.GetInfo()
	if err != nil {
		t.Fatalf("a.GetInfo(): err: %s", err)
	}

	if got, want := info.(map[string]string)["name"], "test"; got != want {
		t.Errorf(`a.GetInfo(): info["name"] = %q want %q`, got, want)
	}
	if got, want := a.MockGetInfo.CalledCount, 2; got != want {
		t.Errorf(`a.MockGetInfo.CalledCount = %d want %d`, got, want)
	}
}

func TestMockAdapter_AddSaveChan(t *testing.T) {
	a := &MockAdapter{}
	c := make(chan *cs.Segment)

	a.AddDidSaveChannel(c)

	if got, want := a.MockAddDidSaveChannel.CalledCount, 1; got != want {
		t.Errorf(`a.MockAddDidSaveChannel.CalledCount = %d want %d`, got, want)
	}
	if got, want := a.MockAddDidSaveChannel.CalledWith, []chan *cs.Segment{c}; !reflect.DeepEqual(got, want) {
		t.Errorf("a.MockAddDidSaveChannel.CalledWith = %p\n want %p", got, want)
	}
	if got, want := a.MockAddDidSaveChannel.LastCalledWith, c; got != want {
		t.Errorf("a.MockAddDidSaveChannel.LastCalledWith = %p\n want %p", got, want)
	}
}

func TestMockAdapter_SaveSegment(t *testing.T) {
	a := &MockAdapter{}
	s := cstesting.RandomSegment()

	err := a.SaveSegment(s)
	if err != nil {
		t.Fatalf("a.SaveSegment(): err: %s", err)
	}

	a.MockSaveSegment.Fn = func(s *cs.Segment) error { return nil }
	err = a.SaveSegment(s)
	if err != nil {
		t.Fatalf("a.SaveSegment(): err: %s", err)
	}

	if got, want := a.MockSaveSegment.CalledCount, 2; got != want {
		t.Errorf(`a.MockSaveSegment.CalledCount = %d want %d`, got, want)
	}
	if got, want := a.MockSaveSegment.CalledWith, []*cs.Segment{s, s}; !reflect.DeepEqual(got, want) {
		gotJS, _ := json.MarshalIndent(got, "", "  ")
		wantJS, _ := json.MarshalIndent(want, "", "  ")
		t.Errorf("a.MockSaveSegment.CalledWith = %s\n want %s", gotJS, wantJS)
	}
	if got, want := a.MockSaveSegment.LastCalledWith, s; got != want {
		gotJS, _ := json.MarshalIndent(got, "", "  ")
		wantJS, _ := json.MarshalIndent(want, "", "  ")
		t.Errorf("a.MockSaveSegment.LastCalledWith = %s\n want %s", gotJS, wantJS)
	}
}

func TestMockAdapter_GetSegment(t *testing.T) {
	a := &MockAdapter{}

	linkHash1 := testutil.RandomHash()
	_, err := a.GetSegment(linkHash1)
	if err != nil {
		t.Fatalf("a.GetSegment(): err: %s", err)
	}

	s1 := cstesting.RandomSegment()
	a.MockGetSegment.Fn = func(linkHash *types.Bytes32) (*cs.Segment, error) { return s1, nil }
	linkHash2 := testutil.RandomHash()
	s2, err := a.GetSegment(linkHash2)
	if err != nil {
		t.Fatalf("a.GetSegment(): err: %s", err)
	}

	if got, want := s2, s1; got != want {
		gotJS, _ := json.MarshalIndent(got, "", "  ")
		wantJS, _ := json.MarshalIndent(want, "", "  ")
		t.Errorf("s2 = %s\n want %s", gotJS, wantJS)
	}
	if got, want := a.MockGetSegment.CalledCount, 2; got != want {
		t.Errorf(`a.MockGetSegment.CalledCount = %d want %d`, got, want)
	}
	if got, want := a.MockGetSegment.CalledWith, []*types.Bytes32{linkHash1, linkHash2}; !reflect.DeepEqual(got, want) {
		t.Errorf("a.MockGetSegment.CalledWith = %q\n want %q", got, want)
	}
	if got, want := *a.MockGetSegment.LastCalledWith, *linkHash2; got != want {
		t.Errorf("a.MockGetSegment.LastCalledWith = %q want %q", got, want)
	}
}

func TestMockAdapter_DeleteSegment(t *testing.T) {
	a := &MockAdapter{}

	linkHash1 := testutil.RandomHash()
	_, err := a.DeleteSegment(linkHash1)
	if err != nil {
		t.Fatalf("a.DeleteSegment(): err: %s", err)
	}

	s1 := cstesting.RandomSegment()
	a.MockDeleteSegment.Fn = func(linkHash *types.Bytes32) (*cs.Segment, error) { return s1, nil }
	linkHash2 := testutil.RandomHash()
	s2, err := a.DeleteSegment(linkHash2)
	if err != nil {
		t.Fatalf("a.DeleteSegment(): err: %s", err)
	}

	if got, want := s2, s1; got != want {
		gotJS, _ := json.MarshalIndent(got, "", "  ")
		wantJS, _ := json.MarshalIndent(want, "", "  ")
		t.Errorf("s2 = %s\n want %s", gotJS, wantJS)
	}
	if got, want := a.MockDeleteSegment.CalledCount, 2; got != want {
		t.Errorf(`a.MockDeleteSegment.CalledCount = %d want %d`, got, want)
	}
	if got, want := a.MockDeleteSegment.CalledWith, []*types.Bytes32{linkHash1, linkHash2}; !reflect.DeepEqual(got, want) {
		t.Errorf("a.MockDeleteSegment.CalledWith = %q\n want %q", got, want)
	}
	if got, want := a.MockDeleteSegment.LastCalledWith, linkHash2; got != want {
		t.Errorf("a.MockDeleteSegment.LastCalledWith = %q want %q", got, want)
	}
}

func TestMockAdapter_FindSegments(t *testing.T) {
	a := &MockAdapter{}

	_, err := a.FindSegments(nil)
	if err != nil {
		t.Fatalf("a.FindSegments(): err: %s", err)
	}

	s := cstesting.RandomSegment()
	a.MockFindSegments.Fn = func(*store.SegmentFilter) (cs.SegmentSlice, error) { return cs.SegmentSlice{s}, nil }
	prevLinkHash := testutil.RandomHash().String()
	f := store.SegmentFilter{PrevLinkHash: &prevLinkHash}
	s1, err := a.FindSegments(&f)
	if err != nil {
		t.Fatalf("a.FindSegments(): err: %s", err)
	}

	if got, want := s1, (cs.SegmentSlice{s}); !reflect.DeepEqual(got, want) {
		gotJS, _ := json.MarshalIndent(got, "", "  ")
		wantJS, _ := json.MarshalIndent(want, "", "  ")
		t.Errorf("s1 = %s\n want %s", gotJS, wantJS)
	}
	if got, want := a.MockFindSegments.CalledCount, 2; got != want {
		t.Errorf(`a.MockFindSegments.CalledCount = %d want %d`, got, want)
	}
	if got, want := a.MockFindSegments.CalledWith, []*store.SegmentFilter{nil, &f}; !reflect.DeepEqual(got, want) {
		t.Errorf("a.MockFindSegments.CalledWith = %q\n want %q", got, want)
	}
	if got, want := a.MockFindSegments.LastCalledWith, &f; got != want {
		t.Errorf("a.MockFindSegments.LastCalledWith = %q\n want %q", got, want)
	}
}

func TestMockAdapter_GetMapIDs(t *testing.T) {
	a := &MockAdapter{}

	_, err := a.GetMapIDs(nil)
	if err != nil {
		t.Fatalf("a.GetMapIDs(): err: %s", err)
	}

	a.MockGetMapIDs.Fn = func(*store.MapFilter) ([]string, error) { return []string{"one", "two"}, nil }
	// FIXME test with process
	filter := store.MapFilter{
		Pagination: store.Pagination{Offset: 10},
	}
	s, err := a.GetMapIDs(&filter)
	if err != nil {
		t.Fatalf("a.GetMapIDs(): err: %s", err)
	}

	if got, want := s, []string{"one", "two"}; !reflect.DeepEqual(got, want) {
		t.Errorf("s1 = %q\n want %q", got, want)
	}
	if got, want := a.MockGetMapIDs.CalledCount, 2; got != want {
		t.Errorf(`a.MockGetMapIDs.CalledCount = %d want %d`, got, want)
	}
	if got, want := a.MockGetMapIDs.CalledWith, []*store.MapFilter{nil, &filter}; !reflect.DeepEqual(got, want) {
		t.Errorf("a.MockGetMapIDs.CalledWith = %q\n want %q", got, want)
	}
	if got, want := a.MockGetMapIDs.LastCalledWith, &filter; got != want {
		t.Errorf("a.MockGetMapIDs.LastCalledWith = %q\n want %q", got, want)
	}
}

func TestMockAdapter_GetValue(t *testing.T) {
	a := &MockAdapter{}

	k1 := testutil.RandomKey()
	_, err := a.GetValue(k1)
	if err != nil {
		t.Fatalf("a.GetValue(): err: %s", err)
	}

	v1 := testutil.RandomValue()
	a.MockGetValue.Fn = func(key []byte) ([]byte, error) { return v1, nil }
	k2 := testutil.RandomKey()
	v2, err := a.GetValue(k2)
	if err != nil {
		t.Fatalf("a.GetValue(): err: %s", err)
	}

	if got, want := v2, v1; bytes.Compare(got, want) != 0 {
		t.Errorf("v2 = %s\n want %s", got, want)
	}
	if got, want := a.MockGetValue.CalledCount, 2; got != want {
		t.Errorf(`a.MockGetValue.CalledCount = %d want %d`, got, want)
	}
	if got, want := a.MockGetValue.CalledWith, [][]byte{k1, k2}; !reflect.DeepEqual(got, want) {
		t.Errorf("a.MockGetValue.CalledWith = %q\n want %q", got, want)
	}
	if got, want := a.MockGetValue.LastCalledWith, k2; bytes.Compare(got, want) != 0 {
		t.Errorf("a.MockGetValue.LastCalledWith = %q want %q", got, want)
	}
}

func TestMockAdapter_DeleteValue(t *testing.T) {
	a := &MockAdapter{}

	k1 := testutil.RandomKey()
	_, err := a.DeleteValue(k1)
	if err != nil {
		t.Fatalf("a.DeleteValue(): err: %s", err)
	}

	v1 := testutil.RandomValue()
	a.MockDeleteValue.Fn = func(key []byte) ([]byte, error) { return v1, nil }
	k2 := testutil.RandomKey()
	v2, err := a.DeleteValue(k2)
	if err != nil {
		t.Fatalf("a.DeleteValue(): err: %s", err)
	}

	if got, want := v2, v1; bytes.Compare(got, want) != 0 {
		t.Errorf("v2 = %s\n want %s", got, want)
	}
	if got, want := a.MockDeleteValue.CalledCount, 2; got != want {
		t.Errorf(`a.MockDeleteValue.CalledCount = %d want %d`, got, want)
	}
	if got, want := a.MockDeleteValue.CalledWith, [][]byte{k1, k2}; !reflect.DeepEqual(got, want) {
		t.Errorf("a.MockDeleteValue.CalledWith = %s\n want %s", got, want)
	}
	if got, want := a.MockDeleteValue.LastCalledWith, k2; bytes.Compare(got, want) != 0 {
		t.Errorf("a.MockDeleteValue.LastCalledWith = %s want %s", got, want)
	}
}

func TestMockAdapter_SaveValue(t *testing.T) {
	a := &MockAdapter{}
	k := testutil.RandomKey()
	v := testutil.RandomValue()

	err := a.SaveValue(k, v)
	if err != nil {
		t.Fatalf("a.SaveValue(): err: %s", err)
	}

	a.MockSaveValue.Fn = func(key, value []byte) error { return nil }
	err = a.SaveValue(k, v)
	if err != nil {
		t.Fatalf("a.SaveValue(): err: %s", err)
	}

	if got, want := a.MockSaveValue.CalledCount, 2; got != want {
		t.Errorf(`a.MockSaveValue.CalledCount = %d want %d`, got, want)
	}
	if got, want := a.MockSaveValue.CalledWith, [][][]byte{[][]byte{k, v}, [][]byte{k, v}}; !reflect.DeepEqual(got, want) {
		t.Errorf("a.MockSaveValue.CalledWith = %s\n want %s", got, want)
	}
	if got, want := a.MockSaveValue.LastCalledWith, [][]byte{k, v}; !reflect.DeepEqual(got, want) {
		t.Errorf("a.MockSaveValue.LastCalledWith = %s\n want %s", got, want)
	}
}

func TestMockAdapter_BatchSaveValue(t *testing.T) {
	a := &MockAdapter{}
	batch, err := a.NewBatch()
	if err != nil {
		t.Fatalf("a.NewBatch(): err: %s", err)
	}
	b := batch.(*MockBatch)

	k := testutil.RandomKey()
	v := testutil.RandomValue()

	err = b.SaveValue(k, v)
	if err != nil {
		t.Fatalf("b.SaveValue(): err: %s", err)
	}

	b.MockSaveValue.Fn = func(key, value []byte) error { return nil }
	err = b.SaveValue(k, v)
	if err != nil {
		t.Fatalf("b.SaveValue(): err: %s", err)
	}

	if got, want := b.MockSaveValue.CalledCount, 2; got != want {
		t.Errorf(`a.MockSaveValue.CalledCount = %d want %d`, got, want)
	}
	if got, want := b.MockSaveValue.CalledWith, [][][]byte{[][]byte{k, v}, [][]byte{k, v}}; !reflect.DeepEqual(got, want) {
		t.Errorf("a.MockSaveValue.CalledWith = %s\n want %s", got, want)
	}
	if got, want := b.MockSaveValue.LastCalledWith, [][]byte{k, v}; !reflect.DeepEqual(got, want) {
		t.Errorf("a.MockSaveValue.LastCalledWith = %s\n want %s", got, want)
	}
}

func TestMockAdapter_BatchDeleteValue(t *testing.T) {
	a := &MockAdapter{}
	batch, err := a.NewBatch()
	if err != nil {
		t.Fatalf("a.NewBatch(): err: %s", err)
	}
	b := batch.(*MockBatch)

	k1 := testutil.RandomKey()
	_, err = b.DeleteValue(k1)
	if err != nil {
		t.Fatalf("a.DeleteValue(): err: %s", err)
	}

	v1 := testutil.RandomValue()
	b.MockDeleteValue.Fn = func(key []byte) ([]byte, error) { return v1, nil }
	k2 := testutil.RandomKey()
	v2, err := b.DeleteValue(k2)
	if err != nil {
		t.Fatalf("a.DeleteValue(): err: %s", err)
	}

	if got, want := v2, v1; bytes.Compare(got, want) != 0 {
		t.Errorf("v2 = %s\n want %s", got, want)
	}
	if got, want := b.MockDeleteValue.CalledCount, 2; got != want {
		t.Errorf(`b.MockDeleteValue.CalledCount = %d want %d`, got, want)
	}
	if got, want := b.MockDeleteValue.CalledWith, [][]byte{k1, k2}; !reflect.DeepEqual(got, want) {
		t.Errorf("b.MockDeleteValue.CalledWith = %s\n want %s", got, want)
	}
	if got, want := b.MockDeleteValue.LastCalledWith, k2; bytes.Compare(got, want) != 0 {
		t.Errorf("b.MockDeleteValue.LastCalledWith = %s want %s", got, want)
	}
}

func TestMockAdapter_BatchSaveSegment(t *testing.T) {
	a := &MockAdapter{}
	batch, err := a.NewBatch()
	if err != nil {
		t.Fatalf("a.NewBatch(): err: %s", err)
	}
	b := batch.(*MockBatch)

	s := cstesting.RandomSegment()

	err = b.SaveSegment(s)
	if err != nil {
		t.Fatalf("a.SaveSegment(): err: %s", err)
	}

	b.MockSaveSegment.Fn = func(s *cs.Segment) error { return nil }
	err = b.SaveSegment(s)
	if err != nil {
		t.Fatalf("a.SaveSegment(): err: %s", err)
	}

	if got, want := b.MockSaveSegment.CalledCount, 2; got != want {
		t.Errorf(`b.MockSaveSegment.CalledCount = %d want %d`, got, want)
	}
	if got, want := b.MockSaveSegment.CalledWith, []*cs.Segment{s, s}; !reflect.DeepEqual(got, want) {
		gotJS, _ := json.MarshalIndent(got, "", "  ")
		wantJS, _ := json.MarshalIndent(want, "", "  ")
		t.Errorf("b.MockSaveSegment.CalledWith = %s\n want %s", gotJS, wantJS)
	}
	if got, want := b.MockSaveSegment.LastCalledWith, s; got != want {
		gotJS, _ := json.MarshalIndent(got, "", "  ")
		wantJS, _ := json.MarshalIndent(want, "", "  ")
		t.Errorf("b.MockSaveSegment.LastCalledWith = %s\n want %s", gotJS, wantJS)
	}
}

func TestMockAdapter_BatchDeleteSegment(t *testing.T) {
	a := &MockAdapter{}
	batch, err := a.NewBatch()
	if err != nil {
		t.Fatalf("a.NewBatch(): err: %s", err)
	}
	b := batch.(*MockBatch)

	linkHash1 := testutil.RandomHash()
	_, err = b.DeleteSegment(linkHash1)
	if err != nil {
		t.Fatalf("a.DeleteSegment(): err: %s", err)
	}

	s1 := cstesting.RandomSegment()
	b.MockDeleteSegment.Fn = func(linkHash *types.Bytes32) (*cs.Segment, error) { return s1, nil }
	linkHash2 := testutil.RandomHash()
	s2, err := b.DeleteSegment(linkHash2)
	if err != nil {
		t.Fatalf("a.DeleteSegment(): err: %s", err)
	}

	if got, want := s2, s1; got != want {
		gotJS, _ := json.MarshalIndent(got, "", "  ")
		wantJS, _ := json.MarshalIndent(want, "", "  ")
		t.Errorf("s2 = %s\n want %s", gotJS, wantJS)
	}
	if got, want := b.MockDeleteSegment.CalledCount, 2; got != want {
		t.Errorf(`b.MockDeleteSegment.CalledCount = %d want %d`, got, want)
	}
	if got, want := b.MockDeleteSegment.CalledWith, []*types.Bytes32{linkHash1, linkHash2}; !reflect.DeepEqual(got, want) {
		t.Errorf("b.MockDeleteSegment.CalledWith = %q\n want %q", got, want)
	}
	if got, want := b.MockDeleteSegment.LastCalledWith, linkHash2; got != want {
		t.Errorf("b.MockDeleteSegment.LastCalledWith = %q want %q", got, want)
	}
}
