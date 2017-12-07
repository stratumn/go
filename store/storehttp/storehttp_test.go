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

package storehttp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/cs/cstesting"
	"github.com/stratumn/sdk/jsonhttp"
	"github.com/stratumn/sdk/jsonws"
	"github.com/stratumn/sdk/jsonws/jsonwstesting"
	"github.com/stratumn/sdk/store"
	"github.com/stratumn/sdk/store/storetesting"
	"github.com/stratumn/sdk/testutil"
	"github.com/stratumn/sdk/types"
)

const zeros = "0000000000000000000000000000000000000000000000000000000000000000"

func TestRoot(t *testing.T) {
	s, a := createServer()
	a.MockGetInfo.Fn = func() (interface{}, error) { return "test", nil }

	var body map[string]interface{}
	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/", nil, &body)
	if err != nil {
		t.Fatalf("testutil.RequestJSON(): err: %s", err)
	}

	if got, want := w.Code, http.StatusOK; got != want {
		t.Errorf("w.StatusCode = %d want %d", got, want)
	}
	if got, want := body["adapter"].(string), "test"; got != want {
		t.Errorf(`body["adapter"] = %q want %q`, got, want)
	}
	if got, want := a.MockGetInfo.CalledCount, 1; got != want {
		t.Errorf("a.MockGetInfo.CalledCount = %d want %d", got, want)
	}
}

func TestRoot_err(t *testing.T) {
	s, a := createServer()
	a.MockGetInfo.Fn = func() (interface{}, error) { return "test", errors.New("error") }

	var body map[string]interface{}
	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/", nil, &body)
	if err != nil {
		t.Fatalf("testutil.RequestJSON(): err: %s", err)
	}

	if got, want := w.Code, jsonhttp.NewErrInternalServer("").Status(); got != want {
		t.Errorf("w.Code = %d want %d", got, want)
	}
	if got, want := body["error"].(string), jsonhttp.NewErrInternalServer("").Error(); got != want {
		t.Errorf(`body["error"] = %q want %q`, got, want)
	}
	if got, want := a.MockGetInfo.CalledCount, 1; got != want {
		t.Errorf("a.MockGetInfo.CalledCount = %d want %d", got, want)
	}
}

func TestCreateLink(t *testing.T) {
	s, a := createServer()
	a.MockCreateLink.Fn = func(l *cs.Link) (*types.Bytes32, error) { return l.Hash() }

	l1 := cstesting.RandomLink()
	var s1 cs.Segment
	w, err := testutil.RequestJSON(s.ServeHTTP, "POST", "/links", l1, &s1)
	if err != nil {
		t.Fatalf("testutil.RequestJSON(): err: %s", err)
	}

	if !reflect.DeepEqual(a.MockCreateLink.LastCalledWith, l1) {
		got, _ := json.MarshalIndent(a.MockCreateLink.LastCalledWith, "", "  ")
		want, _ := json.MarshalIndent(l1, "", "  ")
		t.Errorf("a.MockCreateLink.LastCalledWith = %s\nwant %s", got, want)
	}
	if got, want := w.Code, http.StatusOK; got != want {
		t.Errorf("w.Code = %d want %d", got, want)
	}
	if !reflect.DeepEqual(&s1.Link, l1) {
		got, _ := json.MarshalIndent(&s1.Link, "", "  ")
		want, _ := json.MarshalIndent(l1, "", "  ")
		t.Errorf("s2 = %s\nwant %s", got, want)
	}
	if got, want := a.MockCreateLink.CalledCount, 1; got != want {
		t.Errorf("a.MockCreateLink.CalledCount = %d want %d", got, want)
	}
}

