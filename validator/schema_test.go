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

func TestSchemaValidator(t *testing.T) {
	schema := []byte(testSchema)
	sv, err := newSchemaValidator("p1", "sell", schema)
	require.NoError(t, err)

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
