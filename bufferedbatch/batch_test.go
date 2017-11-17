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
	"bytes"
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

func TestBatch_SaveValue(t *testing.T) {
	a := &storetesting.MockAdapter{}
	batch := NewBatch(a)

	k := testutil.RandomKey()
	v := testutil.RandomValue()

	wantedErr := errors.New("error on MockSaveValue")
	a.MockSaveValue.Fn = func(k, v []byte) error { return wantedErr }

	if err := batch.SaveValue(k, v); err != nil {
		t.Fatalf("a.SaveValue(): err: %s", err)
	}
	if got, want := a.MockSaveValue.CalledCount, 0; got != want {
		t.Errorf("a.MockSaveValue.CalledCount = %d want %d", got, want)
	}
	if got, want := len(batch.ValueOps), 1; got != want {
		t.Errorf("len(batch.ValueOps) = %d want %d", got, want)
	}
	if got, want := len(batch.SegmentOps), 0; got != want {
		t.Errorf("len(batch.SegmentOps) = %d want %d", got, want)
	}
}

func TestBatch_DeleteValue(t *testing.T) {
	a := &storetesting.MockAdapter{}
	batch := NewBatch(a)

	k1, v1 := testutil.RandomKey(), testutil.RandomValue()
	k2, v2 := testutil.RandomKey(), testutil.RandomValue()

	if err := batch.SaveValue(k1, v1); err != nil {
		t.Fatalf("batch.SaveValue(): err: %s", err)
	}
	if err := batch.SaveValue(k2, v2); err != nil {
		t.Fatalf("batch.SaveValue(): err: %s", err)
	}
	if err := batch.SaveValue(k1, v1); err != nil {
		t.Fatalf("batch.SaveValue(): err: %s", err)
	}

	value, err := batch.DeleteValue(k1)
	if err != nil {
		t.Fatalf("batch.DeleteValue(): err: %s", err)
	}
	if got, want := string(value), string(v1); !reflect.DeepEqual(got, want) {
		t.Errorf("value = %v want %v", got, want)
	}
	if got, want := len(batch.ValueOps), 2; got != want {
		t.Errorf("len(batch.ValueOps) = %d want %d", got, want)
	}
	if got, want := len(batch.SegmentOps), 0; got != want {
		t.Errorf("len(batch.SegmentOps) = %d want %d", got, want)
	}

	v3 := testutil.RandomValue()
	a.MockGetValue.Fn = func([]byte) ([]byte, error) { return v3, errors.New("Unit test error") }
	value, err = batch.DeleteValue(k1)
	if err == nil {
		t.Fatalf("batch.DeleteValue() should return an error")
	}
	if got, want := string(value), string(v3); !reflect.DeepEqual(got, want) {
		t.Errorf("value = %v want %v", got, want)
	}
	if got, want := len(batch.ValueOps), 2; got != want {
		t.Errorf("len(batch.ValueOps) = %d want %d", got, want)
	}
	if got, want := len(batch.SegmentOps), 0; got != want {
		t.Errorf("len(batch.SegmentOps) = %d want %d", got, want)
	}
}

func TestBatch_SaveSegment(t *testing.T) {
	a := &storetesting.MockAdapter{}
	batch := NewBatch(a)

	s := cstesting.RandomSegment()

	wantedErr := errors.New("error on MockSaveValue")
	a.MockSaveValue.Fn = func(k, v []byte) error { return wantedErr }

	if err := batch.SaveSegment(s); err != nil {
		t.Fatalf("batch.SaveSegment(): err: %s", err)
	}
	if got, want := a.MockSaveSegment.CalledCount, 0; got != want {
		t.Errorf("batch.MockSaveValue.CalledCount = %d want %d", got, want)
	}
	if got, want := len(batch.ValueOps), 0; got != want {
		t.Errorf("len(batch.ValueOps) = %d want %d", got, want)
	}
	if got, want := len(batch.SegmentOps), 1; got != want {
		t.Errorf("len(batch.SegmentOps) = %d want %d", got, want)
	}

	s.Link.Meta["mapId"] = ""
	if err := batch.SaveSegment(s); err == nil {
		t.Fatal("batch.SaveSegment() should return an error")
	}
}

