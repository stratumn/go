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
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/cs/cstesting"
	"github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/testutil"
	"github.com/stratumn/go-indigocore/types"
)

func TestMockAdapter_GetInfo(t *testing.T) {
	a := &MockAdapter{}

	if _, err := a.GetInfo(context.Background()); err != nil {
		t.Fatalf("a.GetInfo(): err: %s", err)
	}

	a.MockGetInfo.Fn = func() (interface{}, error) { return map[string]string{"name": "test"}, nil }
	info, err := a.GetInfo(context.Background())
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

func TestMockAdapter_AddStoreEventChan(t *testing.T) {
	a := &MockAdapter{}
	c := make(chan *store.Event)

	a.AddStoreEventChannel(c)

	if got, want := a.MockAddStoreEventChannel.CalledCount, 1; got != want {
		t.Errorf(`a.MockAddStoreEventChannel.CalledCount = %d want %d`, got, want)
	}
	if got, want := a.MockAddStoreEventChannel.CalledWith, []chan *store.Event{c}; !reflect.DeepEqual(got, want) {
		t.Errorf("a.MockAddStoreEventChannel.CalledWith = %p\n want %p", got, want)
	}
	if got, want := a.MockAddStoreEventChannel.LastCalledWith, c; got != want {
		t.Errorf("a.MockAddStoreEventChannel.LastCalledWith = %p\n want %p", got, want)
	}
}

func TestMockAdapter_CreateLink(t *testing.T) {
	a := &MockAdapter{}
	l := cstesting.RandomLink()

	_, err := a.CreateLink(context.Background(), l)
	if err != nil {
		t.Fatalf("a.CreateLink(): err: %s", err)
	}

	a.MockCreateLink.Fn = func(l *cs.Link) (*types.Bytes32, error) { return nil, nil }
	_, err = a.CreateLink(context.Background(), l)
	if err != nil {
		t.Fatalf("a.CreateLink(): err: %s", err)
	}

	if got, want := a.MockCreateLink.CalledCount, 2; got != want {
		t.Errorf(`a.MockCreateLink.CalledCount = %d want %d`, got, want)
	}
	if got, want := a.MockCreateLink.CalledWith, []*cs.Link{l, l}; !reflect.DeepEqual(got, want) {
		gotJS, _ := json.MarshalIndent(got, "", "  ")
		wantJS, _ := json.MarshalIndent(want, "", "  ")
		t.Errorf("a.MockCreateLink.CalledWith = %s\n want %s", gotJS, wantJS)
	}
	if got, want := a.MockCreateLink.LastCalledWith, l; got != want {
		gotJS, _ := json.MarshalIndent(got, "", "  ")
		wantJS, _ := json.MarshalIndent(want, "", "  ")
		t.Errorf("a.MockCreateLink.LastCalledWith = %s\n want %s", gotJS, wantJS)
	}
}

func TestMockAdapter_GetSegment(t *testing.T) {
	a := &MockAdapter{}

	linkHash1 := testutil.RandomHash()
	_, err := a.GetSegment(context.Background(), linkHash1)
	if err != nil {
		t.Fatalf("a.GetSegment(): err: %s", err)
	}

	s1 := cstesting.RandomSegment()
	a.MockGetSegment.Fn = func(linkHash *types.Bytes32) (*cs.Segment, error) { return s1, nil }
	linkHash2 := testutil.RandomHash()
	s2, err := a.GetSegment(context.Background(), linkHash2)
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

func TestMockAdapter_FindSegments(t *testing.T) {
	a := &MockAdapter{}

	_, err := a.FindSegments(context.Background(), nil)
	if err != nil {
		t.Fatalf("a.FindSegments(): err: %s", err)
	}

	s := cstesting.RandomSegment()
	a.MockFindSegments.Fn = func(*store.SegmentFilter) (cs.SegmentSlice, error) { return cs.SegmentSlice{s}, nil }
	prevLinkHash := testutil.RandomHash().String()
	f := store.SegmentFilter{PrevLinkHash: &prevLinkHash}
	s1, err := a.FindSegments(context.Background(), &f)
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

	_, err := a.GetMapIDs(context.Background(), nil)
	if err != nil {
		t.Fatalf("a.GetMapIDs(): err: %s", err)
	}

	a.MockGetMapIDs.Fn = func(*store.MapFilter) ([]string, error) { return []string{"one", "two"}, nil }
	// FIXME test with process
	filter := store.MapFilter{
		Pagination: store.Pagination{Offset: 10},
	}
	s, err := a.GetMapIDs(context.Background(), &filter)
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
	a := &MockKeyValueStore{}

	k1 := testutil.RandomKey()
	_, err := a.GetValue(context.Background(), k1)
	if err != nil {
		t.Fatalf("a.GetValue(): err: %s", err)
	}

	v1 := testutil.RandomValue()
	a.MockGetValue.Fn = func(key []byte) ([]byte, error) { return v1, nil }
	k2 := testutil.RandomKey()
	v2, err := a.GetValue(context.Background(), k2)
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
	a := &MockKeyValueStore{}

	k1 := testutil.RandomKey()
	_, err := a.DeleteValue(context.Background(), k1)
	if err != nil {
		t.Fatalf("a.DeleteValue(): err: %s", err)
	}

	v1 := testutil.RandomValue()
	a.MockDeleteValue.Fn = func(key []byte) ([]byte, error) { return v1, nil }
	k2 := testutil.RandomKey()
	v2, err := a.DeleteValue(context.Background(), k2)
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

func TestMockAdapter_SetValue(t *testing.T) {
	a := &MockKeyValueStore{}
	k := testutil.RandomKey()
	v := testutil.RandomValue()

	err := a.SetValue(context.Background(), k, v)
	if err != nil {
		t.Fatalf("a.SetValue(): err: %s", err)
	}

	a.MockSetValue.Fn = func(key, value []byte) error { return nil }
	err = a.SetValue(context.Background(), k, v)
	if err != nil {
		t.Fatalf("a.SetValue(): err: %s", err)
	}

	if got, want := a.MockSetValue.CalledCount, 2; got != want {
		t.Errorf(`a.MockSetValue.CalledCount = %d want %d`, got, want)
	}
	if got, want := a.MockSetValue.CalledWith, [][][]byte{{k, v}, {k, v}}; !reflect.DeepEqual(got, want) {
		t.Errorf("a.MockSetValue.CalledWith = %s\n want %s", got, want)
	}
	if got, want := a.MockSetValue.LastCalledWith, [][]byte{k, v}; !reflect.DeepEqual(got, want) {
		t.Errorf("a.MockSetValue.LastCalledWith = %s\n want %s", got, want)
	}
}
