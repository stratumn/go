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
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
)

const validAuctionJSONPKIConfig = `
{
	"alice.vandenbudenmayer@stratumn.com": {
		"keys": ["TESTKEY1"],
		"roles": ["employee"]
	},
	"Bob Wagner": {
		"keys": ["hmxvE+c9PwGUSEVZQ10RPaTP5SkuTR60pJ+Bhwqih48="],
		"roles": ["manager", "it"]
	}
}
`

const validAuctionJSONTypesConfig = `
{
	"init": {
		"signatures": ["Alice Van den Budenmayer"],
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
			"required": ["seller", "lot", "initialPrice"]
		}
	},
	"bid": {
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
			"required": ["buyer", "bidPrice"]
		}
	}
}
`

const validChatJSONPKIConfig = `
{
	"Bob Wagner": {
		"keys": ["hmxvE+c9PwGUSEVZQ10RPaTP5SkuTR60pJ+Bhwqih48="],
		"roles": ["manager", "it"]
	}
}
`

const validChatJSONTypesConfig = `
{
	"message": {
		"signatures": null,
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
			"required": ["to", "content"]
		}
	},
	"init": {
		"signatures": ["manager", "it"]
	}
}
`

func createValidatorJSON(name, pki, types string) string {
	return fmt.Sprintf(`"%s": {"pki": %s,"types": %s}`, name, pki, types)
}

var validAuctionJSONConfig = createValidatorJSON("auction", validAuctionJSONPKIConfig, validAuctionJSONTypesConfig)
var validChatJSONConfig = createValidatorJSON("chat", validChatJSONPKIConfig, validChatJSONTypesConfig)
var validJSONConfig = fmt.Sprintf(`{%s,%s}`, validAuctionJSONConfig, validChatJSONConfig)

func createTempFile(t *testing.T, data string) string {
	tmpfile, err := ioutil.TempFile("", "validator-tmpfile")
	require.NoError(t, err, "ioutil.TempFile()")

	_, err = tmpfile.WriteString(data)
	require.NoError(t, err, "tmpfile.WriteString()")
	return tmpfile.Name()
}
