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

package storehttp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-chainscript/chainscripttest"
	"github.com/stratumn/go-core/jsonhttp"
	"github.com/stratumn/go-core/jsonws"
	"github.com/stratumn/go-core/jsonws/jsonwstesting"
	"github.com/stratumn/go-core/store"
	"github.com/stratumn/go-core/store/storetesting"
	"github.com/stratumn/go-core/testutil"
	"github.com/stratumn/go-core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoot(t *testing.T) {
	s, a := createServer()
	a.MockGetInfo.Fn = func() (interface{}, error) { return "test", nil }

	var body map[string]interface{}
	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/", nil, &body)
	require.NoError(t, err, "testutil.RequestJSON()")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "test", body["adapter"].(string))
	assert.Equal(t, 1, a.MockGetInfo.CalledCount)
}

func TestRoot_err(t *testing.T) {
	s, a := createServer()
	a.MockGetInfo.Fn = func() (interface{}, error) { return "test", errors.New("error") }

	var body map[string]interface{}
	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/", nil, &body)
	require.NoError(t, err, "testutil.RequestJSON()")

	assert.Equal(t, jsonhttp.NewErrInternalServer("").Status(), w.Code)
	assert.Equal(t, jsonhttp.NewErrInternalServer("").Error(), body["error"].(string))
	assert.Equal(t, 1, a.MockGetInfo.CalledCount)
}

func TestCreateLink(t *testing.T) {
	s, a := createServer()
	a.MockCreateLink.Fn = func(l *chainscript.Link) (chainscript.LinkHash, error) { return l.Hash() }

	l1 := chainscripttest.RandomLink(t)
	var s1 chainscript.Segment
	w, err := testutil.RequestJSON(s.ServeHTTP, "POST", "/links", l1, &s1)
	require.NoError(t, err, "testutil.RequestJSON()")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 1, a.MockCreateLink.CalledCount)
	chainscripttest.LinksEqual(t, l1, a.MockCreateLink.LastCalledWith)
	chainscripttest.LinksEqual(t, l1, s1.Link)
}

func TestCreateLink_err(t *testing.T) {
	s, a := createServer()
	a.MockCreateLink.Fn = func(l *chainscript.Link) (chainscript.LinkHash, error) { return nil, errors.New("test") }

	var body map[string]interface{}
	w, err := testutil.RequestJSON(s.ServeHTTP, "POST", "/links", chainscripttest.RandomLink(t), &body)
	require.NoError(t, err, "testutil.RequestJSON()")

	assert.Equal(t, jsonhttp.NewErrInternalServer("").Status(), w.Code)
	assert.Equal(t, jsonhttp.NewErrInternalServer("").Error(), body["error"].(string))
	assert.Equal(t, 1, a.MockCreateLink.CalledCount)
}

func TestCreateLink_invalidJSON(t *testing.T) {
	s, a := createServer()

	var body map[string]interface{}
	w, err := testutil.RequestJSON(s.ServeHTTP, "POST", "/links", "azertyuio", &body)
	require.NoError(t, err, "testutil.RequestJSON()")

	assert.Equal(t, jsonhttp.NewErrBadRequest("").Status(), w.Code)
	assert.Equal(t, "json: cannot unmarshal string into Go value of type chainscript.Link", body["error"].(string))
	assert.Zero(t, a.MockCreateLink.CalledCount)
}

func TestAddEvidence(t *testing.T) {
	s, a := createServer()
	a.MockAddEvidence.Fn = func(chainscript.LinkHash, *chainscript.Evidence) error { return nil }

	link := chainscripttest.RandomLink(t)
	linkHash, _ := link.Hash()
	e := chainscripttest.RandomEvidence(t)
	var body map[string]interface{}
	w, err := testutil.RequestJSON(s.ServeHTTP, "POST", "/evidences/"+linkHash.String(), e, &body)
	require.NoError(t, err, "testutil.RequestJSON()")

	assert.Equal(t, e.Provider, a.MockAddEvidence.LastCalledWith.Provider)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 1, a.MockAddEvidence.CalledCount)
}

