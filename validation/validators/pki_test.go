// Copyright 2016-2018 Stratumn SAS. All rights reserved.
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
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-chainscript/chainscripttest"
	"github.com/stratumn/go-core/validation/validationtesting"
	"github.com/stratumn/go-core/validation/validators"
	"github.com/stratumn/go-crypto/keys"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPKI(t *testing.T) {
	t.Run("Validate()", func(t *testing.T) {
		t.Run("empty PKI", func(t *testing.T) {
			var pki validators.PKI
			err := pki.Validate()
			assert.NoError(t, err)
		})

		t.Run("missing public key", func(t *testing.T) {
			pki := validators.PKI{
				"alice": &validators.Identity{
					Roles: []string{"bob's friend"},
				},
			}

			err := pki.Validate()
			assert.Equal(t, 0, strings.Index(err.Error(), validators.ErrMissingKeys.Error()))
		})

		t.Run("invalid public key", func(t *testing.T) {
			pki := validators.PKI{
				"alice": &validators.Identity{
					Keys:  []string{validationtesting.AlicePublicKey, "-----BEGIN QUANTUM PUBLIC KEY-----\nOR NOT"},
					Roles: []string{"crypto-nerd", "likes-bob"},
				},
			}

			err := pki.Validate()
			assert.Equal(t, 0, strings.Index(err.Error(), validators.ErrInvalidIdentity.Error()))
		})

		t.Run("valid pki", func(t *testing.T) {
			pki := validators.PKI{
				"alice": &validators.Identity{
					Keys: []string{validationtesting.AlicePublicKey},
				},
			}

			err := pki.Validate()
			assert.NoError(t, err)
		})
	})
}

func TestPKIValidator(t *testing.T) {
	process := "p1"
	linkStep := "test"
	psv, err := validators.NewProcessStepValidator(process, linkStep)
	require.NoError(t, err)

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

	pki := validators.PKI{
		"Alice Van den Budenmayer": &validators.Identity{
			Keys:  []string{string(link1.Signatures[0].PublicKey)},
			Roles: []string{"employee"},
		},
		"Bob Wagner": &validators.Identity{
			Keys:  []string{string(link2.Signatures[0].PublicKey)},
			Roles: []string{"manager", "it"},
		},
	}

	t.Run("Validate()", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			name               string
			link               *chainscript.Link
			requiredSignatures []string
			valid              bool
			err                error
		}{{
			name:               "valid-link",
			valid:              true,
			link:               chainscripttest.NewLinkBuilder(t).WithProcess(process).WithStep(linkStep).WithSignature(t, "").Build(),
			requiredSignatures: nil,
		}, {
			name:               "required-signature-pubkey",
			valid:              true,
			link:               link1,
			requiredSignatures: []string{string(link1.Signatures[0].PublicKey)},
		}, {
			name:               "required-signature-name",
			valid:              true,
			link:               link1,
			requiredSignatures: []string{"Alice Van den Budenmayer"},
		}, {
			name:               "required-signature-role",
			valid:              true,
			link:               link1,
			requiredSignatures: []string{"employee"},
		}, {
			name:               "required-signature-extra",
			valid:              true,
			link:               chainscripttest.NewLinkBuilder(t).From(t, link1).WithSignature(t, "").Build(),
			requiredSignatures: []string{"employee"},
		}, {
			name:               "required-signature-multi",
			valid:              true,
			link:               chainscripttest.NewLinkBuilder(t).From(t, link1).WithSignatureFromKey(t, priv2Bytes, "").Build(),
			requiredSignatures: []string{"employee", "it", "Bob Wagner"},
		}, {
			name:               "required-signature-fails",
			valid:              false,
			err:                validators.ErrMissingSignature,
			link:               chainscripttest.NewLinkBuilder(t).WithProcess(process).WithStep(linkStep).WithSignature(t, "").Build(),
			requiredSignatures: []string{"Alice Van den Budenmayer"},
		}}

		for _, tt := range testCases {
			t.Run(tt.name, func(t *testing.T) {
				sv := validators.NewPKIValidator(psv, tt.requiredSignatures, pki)
				err := sv.Validate(context.Background(), nil, tt.link)

				if tt.valid {
					assert.NoError(t, err)
				} else {
					assert.EqualError(t, errors.Cause(err), tt.err.Error())
				}
			})
		}
	})

	t.Run("Hash()", func(t *testing.T) {
		pki1 := validators.PKI{
			"Alice": &validators.Identity{
				Keys:  []string{string(link1.Signatures[0].PublicKey)},
				Roles: []string{"employee"},
			},
		}

		pki2 := validators.PKI{
			"Alice": &validators.Identity{
				Keys:  []string{string(link2.Signatures[0].PublicKey)},
				Roles: []string{"employee"},
			},
		}

		v1 := validators.NewPKIValidator(psv, []string{"a", "b"}, pki1)
		v2 := validators.NewPKIValidator(psv, []string{"a", "b"}, pki2)
		v3 := validators.NewPKIValidator(psv, []string{"c", "d"}, pki1)

		hash1, err := v1.Hash()
		require.NoError(t, err)
		assert.NotNil(t, hash1)

		hash2, err := v2.Hash()
		require.NoError(t, err)
		assert.NotNil(t, hash2)

		hash3, err := v3.Hash()
		require.NoError(t, err)
		assert.NotNil(t, hash3)

		assert.NotEqual(t, hash1, hash2)
		assert.NotEqual(t, hash1, hash3)
		assert.NotEqual(t, hash2, hash3)
	})
}
