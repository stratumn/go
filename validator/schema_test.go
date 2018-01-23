package validator

import (
	"testing"

	"github.com/stratumn/sdk/cs/cstesting"

	"github.com/stratumn/sdk/cs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSchema = `
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
	validSchema := []byte(testSchema)
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
		name:          "missing-process",
		process:       "",
		linkType:      linkType,
		schema:        validSchema,
		valid:         false,
		expectedError: ErrMissingProcess,
	}, {
		name:          "missing-link-type",
		process:       process,
		linkType:      "",
		schema:        validSchema,
		valid:         false,
		expectedError: ErrMissingLinkType,
	}, {
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
	schema := []byte(testSchema)
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
