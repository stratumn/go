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
	"crypto/x509"
	"testing"

	"github.com/stratumn/go-crypto/signatures"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/stratumn/go-crypto/encoding"
	"github.com/stratumn/go-crypto/keys"
	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/cs/cstesting"
)

func TestGetSignatures(t *testing.T) {
	signatures := cs.Signatures{
		&cs.Signature{
			PublicKey: "one",
		},
		&cs.Signature{
			PublicKey: "two",
		},
	}
	got := signatures.Get("one")
	assert.EqualValues(t, signatures[0], got, "signatures.Get()")
}

func TestGetSignatures_NotFound(t *testing.T) {
	signatures := cs.Signatures{
		&cs.Signature{
			PublicKey: "one",
		},
		&cs.Signature{
			PublicKey: "two",
		},
	}
	got := signatures.Get("wrong")
	assert.Nil(t, got, "s.Get(()")
}

func TestNewSignature(t *testing.T) {
	_, privPEM, err := keys.GenerateKey(keys.ED25519)
	require.NoError(t, err)
	link := cstesting.RandomLink()

	t.Run("Valid signature", func(t *testing.T) {
		payloadPath := "[state,meta]"
		sig, err := cs.NewSignature(payloadPath, privPEM, link)
		require.NoError(t, err)
		assert.NoError(t, sig.Verify(link), "signature verification failed")
	})

	t.Run("Bad payload", func(t *testing.T) {
		payloadPath := ""
		_, err := cs.NewSignature(payloadPath, privPEM, link)
		assert.EqualError(t, err, cs.ErrBadJMESPATHQuery+": SyntaxError: Incomplete expression")
	})

	t.Run("Canonicaljson failed", func(t *testing.T) {
		payloadPath := "[state,meta]"
		link.State["lol"] = func() {}
		_, err := cs.NewSignature(payloadPath, privPEM, link)
		assert.EqualError(t, err, "canonicaljson: unsupported type: func()")
	})

}

func TestSignatureValidator(t *testing.T) {
	payload := cstesting.RandomLink()
	type testCase struct {
		name               string
		signature          func() *cs.Signature
		valid              bool
		err                string
		requiredSignatures []string
	}

	testCases := []testCase{
		{
			name:      "valid-link",
			valid:     true,
			signature: func() *cs.Signature { return cstesting.RandomSignature(payload) },
		},
		{
			name:  "unsupported-signature-type",
			valid: false,
			err:   x509.ErrUnsupportedAlgorithm.Error(),
			signature: func() *cs.Signature {
				s := cstesting.RandomSignature(payload)
				s.Type = "test"
				return s
			},
		},
		{
			name:  "wrong-jmespath-query",
			valid: false,
			err:   cs.ErrBadJMESPATHQuery + ": SyntaxError: Incomplete expression",
			signature: func() *cs.Signature {
				s := cstesting.RandomSignature(payload)
				s.Payload = ""
				return s
			},
		},
		{
			name:  "empty-jmespath-query",
			valid: false,
			err:   cs.ErrEmptyJMESPATHResult,
			signature: func() *cs.Signature {
				s := cstesting.RandomSignature(payload)
				s.Payload = "notfound"
				return s
			},
		},
		{
			name:  "wrong-signature",
			valid: false,
			err:   "invalid ed25519 signature: signature verification failed",
			signature: func() *cs.Signature {
				s := cstesting.RandomSignature(payload)
				wrongSigPEM, _ := encoding.EncodePEM([]byte("test"), signatures.SignaturePEMLabel)
				s.Signature = string(wrongSigPEM)
				return s
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.signature().Verify(payload)
			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.err)
			}
		})
	}

}
