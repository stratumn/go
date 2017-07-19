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

package storetestcases

import (
	"encoding/json"
	"reflect"
	"testing"

	"bytes"

	"github.com/stratumn/sdk/cs/cstesting"
	"github.com/stratumn/sdk/store"
	"github.com/stratumn/sdk/testutil"
)

// TestBatchSaveSegment tests what happens
// when you write a segment in a Batch
func (f Factory) TestBatchSaveSegment(t *testing.T) {
	a, b := f.initBatch(t)
	defer f.free(a)

	s1 := cstesting.RandomSegment()
	b.SaveSegment(s1)

	s2, err := a.GetSegment(s1.GetLinkHash())
	if err != nil {
		t.Fatalf("a.GetSegment(): err: %s", err)
	}

	if got := s2; got != nil {
		t.Error("s2 != nil want nil")
	}
}

// TestBatchSaveValue tests what happens
// when you write a value in a Batch
func (f Factory) TestBatchSaveValue(t *testing.T) {
	a, b := f.initBatch(t)
	defer f.free(a)

	k := testutil.RandomKey()
	v1 := testutil.RandomValue()
	err := b.SaveValue(k, v1)
	if err != nil {
		t.Fatalf("b.SaveValue(): err: %s", err)
	}

	v2, err := a.GetValue(k)
	if err != nil {
		t.Fatalf("a.GetValue(): err: %s", err)
	}

	if got := v2; got != nil {
		t.Error("v2 != nil want nil")
	}
}

// TestBatchDeleteSegment tests what happens when you delete a segment from
// a batch.
func (f Factory) TestBatchDeleteSegment(t *testing.T) {
	a, b := f.initBatch(t)
	defer f.free(a)

	s1 := cstesting.RandomSegment()
	a.SaveSegment(s1)

	linkHash := s1.GetLinkHash()

	s2, err := b.DeleteSegment(linkHash)
	if err != nil {
		t.Fatalf("b.DeleteSegment(): err: %s", err)
	}

	if got := s2; got == nil {
		t.Fatal("s2 = nil want *cs.Segment")
	}

	delete(s2.Meta, "evidence")

	if got, want := s2, s1; !reflect.DeepEqual(want, got) {
		gotJS, _ := json.MarshalIndent(got, "", "  ")
		wantJS, _ := json.MarshalIndent(want, "", "  ")
		t.Errorf("s2 = %s\n want%s", gotJS, wantJS)
	}

	s2, err = a.GetSegment(linkHash)
	if err != nil {
		t.Fatalf("a.GetSegment(): err: %s", err)
	}
	if got := s2; got == nil {
		gotJS, _ := json.MarshalIndent(got, "", "  ")
		t.Fatalf("s2 = %s\n want %s", gotJS, s2)
	}
}

// TestBatchDeleteValue tests what happens when you delete a value from
// a batch.
func (f Factory) TestBatchDeleteValue(t *testing.T) {
	a, b := f.initBatch(t)
	defer f.free(a)

	k := testutil.RandomKey()
	v1 := testutil.RandomValue()
	a.SaveValue(k, v1)

	v2, err := b.DeleteValue(k)
	if err != nil {
		t.Fatalf("b.DeleteValue(): err: %s", err)
	}

	if got := v2; got == nil {
		t.Fatal("s2 = nil want []byte")
	}

	if got, want := v2, v1; bytes.Compare(got, want) != 0 {
		t.Errorf("s2 = %s\n want%s", got, want)
	}

	v2, err = a.GetValue(k)
	if err != nil {
		t.Fatalf("a.GetValue(): err: %s", err)
	}
	if got := v2; got == nil {
		t.Fatalf("s2 = %s\n want %s", got, v2)
	}
}

