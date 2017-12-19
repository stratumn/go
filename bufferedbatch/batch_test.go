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

package bufferedbatch

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/cs/cstesting"
	"github.com/stratumn/sdk/store"
	"github.com/stratumn/sdk/store/storetesting"
	"github.com/stratumn/sdk/testutil"
	"github.com/stratumn/sdk/types"
)

func TestBatch_CreateLink(t *testing.T) {
	a := &storetesting.MockAdapter{}
	batch := NewBatch(a)

	l := cstesting.RandomLink()

	wantedErr := errors.New("error on MockCreateLink")
	a.MockCreateLink.Fn = func(link *cs.Link) (*types.Bytes32, error) { return nil, wantedErr }

	if _, err := batch.CreateLink(l); err != nil {
		t.Fatalf("batch.CreateLink(): err: %s", err)
	}
	if got, want := a.MockCreateLink.CalledCount, 0; got != want {
		t.Errorf("batch.MockCreateLink.CalledCount = %d want %d", got, want)
	}
	if got, want := len(batch.Links), 1; got != want {
		t.Errorf("len(batch.Links) = %d want %d", got, want)
	}

	l.Meta["mapId"] = ""
	if _, err := batch.CreateLink(l); err == nil {
		t.Fatal("batch.CreateLink() should return an error when mapId is missing")
	}
}

func TestBatch_GetSegment(t *testing.T) {
	a := &storetesting.MockAdapter{}
	batch := NewBatch(a)

	storedLink := cstesting.RandomLink()
	storedLinkHash, _ := storedLink.Hash()
	batchLink1 := cstesting.RandomLink()
	batchLink2 := cstesting.RandomLink()

	batchLinkHash1, _ := batch.CreateLink(batchLink1)
	batchLinkHash2, _ := batch.CreateLink(batchLink2)

	notFoundErr := errors.New("Unit test error")
	a.MockGetSegment.Fn = func(linkHash *types.Bytes32) (*cs.Segment, error) {
		if *storedLinkHash == *linkHash {
			return storedLink.Segmentify(), nil
		}

		return nil, notFoundErr
	}

	var segment *cs.Segment
	var err error

	segment, err = batch.GetSegment(batchLinkHash1)
	if err != nil {
		t.Fatalf("batch.GetSegment(): err: %s", err)
	}
	if got, want := segment.Link, *batchLink1; !reflect.DeepEqual(got, want) {
		t.Errorf("link = %v want %v", got, want)
	}

	segment, err = batch.GetSegment(batchLinkHash2)
	if err != nil {
		t.Fatalf("batch.GetSegment(): err: %s", err)
	}
	if got, want := segment.Link, *batchLink2; !reflect.DeepEqual(got, want) {
		t.Errorf("link = %v want %v", got, want)
	}

	segment, err = batch.GetSegment(storedLinkHash)
	if err != nil {
		t.Fatalf("batch.GetSegment(): err: %s", err)
	}
	if got, want := segment.Link, *storedLink; !reflect.DeepEqual(got, want) {
		t.Errorf("link = %v want %v", got, want)
	}

	segment, err = batch.GetSegment(testutil.RandomHash())
	if got, want := err, notFoundErr; got != want {
		t.Errorf("GetSegment should return an error: %s want %s", got, want)
	}
}

func TestBatch_FindSegments(t *testing.T) {
	a := &storetesting.MockAdapter{}
	batch := NewBatch(a)

	storedLink := cstesting.RandomLink()
	storedLink.Meta["process"] = "Foo"
	l1 := cstesting.RandomLink()
	l1.Meta["process"] = "Foo"
	l2 := cstesting.RandomLink()
	l2.Meta["process"] = "Bar"

	batch.CreateLink(l1)
	batch.CreateLink(l2)

	notFoundErr := errors.New("Unit test error")
	a.MockFindSegments.Fn = func(filter *store.SegmentFilter) (cs.SegmentSlice, error) {
		if filter.Process == "Foo" {
			return cs.SegmentSlice{storedLink.Segmentify()}, nil
		}
		if filter.Process == "Bar" {
			return cs.SegmentSlice{}, nil
		}

		return nil, notFoundErr
	}

	var segments cs.SegmentSlice
	var err error

	segments, err = batch.FindSegments(&store.SegmentFilter{Pagination: store.Pagination{Limit: store.DefaultLimit}, Process: "Foo"})
	if err != nil {
		t.Fatalf("batch.FindSegments(): err: %s", err)
	}
	if got, want := len(segments), 2; got != want {
		t.Errorf("segment slice length = %d want %d", got, want)
	}

	segments, err = batch.FindSegments(&store.SegmentFilter{Pagination: store.Pagination{Limit: store.DefaultLimit}, Process: "Bar"})
	if err != nil {
		t.Fatalf("batch.FindSegments(): err: %s", err)
	}
	if got, want := len(segments), 1; got != want {
		t.Errorf("segment slice length = %d want %d", got, want)
	}

	_, err = batch.FindSegments(&store.SegmentFilter{Pagination: store.Pagination{Limit: store.DefaultLimit}, Process: "NotFound"})
	if got, want := err, notFoundErr; got != want {
		t.Errorf("FindSegments should return an error: %s want %s", got, want)
	}
}