func TestCreateLink_err(t *testing.T) {
	s, a := createServer()
	a.MockCreateLink.Fn = func(l *cs.Link) (*types.Bytes32, error) { return nil, errors.New("test") }

	var body map[string]interface{}
	w, err := testutil.RequestJSON(s.ServeHTTP, "POST", "/links", cstesting.RandomLink(), &body)
	if err != nil {
		t.Fatalf("testutil.RequestJSON(): err: %s", err)
	}

	if got, want := w.Code, jsonhttp.NewErrInternalServer("").Status(); got != want {
		t.Errorf("w.Code = %d want %d", got, want)
	}
	if got, want := body["error"].(string), jsonhttp.NewErrInternalServer("").Error(); got != want {
		t.Errorf(`body["error"] = %q want %q`, got, want)
	}
	if got, want := a.MockCreateLink.CalledCount, 1; got != want {
		t.Errorf("a.MockCreateLink.CalledCount = %d want %d", got, want)
	}
}

func TestCreateLink_invalidLink(t *testing.T) {
	s, a := createServer()

	l1 := cstesting.RandomLink()
	l1.Meta["process"] = ""
	var body map[string]interface{}
	w, err := testutil.RequestJSON(s.ServeHTTP, "POST", "/links", l1, &body)
	if err != nil {
		t.Fatalf("testutil.RequestJSON(): err: %s", err)
	}

	if got, want := w.Code, jsonhttp.NewErrBadRequest("").Status(); got != want {
		t.Errorf("w.Code = %d want %d", got, want)
	}
	if got, want := body["error"].(string), "link.meta.process should be a non empty string"; got != want {
		t.Errorf(`body["error"] = %q want %q`, got, want)
	}
	if got, want := a.MockCreateLink.CalledCount, 0; got != want {
		t.Errorf("a.MockCreateLink.CalledCount = %d want %d", got, want)
	}
}

func TestCreateLink_invalidJSON(t *testing.T) {
	s, a := createServer()

	var body map[string]interface{}
	w, err := testutil.RequestJSON(s.ServeHTTP, "POST", "/links", "azertyuio", &body)
	if err != nil {
		t.Fatalf("testutil.RequestJSON(): err: %s", err)
	}

	if got, want := w.Code, jsonhttp.NewErrBadRequest("").Status(); got != want {
		t.Errorf("w.Code = %d want %d", got, want)
	}
	if got, want := body["error"].(string), "json: cannot unmarshal string into Go value of type cs.Link"; got != want {
		t.Errorf(`body["error"] = %q want %q`, got, want)
	}
	if got, want := a.MockCreateLink.CalledCount, 0; got != want {
		t.Errorf("a.MockCreateLink.CalledCount = %d want %d", got, want)
	}
}

func TestAddEvidence(t *testing.T) {
	s, a := createServer()
	a.MockAddEvidence.Fn = func(*types.Bytes32, *cs.Evidence) error { return nil }

	link := cstesting.RandomLink()
	linkHash, _ := link.HashString()
	e := cstesting.RandomEvidence()
	var body map[string]interface{}
	w, err := testutil.RequestJSON(s.ServeHTTP, "POST", "/evidences/"+linkHash, e, &body)
	if err != nil {
		t.Fatalf("testutil.RequestJSON(): err: %s", err)
	}

	if got, want := a.MockAddEvidence.LastCalledWith.Provider, e.Provider; got != want {
		t.Errorf("a.MockAddEvidence.LastCalledWith = %s\nwant %s", got, want)
	}
	if got, want := w.Code, http.StatusOK; got != want {
		t.Errorf("w.Code = %d want %d", got, want)
	}

	if got, want := a.MockAddEvidence.CalledCount, 1; got != want {
		t.Errorf("a.MockAddEvidence.CalledCount = %d want %d", got, want)
	}
}

