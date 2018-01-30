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

package validator

import (
	"testing"

	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/cs/cstesting"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignatureValidator(t *testing.T) {
	cfg, err := newSignatureValidatorConfig("p1", "test")
	require.NoError(t, err)
	sv := newSignatureValidator(cfg)

	createValidLink := func() *cs.Link {
		l := cstesting.RandomLink()
		l.Meta["process"] = "p1"
		l.Meta["action"] = "test"
		l.Signatures = append(l.Signatures, &cs.Signature{})
		return l
	}

	createInvalidLink := func() *cs.Link {
		l := createValidLink()
		l.Signatures = nil
		return l
	}

	type testCase struct {
		name  string
		link  func() *cs.Link
		valid bool
	}

	testCases := []testCase{{
		name:  "process-not-matched",
		valid: true,
		link: func() *cs.Link {
			l := createInvalidLink()
			l.Meta["process"] = "p2"
			return l
		},
	}, {
		name:  "type-not-matched",
		valid: true,
		link: func() *cs.Link {
			l := createInvalidLink()
			l.Meta["action"] = "buy"
			return l
		},
	}, {
		name:  "valid-link",
		valid: true,
		link:  createValidLink,
	}, {
		name:  "invalid-link",
		valid: false,
		link:  createInvalidLink,
	}}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			err := sv.Validate(nil, tt.link())
			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, ErrMissingSignature.Error())
			}
		})
	}

}
