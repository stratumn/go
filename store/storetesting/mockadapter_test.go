// Copyright 2016 Stratumn SAS. All rights reserved.
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
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stratumn/go/cs"
	"github.com/stratumn/go/cs/cstesting"
	"github.com/stratumn/go/store"
	"github.com/stratumn/go/testutil"
	"github.com/stratumn/go/types"
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
		t.Errorf("s2 = %s\n want", gotJS, wantJS)
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
		t.Errorf("s2 = %s\n want", gotJS, wantJS)
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
	a.MockFindSegments.Fn = func(*store.Filter) (cs.SegmentSlice, error) { return cs.SegmentSlice{s}, nil }
	f := store.Filter{PrevLinkHash: testutil.RandomHash()}
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
	if got, want := a.MockFindSegments.CalledWith, []*store.Filter{nil, &f}; !reflect.DeepEqual(got, want) {
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

	a.MockGetMapIDs.Fn = func(*store.Pagination) ([]string, error) { return []string{"one", "two"}, nil }
	p := store.Pagination{Offset: 10}
	s, err := a.GetMapIDs(&p)
	if err != nil {
		t.Fatalf("a.GetMapIDs(): err: %s", err)
	}

	if got, want := s, []string{"one", "two"}; !reflect.DeepEqual(got, want) {
		t.Errorf("s1 = %q\n want %q", got, want)
	}
	if got, want := a.MockGetMapIDs.CalledCount, 2; got != want {
		t.Errorf(`a.MockGetMapIDs.CalledCount = %d want %d`, got, want)
	}
	if got, want := a.MockGetMapIDs.CalledWith, []*store.Pagination{nil, &p}; !reflect.DeepEqual(got, want) {
		t.Errorf("a.MockGetMapIDs.CalledWith = %q\n want %q", got, want)
	}
	if got, want := a.MockGetMapIDs.LastCalledWith, &p; got != want {
		t.Errorf("a.MockGetMapIDs.LastCalledWith = %q\n want %q", got, want)
	}
}
