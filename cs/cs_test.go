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
	"sort"
	"testing"

	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/cs/cstesting"
	"github.com/stratumn/sdk/testutil"
	"github.com/stratumn/sdk/types"
	"github.com/stretchr/testify/assert"
)

func TestSegmentGetLinkHash(t *testing.T) {
	l := cstesting.RandomLink()
	lh, _ := l.Hash()
	s := l.Segmentify()
	assert.EqualValues(t, lh, s.GetLinkHash(), "s.GetLinkHash()")
}
func TestSegmentGetLinkHashString(t *testing.T) {
	l := cstesting.RandomLink()
	lh, _ := l.HashString()
	s := l.Segmentify()
	assert.EqualValues(t, lh, s.GetLinkHashString(), "s.GetLinkHashString()")
}

func TestSegmentHashLink(t *testing.T) {
	l := cstesting.RandomLink()
	s := l.Segmentify()
	_, err := s.HashLink()
	assert.NoError(t, err, "s.HashLink()")
}

func TestSegmentSetLinkHash(t *testing.T) {
	l := cstesting.RandomLink()
	lh, _ := l.Hash()
	s := &cs.Segment{
		Link: *l,
	}
	err := s.SetLinkHash()
	assert.NoError(t, err, "s.SetLinkHash()")
	assert.EqualValues(t, lh, s.GetLinkHash(), "s.GetLinkHash()")
}

func TestLinkValidate_valid(t *testing.T) {
	l := cstesting.RandomLink()
	err := l.Validate(nil)
	assert.NoError(t, err, "l.Validate()")
}

func TestSegmentValidate_valid(t *testing.T) {
	s := cstesting.RandomSegment()
	err := s.Validate(nil)
	assert.NoError(t, err, "s.Validate()")
}

func TestSegmentValidate_invalidLinkHash(t *testing.T) {
	s := &cs.Segment{
		Link: *cstesting.RandomLink(),
		Meta: cs.SegmentMeta{
			LinkHash: testutil.RandomString(24),
		},
	}
	err := s.Validate(nil)
	assert.Error(t, err)
}

func TestLinkValidate_processNil(t *testing.T) {
	l := cstesting.RandomLink()
	delete(l.Meta, "process")
	testLinkValidateError(t, l, nil, "link.meta.process should be a non empty string")
}

func TestLinkValidate_processEmpty(t *testing.T) {
	l := cstesting.RandomLink()
	l.Meta["process"] = ""
	testLinkValidateError(t, l, nil, "link.meta.process should be a non empty string")
}

func TestLinkValidate_processWrongType(t *testing.T) {
	l := cstesting.RandomLink()
	l.Meta["process"] = true
	testLinkValidateError(t, l, nil, "link.meta.process should be a non empty string")
}

func TestLinkValidate_mapIDNil(t *testing.T) {
	l := cstesting.RandomLink()
	delete(l.Meta, "mapId")
	testLinkValidateError(t, l, nil, "link.meta.mapId should be a non empty string")
}

func TestLinkValidate_mapIDEmpty(t *testing.T) {
	l := cstesting.RandomLink()
	l.Meta["mapId"] = ""
	testLinkValidateError(t, l, nil, "link.meta.mapId should be a non empty string")
}

func TestLinkValidate_mapIDWrongType(t *testing.T) {
	l := cstesting.RandomLink()
	l.Meta["mapId"] = true
	testLinkValidateError(t, l, nil, "link.meta.mapId should be a non empty string")
}

func TestLinkValidate_prevLinkHashNil(t *testing.T) {
	l := cstesting.RandomLink()
	delete(l.Meta, "prevLinkHash")
	err := l.Validate(nil)
	assert.NoError(t, err, "l.Validate()")
}

func TestLinkValidate_prevLinkHashEmpty(t *testing.T) {
	l := cstesting.RandomLink()
	l.Meta["prevLinkHash"] = ""
	testLinkValidateError(t, l, nil, "link.meta.prevLinkHash should be a non empty string")
}

func TestLinkValidate_prevLinkHashWrongType(t *testing.T) {
	l := cstesting.RandomLink()
	l.Meta["prevLinkHash"] = []string{}
	testLinkValidateError(t, l, nil, "link.meta.prevLinkHash should be a non empty string")
}

func TestLinkValidate_tagsNil(t *testing.T) {
	l := cstesting.RandomLink()
	delete(l.Meta, "tags")
	err := l.Validate(nil)
	assert.NoError(t, err, "l.Validate()")
}