func TestAddEvidence_err(t *testing.T) {
	s, a := createServer()
	a.MockAddEvidence.Fn = func(chainscript.LinkHash, *chainscript.Evidence) error { return errors.New("test") }

	link := chainscripttest.RandomLink(t)
	linkHash, _ := link.Hash()
	e := chainscripttest.RandomEvidence(t)
	var body map[string]interface{}
	w, err := testutil.RequestJSON(s.ServeHTTP, "POST", "/evidences/"+linkHash.String(), e, &body)
	require.NoError(t, err, "testutil.RequestJSON()")

	assert.Equal(t, jsonhttp.NewErrInternalServer("").Status(), w.Code)
	assert.Equal(t, jsonhttp.NewErrInternalServer("").Error(), body["error"].(string))
	assert.Equal(t, 1, a.MockAddEvidence.CalledCount)
}

func TestGetSegment(t *testing.T) {
	s, a := createServer()
	s1 := chainscripttest.RandomSegment(t)
	a.MockGetSegment.Fn = func(chainscript.LinkHash) (*chainscript.Segment, error) { return s1, nil }

	var s2 chainscript.Segment
	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/segments/"+s1.LinkHash().String(), nil, &s2)
	require.NoError(t, err, "testutil.RequestJSON()")

	assert.Equal(t, s1.LinkHash(), a.MockGetSegment.LastCalledWith)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 1, a.MockGetSegment.CalledCount)
	chainscripttest.SegmentsEqual(t, s1, &s2)
}

func TestGetSegment_notFound(t *testing.T) {
	s, a := createServer()
	unknownHash := chainscripttest.RandomHash()

	var body map[string]interface{}
	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/segments/"+unknownHash.String(), nil, &body)
	require.NoError(t, err, "testutil.RequestJSON()")

	assert.Equal(t, unknownHash, a.MockGetSegment.LastCalledWith)
	assert.Equal(t, jsonhttp.NewErrNotFound("").Status(), w.Code)
	assert.Equal(t, jsonhttp.NewErrNotFound("").Error(), body["error"].(string))
	assert.Equal(t, 1, a.MockGetSegment.CalledCount)
}

func TestGetSegment_err(t *testing.T) {
	s, a := createServer()
	lh := chainscripttest.RandomHash()
	a.MockGetSegment.Fn = func(chainscript.LinkHash) (*chainscript.Segment, error) { return nil, errors.New("error") }

	var body map[string]interface{}
	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/segments/"+lh.String(), nil, &body)
	require.NoError(t, err, "testutil.RequestJSON()")

	assert.Equal(t, lh, a.MockGetSegment.LastCalledWith)
	assert.Equal(t, jsonhttp.NewErrInternalServer("").Status(), w.Code)
	assert.Equal(t, jsonhttp.NewErrInternalServer("").Error(), body["error"].(string))
	assert.Equal(t, 1, a.MockGetSegment.CalledCount)
}

func TestFindSegments(t *testing.T) {
	s, a := createServer()
	s1 := &types.PaginatedSegments{}
	for i := 0; i < 10; i++ {
		s1.Segments = append(s1.Segments, chainscripttest.RandomSegment(t))
	}
	a.MockFindSegments.Fn = func(*store.SegmentFilter) (*types.PaginatedSegments, error) { return s1, nil }

	s2 := &types.PaginatedSegments{}
	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/segments?offset=1&limit=2&mapIds%5B%5D=123&tags%5B%5D=one&tags%5B%5D=two", nil, &s2)
	require.NoError(t, err, "testutil.RequestJSON()")

	assert.Equal(t, http.StatusOK, w.Code)
	testutil.PaginatedSegmentsEqual(t, s1, s2)
	assert.Equal(t, 1, a.MockFindSegments.CalledCount)

	f := a.MockFindSegments.LastCalledWith
	assert.Equal(t, 1, f.Offset)
	assert.Equal(t, 2, f.Limit)
	assert.Equal(t, []string{"123"}, f.MapIDs)
	assert.Equal(t, []string{"one", "two"}, f.Tags)
	assert.False(t, f.WithoutParent)
	assert.Nil(t, f.PrevLinkHash)
	assert.Nil(t, f.LinkHashes)
}