func TestBatch_GetMapIDs(t *testing.T) {
	a := &storetesting.MockAdapter{}
	batch := NewBatch(a)

	storedLink1 := cstesting.RandomLink()
	storedLink1.Meta["mapId"] = "Foo1"
	storedLink1.Meta["process"] = "FooProcess"
	storedLink2 := cstesting.RandomLink()
	storedLink2.Meta["mapId"] = "Bar"
	storedLink2.Meta["process"] = "BarProcess"

	batchLink1 := cstesting.RandomLink()
	batchLink1.Meta["mapId"] = "Foo2"
	batchLink1.Meta["process"] = "FooProcess"
	batchLink2 := cstesting.RandomLink()
	batchLink2.Meta["mapId"] = "Yin"
	batchLink2.Meta["process"] = "YinProcess"

	batch.CreateLink(batchLink1)
	batch.CreateLink(batchLink2)

	a.MockGetMapIDs.Fn = func(filter *store.MapFilter) ([]string, error) {
		if filter.Process == storedLink1.Meta["process"] {
			return []string{storedLink1.Meta["mapId"].(string)}, nil
		}
		if filter.Process == storedLink2.Meta["process"] {
			return []string{storedLink2.Meta["mapId"].(string)}, nil
		}

		return []string{
			storedLink1.Meta["mapId"].(string),
			storedLink2.Meta["mapId"].(string),
		}, nil
	}

	var mapIDs []string
	var err error

	mapIDs, err = batch.GetMapIDs(&store.MapFilter{Pagination: store.Pagination{Limit: store.DefaultLimit}})
	if err != nil {
		t.Fatalf("batch.GetMapIDs(): err: %s", err)
	}
	if got, want := len(mapIDs), 4; got != want {
		t.Errorf("mapIds length = %d want %d / values = %v", got, want, mapIDs)
	}

	processFilter := &store.MapFilter{
		Process:    "FooProcess",
		Pagination: store.Pagination{Limit: store.DefaultLimit},
	}
	mapIDs, err = batch.GetMapIDs(processFilter)
	if err != nil {
		t.Fatalf("batch.GetMapIDs(): err: %s", err)
	}
	if got, want := len(mapIDs), 2; got != want {
		t.Errorf("mapIds length = %d want %d / values = %v", got, want, mapIDs)
	}
	for _, mapID := range []string{
		storedLink1.GetMapID(),
		batchLink1.GetMapID(),
	} {
		if mapIDs[0] != mapID && mapIDs[1] != mapID {
			t.Errorf("Invalid mapId returned: %v", mapID)
		}
	}
}

func TestBatch_GetMapIDsWithStoreReturningAnErrorOnGetMapIDs(t *testing.T) {
	a := &storetesting.MockAdapter{}
	batch := NewBatch(a)

	wantedMapIds := []string{"Foo", "Bar"}
	notFoundErr := errors.New("Unit test error")
	a.MockGetMapIDs.Fn = func(filter *store.MapFilter) ([]string, error) {
		return wantedMapIds, notFoundErr
	}

	if mapIDs, err := batch.GetMapIDs(&store.MapFilter{}); err == nil {
		t.Fatal("batch.GetMapIDs() should return an error")
	} else if got, want := len(mapIDs), len(wantedMapIds); got != want {
		t.Fatalf("mapIds length = %d want %d", got, want)
	} else if got, want := mapIDs, wantedMapIds; !reflect.DeepEqual(got, want) {
		t.Fatalf("mapIds = %v want %v", got, want)
	}
}

func TestBatch_WriteLink(t *testing.T) {
	a := &storetesting.MockAdapter{}
	l := cstesting.RandomLink()

	batch := NewBatch(a)

	_, err := batch.CreateLink(l)
	if err != nil {
		t.Fatalf("batch.CreateLink(): err: %s", err)
	}

	err = batch.Write()
	if err != nil {
		t.Fatalf("batch.Write(): err: %s", err)
	}

	if got, want := a.MockCreateLink.CalledCount, 1; got != want {
		t.Errorf("batch.Write(): expected to have called CreateLink %d time, got %d", want, got)
	}

	if got, want := a.MockCreateLink.LastCalledWith, l; got != want {
		t.Errorf("batch.Write(): expected to have called CreateLink with %v, got %v", want, got)
	}
}

func TestBatch_WriteLinkWithFailure(t *testing.T) {
	a := &storetesting.MockAdapter{}
	mockError := errors.New("Error")

	la := cstesting.RandomLink()
	lb := cstesting.RandomLink()

	a.MockCreateLink.Fn = func(l *cs.Link) (*types.Bytes32, error) {
		if l == la {
			return nil, mockError
		}
		return l.Hash()
	}

	batch := NewBatch(a)

	_, err := batch.CreateLink(la)
	if err != nil {
		t.Fatalf("batch.CreateLink(): err: %s", err)
	}

	_, err = batch.CreateLink(lb)
	if err != nil {
		t.Fatalf("batch.CreateLink(): err: %s", err)
	}

	if got, want := batch.Write(), mockError; got != want {
		t.Errorf("batch.Write returned %v want %v", got, want)
	}
}
