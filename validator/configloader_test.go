package validator

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const validJSONConfig = `
{
  "auction": [
    {
      "type": "init",
      "schema": {
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
      }
    },
    {
      "type": "bid",
      "schema": {
        "type": "object",
        "properties": {
          "buyer": {
            "type": "string"
          },
          "bidPrice": {
            "type": "integer",
            "minimum": 0
          }
        },
        "required": [
          "buyer",
          "bidPrice"
        ]
      }
    }
  ],
  "chat": [
    {
      "type": "message",
      "schema": {
        "type": "object",
        "properties": {
          "to": {
            "type": "string"
          },
          "content": {
            "type": "string"
          }
        },
        "required": [
          "to",
          "content"
        ]   
      }
    }
  ]
}
`

func TestLoadConfig_Success(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "valid-config")
	require.NoError(t, err, "ioutil.TempFile()")

	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.WriteString(validJSONConfig)
	require.NoError(t, err, "tmpfile.WriteString()")

	cfg, err := LoadConfig(tmpfile.Name())

	assert.NoError(t, err, "LoadConfig()")
	assert.NotNil(t, cfg)

	assert.Len(t, cfg.SchemaConfigs, 3)
}

const invalidJSONConfig = `
{
  "auction": [
  {
    "type": "init"
  },
  {
    "type": "bid",
    "schema": {
      "type": "object",
      "properties": {
        "buyer": {
    	  "type": "string"
        }
      },
      "required": [
        "buyer"
      ]
    }
  }]
}
`

func TestLoadConfig_Error(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "invalid-config")
	require.NoError(t, err, "ioutil.TempFile()")

	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.WriteString(invalidJSONConfig)
	require.NoError(t, err, "tmpfile.WriteString()")

	cfg, err := LoadConfig(tmpfile.Name())

	assert.Nil(t, cfg)
	assert.EqualError(t, err, ErrMissingSchema.Error())
}