func TestAddEvidence_err(t *testing.T) {
	s, a := createServer()
	a.MockAddEvidence.Fn = func(*types.Bytes32, *cs.Evidence) error { return errors.New("test") }

	link := cstesting.RandomLink()
	linkHash, _ := link.HashString()
	e := cstesting.RandomEvidence()
	var body map[string]interface{}
	w, err := testutil.RequestJSON(s.ServeHTTP, "POST", "/evidences/"+linkHash, e, &body)
	if err != nil {
		t.Fatalf("testutil.RequestJSON(): err: %s", err)
	}
	if got, want := w.Code, jsonhttp.NewErrBadRequest("").Status(); got != want {
		t.Errorf("w.Code = %d want %d", got, want)
	}
	if got, want := body["error"].(string), jsonhttp.NewErrBadRequest("test").Error(); got != want {
		t.Errorf(`body["error"] = %q want %q`, got, want)
	}
	if got, want := a.MockAddEvidence.CalledCount, 1; got != want {
		t.Errorf("a.MockAddEvidence.CalledCount = %d want %d", got, want)
	}
}

func TestGetSegment(t *testing.T) {
	s, a := createServer()
	s1 := cstesting.RandomSegment()
	a.MockGetSegment.Fn = func(*types.Bytes32) (*cs.Segment, error) { return s1, nil }

	var s2 cs.Segment
	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/segments/"+zeros, nil, &s2)
	if err != nil {
		t.Fatalf("testutil.RequestJSON(): err: %s", err)
	}

	if got, want := a.MockGetSegment.LastCalledWith.String(), zeros; got != want {
		t.Errorf("a.MockGetSegment.LastCalledWith = %q\nwant %q", got, want)
	}
	if got, want := w.Code, http.StatusOK; got != want {
		t.Errorf("w.Code = %d want %d", got, want)
	}
	if !reflect.DeepEqual(&s2, s1) {
		got, _ := json.MarshalIndent(s2, "", "  ")
		want, _ := json.MarshalIndent(s1, "", "  ")
		t.Errorf("s2 = %s\nwant %s", got, want)
	}
	if got, want := a.MockGetSegment.CalledCount, 1; got != want {
		t.Errorf("a.MockGetSegment.CalledCount = %d want %d", got, want)
	}
}

func TestGetSegment_notFound(t *testing.T) {
	s, a := createServer()

	var body map[string]interface{}
	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/segments/"+zeros, nil, &body)
	if err != nil {
		t.Fatalf("testutil.RequestJSON(): err: %s", err)
	}

	if got, want := a.MockGetSegment.LastCalledWith.String(), zeros; got != want {
		t.Errorf("a.MockGetSegment.LastCalledWith = %q\nwant %q", got, want)
	}
	if got, want := w.Code, jsonhttp.NewErrNotFound("").Status(); got != want {
		t.Errorf("w.Code = %d want %d", got, want)
	}
	if got, want := body["error"].(string), jsonhttp.NewErrNotFound("").Error(); got != want {
		t.Errorf(`body["error"] = %q want %q`, got, want)
	}
	if got, want := a.MockGetSegment.CalledCount, 1; got != want {
		t.Errorf("a.MockGetSegment.CalledCount = %d want %d", got, want)
	}
}

func TestGetSegment_err(t *testing.T) {
	s, a := createServer()
	a.MockGetSegment.Fn = func(*types.Bytes32) (*cs.Segment, error) { return nil, errors.New("error") }

	var body map[string]interface{}
	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/segments/"+zeros, nil, &body)
	if err != nil {
		t.Fatalf("testutil.RequestJSON(): err: %s", err)
	}

	if got, want := a.MockGetSegment.LastCalledWith.String(), zeros; got != want {
		t.Errorf("a.MockGetSegment.LastCalledWith = %q\nwant %q", got, want)
	}
	if got, want := w.Code, jsonhttp.NewErrInternalServer("").Status(); got != want {
		t.Errorf("w.Code = %d want %d", got, want)
	}
	if got, want := body["error"].(string), jsonhttp.NewErrInternalServer("").Error(); got != want {
		t.Errorf(`body["error"] = %q want %q`, got, want)
	}
	if got, want := a.MockGetSegment.CalledCount, 1; got != want {
		t.Errorf("a.MockGetSegment.CalledCount = %d want %d", got, want)
	}
}