func TestLinkValidate_tagsWrongType(t *testing.T) {
	l := cstesting.RandomLink()
	l.Meta["tags"] = 2.4
	testLinkValidateError(t, l, nil, "link.meta.tags should be an array of non empty string")
}

func TestLinkValidate_tagsWrongElementType(t *testing.T) {
	l := cstesting.RandomLink()
	l.Meta["tags"] = []interface{}{1, true, 3}
	testLinkValidateError(t, l, nil, "link.meta.tags should be an array of non empty string")
}

func TestLinkValidate_tagsEmpty(t *testing.T) {
	l := cstesting.RandomLink()
	l.Meta["tags"] = []interface{}{"test", ""}
	testLinkValidateError(t, l, nil, "link.meta.tags should be an array of non empty string")
}

func TestLinkValidate_priorityNil(t *testing.T) {
	l := cstesting.RandomLink()
	delete(l.Meta, "priority")
	err := l.Validate(nil)
	assert.NoError(t, err, "l.Validate()")
}

func TestLinkValidate_priorityWrongType(t *testing.T) {
	l := cstesting.RandomLink()
	l.Meta["priority"] = false
	testLinkValidateError(t, l, nil, "link.meta.priority should be a float64")
}

func TestLinkValidate_refBadType(t *testing.T) {
	l := cstesting.RandomLink()
	l.Meta["refs"] = []interface{}{"foo", "bar"}
	testLinkValidateError(t, l, nil, "link.meta.refs[0] should be a map")
}

func TestLinkValidate_refGoodLink(t *testing.T) {
	l := cstesting.RandomLink()
	ref := cstesting.RandomLink()
	appendRefSegment(l, ref)
	err := l.Validate(nil)
	assert.NoError(t, err, "l.Validate()")
}

func TestLinkValidate_refBadLink(t *testing.T) {
	l := cstesting.RandomLink()
	ref := cstesting.RandomLink()
	ref.Meta["process"] = ""
	appendRefSegment(l, ref)
	testLinkValidateErrorWrapper(t, l, nil, "invalid link.meta.refs[0].segment")
}

func TestLinkValidate_refBadLinkFormat(t *testing.T) {
	l := cstesting.RandomLink()
	ref := cstesting.RandomLink()
	appendRefSegment(l, ref)
	l.Meta["refs"] = append(l.Meta["refs"].([]interface{}), map[string]interface{}{"segment": "foobar"})
	testLinkValidateError(t, l, nil, "link.meta.refs[1].segment should be a valid json segment")
}

func TestLinkValidate_refMissingProcess(t *testing.T) {
	l := cstesting.RandomLink()
	appendRefLink(l, "", testutil.RandomHash().String())
	testLinkValidateError(t, l, nil, "link.meta.refs[0].process should be a non empty string")
}

func TestLinkValidate_refMissingLinkHash(t *testing.T) {
	l := cstesting.RandomLink()
	appendRefLink(l, testutil.RandomString(24), "")
	testLinkValidateError(t, l, nil, "link.meta.refs[0].linkHash should be a non empty string")
}

func TestLinkValidate_refLinkHashBadType(t *testing.T) {
	l := cstesting.RandomLink()
	appendRefLink(l, testutil.RandomString(24), "FooBar")
	testLinkValidateError(t, l, nil, "link.meta.refs[0].linkHash should be a bytes32 field")
}

func TestLinkValidate_refGoodLinkNotChecked(t *testing.T) {
	l := cstesting.RandomLink()
	appendRefLink(l, l.Meta["process"].(string), testutil.RandomHash().String())
	err := l.Validate(nil)
	assert.NoError(t, err)
}

func TestLinkValidate_refGoodLinkChecked(t *testing.T) {
	l := cstesting.RandomLink()
	appendRefLink(l, l.Meta["process"].(string), testutil.RandomHash().String())
	err := l.Validate(func(linkHash *types.Bytes32) (*cs.Segment, error) {
		return cstesting.RandomSegment(), nil
	})
	assert.NoError(t, err)
}

func TestLinkValidate_refGoodLinkNotFound(t *testing.T) {
	l := cstesting.RandomLink()
	appendRefLink(l, l.Meta["process"].(string), testutil.RandomHash().String())
	testLinkValidateErrorWrapper(t, l, func(linkHash *types.Bytes32) (*cs.Segment, error) {
		return nil, errors.New("Bad mood")
	}, "Bad mood")
}

