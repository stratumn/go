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
	"errors"
	"math"
	"reflect"
	"sort"
	"testing"

	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/cs/cstesting"
	"github.com/stratumn/sdk/testutil"
	"github.com/stratumn/sdk/types"
)

func TestSegmentGetLinkHash(t *testing.T) {
	s := cstesting.RandomSegment()
	wantStr := "0123456789012345678901234567890123456789012345678901234567890123"
	s.Meta.LinkHash = wantStr
	got := s.GetLinkHash()
	want, _ := types.NewBytes32FromString(wantStr)
	if *got != *want {
		t.Errorf("s.GetLinkHash() = %q want %q", got, want)
	}
}
func TestSegmentGetLinkHashString(t *testing.T) {
	s := cstesting.RandomSegment()
	wantStr := "0123456789012345678901234567890123456789012345678901234567890123"
	s.Meta.LinkHash = wantStr
	got := s.GetLinkHashString()
	if got != wantStr {
		t.Errorf("s.GetLinkHashString() = %s want %s", got, wantStr)
	}
}

func TestSegmentHashLink(t *testing.T) {
	s := cstesting.RandomSegment()
	_, err := s.HashLink()
	if err != nil {
		t.Errorf("s.HashLink() = %q want nil", err)
	}
}

func TestSegmentSetLinkHash(t *testing.T) {
	s := cstesting.RandomSegment()
	if err := s.SetLinkHash(); err != nil {
		t.Errorf("s.SetLinkHash() = %q want nil", err)
	}
}

func TestSegmentValidate_valid(t *testing.T) {
	s := cstesting.RandomSegment()
	if err := s.Validate(nil); err != nil {
		t.Errorf("s.Validate() = %q want nil", err)
	}
}

func TestSegmentValidate_linkHashEmpty(t *testing.T) {
	s := cstesting.RandomSegment()
	s.Meta.LinkHash = ""
	if err := s.Validate(nil); err == nil {
		t.Error("s.Validate() = nil want Error")
	} else if got, want := err.Error(), "meta.linkHash should be a non empty string"; got != want {
		t.Errorf("s.Validate() = %q want %q", got, want)
	}
}

func TestSegmentValidate_processNil(t *testing.T) {
	s := cstesting.RandomSegment()
	delete(s.Link.Meta, "process")
	testSegmentValidateError(t, s, nil, "link.meta.process should be a non empty string")
}

func TestSegmentValidate_processEmpty(t *testing.T) {
	s := cstesting.RandomSegment()
	s.Link.Meta["process"] = ""
	testSegmentValidateError(t, s, nil, "link.meta.process should be a non empty string")
}

func TestSegmentValidate_processWrongType(t *testing.T) {
	s := cstesting.RandomSegment()
	s.Link.Meta["process"] = true
	testSegmentValidateError(t, s, nil, "link.meta.process should be a non empty string")
}

func TestSegmentValidate_mapIDNil(t *testing.T) {
	s := cstesting.RandomSegment()
	delete(s.Link.Meta, "mapId")
	testSegmentValidateError(t, s, nil, "link.meta.mapId should be a non empty string")
}

func TestSegmentValidate_mapIDEmpty(t *testing.T) {
	s := cstesting.RandomSegment()
	s.Link.Meta["mapId"] = ""
	testSegmentValidateError(t, s, nil, "link.meta.mapId should be a non empty string")
}

func TestSegmentValidate_mapIDWrongType(t *testing.T) {
	s := cstesting.RandomSegment()
	s.Link.Meta["mapId"] = true
	testSegmentValidateError(t, s, nil, "link.meta.mapId should be a non empty string")
}

func TestSegmentValidate_prevLinkHashNil(t *testing.T) {
	s := cstesting.RandomSegment()
	delete(s.Link.Meta, "prevLinkHash")
	s.SetLinkHash()
	if err := s.Validate(nil); err != nil {
		t.Errorf("s.Validate() = %q want nil", err)
	}
}

func TestSegmentValidate_prevLinkHashEmpty(t *testing.T) {
	s := cstesting.RandomSegment()
	s.Link.Meta["prevLinkHash"] = ""
	testSegmentValidateError(t, s, nil, "link.meta.prevLinkHash should be a non empty string")
}

func TestSegmentValidate_prevLinkHashWrongType(t *testing.T) {
	s := cstesting.RandomSegment()
	s.Link.Meta["prevLinkHash"] = []string{}
	testSegmentValidateError(t, s, nil, "link.meta.prevLinkHash should be a non empty string")
}