func TestFindSegments(t *testing.T) {
	s, a := createServer()
	var s1 cs.SegmentSlice
	for i := 0; i < 10; i++ {
		s1 = append(s1, cstesting.RandomSegment())
	}
	a.MockFindSegments.Fn = func(*store.SegmentFilter) (cs.SegmentSlice, error) { return s1, nil }

	var s2 cs.SegmentSlice
	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/segments?offset=1&limit=2&mapIds%5B%5D=123&prevLinkHash="+zeros+"&tags%5B%5D=one&tags%5B%5D=two", nil, &s2)
	if err != nil {
		t.Fatalf("testutil.RequestJSON(): err: %s", err)
	}

	if got, want := w.Code, http.StatusOK; got != want {
		t.Errorf("w.Code = %d want %d", got, want)
	}
	if !reflect.DeepEqual(s2, s1) {
		got, _ := json.MarshalIndent(s2, "", "  ")
		want, _ := json.MarshalIndent(s1, "", "  ")
		t.Errorf("s2 = %s\nwant %s", got, want)
	}
	if got, want := a.MockFindSegments.CalledCount, 1; got != want {
		t.Errorf("a.MockFindSegments.CalledCount = %d want %d", got, want)
	}

	f := a.MockFindSegments.LastCalledWith
	if got, want := f.Offset, 1; got != want {
		t.Errorf("a.MockFindSegments.LastCalledWith.Offset = %d want %d", got, want)
	}
	if got, want := f.Limit, 2; got != want {
		t.Errorf("a.MockFindSegments.LastCalledWith.Limit = %d want %d", got, want)
	}
	if got, want := len(f.MapIDs), 1; got != want {
		t.Errorf("a.MockFindSegments.LastCalledWith.MapIDs = %q want %q", got, want)
	} else if got, want := f.MapIDs[0], "123"; got != want {
		t.Errorf("a.MockFindSegments.LastCalledWith.MapIDs = %q want %q", got, want)
	}
	if got := f.PrevLinkHash; got == nil {
		t.Errorf("a.MockFindSegments.LastCalledWith.PrevLinkHash = %q is nil", got)
	} else if got, want := *f.PrevLinkHash, zeros; got != want {
		t.Errorf("a.MockFindSegments.LastCalledWith.PrevLinkHash = %q want %q", got, want)
	}
	if got, want := f.Tags, []string{"one", "two"}; !reflect.DeepEqual(got, want) {
		t.Errorf("a.MockFindSegments.LastCalledWith.Tags = %v want %v", got, want)
	}
}

