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
)

const AlicePrivateKey = `XyvgStfhWJ3uq/peoh/VyIIl0kTkCKQ1Fv5VQC4f5w3a+HASh6jEKKWYBvrALvSZpOaiR+c0ak3O0Oqa9kJ44w==`
const AlicePublicKey = `2vhwEoeoxCilmAb6wC70maTmokfnNGpNztDqmvZCeOM=`

var ValidAuctionJSONPKIConfig = fmt.Sprintf(`
{
	"alice.vandenbudenmayer@stratumn.com": {
		"keys": ["%s"],
		"roles": ["employee"]
	},
	"Bob Wagner": {
		"keys": ["hmxvE+c9PwGUSEVZQ10RPaTP5SkuTR60pJ+Bhwqih48="],
		"roles": ["manager", "it"]
	}
}
`, AlicePublicKey)

const ValidAuctionJSONTypesConfig = `
{
	"init": {
		"signatures": ["alice.vandenbudenmayer@stratumn.com"],
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
		},
		"transitions": [""]
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
		},
		"transitions": ["init", "bid"]
	}
}
`

const ValidChatJSONPKIConfig = `
{
	"Bob Wagner": {
		"keys": ["hmxvE+c9PwGUSEVZQ10RPaTP5SkuTR60pJ+Bhwqih48="],
		"roles": ["manager", "it"]
	}
}
`

const ValidChatJSONTypesConfig = `
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
		},
		"transitions": ["init", "message"]
	},
	"init": {
		"signatures": ["manager", "it"],
		"transitions": [""]
	}
}
`

func createValidatorJSON(name, pki, types string) string {
	return fmt.Sprintf(`"%s": {"pki": %s,"types": %s}`, name, pki, types)
}

var ValidAuctionJSONConfig = createValidatorJSON("auction", ValidAuctionJSONPKIConfig, ValidAuctionJSONTypesConfig)
var ValidChatJSONConfig = createValidatorJSON("chat", ValidChatJSONPKIConfig, ValidChatJSONTypesConfig)
var ValidJSONConfig = fmt.Sprintf(`{%s,%s}`, ValidAuctionJSONConfig, ValidChatJSONConfig)