func TestBatch_DeleteSegment(t *testing.T) {
	a := &storetesting.MockAdapter{}
	batch := NewBatch(a)

	s1 := cstesting.RandomSegment()
	s2 := cstesting.RandomSegment()

	if err := batch.SaveSegment(s1); err != nil {
		t.Fatalf("batch.SaveSegment(): err: %s", err)
	}
	if err := batch.SaveSegment(s2); err != nil {
		t.Fatalf("batch.SaveSegment(): err: %s", err)
	}
	if err := batch.SaveSegment(s1); err != nil {
		t.Fatalf("batch.SaveSegment(): err: %s", err)
	}

	segment, err := batch.DeleteSegment(s1.GetLinkHash())
	if err != nil {
		t.Fatalf("batch.DeleteSegment(): err: %s", err)
	}
	if got, want := segment, s1; !reflect.DeepEqual(got, want) {
		t.Errorf("value = %v want %v", got, want)
	}
	if got, want := len(batch.ValueOps), 0; got != want {
		t.Errorf("len(batch.ValueOps) = %d want %d", got, want)
	}
	if got, want := len(batch.SegmentOps), 2; got != want {
		t.Errorf("len(batch.SegmentOps) = %d want %d", got, want)
	}

	s3 := cstesting.RandomSegment()
	a.MockGetSegment.Fn = func(linkHash *types.Bytes32) (segment *cs.Segment, err error) {
		return s3, errors.New("Unit test error")
	}
	segment, err = batch.DeleteSegment(s1.GetLinkHash())
	if err == nil {
		t.Fatal("batch.DeleteValue() should return an error")
	}
	if got, want := segment, s3; !reflect.DeepEqual(got, want) {
		t.Errorf("value = %v want %v", got, want)
	}
	if got, want := len(batch.ValueOps), 0; got != want {
		t.Errorf("len(batch.ValueOps) = %d want %d", got, want)
	}
	if got, want := len(batch.SegmentOps), 2; got != want {
		t.Errorf("len(batch.SegmentOps) = %d want %d", got, want)
	}

}

func TestBatch_GetSegment(t *testing.T) {
	a := &storetesting.MockAdapter{}
	batch := NewBatch(a)

	sa := cstesting.RandomSegment()
	sb := cstesting.RandomSegment()
	s1 := cstesting.RandomSegment()
	s2 := cstesting.RandomSegment()

	batch.SaveSegment(s1)
	batch.SaveSegment(s2)
	batch.DeleteSegment(s2.GetLinkHash())
	batch.DeleteSegment(sb.GetLinkHash())

	notFoundErr := errors.New("Unit test error")
	a.MockGetSegment.Fn = func(linkHash *types.Bytes32) (*cs.Segment, error) {
		if sa.GetLinkHashString() == linkHash.String() {
			return sa, nil
		}
		if sb.GetLinkHashString() == linkHash.String() {
			return sb, nil
		}
		return nil, notFoundErr
	}

	var segment *cs.Segment
	var err error

	segment, err = batch.GetSegment(s1.GetLinkHash())
	if err != nil {
		t.Fatalf("batch.GetSegment(): err: %s", err)
	}
	if got, want := segment, s1; !reflect.DeepEqual(got, want) {
		t.Errorf("segment = %v want %v", got, want)
	}

	segment, err = batch.GetSegment(s2.GetLinkHash())
	if err != nil {
		t.Fatalf("batch.GetSegment(): err: %s", err)
	}
	if segment != nil {
		t.Errorf("segment = %v want nil", segment)
	}

	segment, err = batch.GetSegment(sa.GetLinkHash())
	if err != nil {
		t.Fatalf("batch.GetSegment(): err: %s", err)
	}
	if got, want := segment, sa; !reflect.DeepEqual(got, want) {
		t.Errorf("segment = %v want %v", got, want)
	}

	segment, err = batch.GetSegment(sb.GetLinkHash())
	if err != nil {
		t.Fatalf("batch.GetSegment(): err: %s", err)
	}
	if segment != nil {
		t.Errorf("segment = %v want nil", segment)
	}

	segment, err = batch.GetSegment(cstesting.RandomSegment().GetLinkHash())
	if got, want := err, notFoundErr; got != want {
		t.Errorf("GetSegment should return an error: %s want %s", got, want)
	}
}

