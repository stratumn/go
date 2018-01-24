package validator

import (
	"testing"

	"github.com/stratumn/sdk/cs/cstesting"
	"github.com/stretchr/testify/assert"
)

func TestMultiValidator_New(t *testing.T) {
	mv := NewMultiValidator(&MultiValidatorConfig{
		SchemaConfigs: []*schemaValidatorConfig{
			&schemaValidatorConfig{},
			&schemaValidatorConfig{},
		},
	})

	assert.Len(t, mv.(*multiValidator).validators, 2)
}

func TestMultiValidator_Hash(t *testing.T) {
	mv1 := NewMultiValidator(&MultiValidatorConfig{
		SchemaConfigs: []*schemaValidatorConfig{
			&schemaValidatorConfig{Process: "p"},
		},
	})

	h1 := mv1.Hash()
	assert.NotNil(t, h1)

	mv2 := NewMultiValidator(&MultiValidatorConfig{
		SchemaConfigs: []*schemaValidatorConfig{
			&schemaValidatorConfig{Process: "p"},
		},
	})
	h2 := mv2.Hash()
	assert.EqualValues(t, h1, h2)

	mv3 := NewMultiValidator(&MultiValidatorConfig{
		SchemaConfigs: []*schemaValidatorConfig{
			&schemaValidatorConfig{Process: "p2"},
		},
	})
	h3 := mv3.Hash()
	assert.False(t, h1.Equals(h3))
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

	mv := NewMultiValidator(&MultiValidatorConfig{
		SchemaConfigs: []*schemaValidatorConfig{svCfg1, svCfg2},
	})

	t.Run("Validate succeeds when all children succeed", func(t *testing.T) {
		err := mv.Validate(nil, cstesting.RandomLink())
		assert.NoError(t, err)
	})

	t.Run("Validate fails if one of the children fails", func(t *testing.T) {
		l := cstesting.RandomLink()
		l.Meta["process"] = "p"
		l.Meta["action"] = "a2"

		err := mv.Validate(nil, l)
		assert.Error(t, err)
	})
}
