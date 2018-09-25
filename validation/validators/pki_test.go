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

package validators_test

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"testing"

	"github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-chainscript/chainscripttest"
	"github.com/stratumn/go-crypto/keys"
	"github.com/stratumn/go-core/validation/validators"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ed25519"
)

func TestPKIValidator(t *testing.T) {
	t.Parallel()

	process := "p1"
	linkStep := "test"
	_, priv1, _ := keys.NewEd25519KeyPair()
	priv1Bytes, _ := keys.EncodeSecretkey(priv1)
	_, priv2, _ := keys.NewEd25519KeyPair()
	priv2Bytes, _ := keys.EncodeSecretkey(priv2)

	link1 := chainscripttest.NewLinkBuilder(t).
		WithProcess(process).
		WithStep(linkStep).
		WithSignatureFromKey(t, priv1Bytes, "").
		Build()
	link2 := chainscripttest.NewLinkBuilder(t).
		WithProcess(process).
		WithStep(linkStep).
		WithSignatureFromKey(t, priv2Bytes, "").
		Build()

	pki := &validators.PKI{
		"Alice Van den Budenmayer": &validators.Identity{
			Keys:  []string{string(link1.Signatures[0].PublicKey)},
			Roles: []string{"employee"},
		},
		"Bob Wagner": &validators.Identity{
			Keys:  []string{string(link2.Signatures[0].PublicKey)},
			Roles: []string{"manager", "it"},
		},
	}

	type testCase struct {
		name               string
		link               *chainscript.Link
		valid              bool
		err                string
		requiredSignatures []string
	}

	testCases := []testCase{
		{
			name:  "valid-link",
			valid: true,
			link: chainscripttest.NewLinkBuilder(t).
				WithProcess(process).
				WithStep(linkStep).
				WithSignature(t, "").
				Build(),
		},
		{
			name:               "required-signature-pubkey",
			valid:              true,
			link:               link1,
			requiredSignatures: []string{string(link1.Signatures[0].PublicKey)},
		},
		{
			name:               "required-signature-name",
			valid:              true,
			link:               link1,
			requiredSignatures: []string{"Alice Van den Budenmayer"},
		},
		{
			name:               "required-signature-role",
			valid:              true,
			link:               link1,
			requiredSignatures: []string{"employee"},
		},
		{
			name:               "required-signature-extra",
			valid:              true,
			link:               chainscripttest.NewLinkBuilder(t).From(t, link1).WithSignature(t, "").Build(),
			requiredSignatures: []string{"employee"},
		},
		{
			name:               "required-signature-multi",
			valid:              true,
			link:               chainscripttest.NewLinkBuilder(t).From(t, link1).WithSignatureFromKey(t, priv2Bytes, "").Build(),
			requiredSignatures: []string{"employee", "it", "Bob Wagner"},
		},
		{
			name:               "required-signature-fails",
			valid:              false,
			err:                "Missing signatory for validator test of process p1: signature from Alice Van den Budenmayer is required",
			link:               chainscripttest.NewLinkBuilder(t).WithProcess(process).WithStep(linkStep).WithSignature(t, "").Build(),
			requiredSignatures: []string{"Alice Van den Budenmayer"},
		},
	}

	for _, tt := range testCases {
		baseCfg, err := validators.NewValidatorBaseConfig(process, linkStep)
		require.NoError(t, err)
		sv := validators.NewPKIValidator(baseCfg, tt.requiredSignatures, pki)

		t.Run(tt.name, func(t *testing.T) {
			err := sv.Validate(context.Background(), nil, tt.link)
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

	pki1 := &validators.PKI{
		"Alice": &validators.Identity{
			Keys:  []string{pub1},
			Roles: []string{"employee"},
		},
	}
	pki2 := &validators.PKI{
		"Bob": &validators.Identity{
			Keys:  []string{pub2},
			Roles: []string{"manager", "it"},
		},
	}

	baseCfg, err := validators.NewValidatorBaseConfig("foo", "bar")
	require.NoError(t, err)
	v1 := validators.NewPKIValidator(baseCfg, []string{"a", "b"}, pki1)
	v2 := validators.NewPKIValidator(baseCfg, []string{"a", "b"}, pki2)
	v3 := validators.NewPKIValidator(baseCfg, []string{"c", "d"}, pki1)

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
