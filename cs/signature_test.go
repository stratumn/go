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
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/cs/cstesting"
	sig "github.com/stratumn/go-indigocore/cs/signatures"
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
			err:   "Unhandled signature scheme [test]: " + sig.ErrUnsupportedSignatureType.Error(),
			signature: func() *cs.Signature {
				s := cstesting.RandomSignature(payload)
				s.Type = "test"
				return s
			},
		},
		{
			name:  "wrong-jmespath-query",
			valid: false,
			err:   "failed to execute JMESPATH query: SyntaxError: Incomplete expression",
			signature: func() *cs.Signature {
				s := cstesting.RandomSignature(payload)
				s.Payload = ""
				return s
			},
		},
		{
			name:  "empty-jmespath-query",
			valid: false,
			err:   "JMESPATH query does not match any link data",
			signature: func() *cs.Signature {
				s := cstesting.RandomSignature(payload)
				s.Payload = "notfound"
				return s
			},
		},
		{
			name:  "wrong-signature",
			valid: false,
			err:   sig.ErrInvalidSignature.Error(),
			signature: func() *cs.Signature {
				s := cstesting.RandomSignature(payload)
				s.Signature = "test"
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
