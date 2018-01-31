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
	process := "p1"
	action := "test"
	cfg, err := newSignatureValidatorConfig(process, action)
	require.NoError(t, err)
	sv := newSignatureValidator(cfg)

	createValidLink := func() *cs.Link {
		l := cstesting.RandomLink()
		l.Meta["process"] = process
		l.Meta["action"] = action
		return cstesting.SignLink(l)
	}

	type testCase struct {
		name  string
		link  func() *cs.Link
		valid bool
		err   string
	}

	testCases := []testCase{
		{
			name:  "valid-link",
			valid: true,
			link:  createValidLink,
		},
		{
			name:  "process-not-matched",
			valid: true,
			link: func() *cs.Link {
				l := createValidLink()
				l.Meta["process"] = "p2"
				return l
			},
		},
		{
			name:  "type-not-matched",
			valid: true,
			link: func() *cs.Link {
				l := createValidLink()
				l.Meta["action"] = "buy"
				return l
			},
		},
		{
			name:  "empty-signatures",
			valid: false,
			err:   ErrMissingSignature.Error(),
			link: func() *cs.Link {
				l := createValidLink()
				l.Signatures = nil
				return l
			},
		},
		{
			name:  "unsupported-signature-type",
			valid: false,
			err:   ErrUnsupportedSignatureType.Error(),
			link: func() *cs.Link {
				l := createValidLink()
				l.Signatures[0].Type = "test"
				return l
			},
		},
		{
			name:  "wrong-jmespath-query",
			valid: false,
			err:   "failed to execute jmespath query: SyntaxError: Incomplete expression",
			link: func() *cs.Link {
				l := createValidLink()
				l.Signatures[0].Payload = ""
				return l
			},
		},
		{
			name:  "empty-jmespath-query",
			valid: false,
			err:   ErrEmptyPayload.Error(),
			link: func() *cs.Link {
				l := createValidLink()
				l.Signatures[0].Payload = "notfound"
				return l
			},
		},
		{
			name:  "bad-public-key-length",
			valid: false,
			err:   "Ed25519 public key length must be 32, got 3",
			link: func() *cs.Link {
				l := createValidLink()
				l.Signatures[0].PublicKey = "test"
				return l
			},
		},
		{
			name:  "wrong-signature",
			valid: false,
			err:   "signature verification failed",
			link: func() *cs.Link {
				l := createValidLink()
				l.Signatures[0].Signature = "test"
				return l
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			err := sv.Validate(nil, tt.link())
			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.err)
			}
		})
	}

}