// TestBatchWriteSaveSegment tests what happens when you write a Batch with a saved segment.
func (f Factory) TestBatchWriteSaveSegment(t *testing.T) {
	a, b := f.initBatch(t)
	defer f.free(a)

	s1 := cstesting.RandomSegment()
	err := b.SaveSegment(s1)
	if err != nil {
		t.Fatalf("b.SaveSegment(): err: %s", err)
	}

	err = b.Write()
	if err != nil {
		t.Fatalf("b.Write(): err: %s", err)
	}

	s2, err := a.GetSegment(s1.GetLinkHash())
	if err != nil {
		t.Fatalf("a.GetSegment(): err: %s", err)
	}

	if got := s2; got == nil {
		t.Fatal("s2 = nil want *cs.Segment")
	}

	delete(s2.Meta, "evidence")

	if got, want := s2, s1; !reflect.DeepEqual(want, got) {
		gotJS, _ := json.MarshalIndent(got, "", "  ")
		wantJS, _ := json.MarshalIndent(want, "", "  ")
		t.Errorf("s2 = %s\n want%s", gotJS, wantJS)
	}
}

// TestBatchWriteSaveValue tests what happens when you write a Batch with a saved segment.
func (f Factory) TestBatchWriteSaveValue(t *testing.T) {
	a, b := f.initBatch(t)
	defer f.free(a)

	k := testutil.RandomKey()
	v1 := testutil.RandomValue()
	err := b.SaveValue(k, v1)
	if err != nil {
		t.Fatalf("b.SaveValue(): err: %s", err)
	}

	err = b.Write()
	if err != nil {
		t.Fatalf("b.Write(): err: %s", err)
	}

	v2, err := a.GetValue(k)
	if err != nil {
		t.Fatalf("a.GetValue(): err: %s", err)
	}

	if got := v2; got == nil {
		t.Fatal("s2 = nil want []byte")
	}

	if got, want := v2, v1; bytes.Compare(got, want) != 0 {
		t.Errorf("s2 = %s\n want%s", got, want)
	}
}

// TestBatchWriteDeleteSegment tests what happens when you write a Batch with a deleted segment.
func (f Factory) TestBatchWriteDeleteSegment(t *testing.T) {
	a, b := f.initBatch(t)
	defer f.free(a)

	s1 := cstesting.RandomSegment()
	a.SaveSegment(s1)

	linkHash := s1.GetLinkHash()
	s2, err := b.DeleteSegment(linkHash)
	if err != nil {
		t.Fatalf("a.DeleteSegment(): err: %s", err)
	}
	err = b.Write()
	if err != nil {
		t.Fatalf("b.Write(): err: %s", err)
	}

	s2, err = a.GetSegment(linkHash)
	if err != nil {
		t.Fatalf("a.GetSegment(): err: %s", err)
	}
	if got := s2; got != nil {
		gotJS, _ := json.MarshalIndent(got, "", "  ")
		t.Errorf("s2 = %s\n want nil", gotJS)
	}
}

// TestBatchWriteDeleteValue tests what happens when you write a Batch with a deleted value.
func (f Factory) TestBatchWriteDeleteValue(t *testing.T) {
	a, b := f.initBatch(t)
	defer f.free(a)

	k := testutil.RandomKey()
	v1 := testutil.RandomValue()
	err := a.SaveValue(k, v1)
	if err != nil {
		t.Fatalf("b.SaveValue(): err: %s", err)
	}

	v2, err := b.DeleteValue(k)
	if err != nil {
		t.Fatalf("a.DeleteValue(): err: %s", err)
	}
	err = b.Write()
	if err != nil {
		t.Fatalf("b.Write(): err: %s", err)
	}

	v2, err = a.GetValue(k)
	if err != nil {
		t.Fatalf("a.GetValue(): err: %s", err)
	}
	if got := v2; got != nil {
		t.Errorf("s2 = %s\n want nil", got)
	}
}

func (f Factory) initBatch(t *testing.T) (store.Adapter, store.Batch) {
	a := f.initAdapter(t)

	b, err := a.NewBatch()
	if err != nil {
		t.Fatalf("a.NewBatch(): err: %s", err)
	}
	if b == nil {
		t.Fatal("b = nil want store.Batch")
	}

	return a, b
}
