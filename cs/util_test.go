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
	"encoding/json"
	"strings"
	"testing"

	"github.com/stratumn/sdk/cs"
)

func strEqual(lhs, rhs string) bool {
	return lhs == rhs
}

func innerTestSegmentValidate(t *testing.T, s *cs.Segment, getSegment cs.GetSegmentFunc, want string, strComp func(lhs, rhs string) bool) {
	s.SetLinkHash()
	if err := s.Validate(getSegment); err == nil {
		t.Error("s.Validate() = nil want Error")
	} else if got := err.Error(); !strComp(got, want) {
		t.Errorf("s.Validate() = %q want %q", got, want)
	}
}

func testSegmentValidateError(t *testing.T, s *cs.Segment, getSegment cs.GetSegmentFunc, want string) {
	innerTestSegmentValidate(t, s, getSegment, want, strEqual)
}

func testSegmentValidateErrorWrapper(t *testing.T, s *cs.Segment, getSegment cs.GetSegmentFunc, want string) {
	innerTestSegmentValidate(t, s, getSegment, want, strings.Contains)
}

func appendRefSegment(s, ref *cs.Segment) {
	var refs []interface{}
	var present bool
	if refs, present = s.Link.Meta["refs"].([]interface{}); !present {
		refs = []interface{}{}
	}
	marshalledRef, _ := json.Marshal(ref)
	refs = append(refs, map[string]interface{}{"segment": string(marshalledRef)})
	s.Link.Meta["refs"] = refs
}

func appendRefLink(s *cs.Segment, process, linkHash string) {
	var refs []interface{}
	var present bool
	if refs, present = s.Link.Meta["refs"].([]interface{}); !present {
		refs = []interface{}{}
	}
	refs = append(refs, map[string]interface{}{
		"process":  process,
		"linkHash": linkHash,
	})
	s.Link.Meta["refs"] = refs
}
