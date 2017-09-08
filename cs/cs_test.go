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

package cs_test

import (
	"math"
	"reflect"
	"sort"
	"testing"

	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/cs/cstesting"
	"github.com/stratumn/sdk/types"
)

func TestSegmentGetLinkHash(t *testing.T) {
	s := cstesting.RandomSegment()
	wantStr := "0123456789012345678901234567890123456789012345678901234567890123"
	s.Meta["linkHash"] = wantStr
	got := s.GetLinkHash()
	want, _ := types.NewBytes32FromString(wantStr)
	if *got != *want {
		t.Errorf("s.GetLinkHash() = %q want %q", got, want)
	}
}

func TestSegmentGetLinkHashString(t *testing.T) {
	s := cstesting.RandomSegment()
	want := "0123456789012345678901234567890123456789012345678901234567890123"
	s.Meta["linkHash"] = want
	got := s.GetLinkHashString()
	if got != want {
		t.Errorf("s.GetLinkHash() = %q want %q", got, want)
	}
}

func TestSegmentValidate_valid(t *testing.T) {
	s := cstesting.RandomSegment()
	if err := s.Validate(); err != nil {
		t.Errorf("s.Validate() = %q want nil", err)
	}
}

func TestSegmentValidate_linkHashNil(t *testing.T) {
	s := cstesting.RandomSegment()
	delete(s.Meta, "linkHash")
	testSegmentValidateError(t, s, "meta.linkHash should be a non empty string")
}

func TestSegmentValidate_linkHashEmpty(t *testing.T) {
	s := cstesting.RandomSegment()
	s.Meta["linkHash"] = ""
	testSegmentValidateError(t, s, "meta.linkHash should be a non empty string")
}

func TestSegmentValidate_linkHashWrongType(t *testing.T) {
	s := cstesting.RandomSegment()
	s.Meta["linkHash"] = 3
	testSegmentValidateError(t, s, "meta.linkHash should be a non empty string")
}

func TestSegmentValidate_processNil(t *testing.T) {
	s := cstesting.RandomSegment()
	delete(s.Link.Meta, "process")
	testSegmentValidateError(t, s, "link.meta.process should be a non empty string")
}

func TestSegmentValidate_processEmpty(t *testing.T) {
	s := cstesting.RandomSegment()
	s.Link.Meta["process"] = ""
	testSegmentValidateError(t, s, "link.meta.process should be a non empty string")
}

func TestSegmentValidate_processWrongType(t *testing.T) {
	s := cstesting.RandomSegment()
	s.Link.Meta["process"] = true
	testSegmentValidateError(t, s, "link.meta.process should be a non empty string")
}

func TestSegmentValidate_mapIDNil(t *testing.T) {
	s := cstesting.RandomSegment()
	delete(s.Link.Meta, "mapId")
	testSegmentValidateError(t, s, "link.meta.mapId should be a non empty string")
}

func TestSegmentValidate_mapIDEmpty(t *testing.T) {
	s := cstesting.RandomSegment()
	s.Link.Meta["mapId"] = ""
	testSegmentValidateError(t, s, "link.meta.mapId should be a non empty string")
}

func TestSegmentValidate_mapIDWrongType(t *testing.T) {
	s := cstesting.RandomSegment()
	s.Link.Meta["mapId"] = true
	testSegmentValidateError(t, s, "link.meta.mapId should be a non empty string")
}

func TestSegmentValidate_prevLinkHashNil(t *testing.T) {
	s := cstesting.RandomSegment()
	delete(s.Link.Meta, "prevLinkHash")
	if err := s.Validate(); err != nil {
		t.Errorf("s.Validate() = %q want nil", err)
	}
}

func TestSegmentValidate_prevLinkHashEmpty(t *testing.T) {
	s := cstesting.RandomSegment()
	s.Link.Meta["prevLinkHash"] = ""
	testSegmentValidateError(t, s, "link.meta.prevLinkHash should be a non empty string")
}

