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
	"math/rand"
	"sort"
	"testing"

	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/cs/cstesting"
	"github.com/stratumn/go-indigocore/testutil"
	"github.com/stratumn/go-indigocore/types"
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

func TestLinkValidate_processEmpty(t *testing.T) {
	l := cstesting.RandomLink()
	l.Meta.Process = ""
	testLinkValidateError(t, l, nil, "link.meta.process should be a non empty string")
}

func TestLinkValidate_mapIDEmpty(t *testing.T) {
	l := cstesting.RandomLink()
	l.Meta.MapID = ""
	testLinkValidateError(t, l, nil, "link.meta.mapId should be a non empty string")
}

func TestLinkValidate_prevLinkHashEmpty(t *testing.T) {
	l := cstesting.RandomLink()
	l.Meta.PrevLinkHash = ""
	err := l.Validate(nil)
	assert.NoError(t, err, "l.Validate()")
}

func TestLinkValidate_tagsNil(t *testing.T) {
	l := cstesting.RandomLink()
	l.Meta.Tags = nil
	err := l.Validate(nil)
	assert.NoError(t, err, "l.Validate()")
}

func TestLinkValidate_tagsEmpty(t *testing.T) {
	l := cstesting.RandomLink()
	l.Meta.Tags = []string{"test", ""}
	testLinkValidateError(t, l, nil, "link.meta.tags should be an array of non empty string")
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
	ref.Meta.Process = ""
	appendRefSegment(l, ref)
	testLinkValidateErrorWrapper(t, l, nil, "invalid link.meta.refs[0].segment")
}

func TestLinkValidate_refMissingProcess(t *testing.T) {
	l := cstesting.RandomLink()
	appendRefLink(l, "", testutil.RandomHash().String())
	testLinkValidateError(t, l, nil, "link.meta.refs[0].process should be a non empty string")
}

func TestLinkValidate_refMissingLinkHash(t *testing.T) {
	l := cstesting.RandomLink()
	appendRefLink(l, testutil.RandomString(24), "")
	testLinkValidateError(t, l, nil, "link.meta.refs[0].linkHash should be a bytes32 field")
}

func TestLinkValidate_refLinkHashBadType(t *testing.T) {
	l := cstesting.RandomLink()
	appendRefLink(l, testutil.RandomString(24), "FooBar")
	testLinkValidateError(t, l, nil, "link.meta.refs[0].linkHash should be a bytes32 field")
}

func TestLinkValidate_refGoodLinkNotChecked(t *testing.T) {
	l := cstesting.RandomLink()
	appendRefLink(l, l.Meta.Process, testutil.RandomHash().String())
	err := l.Validate(nil)
	assert.NoError(t, err)
}

func TestLinkValidate_refGoodLinkChecked(t *testing.T) {
	l := cstesting.RandomLink()
	appendRefLink(l, l.Meta.Process, testutil.RandomHash().String())
	err := l.Validate(func(linkHash *types.Bytes32) (*cs.Segment, error) {
		return cstesting.RandomSegment(), nil
	})
	assert.NoError(t, err)
}

func TestLinkValidate_refGoodLinkNotFound(t *testing.T) {
	l := cstesting.RandomLink()
	appendRefLink(l, l.Meta.Process, testutil.RandomHash().String())
	testLinkValidateErrorWrapper(t, l, func(linkHash *types.Bytes32) (*cs.Segment, error) {
		return nil, errors.New("Bad mood")
	}, "Bad mood")
}

func TestLinkValidate_refGoodNilLink(t *testing.T) {
	l := cstesting.RandomLink()
	appendRefLink(l, l.Meta.Process, testutil.RandomHash().String())
	testLinkValidateError(t, l, func(linkHash *types.Bytes32) (*cs.Segment, error) {
		return nil, nil
	}, "link.meta.refs[0] segment is nil")
}

func TestLinkValidate_validSignature(t *testing.T) {
	l := cstesting.SignLink(cstesting.RandomLink())
	err := l.Validate(nil)
	assert.NoError(t, err, "l.Validate()")
}

func TestLinkValidate_emptySignatureType(t *testing.T) {
	l := cstesting.RandomLink()
	l.Signatures = append(l.Signatures, &cs.Signature{
		Type: "",
	})
	testLinkValidateError(t, l, nil, "signature.Type cannot be empty")
}

func TestLinkValidate_wrongPublicKeyFormat(t *testing.T) {
	l := cstesting.RandomLink()
	l.Signatures = append(l.Signatures, &cs.Signature{
		Type:      "ok",
		PublicKey: "*test*",
	})
	testLinkValidateError(t, l, nil, "signature.PublicKey [*test*] has to be a base64-encoded string")
}

