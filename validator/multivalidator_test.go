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
)

func TestMultiValidator_New(t *testing.T) {
	mv := NewMultiValidator(&MultiValidatorConfig{
		SchemaConfigs: []*schemaValidatorConfig{
			&schemaValidatorConfig{},
			&schemaValidatorConfig{},
		},
		SignatureConfigs: []*signatureValidatorConfig{
			&signatureValidatorConfig{},
		},
	})

	assert.Len(t, mv.(*multiValidator).validators, 3)
}

func TestMultiValidator_Hash(t *testing.T) {
	baseConfig1 := &validatorBaseConfig{Process: "p"}
	baseConfig2 := &validatorBaseConfig{Process: "p2"}

	t.Run("With schema validator", func(t *testing.T) {
		mv1 := NewMultiValidator(&MultiValidatorConfig{
			SchemaConfigs: []*schemaValidatorConfig{
				&schemaValidatorConfig{
					validatorBaseConfig: baseConfig1,
				}},
		})

		h1, err := mv1.Hash()
		assert.NoError(t, err)
		assert.NotNil(t, h1)

		mv2 := NewMultiValidator(&MultiValidatorConfig{
			SchemaConfigs: []*schemaValidatorConfig{
				&schemaValidatorConfig{
					validatorBaseConfig: baseConfig1,
				}},
		})

		h2, err := mv2.Hash()
		assert.NoError(t, err)
		assert.EqualValues(t, h1, h2)

		mv3 := NewMultiValidator(&MultiValidatorConfig{
			SchemaConfigs: []*schemaValidatorConfig{
				&schemaValidatorConfig{
					validatorBaseConfig: baseConfig2,
				}},
		})

		h3, err := mv3.Hash()
		assert.NoError(t, err)
		assert.False(t, h1.Equals(h3))
	})

	t.Run("With signature validator", func(t *testing.T) {
		mv1 := NewMultiValidator(&MultiValidatorConfig{
			SignatureConfigs: []*signatureValidatorConfig{
				&signatureValidatorConfig{
					validatorBaseConfig: baseConfig1,
				}},
		})

		h1, err := mv1.Hash()
		assert.NoError(t, err)
		assert.NotNil(t, h1)

		mv2 := NewMultiValidator(&MultiValidatorConfig{
			SignatureConfigs: []*signatureValidatorConfig{
				&signatureValidatorConfig{
					validatorBaseConfig: baseConfig1,
				}},
		})

		h2, err := mv2.Hash()
		assert.NoError(t, err)
		assert.EqualValues(t, h1, h2)

		mv3 := NewMultiValidator(&MultiValidatorConfig{
			SignatureConfigs: []*signatureValidatorConfig{
				&signatureValidatorConfig{
					validatorBaseConfig: baseConfig2,
				}},
		})

		h3, err := mv3.Hash()
		assert.NoError(t, err)
		assert.False(t, h1.Equals(h3))
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
	svCfg1, _ := newSchemaValidatorConfig("p", "a1", []byte(testMessageSchema))
	svCfg2, _ := newSchemaValidatorConfig("p", "a2", []byte(testMessageSchema))

	sigVCfg1, _ := newSignatureValidatorConfig("p", "a1")
	sigVCfg2, _ := newSignatureValidatorConfig("p", "a2")

	mv := NewMultiValidator(&MultiValidatorConfig{
		SchemaConfigs:    []*schemaValidatorConfig{svCfg1, svCfg2},
		SignatureConfigs: []*signatureValidatorConfig{sigVCfg1, sigVCfg2},
	})

	t.Run("Validate succeeds when all children succeed", func(t *testing.T) {
		l := cstesting.RandomLink()
		l.Signatures = append(l.Signatures, &cs.Signature{})
		err := mv.Validate(nil, l)
		assert.NoError(t, err)
	})

	t.Run("Validate fails if one of the children fails (schema)", func(t *testing.T) {
		l := cstesting.RandomLink()
		l.Meta["process"] = "p"
		l.Meta["action"] = "a2"

		err := mv.Validate(nil, l)
		assert.EqualError(t, err, "link validation failed: [message: message is required]")
	})

	t.Run("Validate fails if one of the children fails (signature)", func(t *testing.T) {
		l := cstesting.RandomLink()
		l.Meta["process"] = "p"
		l.Meta["action"] = "a2"
		l.State["message"] = "test"

		err := mv.Validate(nil, l)
		assert.EqualError(t, err, ErrMissingSignature.Error())
	})
}