func TestFindSegments_multipleMapIDs(t *testing.T) {
	s, a := createServer()
	var s1 cs.SegmentSlice
	for i := 0; i < 10; i++ {
		s1 = append(s1, cstesting.RandomSegment())
	}
	a.MockFindSegments.Fn = func(*store.SegmentFilter) (cs.SegmentSlice, error) { return s1, nil }

	var s2 cs.SegmentSlice
	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/segments?offset=1&limit=2&mapIds[]=123&mapIds[]=456&prevLinkHash="+zeros+"&tags[]=one&tags%5B%5D=two", nil, &s2)
	if err != nil {
		t.Fatalf("testutil.RequestJSON(): err: %s", err)
	}

	if got, want := w.Code, http.StatusOK; got != want {
		t.Errorf("w.Code = %d want %d", got, want)
	}
	if !reflect.DeepEqual(s2, s1) {
		got, _ := json.MarshalIndent(s2, "", "  ")
		want, _ := json.MarshalIndent(s1, "", "  ")
		t.Errorf("s2 = %s\nwant %s", got, want)
	}
	if got, want := a.MockFindSegments.CalledCount, 1; got != want {
		t.Errorf("a.MockFindSegments.CalledCount = %d want %d", got, want)
	}

	f := a.MockFindSegments.LastCalledWith
	if got, want := f.Offset, 1; got != want {
		t.Errorf("a.MockFindSegments.LastCalledWith.Offset = %d want %d", got, want)
	}
	if got, want := f.Limit, 2; got != want {
		t.Errorf("a.MockFindSegments.LastCalledWith.Limit = %d want %d", got, want)
	}
	if got, want := len(f.MapIDs), 2; got != want {
		t.Errorf("a.MockFindSegments.LastCalledWith.MapIDs = %q want %q", got, want)
	} else if got, want := f.MapIDs[0], "123"; got != want {
		t.Errorf("a.MockFindSegments.LastCalledWith.MapIDs = %q want %q", got, want)
	} else if got, want := f.MapIDs[1], "456"; got != want {
		t.Errorf("a.MockFindSegments.LastCalledWith.MapIDs = %q want %q", got, want)
	}
	if got := f.PrevLinkHash; got == nil {
		t.Errorf("a.MockFindSegments.LastCalledWith.PrevLinkHash = %q is nil", got)
	} else if got, want := *f.PrevLinkHash, zeros; got != want {
		t.Errorf("a.MockFindSegments.LastCalledWith.PrevLinkHash = %q want %q", got, want)
	}
	if got, want := f.Tags, []string{"one", "two"}; !reflect.DeepEqual(got, want) {
		t.Errorf("a.MockFindSegments.LastCalledWith.Tags = %v want %v", got, want)
	}
}

func TestFindSegments_defaultLimit(t *testing.T) {
	s, a := createServer()
	var s1 cs.SegmentSlice
	for i := 0; i < 2; i++ {
		s1 = append(s1, cstesting.RandomSegment())
	}
	a.MockFindSegments.Fn = func(*store.SegmentFilter) (cs.SegmentSlice, error) { return s1, nil }

	var s2 cs.SegmentSlice
	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/segments?offset=1&&mapIds%5B%5D=123&prevLinkHash="+zeros+"&tags[]=one&tags[]=two", nil, &s2)
	if err != nil {
		t.Fatalf("testutil.RequestJSON(): err: %s", err)
	}

	if got, want := w.Code, http.StatusOK; got != want {
		t.Errorf("w.Code = %d want %d", got, want)
	}
	if !reflect.DeepEqual(s2, s1) {
		got, _ := json.MarshalIndent(s2, "", "  ")
		want, _ := json.MarshalIndent(s1, "", "  ")
		t.Errorf("s2 = %s\nwant %s", got, want)
	}
	if got, want := a.MockFindSegments.CalledCount, 1; got != want {
		t.Errorf("a.MockFindSegments.CalledCount = %d want %d", got, want)
	}

	f := a.MockFindSegments.LastCalledWith
	if got, want := f.Offset, 1; got != want {
		t.Errorf("a.MockFindSegments.LastCalledWith.Offset = %d want %d", got, want)
	}
	if got, want := f.Limit, store.DefaultLimit; got != want {
		t.Errorf("a.MockFindSegments.LastCalledWith.Limit = %d want %d", got, want)
	}
	if got, want := len(f.MapIDs), 1; got != want {
		t.Errorf("a.MockFindSegments.LastCalledWith.MapIDs = %q want %q", got, want)
	} else if got, want := f.MapIDs[0], "123"; got != want {
		t.Errorf("a.MockFindSegments.LastCalledWith.MapIDs = %q want %q", got, want)
	}
	if got := f.PrevLinkHash; got == nil {
		t.Errorf("a.MockFindSegments.LastCalledWith.PrevLinkHash = %q is nil", got)
	} else if got, want := *f.PrevLinkHash, zeros; got != want {
		t.Errorf("a.MockFindSegments.LastCalledWith.PrevLinkHash = %q want %q", got, want)
	}
	if got, want := f.Tags, []string{"one", "two"}; !reflect.DeepEqual(got, want) {
		t.Errorf("a.MockFindSegments.LastCalledWith.Tags = %v want %v", got, want)
	}
}

