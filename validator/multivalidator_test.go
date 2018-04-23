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
	"crypto/sha256"
	"fmt"
	"testing"

	"github.com/stratumn/go-indigocore/cs/cstesting"
	"github.com/stratumn/go-indigocore/testutil"
	"github.com/stretchr/testify/assert"
)

const validJSON = `
{
	"pki": {
	},
	"validators": {
	}
    }
`

func TestMultiValidator_New(t *testing.T) {
	mv := NewMultiValidator([]Validator{})
	assert.Len(t, mv.(*multiValidator).validators, 0)
}

func TestMultiValidator_Hash(t *testing.T) {
	t.Parallel()

	t.Run("Produces different hashes based on internal validators", func(t *testing.T) {
		baseConfig := &validatorBaseConfig{Process: "p"}
		testHash1 := testutil.RandomHash()
		testHash2 := testutil.RandomHash()

		type testCase struct {
			name string
			v1   Validator
			v2   Validator
			v3   Validator
		}

		testCases := []testCase{
			{
				name: "With schema validator",
				v1:   &schemaValidator{Config: baseConfig, SchemaHash: *testHash1},
				v2:   &schemaValidator{Config: baseConfig, SchemaHash: *testHash1},
				v3:   &schemaValidator{Config: baseConfig, SchemaHash: *testHash2},
			},
			{
				name: "With pki validator",
				v1:   &pkiValidator{Config: baseConfig, PKI: &PKI{"a": &Identity{}}},
				v2:   &pkiValidator{Config: baseConfig, PKI: &PKI{"a": &Identity{}}},
				v3:   &pkiValidator{Config: baseConfig, PKI: &PKI{"b": &Identity{}}},
			},
			{
				name: "With transition validator",
				v1:   &transitionValidator{Config: baseConfig, Transitions: []string{"one"}},
				v2:   &transitionValidator{Config: baseConfig, Transitions: []string{"one"}},
				v3:   &transitionValidator{Config: baseConfig, Transitions: []string{"two"}},
			},
			{
				name: "With script validator",
				v1:   &scriptValidator{Config: baseConfig, ScriptHash: *testHash1},
				v2:   &scriptValidator{Config: baseConfig, ScriptHash: *testHash1},
				v3:   &scriptValidator{Config: baseConfig, ScriptHash: *testHash2},
			},
		}

		for _, tt := range testCases {
			t.Run(tt.name, func(t *testing.T) {
				mv1 := NewMultiValidator([]Validator{tt.v1})

				h1, err := mv1.Hash()
				assert.NoError(t, err)
				assert.NotNil(t, h1)

				mv2 := NewMultiValidator([]Validator{tt.v2})

				h2, err := mv2.Hash()
				assert.NoError(t, err)
				assert.True(t, h1.Equals(h2))

				mv3 := NewMultiValidator([]Validator{tt.v3})

				h3, err := mv3.Hash()
				assert.NoError(t, err)
				assert.False(t, h1.Equals(h3))
			})
		}
	})

	t.Run("Uses child validators' Hash() function", func(t *testing.T) {
		baseConfig := &validatorBaseConfig{Process: "p"}
		schemaValidator := &schemaValidator{Config: baseConfig, SchemaHash: *testutil.RandomHash()}
		transitionValidator := newTransitionValidator(baseConfig, []string{"king"})
		pkiValidator := newPkiValidator(baseConfig, []string{"romeo"}, &PKI{})
		scriptValidator := &scriptValidator{Config: baseConfig, ScriptHash: *testutil.RandomHash()}

		mv := NewMultiValidator([]Validator{schemaValidator, transitionValidator, pkiValidator, scriptValidator})
		mvHash, err := mv.Hash()
		assert.NoError(t, err)

		b := make([]byte, 0)
		for _, validator := range []Validator{schemaValidator, transitionValidator, pkiValidator, scriptValidator} {
			validatorHash, err := validator.Hash()
			assert.NoError(t, err)
			b = append(b, validatorHash[:]...)
		}
		sum := sha256.Sum256(b)
		assert.True(t, mvHash.EqualsBytes(sum[:]))
	})
}

const testMessageSchema = `
{
	"type": "object",
	"properties": {
		"message": {
			"type": "string"
		}
	},
	"required": [
		"message"
	]
}`

func TestMultiValidator_Validate(t *testing.T) {
	t.Parallel()
	baseConfig1, _ := newValidatorBaseConfig("p", "a1")
	baseConfig2, _ := newValidatorBaseConfig("p", "a2")
	baseConfig3, _ := newValidatorBaseConfig("p", "a1")
	baseConfig4, _ := newValidatorBaseConfig("p", "a2")

	svCfg1, _ := newSchemaValidator(baseConfig1, []byte(testMessageSchema))
	svCfg2, _ := newSchemaValidator(baseConfig2, []byte(testMessageSchema))

	sigVCfg1 := newPkiValidator(baseConfig3, []string{"alice"}, &PKI{
		"alice": &Identity{
			Keys: []string{"TESTKEY1"},
		},
	})
	sigVCfg2 := newPkiValidator(baseConfig4, []string{}, &PKI{})

	mv := multiValidator{
		validators: []Validator{svCfg1, svCfg2, sigVCfg1, sigVCfg2},
	}

	testState := map[string]interface{}{"message": "test"}

	t.Run("Validate succeeds when all children succeed", func(t *testing.T) {
		l := cstesting.NewLinkBuilder().
			WithProcess("p").
			WithType("a1").
			WithState(testState).
			Sign().
			Build()
		l.Signatures[0].PublicKey = "TESTKEY1"

		err := mv.Validate(context.Background(), nil, l)
		assert.NoError(t, err)
	})

	t.Run("Validate fails if no validator matches the given segment", func(t *testing.T) {
		l := cstesting.NewLinkBuilder().
			WithType("nomatch").
			Build()
		process := l.Meta.Process

		err := mv.Validate(context.Background(), nil, l)
		assert.EqualError(t, err, fmt.Sprintf("Validation failed: link with process: [%s] and type: [nomatch] does not match any validator", process))
	})

	t.Run("Validate fails if one of the children fails (schema)", func(t *testing.T) {
		l := cstesting.NewLinkBuilder().
			WithProcess("p").
			WithType("a2").
			Build()

		err := mv.Validate(context.Background(), nil, l)
		assert.EqualError(t, err, "link validation failed: [message: message is required]")
	})

	t.Run("Validate fails if one of the children fails (pki)", func(t *testing.T) {
		l := cstesting.NewLinkBuilder().
			WithProcess("p").
			WithType("a1").
			WithState(testState).
			Sign().
			Build()

		err := mv.Validate(context.Background(), nil, l)
		assert.Error(t, err)
	})
}