func TestFindSegments_multipleMapIDs(t *testing.T) {
	s, a := createServer()
	s1 := &types.PaginatedSegments{}
	for i := 0; i < 10; i++ {
		s1.Segments = append(s1.Segments, chainscripttest.RandomSegment(t))
	}
	a.MockFindSegments.Fn = func(*store.SegmentFilter) (*types.PaginatedSegments, error) { return s1, nil }

	s2 := &types.PaginatedSegments{}
	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/segments?offset=1&limit=2&mapIds[]=123&mapIds[]=456&tags[]=one&tags%5B%5D=two", nil, &s2)
	require.NoError(t, err, "testutil.RequestJSON()")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 1, a.MockFindSegments.CalledCount)
	testutil.PaginatedSegmentsEqual(t, s1, s2)

	f := a.MockFindSegments.LastCalledWith
	assert.Equal(t, 1, f.Offset)
	assert.Equal(t, 2, f.Limit)
	assert.Equal(t, []string{"123", "456"}, f.MapIDs)
	assert.Equal(t, []string{"one", "two"}, f.Tags)
	assert.False(t, f.WithoutParent)
	assert.Nil(t, f.PrevLinkHash)
	assert.Nil(t, f.LinkHashes)
}

func TestFindSegments_multipleLinkHashes(t *testing.T) {
	s, a := createServer()

	s1 := &types.PaginatedSegments{}
	for i := 0; i < 10; i++ {
		s1.Segments = append(s1.Segments, chainscripttest.RandomSegment(t))
	}

	linkHash1 := s1.Segments[0].LinkHash()
	linkHash2 := s1.Segments[1].LinkHash()

	a.MockFindSegments.Fn = func(*store.SegmentFilter) (*types.PaginatedSegments, error) { return s1, nil }

	s2 := &types.PaginatedSegments{}
	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/segments?offset=1&limit=2&linkHashes[]="+linkHash1.String()+"&linkHashes%5B%5D="+linkHash2.String(), nil, &s2)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, w.Code)
	testutil.PaginatedSegmentsEqual(t, s1, s2)
	assert.Equal(t, 1, a.MockFindSegments.CalledCount)

	f := a.MockFindSegments.LastCalledWith
	assert.Equal(t, 1, f.Offset)
	assert.Equal(t, 2, f.Limit)
	assert.Equal(t, 2, len(f.LinkHashes))
	assert.Equal(t, f.LinkHashes[0], linkHash1)
	assert.Equal(t, f.LinkHashes[1], linkHash2)
}

func TestFindSegments_prevLinkHash(t *testing.T) {
	s, a := createServer()
	s1 := &types.PaginatedSegments{
		Segments: []*chainscript.Segment{chainscripttest.RandomSegment(t)},
	}
	a.MockFindSegments.Fn = func(*store.SegmentFilter) (*types.PaginatedSegments, error) { return s1, nil }

	s2 := &types.PaginatedSegments{}
	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/segments?prevLinkHash="+s1.Segments[0].LinkHash().String(), nil, &s2)
	require.NoError(t, err, "testutil.RequestJSON()")

	assert.Equal(t, http.StatusOK, w.Code)
	testutil.PaginatedSegmentsEqual(t, s1, s2)
	assert.Equal(t, 1, a.MockFindSegments.CalledCount)

	f := a.MockFindSegments.LastCalledWith
	assert.False(t, f.WithoutParent)
	assert.Equal(t, s1.Segments[0].LinkHash(), f.PrevLinkHash)
}

func TestFindSegments_withoutParent(t *testing.T) {
	s, a := createServer()
	s1 := &types.PaginatedSegments{
		Segments: []*chainscript.Segment{chainscripttest.RandomSegment(t)},
	}
	a.MockFindSegments.Fn = func(*store.SegmentFilter) (*types.PaginatedSegments, error) { return s1, nil }

	s2 := &types.PaginatedSegments{}
	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/segments?withoutParent=true", nil, &s2)
	require.NoError(t, err, "testutil.RequestJSON()")

	assert.Equal(t, http.StatusOK, w.Code)
	testutil.PaginatedSegmentsEqual(t, s1, s2)
	assert.Equal(t, 1, a.MockFindSegments.CalledCount)

	f := a.MockFindSegments.LastCalledWith
	assert.True(t, f.WithoutParent)
	assert.Nil(t, f.PrevLinkHash)
}