func TestSegmentValidate_tagsNil(t *testing.T) {
	s := cstesting.RandomSegment()
	delete(s.Link.Meta, "tags")
	s.SetLinkHash()
	if err := s.Validate(nil); err != nil {
		t.Errorf("s.Validate() = %q want nil", err)
	}
}

func TestSegmentValidate_tagsWrongType(t *testing.T) {
	s := cstesting.RandomSegment()
	s.Link.Meta["tags"] = 2.4
	testSegmentValidateError(t, s, nil, "link.meta.tags should be an array of non empty string")
}

func TestSegmentValidate_tagsWrongElementType(t *testing.T) {
	s := cstesting.RandomSegment()
	s.Link.Meta["tags"] = []interface{}{1, true, 3}
	testSegmentValidateError(t, s, nil, "link.meta.tags should be an array of non empty string")
}

func TestSegmentValidate_tagsEmpty(t *testing.T) {
	s := cstesting.RandomSegment()
	s.Link.Meta["tags"] = []interface{}{"test", ""}
	testSegmentValidateError(t, s, nil, "link.meta.tags should be an array of non empty string")
}

func TestSegmentValidate_priorityNil(t *testing.T) {
	s := cstesting.RandomSegment()
	delete(s.Link.Meta, "priority")
	s.SetLinkHash()
	if err := s.Validate(nil); err != nil {
		t.Errorf("s.Validate() = %q want nil", err)
	}
}

func TestSegmentValidate_priorityWrongType(t *testing.T) {
	s := cstesting.RandomSegment()
	s.Link.Meta["priority"] = false
	testSegmentValidateError(t, s, nil, "link.meta.priority should be a float64")
}

func TestSegmentValidate_refBadType(t *testing.T) {
	s := cstesting.RandomSegment()
	s.Link.Meta["refs"] = []interface{}{"foo", "bar"}
	testSegmentValidateError(t, s, nil, "link.meta.refs[0] should be a map")
}

func TestSegmentValidate_refGoodSegment(t *testing.T) {
	s := cstesting.RandomSegment()
	ref := cstesting.RandomSegment()
	appendRefSegment(s, ref)
	s.SetLinkHash()
	if err := s.Validate(nil); err != nil {
		t.Errorf("s.Validate() = %q want nil", err)
	}
}

func TestSegmentValidate_refBadSegment(t *testing.T) {
	s := cstesting.RandomSegment()
	ref := cstesting.RandomSegment()
	ref.Link.Meta["process"] = ""
	appendRefSegment(s, ref)
	testSegmentValidateErrorWrapper(t, s, nil, "invalid link.meta.refs[0].segment")
}

func TestSegmentValidate_refBadSegmentFormat(t *testing.T) {
	s := cstesting.RandomSegment()
	ref := cstesting.RandomSegment()
	appendRefSegment(s, ref)
	s.Link.Meta["refs"] = append(s.Link.Meta["refs"].([]interface{}), map[string]interface{}{"segment": "foobar"})
	testSegmentValidateError(t, s, nil, "link.meta.refs[1].segment should be a valid json segment")
}

func TestSegmentValidate_refMissingProcess(t *testing.T) {
	s := cstesting.RandomSegment()
	appendRefLink(s, "", testutil.RandomHash().String())
	testSegmentValidateError(t, s, nil, "link.meta.refs[0].process should be a non empty string")
}

func TestSegmentValidate_refMissingLinkHash(t *testing.T) {
	s := cstesting.RandomSegment()
	appendRefLink(s, testutil.RandomString(24), "")
	testSegmentValidateError(t, s, nil, "link.meta.refs[0].linkHash should be a non empty string")
}

func TestSegmentValidate_refLinkHashBadType(t *testing.T) {
	s := cstesting.RandomSegment()
	appendRefLink(s, testutil.RandomString(24), "FooBar")
	testSegmentValidateError(t, s, nil, "link.meta.refs[0].linkHash should be a bytes32 field")
}

func TestSegmentValidate_refGoodLinkNotChecked(t *testing.T) {
	s := cstesting.RandomSegment()
	appendRefLink(s, s.Link.Meta["process"].(string), testutil.RandomHash().String())
	s.SetLinkHash()
	if err := s.Validate(nil); err != nil {
		t.Errorf("s.Validate() = %q want nil", err)
	}
}

