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
	"crypto/rand"
	"testing"

	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/cs/cstesting"
	"github.com/stratumn/sdk/validator/signature"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ed25519"
)

func TestSignatureValidator(t *testing.T) {
	process := "p1"
	action := "test"

	createValidLink := func() *cs.Link {
		l := cstesting.RandomLink()
		l.Meta["process"] = process
		l.Meta["action"] = action
		return cstesting.SignLink(l)
	}

	createValidLinkWithKey := func(priv ed25519.PrivateKey) *cs.Link {
		l := cstesting.RandomLink()
		l.Meta["process"] = process
		l.Meta["action"] = action
		return cstesting.SignLinkWithKey(l, priv)
	}

	_, priv1, _ := ed25519.GenerateKey(rand.Reader)
	_, priv2, _ := ed25519.GenerateKey(rand.Reader)
	link1 := createValidLinkWithKey(priv1)
	link2 := createValidLinkWithKey(priv2)

	pki := &PKI{
		link1.Signatures[0].PublicKey: &Identity{
			Name:  "Alice Van den Budenmayer",
			Roles: []string{"employee"},
		},
		link2.Signatures[0].PublicKey: &Identity{
			Name:  "Bob Wagner",
			Roles: []string{"manager", "it"},
		},
	}

	type testCase struct {
		name               string
		link               func() *cs.Link
		valid              bool
		err                string
		requiredSignatures []string
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
			err:   "Unhandled signature scheme [test]: " + signature.ErrUnsupportedSignatureType.Error(),
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
			name:  "wrong-signature",
			valid: false,
			err:   signature.ErrInvalidSignature.Error(),
			link: func() *cs.Link {
				l := createValidLink()
				l.Signatures[0].Signature = "test"
				return l
			},
		},
		{
			name:  "required-signature-pubkey",
			valid: true,
			link: func() *cs.Link {
				return link1
			},
			requiredSignatures: []string{link1.Signatures[0].PublicKey},
		},
		{
			name:  "required-signature-name",
			valid: true,
			link: func() *cs.Link {
				return link1
			},
			requiredSignatures: []string{"alice van den budenmayer"},
		},
		{
			name:  "required-signature-role",
			valid: true,
			link: func() *cs.Link {
				return link1
			},
			requiredSignatures: []string{"employee"},
		},
		{
			name:  "required-signature-extra",
			valid: true,
			link: func() *cs.Link {
				tmpLink := *link1
				return cstesting.SignLink(&tmpLink)
			},
			requiredSignatures: []string{"employee"},
		},
		{
			name:  "required-signature-multi",
			valid: true,
			link: func() *cs.Link {
				tmpLink := *link1
				return cstesting.SignLinkWithKey(&tmpLink, priv2)
			},
			requiredSignatures: []string{"employee", "it", "bob wagner"},
		},
		{
			name:               "required-signature-fails",
			valid:              false,
			err:                "Missing signatory for validator required-signature-fails: A signature from alice van den budenmayer is required",
			link:               createValidLink,
			requiredSignatures: []string{"alice van den budenmayer"},
		},
	}

	for _, tt := range testCases {
		cfg, err := newSignatureValidatorConfig(process, tt.name, action, tt.requiredSignatures, pki)
		require.NoError(t, err)
		sv := newSignatureValidator(cfg)

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
