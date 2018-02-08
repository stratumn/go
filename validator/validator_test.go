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

package validator_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/validator"
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

type testCase struct {
	name  string
	link  *cs.Link
	valid bool
}

var testCases = []testCase{{
	name: "valid-link",
	link: &cs.Link{
		State: map[string]interface{}{
			"buyer":    "alice",
			"bidPrice": 42,
		},
		Meta: map[string]interface{}{
			"process": "auction",
			"action":  "bid",
		},
	},
	valid: true,
}, {
	name: "no-validator-match",
	link: &cs.Link{
		Meta: map[string]interface{}{
			"process": "unknown",
			"action":  "unknown",
		},
	},
	valid: true,
}, {
	name: "missing-required-field",
	link: &cs.Link{
		State: map[string]interface{}{
			"to": "bob",
		},
		Meta: map[string]interface{}{
			"process": "chat",
			"action":  "message",
		},
	},
	valid: false,
}, {
	name: "invalid-field-value",
	link: &cs.Link{
		State: map[string]interface{}{
			"buyer":    "alice",
			"bidPrice": -10,
		},
		Meta: map[string]interface{}{
			"process": "auction",
			"action":  "bid",
		},
	},
	valid: false,
}}

func TestValidator(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "valid-config")
	require.NoError(t, err, "ioutil.TempFile()")

	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.WriteString(validJSONConfig)
	require.NoError(t, err, "tmpfile.WriteString()")

	cfg, err := validator.LoadConfig(tmpfile.Name())
	assert.NoError(t, err, "validator.LoadConfig()")

	v := validator.NewMultiValidator(cfg)
	assert.NotNil(t, v, "validator.NewMultiValidator()")

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Validate(nil, tt.link)
			if tt.valid {
				assert.NoError(t, err, "v.Validate()")
			} else {
				assert.Error(t, err, "v.Validate()")
			}
		})
	}
}