func TestSegmentValidate_refGoodLinkChecked(t *testing.T) {
	s := cstesting.RandomSegment()
	appendRefLink(s, s.Link.Meta["process"].(string), testutil.RandomHash().String())
	s.SetLinkHash()
	if err := s.Validate(func(linkHash *types.Bytes32) (*cs.Segment, error) {
		return cstesting.RandomSegment(), nil
	}); err != nil {
		t.Errorf("s.Validate() = %q want nil", err)
	}
}

func TestSegmentValidate_refGoodLinkNotFound(t *testing.T) {
	s := cstesting.RandomSegment()
	appendRefLink(s, s.Link.Meta["process"].(string), testutil.RandomHash().String())
	testSegmentValidateErrorWrapper(t, s, func(linkHash *types.Bytes32) (*cs.Segment, error) {
		return nil, errors.New("Bad mood")
	}, "Bad mood")
}

func TestSegmentValidate_refGoodNilLink(t *testing.T) {
	s := cstesting.RandomSegment()
	appendRefLink(s, s.Link.Meta["process"].(string), testutil.RandomHash().String())
	testSegmentValidateError(t, s, func(linkHash *types.Bytes32) (*cs.Segment, error) {
		return nil, nil
	}, "link.meta.refs[0] segment is nil")
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
		&cs.Segment{Link: cs.Link{Meta: map[string]interface{}{"priority": 2.0}}, Meta: cs.SegmentMeta{LinkHash: "c"}},
		&cs.Segment{Link: cs.Link{Meta: map[string]interface{}{"priority": 2.0}}, Meta: cs.SegmentMeta{LinkHash: "b"}},
	}

	sort.Sort(slice)
	wantGTE := "a"
	for i, s := range slice {
		got := s.Meta.LinkHash
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

func TestLinkGetProcess(t *testing.T) {
	s := cstesting.RandomSegment()
	want := "hello"
	s.Link.Meta["process"] = want
	if got := s.Link.GetProcess(); got != want {
		t.Errorf("s.Link.GetProcess() = %q want %q", got, want)
	}
}

func TestAddEvidence(t *testing.T) {
	s := cstesting.RandomSegment()
	s.Meta.AddEvidence(TestEvidence)

	if got := s.Meta.Evidences; len(got) != 1 {
		t.Errorf("len(s.Meta.Evidences) = %d want %d", len(got), 1)
	}

	if err := s.Meta.AddEvidence(TestEvidence); err == nil {
		t.Errorf("trying to add an already existing evidence: should have failed")
	}

	e2 := TestEvidence
	e2.Provider = "xyz"
	if err := s.Meta.AddEvidence(e2); err != nil {
		t.Error(err)
	}

	if got := s.Meta.Evidences; len(got) != 2 {
		t.Errorf("len(s.Meta.Evidences) = %d want %d", len(got), 2)
	}
}

func TestGetEvidence(t *testing.T) {
	s := cstesting.RandomSegment()
	s.Meta.AddEvidence(TestEvidence)

	if got := s.Meta.GetEvidence(TestChainId); *got != TestEvidence {
		t.Errorf("s.Meta.GetEvidence() = %v want %v", got, TestEvidence)
	}

	if got := s.Meta.GetEvidence("unknown"); got != nil {
		t.Errorf("s.Meta.GetEvidence() = %v want (nil)", got)
	}
}

func TestFindEvidences(t *testing.T) {
	s := cstesting.RandomSegment()
	e1 := TestEvidence
	s.Meta.AddEvidence(e1)

	e2 := TestEvidence
	e2.Provider = "xyz"
	s.Meta.AddEvidence(e2)

	e3 := TestEvidence
	e3.Provider = "zef"
	e3.Backend = "bcbatchfossilizer"
	s.Meta.AddEvidence(e3)

	if got := s.Meta.FindEvidences(TestEvidence.Backend); len(got) != 2 {
		t.Errorf("len(got) = %q want %q", got, 2)
	}

	if got := s.Meta.FindEvidences("unknown"); len(got) != 0 {
		t.Errorf("len(got) = %d want %d", len(got), 0)
	}
}

func TestEmptySegment(t *testing.T) {
	if got, want := cstesting.RandomSegment().IsEmpty(), false; got != want {
		t.Errorf("IsEmpty = %t want %t", got, want)
	}
	s := cs.Segment{}
	if got, want := s.IsEmpty(), true; got != want {
		t.Errorf("IsEmpty = %t want %t", got, want)
	}
}
