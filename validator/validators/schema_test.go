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
	"testing"

	"github.com/stratumn/go-indigocore/cs/cstesting"
	"github.com/stratumn/go-indigocore/validator/validators"

	"github.com/stratumn/go-indigocore/cs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSellSchema = `
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
}`

func TestSchemaValidatorConfig(t *testing.T) {
	t.Parallel()
	validSchema := []byte(testSellSchema)
	process := "p1"
	linkType := "sell"

	type testCase struct {
		name          string
		process       string
		linkType      string
		schema        []byte
		valid         bool
		expectedError error
	}

	testCases := []testCase{{
		name:     "invalid-schema",
		process:  process,
		linkType: linkType,
		schema:   []byte(`{"type": "object", "properties": {"malformed}}`),
		valid:    false,
	}, {
		name:     "valid-config",
		process:  process,
		linkType: linkType,
		schema:   validSchema,
		valid:    true,
	}}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			baseCfg, _ := validators.NewValidatorBaseConfig(process, tt.linkType)
			sv, err := validators.NewSchemaValidator(baseCfg, tt.schema)

			if tt.valid {
				assert.NotNil(t, sv)
				assert.NoError(t, err)
			} else {
				assert.Nil(t, sv)
				assert.Error(t, err)
				if tt.expectedError != nil {
					assert.EqualError(t, err, tt.expectedError.Error())
				}

			}
		})
	}
}

func TestSchemaValidator(t *testing.T) {
	t.Parallel()
	schema := []byte(testSellSchema)
	baseCfg, err := validators.NewValidatorBaseConfig("p1", "sell")
	require.NoError(t, err)
	sv, err := validators.NewSchemaValidator(baseCfg, schema)
	require.NoError(t, err)

	rightState := map[string]interface{}{
		"seller":       "Alice",
		"lot":          "Secret key",
		"initialPrice": 42,
	}

	badState := map[string]interface{}{
		"lot":          "Secret key",
		"initialPrice": 42,
	}

	type testCase struct {
		name  string
		link  *cs.Link
		valid bool
	}

	testCases := []testCase{{
		name:  "valid-link",
		valid: true,
		link:  cstesting.NewLinkBuilder().WithProcess("p1").WithType("sell").WithState(rightState).Build(),
	}, {
		name:  "invalid-link",
		valid: false,
		link:  cstesting.NewLinkBuilder().WithProcess("p1").WithType("sell").WithState(badState).Build(),
	}}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			err := sv.Validate(context.Background(), nil, tt.link)
			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestSchemaHash(t *testing.T) {
	t.Parallel()
	baseCfg, err := validators.NewValidatorBaseConfig("foo", "bar")
	require.NoError(t, err)
	v1, err1 := validators.NewSchemaValidator(baseCfg, []byte(testSellSchema))
	v2, err2 := validators.NewSchemaValidator(baseCfg, []byte(`{"type": "object","properties": {"seller": {"type": "string"}}, "required": ["seller"]}`))

	hash1, err1 := v1.Hash()
	hash2, err2 := v2.Hash()
	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NotNil(t, hash1)
	assert.NotNil(t, hash2)
	assert.NotEqual(t, hash1.String(), hash2.String())
}
