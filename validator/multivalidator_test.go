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
	"fmt"
	"testing"

	"github.com/stratumn/go-indigocore/cs/cstesting"
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
	baseConfig1 := &validatorBaseConfig{Process: "p"}
	baseConfig2 := &validatorBaseConfig{Process: "p2"}

	type testCase struct {
		name string
		v1   Validator
		v2   Validator
		v3   Validator
	}

	testCases := []testCase{
		{
			name: "With schema validator",
			v1:   &schemaValidator{Config: baseConfig1},
			v2:   &schemaValidator{Config: baseConfig1},
			v3:   &schemaValidator{Config: baseConfig2},
		},
		{
			name: "With pki validator",
			v1:   &pkiValidator{Config: baseConfig1},
			v2:   &pkiValidator{Config: baseConfig1},
			v3:   &pkiValidator{Config: baseConfig2},
		},
		{
			name: "With transition validator",
			v1:   &transitionValidator{Config: baseConfig1},
			v2:   &transitionValidator{Config: baseConfig1},
			v3:   &transitionValidator{Config: baseConfig2},
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
			assert.EqualValues(t, h1, h2)

			mv3 := NewMultiValidator([]Validator{tt.v3})

			h3, err := mv3.Hash()
			assert.NoError(t, err)
			assert.False(t, h1.Equals(h3))
		})
	}
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

	t.Run("Validate succeeds when all children succeed", func(t *testing.T) {
		l := cstesting.SignLink(cstesting.RandomLink())
		l.Meta.Process = "p"
		l.Meta.Type = "a1"
		l.State["message"] = "test"
		l.Signatures[0].PublicKey = "TESTKEY1"

		err := mv.Validate(context.Background(), nil, l)
		assert.NoError(t, err)
	})

	t.Run("Validate fails if no validator matches the given segment", func(t *testing.T) {
		l := cstesting.RandomLink()
		l.Meta.Type = "nomatch"

		process := l.Meta.Process

		err := mv.Validate(context.Background(), nil, l)
		assert.EqualError(t, err, fmt.Sprintf("Validation failed: link with process: [%s] and type: [nomatch] does not match any validator", process))
	})

	t.Run("Validate fails if one of the children fails (schema)", func(t *testing.T) {
		l := cstesting.RandomLink()
		l.Meta.Process = "p"
		l.Meta.Type = "a2"

		err := mv.Validate(context.Background(), nil, l)
		assert.EqualError(t, err, "link validation failed: [message: message is required]")
	})

	t.Run("Validate fails if one of the children fails (pki)", func(t *testing.T) {
		l := cstesting.SignLink(cstesting.RandomLink())
		l.Meta.Process = "p"
		l.Meta.Type = "a1"
		l.State["message"] = "test"

		err := mv.Validate(context.Background(), nil, l)
		assert.Error(t, err)
	})
}
