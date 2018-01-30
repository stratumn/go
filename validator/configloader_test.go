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
      "signatures": true,
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
      "signatures": false,
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
    },
    {
	"type": "init",
	"signatures": true    
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
	assert.Len(t, cfg.SignatureConfigs, 2)
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
	assert.EqualError(t, err, ErrInvalidValidator.Error())
}