func TestLinkValidate_refGoodNilLink(t *testing.T) {
	l := cstesting.RandomLink()
	appendRefLink(l, l.Meta["process"].(string), testutil.RandomHash().String())
	testLinkValidateError(t, l, func(linkHash *types.Bytes32) (*cs.Segment, error) {
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
	l := cstesting.RandomLink()
	want := float64(1.0)
	l.Meta["priority"] = want
	got := l.GetPriority()
	assert.EqualValues(t, want, got, "Invalid priority")
}

func TestLinkGetPriority_nil(t *testing.T) {
	l := cstesting.RandomLink()
	delete(l.Meta, "priority")
	got := l.GetPriority()
	assert.True(t, math.IsInf(got, -1), "Priority should be -Inf")
}

func TestLinkGetMapID(t *testing.T) {
	l := cstesting.RandomLink()
	want := "hello"
	l.Meta["mapId"] = want
	got := l.GetMapID()
	assert.EqualValues(t, want, got, "Invalid map id")
}

func TestLinkGetPrevLinkHash_notNil(t *testing.T) {
	l := cstesting.RandomLink()
	wantStr := "0123456789012345678901234567890123456789012345678901234567890123"
	l.Meta["prevLinkHash"] = wantStr
	got := l.GetPrevLinkHash()
	want, _ := types.NewBytes32FromString(wantStr)
	assert.EqualValues(t, want, got, "Invalid PrevLinkHash")
}

func TestLinkGetPrevLinkHash_nil(t *testing.T) {
	l := cstesting.RandomLink()
	delete(l.Meta, "prevLinkHash")
	got := l.GetPrevLinkHash()
	assert.Nil(t, got, "PrevLinkHash")
}

func TestLinkGetPrevLinkHashString_notNil(t *testing.T) {
	l := cstesting.RandomLink()
	want := "0123456789012345678901234567890123456789012345678901234567890123"
	l.Meta["prevLinkHash"] = want
	got := l.GetPrevLinkHashString()
	assert.EqualValues(t, want, got, "PrevLinkHash")
}

func TestLinkGetPrevLinkHashString_nil(t *testing.T) {
	l := cstesting.RandomLink()
	delete(l.Meta, "prevLinkHash")
	assert.EqualValues(t, "", l.GetPrevLinkHashString(), "Expected empty PrevLinkHash")
}

func TestLinkGetTags_notNil(t *testing.T) {
	l := cstesting.RandomLink()
	want := []string{"one", "two"}
	l.Meta["tags"] = []interface{}{"one", "two"}
	got := l.GetTags()
	assert.EqualValues(t, want, got, "Invalid tags")
}

func TestLinkGetTags_nil(t *testing.T) {
	l := cstesting.RandomLink()
	delete(l.Meta, "tags")
	got := l.GetTags()
	assert.Nil(t, got, "Tags")
}

func TestLinkGetTagMap(t *testing.T) {
	l := cstesting.RandomLink()
	l.Meta["tags"] = []interface{}{"one", "two"}
	tags := l.GetTagMap()
	_, got := tags["one"]
	assert.True(t, got, `tags["one"]`)
	_, got = tags["two"]
	assert.True(t, got, `tags["two"]`)
	_, got = tags["three"]
	assert.False(t, got, `tags["three"]`)
}

func TestLinkGetProcess(t *testing.T) {
	l := cstesting.RandomLink()
	want := "hello"
	l.Meta["process"] = want
	got := l.GetProcess()
	assert.EqualValues(t, want, got, "Invalid processes")
}

func TestAddEvidence(t *testing.T) {
	s := cstesting.RandomSegment()
	s.Meta.AddEvidence(TestEvidence)
	assert.Equal(t, 1, len(s.Meta.Evidences), "Evidences count")

	err := s.Meta.AddEvidence(TestEvidence)
	assert.Error(t, err, "trying to add an already existing evidence: should have failed")

	e2 := TestEvidence
	e2.Provider = "xyz"
	err = s.Meta.AddEvidence(e2)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(s.Meta.Evidences), "Evidences count")
}

func TestGetEvidence(t *testing.T) {
	s := cstesting.RandomSegment()
	s.Meta.AddEvidence(TestEvidence)

	assert.EqualValues(t, TestEvidence, *s.Meta.GetEvidence(TestChainId), "Invalid evidence")
	assert.Nil(t, s.Meta.GetEvidence("unknown"))
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

	assert.Equal(t, 2, len(s.Meta.FindEvidences(TestEvidence.Backend)), "Evidences count")
	assert.Equal(t, 0, len(s.Meta.FindEvidences("unknown")), "Unknown evidences")
}

func TestEmptySegment(t *testing.T) {
	s := cstesting.RandomSegment()
	assert.False(t, s.IsEmpty(), "Segment should not be empty")
	s = &cs.Segment{}
	assert.True(t, s.IsEmpty(), "Segment should be empty")
}