func TestFindSegments_err(t *testing.T) {
	s, a := createServer()
	a.MockFindSegments.Fn = func(*store.SegmentFilter) (cs.SegmentSlice, error) { return nil, errors.New("test") }

	var body map[string]interface{}
	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/segments", nil, &body)
	if err != nil {
		t.Fatalf("testutil.RequestJSON(): err: %s", err)
	}

	if got, want := w.Code, jsonhttp.NewErrInternalServer("").Status(); got != want {
		t.Errorf("w.Code = %d want %d", got, want)
	}
	if got, want := body["error"].(string), jsonhttp.NewErrInternalServer("").Error(); got != want {
		t.Errorf(`body["error"] = %q want %q`, got, want)
	}
	if got, want := a.MockFindSegments.CalledCount, 1; got != want {
		t.Errorf("a.MockFindSegments.CalledCount = %d want %d", got, want)
	}
}

func TestFindSegments_invalidOffset(t *testing.T) {
	s, a := createServer()

	var body map[string]interface{}
	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/segments?offset=a", nil, &body)
	if err != nil {
		t.Fatalf("testutil.RequestJSON(): err: %s", err)
	}

	if got, want := w.Code, newErrOffset("").Status(); got != want {
		t.Errorf("w.Code = %d want %d", got, want)
	}
	if got, want := body["error"].(string), newErrOffset("").Error(); got != want {
		t.Errorf(`body["error"] = %q want %q`, got, want)
	}
	if got, want := a.MockFindSegments.CalledCount, 0; got != want {
		t.Errorf("a.MockFindSegments.CalledCount = %d want %d", got, want)
	}
}

func TestFindSegments_invalidPrevLinkHash(t *testing.T) {
	s, a := createServer()

	var body map[string]interface{}
	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/segments?prevLinkHash=3", nil, &body)
	if err != nil {
		t.Fatalf("testutil.RequestJSON(): err: %s", err)
	}

	if got, want := w.Code, newErrPrevLinkHash("").Status(); got != want {
		t.Errorf("w.Code = %d want %d", got, want)
	}
	if got, want := body["error"].(string), newErrPrevLinkHash("").Error(); got != want {
		t.Errorf(`body["error"] = %q want %q`, got, want)
	}
	if got, want := a.MockFindSegments.CalledCount, 0; got != want {
		t.Errorf("a.MockFindSegments.CalledCount = %d want %d", got, want)
	}
}

func TestGetMapIDs(t *testing.T) {
	s, a := createServer()
	s1 := []string{"one", "two", "three"}
	a.MockGetMapIDs.Fn = func(*store.MapFilter) ([]string, error) { return s1, nil }

	var s2 []string
	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/maps?offset=20&limit=10", nil, &s2)
	if err != nil {
		t.Fatalf("testutil.RequestJSON(): err: %s", err)
	}

	if got, want := w.Code, http.StatusOK; got != want {
		t.Errorf("w.Code = %d want %d", got, want)
	}
	if got, want := s2, s1; !reflect.DeepEqual(got, want) {
		t.Errorf("s2 = %v want %v", got, want)
	}
	if got, want := a.MockGetMapIDs.CalledCount, 1; got != want {
		t.Errorf("a.MockGetMapIDs(pagination).CalledCount = %d want %d", got, want)
	}

	p := a.MockGetMapIDs.LastCalledWith
	if got, want := p.Offset, 20; got != want {
		t.Errorf("a.MockGetMapIDs.LastCalledWith.Offset = %d want %d", got, want)
	}
	if got, want := p.Limit, 10; got != want {
		t.Errorf("a.MockGetMapIDs.LastCalledWith.Limit = %d want %d", got, want)
	}
}