func TestLinkValidate_wrongSignatureFormat(t *testing.T) {
	l := cstesting.RandomLink()
	l.Signatures = append(l.Signatures, &cs.Signature{
		Type:      "ok",
		PublicKey: "AeZ4",
		Signature: "*test*",
	})
	testLinkValidateError(t, l, nil, "signature.Signature [*test*] has to be a base64-encoded string")
}

func TestLinkValidate_badSignature(t *testing.T) {
	l := cstesting.SignLink(cstesting.RandomLink())
	l.Signatures[0].Signature = "test"
	testLinkValidateError(t, l, nil, "signature verification failed")
}

func TestLinkValidate_wrongPaylodExpression(t *testing.T) {
	l := cstesting.RandomLink()
	l.Signatures = append(l.Signatures, &cs.Signature{
		Type:      "ok",
		PublicKey: "deadbeef",
		Signature: "deadbeef",
		Payload:   "",
	})
	testLinkValidateError(t, l, nil, "signature.Payload [] has to be a JMESPATH expression, got: SyntaxError: Incomplete expression")
}

func TestSegmentSliceSort_priority(t *testing.T) {
	slice := cs.SegmentSlice{
		&cs.Segment{Link: cs.Link{Meta: cs.LinkMeta{Priority: 2.3}}},
		&cs.Segment{Link: cs.Link{Meta: cs.LinkMeta{Priority: -1.1}}},
		&cs.Segment{Link: cs.Link{Meta: cs.LinkMeta{Priority: 3.33}}},
		&cs.Segment{Link: cs.Link{Meta: cs.LinkMeta{Data: map[string]interface{}{}}}},
	}

	sort.Sort(slice)
	wantLTE := 100.0
	for i, s := range slice {
		got := s.Link.Meta.Priority
		if got > wantLTE {
			t.Errorf("slice#%d: priority = %f want <= %f", i, got, wantLTE)
		}
		wantLTE = got
	}
}

func TestSegmentSliceSort_linkHash(t *testing.T) {
	slice := cs.SegmentSlice{
		&cs.Segment{Link: cs.Link{Meta: cs.LinkMeta{Priority: 2.0}}, Meta: cs.SegmentMeta{LinkHash: "c"}},
		&cs.Segment{Link: cs.Link{Meta: cs.LinkMeta{Priority: 2.0}}, Meta: cs.SegmentMeta{LinkHash: "b"}},
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

func TestLinkGetPriority(t *testing.T) {
	l := cstesting.RandomLink()
	want := rand.Float64()
	l.Meta.Priority = want
	got := l.Meta.Priority
	assert.EqualValues(t, want, got, "Invalid priority")
}

func TestLinkGetPriority_default(t *testing.T) {
	l := &cs.Link{}
	got := l.Meta.Priority
	assert.Equal(t, 0., got, "Priority should be zero")
}

func TestLinkGetMapID(t *testing.T) {
	l := cstesting.RandomLink()
	want := "hello"
	l.Meta.MapID = want
	got := l.Meta.MapID
	assert.EqualValues(t, want, got, "Invalid map id")
}

func TestLinkGetPrevLinkHash_notNil(t *testing.T) {
	l := cstesting.RandomLink()
	wantStr := "0123456789012345678901234567890123456789012345678901234567890123"
	l.Meta.PrevLinkHash = wantStr
	got := l.Meta.GetPrevLinkHash()
	want, _ := types.NewBytes32FromString(wantStr)
	assert.EqualValues(t, want, got, "Invalid PrevLinkHash")
	assert.EqualValues(t, wantStr, l.Meta.PrevLinkHash, "PrevLinkHash")
}

func TestLinkGetPrevLinkHash_nil(t *testing.T) {
	l := cstesting.RandomLink()
	l.Meta.PrevLinkHash = ""
	got := l.Meta.GetPrevLinkHash()
	assert.Nil(t, got, "PrevLinkHash")
	assert.EqualValues(t, "", l.Meta.PrevLinkHash, "Expected empty PrevLinkHash")
}

func TestLinkGetTags_notNil(t *testing.T) {
	l := cstesting.RandomLink()
	want := []string{"one", "two"}
	l.Meta.Tags = []string{"one", "two"}
	got := l.Meta.Tags
	assert.EqualValues(t, want, got, "Invalid tags")
}

func TestLinkGetTags_nil(t *testing.T) {
	l := cstesting.RandomLink()
	l.Meta.Tags = nil
	got := l.Meta.Tags
	assert.Nil(t, got, "Tags")
}

func TestLinkGetTagMap(t *testing.T) {
	l := cstesting.RandomLink()
	l.Meta.Tags = []string{"one", "two"}
	tags := l.Meta.GetTagMap()
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
	l.Meta.Process = want
	got := l.Meta.Process
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
