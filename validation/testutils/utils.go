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

package testutils

import (
	"encoding/json"
	"fmt"

	"github.com/stratumn/go-core/validation"
	"github.com/stratumn/go-core/validation/validators"
)

// These are the test key pairs of Alice and Bob.
const (
	AlicePrivateKey = "-----BEGIN ED25519 PRIVATE KEY-----\nME4CAQAwBQYDK2VwBEIEQByVNUFScxEsQJbHFjiV49lQ0OSGWqxXGSEV9CfD3RLc\n4HuxduXhOjSyr657IqXX4WBsj++R4pgRmqwJa9PN3W4=\n-----END ED25519 PRIVATE KEY-----\n"
	AlicePublicKey  = `-----BEGIN ED25519 PUBLIC KEY-----\nMCowBQYDK2VwAyEA4HuxduXhOjSyr657IqXX4WBsj++R4pgRmqwJa9PN3W4=\n-----END ED25519 PUBLIC KEY-----\n`

	BobPrivateKey = "-----BEGIN ED25519 PRIVATE KEY-----\nME4CAQAwBQYDK2VwBEIEQBNjYUKZhIQCu1a2DZde6jM5kSltWKqRXkim3MUeWyUT\nPAtD68Uo/tTD6zVSMpxdWb0J1SA7sVHumDI3LZRDGEM=\n-----END ED25519 PRIVATE KEY-----\n"
	BobPublicKey  = `-----BEGIN ED25519 PUBLIC KEY-----\nMCowBQYDK2VwAyEAPAtD68Uo/tTD6zVSMpxdWb0J1SA7sVHumDI3LZRDGEM=\n-----END ED25519 PUBLIC KEY-----\n`
)

// ValidAuctionJSONPKIConfig is a valid PKI schema for the auction process.
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

// ValidAuctionJSONTypesConfig is a valid types schema for the auction process.
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

// ValidChatJSONPKIConfig is a valid PKI schema for the chat process.
var ValidChatJSONPKIConfig = fmt.Sprintf(`
{
	"Bob Wagner": {
		"keys": ["%s"],
		"roles": ["manager", "it"]
	}
}
`, BobPublicKey)

// ValidChatJSONTypesConfig is a valid types schema for the chat process.
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

// CreateValidatorJSON formats a PKI and a types schema into a valid JSON configuration.
func CreateValidatorJSON(name, pki, types string) string {
	return fmt.Sprintf(`"%s": {"pki": %s,"types": %s}`, name, pki, types)
}

// LoadPKI unmarshalls a JSON-formatted PKI into a PKI struct.
func LoadPKI(rawPKI []byte) (*validators.PKI, error) {
	var pki validators.PKI
	if err := json.Unmarshal(rawPKI, &pki); err != nil {
		return nil, err
	}
	return &pki, nil
}

// LoadTypes unmarshalls a JSON-formatted types schema into a TypeSchema struct.
func LoadTypes(rawTypes []byte) (map[string]validation.TypeSchema, error) {
	var types map[string]validation.TypeSchema
	if err := json.Unmarshal(rawTypes, &types); err != nil {
		return nil, err
	}
	return types, nil
}

// These are the validation configuration exported by this package.
var (
	ValidAuctionJSONConfig = CreateValidatorJSON("auction", ValidAuctionJSONPKIConfig, ValidAuctionJSONTypesConfig)
	ValidChatJSONConfig    = CreateValidatorJSON("chat", ValidChatJSONPKIConfig, ValidChatJSONTypesConfig)
	ValidJSONConfig        = fmt.Sprintf(`{%s,%s}`, ValidAuctionJSONConfig, ValidChatJSONConfig)
)