func TestBatch_FindSegments(t *testing.T) {
	a := &storetesting.MockAdapter{}
	batch := NewBatch(a)

	sa := cstesting.RandomSegment()
	sa.Link.Meta["process"] = "Foo"
	sa.SetLinkHash()
	sb := cstesting.RandomSegment()
	sb.Link.Meta["process"] = "Bar"
	sb.SetLinkHash()
	s1 := cstesting.RandomSegment()
	s1.Link.Meta["process"] = "Foo"
	s1.SetLinkHash()
	s2 := cstesting.RandomSegment()
	s2.Link.Meta["process"] = "Bar"
	s2.SetLinkHash()

	batch.SaveSegment(s1)
	batch.SaveSegment(s2)
	batch.DeleteSegment(s2.GetLinkHash())
	batch.DeleteSegment(sb.GetLinkHash())

	notFoundErr := errors.New("Unit test error")
	a.MockFindSegments.Fn = func(filter *store.SegmentFilter) (cs.SegmentSlice, error) {
		if filter.Process == "Foo" {
			return cs.SegmentSlice{sa}, nil
		}
		if filter.Process == "Bar" {
			return cs.SegmentSlice{sb}, nil
		}
		return nil, notFoundErr
	}
	a.MockGetSegment.Fn = func(linkHash *types.Bytes32) (*cs.Segment, error) {
		if sa.GetLinkHashString() == linkHash.String() {
			return sa, nil
		}
		if sb.GetLinkHashString() == linkHash.String() {
			return sb, nil
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
	if got, want := len(segments), 0; got != want {
		t.Errorf("segment slice length = %d want %d", got, want)
	}

	segments, err = batch.FindSegments(&store.SegmentFilter{Pagination: store.Pagination{Limit: store.DefaultLimit}, Process: "NotFound"})
	if got, want := err, notFoundErr; got != want {
		t.Errorf("FindSegments should return an error: %s want %s", got, want)
	}
}

func TestBatch_GetMapIDs(t *testing.T) {
	a := &storetesting.MockAdapter{}
	batch := NewBatch(a)

	sa := cstesting.RandomSegment()
	sa.Link.Meta["mapId"] = "Foo"
	sa.Link.Meta["process"] = "FooProcess"
	sa.SetLinkHash()
	sb := cstesting.RandomSegment()
	sb.Link.Meta["mapId"] = "Bar"
	sb.Link.Meta["process"] = "BarProcess"
	sb.SetLinkHash()
	sc := cstesting.RandomSegment()
	sc.Link.Meta["mapId"] = "Yin"
	sc.Link.Meta["process"] = "YinProcess"
	sc.SetLinkHash()
	s1 := cstesting.RandomSegment()
	s1.Link.Meta["mapId"] = "Foo"
	s1.Link.Meta["process"] = "FooProcess"
	s1.SetLinkHash()
	s2 := cstesting.RandomSegment()
	s2.Link.Meta["mapId"] = "Bar"
	s2.Link.Meta["process"] = "BarProcess"
	s2.SetLinkHash()
	s3 := cstesting.RandomSegment()
	s3.Link.Meta["mapId"] = "Yin"
	s3.Link.Meta["process"] = "YinProcess"
	s3.SetLinkHash()
	s4 := cstesting.RandomSegment()
	s4.Link.Meta["mapId"] = "Yang"
	s4.Link.Meta["process"] = "YangProcess"
	s4.SetLinkHash()

	batch.SaveSegment(s1)
	batch.SaveSegment(s2)
	batch.SaveSegment(s3)
	batch.SaveSegment(s4)

	notFoundErr := errors.New("Unit test error")
	a.MockGetMapIDs.Fn = func(filter *store.MapFilter) ([]string, error) {
		if filter.Process == sa.Link.Meta["process"] {
			return []string{sa.Link.Meta["mapId"].(string)}, nil
		}
		if filter.Process == sb.Link.Meta["process"] {
			return []string{sb.Link.Meta["mapId"].(string)}, nil
		}
		if filter.Process == sc.Link.Meta["process"] {
			return []string{sc.Link.Meta["mapId"].(string)}, nil
		}
		return []string{
			sa.Link.Meta["mapId"].(string),
			sb.Link.Meta["mapId"].(string),
			sc.Link.Meta["mapId"].(string),
		}, nil
	}
	a.MockGetSegment.Fn = func(linkHash *types.Bytes32) (*cs.Segment, error) {
		switch *linkHash {
		case *sa.GetLinkHash():
			return sa, nil
		case *sb.GetLinkHash():
			return sb, nil
		case *sc.GetLinkHash():
			return sc, nil
		}
		return nil, notFoundErr
	}
	a.MockFindSegments.Fn = func(filter *store.SegmentFilter) (cs.SegmentSlice, error) {
		ret := make(cs.SegmentSlice, 0, len(filter.MapIDs))
		for _, mapID := range filter.MapIDs {
			switch mapID {
			case sa.Link.Meta["mapId"].(string):
				ret = append(ret, sa)
			case sb.Link.Meta["mapId"].(string):
				ret = append(ret, sb)
			case sc.Link.Meta["mapId"].(string):
				ret = append(ret, sc)
			}
		}
		if len(ret) == 0 {
			return nil, notFoundErr
		}
		return ret, nil
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

	batch.DeleteSegment(s2.GetLinkHash())
	batch.DeleteSegment(s3.GetLinkHash())
	batch.DeleteSegment(sc.GetLinkHash())

	mapIDs, err = batch.GetMapIDs(&store.MapFilter{Pagination: store.Pagination{Limit: store.DefaultLimit}})
	if err != nil {
		t.Fatalf("batch.GetMapIDs(): err: %s", err)
	}
	if got, want := len(mapIDs), 3; got != want {
		t.Errorf("mapIds length = %d want %d / values = %v", got, want, mapIDs)
	}
	mapIDDict := make(map[string]bool, len(mapIDs))
	for _, m := range mapIDs {
		mapIDDict[m] = true
	}

	for _, m := range []string{
		sa.Link.Meta["mapId"].(string),
		s2.Link.Meta["mapId"].(string),
		s4.Link.Meta["mapId"].(string),
	} {
		if _, exist := mapIDDict[m]; exist == false {
			t.Errorf("mapId missing %s", m)
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

func TestBatch_GetMapIDsWithStoreReturningAnErrorOnFindSegments(t *testing.T) {
	a := &storetesting.MockAdapter{}
	batch := NewBatch(a)

	sa := cstesting.RandomSegment()

	batch.DeleteSegment(sa.GetLinkHash())

	notFoundErr := errors.New("Unit test error")
	a.MockGetMapIDs.Fn = func(filter *store.MapFilter) ([]string, error) {
		return []string{sa.Link.Meta["mapId"].(string)}, nil
	}
	a.MockGetSegment.Fn = func(linkHash *types.Bytes32) (*cs.Segment, error) {
		return sa, nil
	}
	a.MockFindSegments.Fn = func(filter *store.SegmentFilter) (cs.SegmentSlice, error) {
		return nil, notFoundErr
	}

	if _, err := batch.GetMapIDs(&store.MapFilter{Pagination: store.Pagination{Limit: store.DefaultLimit}}); err == nil {
		t.Fatalf("batch.GetMapIDs() should return an error")
	}
}

func TestBatch_GetValue(t *testing.T) {
	a := &storetesting.MockAdapter{}
	batch := NewBatch(a)

	k1, v1 := testutil.RandomKey(), testutil.RandomValue()
	k2, v2 := testutil.RandomKey(), testutil.RandomValue()
	k3, v3 := testutil.RandomKey(), testutil.RandomValue()
	v4 := testutil.RandomValue()

	batch.SaveValue(k1, v1)
	batch.SaveValue(k2, v2)
	batch.SaveValue(k3, v3)
	batch.DeleteValue(k3)
	batch.SaveValue(k2, v4)

	if got, err := batch.GetValue(k1); err != nil {
		t.Errorf("batch.GetValue(): err: %s", err)
	} else if got, want := got, v1; string(got) != string(want) {
		t.Errorf("value = %v want %v", got, want)
	}

	if got, err := batch.GetValue(k2); err != nil {
		t.Errorf("batch.GetValue(): err: %s", err)
	} else if got, want := got, v4; string(got) != string(want) {
		t.Errorf("value = %v want %v", got, want)
	}

	if got, want := a.MockGetValue.CalledCount, 0; got != want {
		t.Errorf("a.MockGetValue.CalledCount = %v want %v", got, want)
	}
	if got, err := batch.GetValue(k3); err != nil {
		t.Errorf("batch.GetValue(): err: %s", err)
	} else if got != nil {
		t.Errorf("value should be nil %v", got)
	}
	if got, want := a.MockGetValue.CalledCount, 1; got != want {
		t.Errorf("a.MockGetValue.CalledCount = %v want %v", got, want)
	}
}

// TestBatch_WriteSegment tests what happens when saving and deleting a segment
func TestBatch_WriteSegment(t *testing.T) {
	a := &storetesting.MockAdapter{}

	s := cstesting.RandomSegment()
	lh := testutil.RandomHash()

	batch := NewBatch(a)

	err := batch.SaveSegment(s)
	if err != nil {
		t.Fatalf("batch.SaveSegment(): err: %s", err)
	}

	_, err = batch.DeleteSegment(lh)
	if err != nil {
		t.Fatalf("batch.DeleteSegment(): err: %s", err)
	}

	err = batch.Write()
	if err != nil {
		t.Fatalf("batch.Write(): err: %s", err)
	}

	if got, want := a.MockSaveSegment.CalledCount, 1; got != want {
		t.Errorf("batch.Write(): expected to have called SaveSegment %d time, got %d", want, got)
	}

	if got, want := a.MockSaveSegment.LastCalledWith, s; got != want {
		t.Errorf("batch.Write(): expected to have called SaveSegment with %v, got %v", want, got)
	}

	if got, want := a.MockDeleteSegment.CalledCount, 1; got != want {
		t.Errorf("batch.Write(): expected to have called DeleteSegment %d time, got %d", want, got)
	}

	if got, want := a.MockDeleteSegment.LastCalledWith, lh; got != want {
		t.Errorf("batch.Write(): expected to have called DeleteSegment with %v, got %v", want, got)
	}
}

// TestBatch_WriteValue tests what happens when saving and deleting a value
func TestBatch_WriteValue(t *testing.T) {
	a := &storetesting.MockAdapter{}

	k := testutil.RandomKey()
	v := testutil.RandomValue()
	k2 := testutil.RandomKey()

	batch := NewBatch(a)

	err := batch.SaveValue(k, v)
	if err != nil {
		t.Fatalf("batch.SaveValue(): err: %s", err)
	}

	_, err = batch.DeleteValue(k2)
	if err != nil {
		t.Fatalf("batch.DeleteValue(): err: %s", err)
	}

	err = batch.Write()
	if err != nil {
		t.Fatalf("batch.Write(): err: %s", err)
	}

	if got, want := a.MockSaveValue.CalledCount, 1; got != want {
		t.Errorf("batch.Write(): expected to have called SaveSegment %d time, got %d", want, got)
	}

	got := a.MockSaveValue.LastCalledWith
	if !bytes.Equal(got[0], k) || !bytes.Equal(got[1], v) {
		t.Errorf("batch.Write(): expected to have called SaveValue with %v, got %v", [][]byte{k, v}, got)
	}

	if got, want := a.MockDeleteValue.CalledCount, 1; got != want {
		t.Errorf("batch.Write(): expected to have called DeleteValue %d time, got %d", want, got)
	}

	if got, want := a.MockDeleteValue.LastCalledWith, k2; !bytes.Equal(got, want) {
		t.Errorf("batch.Write(): expected to have called DeleteValue with %v, got %v", k2, got)
	}
}

// TestBatch_WriteSegmentWithFailure tests what happens when a write fails
func TestBatch_WriteSegmentWithFailure(t *testing.T) {
	a := &storetesting.MockAdapter{}
	mockError := errors.New("Error")

	sa := cstesting.RandomSegment()
	sb := cstesting.RandomSegment()

	a.MockSaveSegment.Fn = func(s *cs.Segment) error {
		if s == sa {
			return mockError
		}
		return nil
	}
	batch := NewBatch(a)

	err := batch.SaveSegment(sa)
	if err != nil {
		t.Fatalf("batch.SaveSegment(): err: %s", err)
	}

	err = batch.SaveSegment(sb)
	if err != nil {
		t.Fatalf("batch.SaveSegment(): err: %s", err)
	}

	if got, want := batch.Write(), mockError; got != want {
		t.Errorf("batch.Write returned %v want %v", got, want)
	}
}

func TestBatch_WriteValueWithFailure(t *testing.T) {
	a := &storetesting.MockAdapter{}
	mockError := errors.New("Error")

	k := testutil.RandomKey()
	va := testutil.RandomValue()
	vb := testutil.RandomValue()

	a.MockSaveValue.Fn = func(key, value []byte) error {
		if bytes.Equal(value, va) {
			return mockError
		}
		return nil
	}
	batch := NewBatch(a)

	err := batch.SaveValue(k, va)
	if err != nil {
		t.Fatalf("batch.SaveValue(): err: %s", err)
	}

	err = batch.SaveValue(k, vb)
	if err != nil {
		t.Fatalf("batch.SaveValue(): err: %s", err)
	}

	if got, want := batch.Write(), mockError; got != want {
		t.Errorf("batch.Write returned %v want %v", got, want)
	}
}