func TestGetMapIDs_err(t *testing.T) {
	s, a := createServer()
	a.MockGetMapIDs.Fn = func(*store.MapFilter) ([]string, error) { return nil, errors.New("test") }

	var body map[string]interface{}
	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/maps", nil, &body)
	if err != nil {
		t.Fatalf("testutil.RequestJSON(): err: %s", err)
	}

	if got, want := w.Code, jsonhttp.NewErrInternalServer("").Status(); got != want {
		t.Errorf("w.Code = %d want %d", got, want)
	}
	if got, want := body["error"].(string), jsonhttp.NewErrInternalServer("").Error(); got != want {
		t.Errorf(`body["error"] = %q want %q`, got, want)
	}
	if got, want := a.MockGetMapIDs.CalledCount, 1; got != want {
		t.Errorf("a.MockGetMapIDs.CalledCount = %d want %d", got, want)
	}
}

func TestGetMapIDs_invalidLimit(t *testing.T) {
	s, a := createServer()

	var body map[string]interface{}
	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/maps?limit=-1", nil, &body)
	if err != nil {
		t.Fatalf("testutil.RequestJSON(): err: %s", err)
	}

	if got, want := w.Code, newErrOffset("").Status(); got != want {
		t.Errorf("w.Code = %d want %d", got, want)
	}
	if got, want := body["error"].(string), newErrLimit("").Error(); got != want {
		t.Errorf(`body["error"] = %q want %q`, got, want)
	}
	if got, want := a.MockGetMapIDs.CalledCount, 0; got != want {
		t.Errorf("a.MockGetMapIDs.CalledCount = %d want %d", got, want)
	}
}

func TestGetMapIDs_limitTooLarge(t *testing.T) {
	s, a := createServer()

	var body map[string]interface{}
	limit := store.MaxLimit + 1
	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", fmt.Sprintf("/maps?limit=%d", limit), nil, &body)
	if err != nil {
		t.Fatalf("testutil.RequestJSON(): err: %s", err)
	}

	if got, want := w.Code, newErrOffset("").Status(); got != want {
		t.Errorf("w.Code = %d want %d", got, want)
	}
	if got, want := body["error"].(string), newErrLimit("").Error(); got != want {
		t.Errorf(`body["error"] = %q want %q`, got, want)
	}
	if got, want := a.MockGetMapIDs.CalledCount, 0; got != want {
		t.Errorf("a.MockGetMapIDs.CalledCount = %d want %d", got, want)
	}
}

func TestNotFound(t *testing.T) {
	s, _ := createServer()

	var body map[string]interface{}
	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/azerty", nil, &body)
	if err != nil {
		t.Fatalf("testutil.RequestJSON(): err: %s", err)
	}

	if got, want := w.Code, jsonhttp.NewErrNotFound("").Status(); got != want {
		t.Errorf("w.Code = %d want %d", got, want)
	}
	if got, want := body["error"].(string), jsonhttp.NewErrNotFound("").Error(); got != want {
		t.Errorf(`body["error"] = %q want %q`, got, want)
	}
}

func TestGetSocket(t *testing.T) {
	link := cstesting.RandomLink()
	event := store.NewSavedLinks()
	event.AddSavedLink(link)

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
			t.Fatalf("connection ready timeout")
		}
		c <- event
	case <-time.After(time.Second):
		t.Fatalf("save channel not added")
	}

	// Wait for message to be broadcasted.
	select {
	case <-doneChan:
		got := conn.MockWriteJSON.LastCalledWith.(*jsonws.Message).Data.([]*cs.Link)
		if len(got) != 1 {
			t.Fatalf("Invalid number of links in json data")
		}
		if !reflect.DeepEqual(got[0], link) {
			gotjs, _ := json.MarshalIndent(got, "", "  ")
			wantjs, _ := json.MarshalIndent(link, "", "  ")
			t.Errorf("conn.MockWriteJSON.LastCalledWith = %s\nwant %s", gotjs, wantjs)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("saved segment not broadcasted")
	}
}
