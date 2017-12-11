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
	"github.com/stretchr/testify/assert"
)

func strEqual(lhs, rhs string) bool {
	return lhs == rhs
}

func innerTestLinkValidate(t *testing.T, l *cs.Link, getSegment cs.GetSegmentFunc, want string, strComp func(lhs, rhs string) bool) {
	err := l.Validate(getSegment)
	assert.Error(t, err, "l.Validate() expected error")
	assert.True(t, strComp(err.Error(), want), "Unexpected error:\n%s\n", want, err.Error())
}

func testLinkValidateError(t *testing.T, l *cs.Link, getSegment cs.GetSegmentFunc, want string) {
	innerTestLinkValidate(t, l, getSegment, want, strEqual)
}

func testLinkValidateErrorWrapper(t *testing.T, l *cs.Link, getSegment cs.GetSegmentFunc, want string) {
	innerTestLinkValidate(t, l, getSegment, want, strings.Contains)
}

func appendRefSegment(l, ref *cs.Link) {
	var refs []interface{}
	var present bool
	if refs, present = l.Meta["refs"].([]interface{}); !present {
		refs = []interface{}{}
	}
	marshalledRef, _ := json.Marshal(ref.Segmentify())
	refs = append(refs, map[string]interface{}{"segment": string(marshalledRef)})
	l.Meta["refs"] = refs
}

func appendRefLink(l *cs.Link, process, linkHash string) {
	var refs []interface{}
	var present bool
	if refs, present = l.Meta["refs"].([]interface{}); !present {
		refs = []interface{}{}
	}
	refs = append(refs, map[string]interface{}{
		"process":  process,
		"linkHash": linkHash,
	})
	l.Meta["refs"] = refs
}
