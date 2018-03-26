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
	"context"
	"crypto/rand"
	"encoding/base64"
	"testing"

	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/cs/cstesting"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ed25519"
)

func TestPKIValidator(t *testing.T) {
	t.Parallel()
	process := "p1"
	linkType := "test"

	createValidLink := func() *cs.Link {
		l := cstesting.RandomLink()
		l.Meta.Process = process
		l.Meta.Type = linkType
		return cstesting.SignLink(l)
	}

	createValidLinkWithKey := func(priv ed25519.PrivateKey) *cs.Link {
		l := cstesting.RandomLink()
		l.Meta.Process = process
		l.Meta.Type = linkType
		return cstesting.SignLinkWithKey(l, priv)
	}

	_, priv1, _ := ed25519.GenerateKey(rand.Reader)
	_, priv2, _ := ed25519.GenerateKey(rand.Reader)
	link1 := createValidLinkWithKey(priv1)
	link2 := createValidLinkWithKey(priv2)

	pki := &PKI{
		"Alice Van den Budenmayer": &Identity{
			Keys:  []string{link1.Signatures[0].PublicKey},
			Roles: []string{"employee"},
		},
		"Bob Wagner": &Identity{
			Keys:  []string{link2.Signatures[0].PublicKey},
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
			requiredSignatures: []string{"Alice Van den Budenmayer"},
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
			requiredSignatures: []string{"employee", "it", "Bob Wagner"},
		},
		{
			name:               "required-signature-fails",
			valid:              false,
			err:                "Missing signatory for validator test of process p1: signature from Alice Van den Budenmayer is required",
			link:               createValidLink,
			requiredSignatures: []string{"Alice Van den Budenmayer"},
		},
	}

	for _, tt := range testCases {
		baseCfg, err := newValidatorBaseConfig(process, linkType)
		require.NoError(t, err)
		sv := newPkiValidator(baseCfg, tt.requiredSignatures, pki)

		t.Run(tt.name, func(t *testing.T) {
			err := sv.Validate(context.Background(), nil, tt.link())
			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.err)
			}
		})
	}

}

func TestPKIHash(t *testing.T) {
	t.Parallel()

	_, priv1, _ := ed25519.GenerateKey(rand.Reader)
	_, priv2, _ := ed25519.GenerateKey(rand.Reader)
	pub1 := base64.StdEncoding.EncodeToString(priv1.Public().(ed25519.PublicKey))
	pub2 := base64.StdEncoding.EncodeToString(priv2.Public().(ed25519.PublicKey))

	pki1 := &PKI{
		"Alice": &Identity{
			Keys:  []string{pub1},
			Roles: []string{"employee"},
		},
	}
	pki2 := &PKI{
		"Bob": &Identity{
			Keys:  []string{pub2},
			Roles: []string{"manager", "it"},
		},
	}

	baseCfg, err := newValidatorBaseConfig("foo", "bar")
	require.NoError(t, err)
	v1 := newPkiValidator(baseCfg, []string{"a", "b"}, pki1)
	v2 := newPkiValidator(baseCfg, []string{"a", "b"}, pki2)
	v3 := newPkiValidator(baseCfg, []string{"c", "d"}, pki1)

	hash1, err1 := v1.Hash()
	hash2, err2 := v2.Hash()
	hash3, err3 := v3.Hash()
	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NoError(t, err3)
	assert.NotNil(t, hash1)
	assert.NotNil(t, hash2)
	assert.NotNil(t, hash3)
	assert.NotEqual(t, hash1.String(), hash2.String())
	assert.NotEqual(t, hash1.String(), hash3.String())
	assert.NotEqual(t, hash2.String(), hash3.String())
}
