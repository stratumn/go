// Copyright 2017-2018 Stratumn SAS. All rights reserved.
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
	"context"
	"testing"

	"github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-chainscript/chainscripttest"
	"github.com/stratumn/go-core/store"
	"github.com/stratumn/go-core/testutil"
	"github.com/stratumn/go-core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockAdapter_GetInfo(t *testing.T) {
	a := &MockAdapter{}

	_, err := a.GetInfo(context.Background())
	require.NoError(t, err)

	a.MockGetInfo.Fn = func() (interface{}, error) { return map[string]string{"name": "test"}, nil }
	info, err := a.GetInfo(context.Background())
	require.NoError(t, err)

	assert.Equal(t, "test", info.(map[string]string)["name"])
	assert.Equal(t, 2, a.MockGetInfo.CalledCount)
}

func TestMockAdapter_AddStoreEventChan(t *testing.T) {
	a := &MockAdapter{}
	c := make(chan *store.Event)

	a.AddStoreEventChannel(c)

	assert.Equal(t, 1, a.MockAddStoreEventChannel.CalledCount)
	assert.Equal(t, []chan *store.Event{c}, a.MockAddStoreEventChannel.CalledWith)
	assert.Equal(t, c, a.MockAddStoreEventChannel.LastCalledWith)
}

func TestMockAdapter_CreateLink(t *testing.T) {
	a := &MockAdapter{}
	l := chainscripttest.RandomLink(t)

	_, err := a.CreateLink(context.Background(), l)
	require.NoError(t, err)

	a.MockCreateLink.Fn = func(l *chainscript.Link) (chainscript.LinkHash, error) { return nil, nil }
	_, err = a.CreateLink(context.Background(), l)
	require.NoError(t, err)

	assert.Equal(t, 2, a.MockCreateLink.CalledCount)
	assert.Equal(t, []*chainscript.Link{l, l}, a.MockCreateLink.CalledWith)
	assert.Equal(t, l, a.MockCreateLink.LastCalledWith)
}

func TestMockAdapter_GetSegment(t *testing.T) {
	a := &MockAdapter{}

	linkHash1 := chainscripttest.RandomHash()
	_, err := a.GetSegment(context.Background(), linkHash1)
	require.NoError(t, err)

	s1, _ := chainscripttest.RandomLink(t).Segmentify()
	a.MockGetSegment.Fn = func(linkHash chainscript.LinkHash) (*chainscript.Segment, error) { return s1, nil }
	linkHash2 := chainscripttest.RandomHash()
	s2, err := a.GetSegment(context.Background(), linkHash2)
	require.NoError(t, err)

	assert.Equal(t, s1, s2)
	assert.Equal(t, 2, a.MockGetSegment.CalledCount)
	assert.Equal(t, []chainscript.LinkHash{linkHash1, linkHash2}, a.MockGetSegment.CalledWith)
	assert.Equal(t, linkHash2, a.MockGetSegment.LastCalledWith)
}

func TestMockAdapter_FindSegments(t *testing.T) {
	a := &MockAdapter{}

	_, err := a.FindSegments(context.Background(), nil)
	require.NoError(t, err)

	s, _ := chainscripttest.RandomLink(t).Segmentify()
	a.MockFindSegments.Fn = func(*store.SegmentFilter) (*types.PaginatedSegments, error) {
		return &types.PaginatedSegments{
			Segments:   types.SegmentSlice{s},
			TotalCount: 1,
		}, nil
	}

	prevLinkHash := chainscripttest.RandomHash()
	f := store.SegmentFilter{PrevLinkHash: prevLinkHash}
	s1, err := a.FindSegments(context.Background(), &f)
	require.NoError(t, err)

	assert.Equal(t, (&types.PaginatedSegments{Segments: types.SegmentSlice{s}, TotalCount: 1}), s1)
	assert.Equal(t, 2, a.MockFindSegments.CalledCount)
	assert.Equal(t, []*store.SegmentFilter{nil, &f}, a.MockFindSegments.CalledWith)
	assert.Equal(t, &f, a.MockFindSegments.LastCalledWith)
}

func TestMockAdapter_GetMapIDs(t *testing.T) {
	a := &MockAdapter{}

	_, err := a.GetMapIDs(context.Background(), nil)
	require.NoError(t, err)

	a.MockGetMapIDs.Fn = func(*store.MapFilter) ([]string, error) { return []string{"one", "two"}, nil }
	filter := store.MapFilter{
		Pagination: store.Pagination{Offset: 10},
	}
	s, err := a.GetMapIDs(context.Background(), &filter)
	require.NoError(t, err)

	assert.Equal(t, []string{"one", "two"}, s)
	assert.Equal(t, 2, a.MockGetMapIDs.CalledCount)
	assert.Equal(t, []*store.MapFilter{nil, &filter}, a.MockGetMapIDs.CalledWith)
	assert.Equal(t, &filter, a.MockGetMapIDs.LastCalledWith)
}

func TestMockAdapter_GetValue(t *testing.T) {
	a := &MockKeyValueStore{}

	k1 := testutil.RandomKey()
	_, err := a.GetValue(context.Background(), k1)
	require.NoError(t, err)

	v1 := testutil.RandomValue()
	a.MockGetValue.Fn = func(key []byte) ([]byte, error) { return v1, nil }
	k2 := testutil.RandomKey()
	v2, err := a.GetValue(context.Background(), k2)
	require.NoError(t, err)

	assert.Equal(t, v1, v2)
	assert.Equal(t, 2, a.MockGetValue.CalledCount)
	assert.Equal(t, [][]byte{k1, k2}, a.MockGetValue.CalledWith)
	assert.Equal(t, k2, a.MockGetValue.LastCalledWith)
}

func TestMockAdapter_DeleteValue(t *testing.T) {
	a := &MockKeyValueStore{}

	k1 := testutil.RandomKey()
	_, err := a.DeleteValue(context.Background(), k1)
	require.NoError(t, err)

	v1 := testutil.RandomValue()
	a.MockDeleteValue.Fn = func(key []byte) ([]byte, error) { return v1, nil }
	k2 := testutil.RandomKey()
	v2, err := a.DeleteValue(context.Background(), k2)
	require.NoError(t, err)

	assert.Equal(t, v1, v2)
	assert.Equal(t, 2, a.MockDeleteValue.CalledCount)
	assert.Equal(t, [][]byte{k1, k2}, a.MockDeleteValue.CalledWith)
	assert.Equal(t, k2, a.MockDeleteValue.LastCalledWith)
}

func TestMockAdapter_SetValue(t *testing.T) {
	a := &MockKeyValueStore{}
	k := testutil.RandomKey()
	v := testutil.RandomValue()

	err := a.SetValue(context.Background(), k, v)
	require.NoError(t, err)

	a.MockSetValue.Fn = func(key, value []byte) error { return nil }
	err = a.SetValue(context.Background(), k, v)
	require.NoError(t, err)

	assert.Equal(t, 2, a.MockSetValue.CalledCount)
	assert.Equal(t, [][][]byte{{k, v}, {k, v}}, a.MockSetValue.CalledWith)
	assert.Equal(t, [][]byte{k, v}, a.MockSetValue.LastCalledWith)
}
