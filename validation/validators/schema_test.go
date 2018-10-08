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
	"testing"

	"github.com/stratumn/go-chainscript/chainscripttest"
	"github.com/stratumn/go-core/validation/validators"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSchemaValidator(t *testing.T) {
	testSchema := []byte(`
	{
		"type": "object",
		"properties": {
			"seller": {
				"type": "string"
			},
			"lot": {
				"type": "string"
			},
			"initialPrice": {
				"type": "integer",
				"minimum": 0
			}
		},
		"required": [
			"seller",
			"lot",
			"initialPrice"
		]
	}`)

	process := "p1"
	step := "sell"
	psv, err := validators.NewProcessStepValidator(process, step)
	require.NoError(t, err)

	t.Run("New()", func(t *testing.T) {
		testCases := []struct {
			name   string
			schema []byte
			valid  bool
		}{{
			name:   "invalid-schema",
			schema: []byte(`{"type": "object", "properties": {"malformed}}`),
			valid:  false,
		}, {
			name:   "valid-schema",
			schema: testSchema,
			valid:  true,
		}}

		for _, tt := range testCases {
			t.Run(tt.name, func(t *testing.T) {
				sv, err := validators.NewSchemaValidator(psv, tt.schema)
				if tt.valid {
					assert.NotNil(t, sv)
					assert.NoError(t, err)
				} else {
					assert.Nil(t, sv)
					assert.Error(t, err)
				}
			})
		}
	})

	t.Run("Validate()", func(t *testing.T) {
		sv, err := validators.NewSchemaValidator(psv, testSchema)
		require.NoError(t, err)

		testCases := []struct {
			name  string
			data  map[string]interface{}
			valid bool
		}{{
			name: "valid-data",
			data: map[string]interface{}{
				"seller":       "Alice",
				"lot":          "Secret key",
				"initialPrice": 42,
			},
			valid: true,
		}, {
			name: "missing-seller",
			data: map[string]interface{}{
				"lot":          "Secret key",
				"initialPrice": 42,
			},
			valid: false,
		}, {
			name: "invalid-integer-constraint",
			data: map[string]interface{}{
				"seller":       "Alice",
				"lot":          "Secret key",
				"initialPrice": -10,
			},
			valid: false,
		}, {
			name: "invalid-field-type",
			data: map[string]interface{}{
				"seller":       "Alice",
				"lot":          10,
				"initialPrice": 42,
			},
			valid: false,
		}}

		for _, tt := range testCases {
			t.Run(tt.name, func(t *testing.T) {
				link := chainscripttest.NewLinkBuilder(t).
					WithProcess(process).
					WithStep(step).
					WithData(t, tt.data).
					Build()

				err := sv.Validate(context.Background(), nil, link)
				if tt.valid {
					assert.NoError(t, err)
				} else {
					assert.Error(t, err)
				}
			})
		}
	})

	t.Run("Hash()", func(t *testing.T) {
		v1, err := validators.NewSchemaValidator(psv, testSchema)
		require.NoError(t, err)

		v2, err := validators.NewSchemaValidator(psv, []byte(`{"type": "object","properties": {"seller": {"type": "string"}}, "required": ["seller"]}`))
		require.NoError(t, err)

		psv2, err := validators.NewProcessStepValidator("test_process", "test_step")
		require.NoError(t, err)

		v3, err := validators.NewSchemaValidator(psv2, testSchema)
		require.NoError(t, err)

		h1, err := v1.Hash()
		require.NoError(t, err)

		h2, err := v2.Hash()
		require.NoError(t, err)

		h3, err := v3.Hash()
		require.NoError(t, err)

		assert.NotEqual(t, h1, h2)
		assert.NotEqual(t, h1, h3)
		assert.NotEqual(t, h2, h3)
	})
}
