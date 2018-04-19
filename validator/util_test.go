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

const (
	AlicePrivateKey = "-----BEGIN ED25519 PRIVATE KEY-----\nBEC0TyVE2Y7+OgPHcSAAIAjUHCVA68swAp235LkQZBIrZnUfW/lss95djRXjIeX+\nezH5bdbVe7s4wbPJRBiej+it\n-----END ED25519 PRIVATE KEY-----\n"
	AlicePublicKey  = `-----BEGIN ED25519 PUBLIC KEY-----\nMCowBQYDK2VwAyEAdR9b+Wyz3l2NFeMh5f57Mflt1tV7uzjBs8lEGJ6P6K0=\n-----END ED25519 PUBLIC KEY-----\n`

	BobPrivateKey = "-----BEGIN ED25519 PRIVATE KEY-----\nBED2FCm0Wxbq0WGpsf+7qNEUe3WXM2rGDey8ZuYn723qJPraxU3A4L+KAsOOc2Hq\nXD7nmG3Bq0+2B2lO5VvcjcSe\n-----END ED25519 PRIVATE KEY-----\n"
	BobPublicKey  = `-----BEGIN ED25519 PUBLIC KEY-----\nMCowBQYDK2VwAyEA+trFTcDgv4oCw45zYepcPueYbcGrT7YHaU7lW9yNxJ4=\n-----END ED25519 PUBLIC KEY-----\n`
)

var ValidAuctionJSONPKIConfig = fmt.Sprintf(`
{
	"alice.vandenbudenmayer@stratumn.com": {
		"keys": ["%s"],
		"roles": ["employee"]
	},
	"Bob Wagner": {
		"keys": ["%s"],
		"roles": ["manager", "it"]
	}
}
`, AlicePublicKey, BobPublicKey)

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
		"transitions": [""],
		"script": {
			"file": "custom_validator.so",
			"type": "go"
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
		},
		"transitions": ["init", "bid"]
	}
}
`

var ValidChatJSONPKIConfig = fmt.Sprintf(`
{
	"Bob Wagner": {
		"keys": ["%s"],
		"roles": ["manager", "it"]
	}
}
`, BobPublicKey)

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
