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

	"github.com/stratumn/sdk/cs/cstesting"

	"github.com/stratumn/sdk/cs"
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
			cfg, err := newSchemaValidatorConfig(
				tt.process,
				tt.linkType,
				tt.schema,
			)

			if tt.valid {
				assert.NotNil(t, cfg)
				assert.NoError(t, err)
			} else {
				assert.Nil(t, cfg)
				assert.Error(t, err)
				if tt.expectedError != nil {
					assert.EqualError(t, err, tt.expectedError.Error())
				}

			}
		})
	}
}

func TestSchemaValidator(t *testing.T) {
	schema := []byte(testSellSchema)
	cfg, err := newSchemaValidatorConfig("p1", "sell", schema)
	require.NoError(t, err)
	sv := newSchemaValidator(cfg)

	createValidLink := func() *cs.Link {
		l := cstesting.RandomLink()
		l.Meta["process"] = "p1"
		l.Meta["action"] = "sell"
		l.State["seller"] = "Alice"
		l.State["lot"] = "Secret key"
		l.State["initialPrice"] = 42
		return l
	}

	createInvalidLink := func() *cs.Link {
		l := createValidLink()
		delete(l.State, "seller")
		return l
	}

	type testCase struct {
		name  string
		link  func() *cs.Link
		valid bool
	}

	testCases := []testCase{{
		name:  "process-not-matched",
		valid: true,
		link: func() *cs.Link {
			l := createInvalidLink()
			l.Meta["process"] = "p2"
			return l
		},
	}, {
		name:  "type-not-matched",
		valid: true,
		link: func() *cs.Link {
			l := createInvalidLink()
			l.Meta["action"] = "buy"
			return l
		},
	}, {
		name:  "missing-action",
		valid: true,
		link: func() *cs.Link {
			l := createInvalidLink()
			delete(l.Meta, "action")
			return l
		},
	}, {
		name:  "valid-link",
		valid: true,
		link:  createValidLink,
	}, {
		name:  "invalid-link",
		valid: false,
		link:  createInvalidLink,
	}}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			err := sv.Validate(nil, tt.link())
			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