func TestSegmentValidate_prevLinkHashWrongType(t *testing.T) {
	s := cstesting.RandomSegment()
	s.Link.Meta["prevLinkHash"] = []string{}
	testSegmentValidateError(t, s, "link.meta.prevLinkHash should be a non empty string")
}

func TestSegmentValidate_tagsNil(t *testing.T) {
	s := cstesting.RandomSegment()
	delete(s.Link.Meta, "tags")
	if err := s.Validate(); err != nil {
		t.Errorf("s.Validate() = %q want nil", err)
	}
}

func TestSegmentValidate_tagsWrongType(t *testing.T) {
	s := cstesting.RandomSegment()
	s.Link.Meta["tags"] = 2.4
	testSegmentValidateError(t, s, "link.meta.tags should be an array of non empty string")
}

func TestSegmentValidate_tagsWrongElementType(t *testing.T) {
	s := cstesting.RandomSegment()
	s.Link.Meta["tags"] = []interface{}{1, true, 3}
	testSegmentValidateError(t, s, "link.meta.tags should be an array of non empty string")
}

func TestSegmentValidate_tagsEmpty(t *testing.T) {
	s := cstesting.RandomSegment()
	s.Link.Meta["tags"] = []interface{}{"test", ""}
	testSegmentValidateError(t, s, "link.meta.tags should be an array of non empty string")
}

func TestSegmentValidate_priorityNil(t *testing.T) {
	s := cstesting.RandomSegment()
	delete(s.Link.Meta, "priority")
	if err := s.Validate(); err != nil {
		t.Errorf("s.Validate() = %q want nil", err)
	}
}

func TestSegmentValidate_priorityWrongType(t *testing.T) {
	s := cstesting.RandomSegment()
	s.Link.Meta["priority"] = false
	testSegmentValidateError(t, s, "link.meta.priority should be a float64")
}

func TestSegmentSliceSort_priority(t *testing.T) {
	slice := cs.SegmentSlice{
		&cs.Segment{Link: cs.Link{Meta: map[string]interface{}{"priority": 2.3}}},
		&cs.Segment{Link: cs.Link{Meta: map[string]interface{}{"priority": -1.1}}},
		&cs.Segment{Link: cs.Link{Meta: map[string]interface{}{"priority": 3.33}}},
	}

	sort.Sort(slice)
	wantLTE := 100.0
	for i, s := range slice {
		got := s.Link.Meta["priority"].(float64)
		if got > wantLTE {
			t.Errorf("slice#%d: priority = %f want <= %f", i, got, wantLTE)
		}
		wantLTE = got
	}
}

func TestSegmentSliceSort_linkHash(t *testing.T) {
	slice := cs.SegmentSlice{
		&cs.Segment{Link: cs.Link{Meta: map[string]interface{}{"priority": 2.0}}, Meta: map[string]interface{}{"linkHash": "c"}},
		&cs.Segment{Link: cs.Link{Meta: map[string]interface{}{"priority": 2.0}}, Meta: map[string]interface{}{"linkHash": "b"}},
	}

	sort.Sort(slice)
	wantGTE := "a"
	for i, s := range slice {
		got := s.Meta["linkHash"].(string)
		if got < wantGTE {
			t.Errorf("slice#%d: linkHash = %q want >= %q", i, got, wantGTE)
		}

		wantGTE = got
	}
}

func TestSegmentSliceSort_noPriority(t *testing.T) {
	slice := cs.SegmentSlice{
		&cs.Segment{Link: cs.Link{Meta: map[string]interface{}{"priority": 2.3}}},
		&cs.Segment{Link: cs.Link{Meta: map[string]interface{}{}}},
		&cs.Segment{Link: cs.Link{Meta: map[string]interface{}{"priority": 3.33}}},
	}

	sort.Sort(slice)
	wantLTE := 100.0
	for i, s := range slice {
		got, ok := s.Link.Meta["priority"].(float64)
		if ok {
			if got > wantLTE {
				t.Errorf("slice#%d: priority = %f want <= %f", i, got, wantLTE)
			}

			wantLTE = got
		} else {
			wantLTE = 0
		}
	}
}