func TestFindSegments_defaultLimit(t *testing.T) {
	s, a := createServer()
	s1 := &types.PaginatedSegments{}
	for i := 0; i < 10; i++ {
		s1.Segments = append(s1.Segments, chainscripttest.RandomSegment(t))
	}
	a.MockFindSegments.Fn = func(*store.SegmentFilter) (*types.PaginatedSegments, error) { return s1, nil }

	s2 := &types.PaginatedSegments{}
	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/segments?offset=1&&mapIds%5B%5D=123&tags[]=one&tags[]=two", nil, &s2)
	require.NoError(t, err, "testutil.RequestJSON()")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 1, a.MockFindSegments.CalledCount)
	testutil.PaginatedSegmentsEqual(t, s1, s2)

	f := a.MockFindSegments.LastCalledWith
	assert.Equal(t, 1, f.Offset)
	assert.Equal(t, store.DefaultLimit, f.Limit)
	assert.Equal(t, []string{"123"}, f.MapIDs)
	assert.Equal(t, []string{"one", "two"}, f.Tags)
}

func TestFindSegments_err(t *testing.T) {
	s, a := createServer()
	a.MockFindSegments.Fn = func(*store.SegmentFilter) (*types.PaginatedSegments, error) {
		return nil, errors.New("test")
	}

	var body map[string]interface{}
	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/segments", nil, &body)
	require.NoError(t, err, "testutil.RequestJSON()")

	assert.Equal(t, jsonhttp.NewErrInternalServer("").Status(), w.Code)
	assert.Equal(t, jsonhttp.NewErrInternalServer("").Error(), body["error"].(string))
	assert.Equal(t, 1, a.MockFindSegments.CalledCount)
}

func TestFindSegments_invalidOffset(t *testing.T) {
	s, a := createServer()

	var body map[string]interface{}
	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/segments?offset=a", nil, &body)
	require.NoError(t, err, "testutil.RequestJSON()")

	assert.Equal(t, newErrOffset("").Status(), w.Code)
	assert.Equal(t, newErrOffset("").Error(), body["error"].(string))
	assert.Zero(t, a.MockFindSegments.CalledCount)
}

func TestFindSegments_invalidPrevLinkHash(t *testing.T) {
	s, a := createServer()

	var body map[string]interface{}
	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/segments?prevLinkHash=3", nil, &body)
	require.NoError(t, err, "testutil.RequestJSON()")

	assert.Equal(t, newErrPrevLinkHash("").Status(), w.Code)
	assert.Equal(t, newErrPrevLinkHash("").Error(), body["error"].(string))
	assert.Zero(t, a.MockFindSegments.CalledCount)
}

func TestFindSegments_invalidLinkHashes(t *testing.T) {
	s, a := createServer()

	var body map[string]interface{}
	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/segments?linkHashes[]=3", nil, &body)
	require.NoError(t, err, "testutil.RequestJSON()")

	assert.Equal(t, newErrLinkHashes("").Status(), w.Code)
	assert.Equal(t, newErrLinkHashes("").Error(), body["error"].(string))
	assert.Equal(t, 0, a.MockFindSegments.CalledCount)
}

func TestGetMapIDs(t *testing.T) {
	s, a := createServer()
	s1 := []string{"one", "two", "three"}
	a.MockGetMapIDs.Fn = func(*store.MapFilter) ([]string, error) { return s1, nil }

	var s2 []string
	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/maps?offset=20&limit=10", nil, &s2)
	require.NoError(t, err, "testutil.RequestJSON()")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 1, a.MockGetMapIDs.CalledCount)
	assert.ElementsMatch(t, s1, s2)

	p := a.MockGetMapIDs.LastCalledWith
	assert.Equal(t, 20, p.Offset)
	assert.Equal(t, 10, p.Limit)
}