func TestLinkGetPriority_notNil(t *testing.T) {
	s := cstesting.RandomSegment()
	want := float64(1.0)
	s.Link.Meta["priority"] = want
	got := s.Link.GetPriority()
	if got != want {
		t.Errorf("s.Link.GetPriority() = %f want %f", got, want)
	}
}

func TestLinkGetPriority_nil(t *testing.T) {
	s := cstesting.RandomSegment()
	delete(s.Link.Meta, "priority")
	if got := s.Link.GetPriority(); !math.IsInf(got, -1) {
		t.Errorf("s.Link.GetPriority() = %f want -Inf", got)
	}
}

func TestLinkGetMapID(t *testing.T) {
	s := cstesting.RandomSegment()
	want := "hello"
	s.Link.Meta["mapId"] = want
	got := s.Link.GetMapID()
	if got != want {
		t.Errorf("s.Link.GetMapID() = %q want %q", got, want)
	}
}

func TestLinkGetPrevLinkHash_notNil(t *testing.T) {
	s := cstesting.RandomSegment()
	wantStr := "0123456789012345678901234567890123456789012345678901234567890123"
	s.Link.Meta["prevLinkHash"] = wantStr
	got := s.Link.GetPrevLinkHash()
	want, _ := types.NewBytes32FromString(wantStr)
	if *got != *want {
		t.Errorf("s.Link.GetPrevLinkHash() = %q want %q", got, want)
	}
}

func TestLinkGetPrevLinkHash_nil(t *testing.T) {
	s := cstesting.RandomSegment()
	delete(s.Link.Meta, "prevLinkHash")
	if got := s.Link.GetPrevLinkHash(); got != nil {
		t.Errorf("s.Link.GetPrevLinkHash() = %q want nil", got)
	}
}

func TestLinkGetPrevLinkHashString_notNil(t *testing.T) {
	s := cstesting.RandomSegment()
	want := "0123456789012345678901234567890123456789012345678901234567890123"
	s.Link.Meta["prevLinkHash"] = want
	got := s.Link.GetPrevLinkHashString()
	if got != want {
		t.Errorf("s.Link.GetPrevLinkHashString() = %q want %q", got, want)
	}
}

func TestLinkGetPrevLinkHashString_nil(t *testing.T) {
	s := cstesting.RandomSegment()
	delete(s.Link.Meta, "prevLinkHash")
	if got, want := s.Link.GetPrevLinkHashString(), ""; got != want {
		t.Errorf("s.Link.GetPrevLinkHashString() = %q want %q", got, want)
	}
}

func TestLinkGetTags_notNil(t *testing.T) {
	s := cstesting.RandomSegment()
	want := []string{"one", "two"}
	s.Link.Meta["tags"] = []interface{}{"one", "two"}
	got := s.Link.GetTags()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("s.Link.GetTags() = %q want %q", got, want)
	}
}

func TestLinkGetTags_nil(t *testing.T) {
	s := cstesting.RandomSegment()
	delete(s.Link.Meta, "tags")
	got := s.Link.GetTags()
	if got != nil {
		t.Errorf("s.Link.GetTags() = %q want nil", got)
	}
}

func TestLinkGetTagMap(t *testing.T) {
	s := cstesting.RandomSegment()
	s.Link.Meta["tags"] = []interface{}{"one", "two"}
	tags := s.Link.GetTagMap()
	if _, got := tags["one"]; !got {
		t.Errorf(`tags["one"] = %v want %v`, got, true)
	}
	if _, got := tags["two"]; !got {
		t.Errorf(`tags["two"] = %v want %v`, got, true)
	}
	if _, got := tags["three"]; got {
		t.Errorf(`tags["three"] = %v want %v`, got, false)
	}
}