func TestGetMapIDs_err(t *testing.T) {
	s, a := createServer()
	a.MockGetMapIDs.Fn = func(*store.MapFilter) ([]string, error) { return nil, errors.New("test") }

	var body map[string]interface{}
	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/maps", nil, &body)
	require.NoError(t, err, "testutil.RequestJSON()")

	assert.Equal(t, jsonhttp.NewErrInternalServer("").Status(), w.Code)
	assert.Equal(t, jsonhttp.NewErrInternalServer("").Error(), body["error"].(string))
	assert.Equal(t, 1, a.MockGetMapIDs.CalledCount)
}

func TestGetMapIDs_invalidLimit(t *testing.T) {
	s, a := createServer()

	var body map[string]interface{}
	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/maps?limit=-1", nil, &body)
	require.NoError(t, err, "testutil.RequestJSON()")

	assert.Equal(t, newErrOffset("").Status(), w.Code)
	assert.Equal(t, newErrLimit("").Error(), body["error"].(string))
	assert.Zero(t, a.MockGetMapIDs.CalledCount)
}

func TestGetMapIDs_limitTooLarge(t *testing.T) {
	s, a := createServer()

	var body map[string]interface{}
	limit := store.MaxLimit + 1
	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", fmt.Sprintf("/maps?limit=%d", limit), nil, &body)
	require.NoError(t, err, "testutil.RequestJSON()")

	assert.Equal(t, newErrOffset("").Status(), w.Code)
	assert.Equal(t, newErrLimit("").Error(), body["error"].(string))
	assert.Zero(t, a.MockGetMapIDs.CalledCount)
}

func TestNotFound(t *testing.T) {
	s, _ := createServer()

	var body map[string]interface{}
	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/azerty", nil, &body)
	require.NoError(t, err, "testutil.RequestJSON()")

	assert.Equal(t, jsonhttp.NewErrNotFound("").Status(), w.Code)
	assert.Equal(t, jsonhttp.NewErrNotFound("").Error(), body["error"].(string))
}

func TestGetSocket(t *testing.T) {
	link := chainscripttest.RandomLink(t)
	event := store.NewSavedLinks(link)

	// Chan that will receive the store event channel.
	sendChan := make(chan chan *store.Event)

	// Chan used to wait for the connection to be ready.
	readyChan := make(chan struct{})

	// Chan used to wait for web socket message.
	doneChan := make(chan struct{})

	conn := jsonwstesting.MockConn{}
	conn.MockReadJSON.Fn = func(interface{}) error {
		readyChan <- struct{}{}
		return nil
	}
	conn.MockWriteJSON.Fn = func(interface{}) error {
		doneChan <- struct{}{}
		return nil
	}

	upgradeHandle := func(w http.ResponseWriter, r *http.Request, h http.Header) (jsonws.PingableConn, error) {
		return &conn, nil
	}

	// Mock adapter to send the store event channel when added.
	a := &storetesting.MockAdapter{}
	a.MockAddStoreEventChannel.Fn = func(c chan *store.Event) {
		sendChan <- c
	}

	s := New(a, &Config{}, &jsonhttp.Config{}, &jsonws.BasicConfig{
		UpgradeHandle: upgradeHandle,
	}, &jsonws.BufferedConnConfig{
		Size:         256,
		WriteTimeout: 10 * time.Second,
		PongTimeout:  70 * time.Second,
		PingInterval: time.Minute,
		MaxMsgSize:   1024,
	})

	go s.Start()
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer s.Shutdown(ctx)
	defer cancel()

	// Register web socket connection.
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/websocket", nil)
	go s.getWebSocket(w, r, nil)

	// Wait for channel to be added.
	select {
	case c := <-sendChan:
		// Wait for connection to be ready.
		select {
		case <-readyChan:
		case <-time.After(time.Second):
			require.Fail(t, "connection ready timeout")
		}
		c <- event
	case <-time.After(time.Second):
		require.Fail(t, "save channel not added")
	}

	// Wait for message to be broadcasted.
	select {
	case <-doneChan:
		got := conn.MockWriteJSON.LastCalledWith.(*jsonws.Message).Data.([]*chainscript.Link)
		require.Len(t, got, 1)
		chainscripttest.LinksEqual(t, link, got[0])
	case <-time.After(2 * time.Second):
		require.Fail(t, "saved segment not broadcasted")
	}
}
