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

func TestLoadConfig_Success(t *testing.T) {

	t.Run("schema & signatures", func(T *testing.T) {
		const validJSONConfig = `
		{
			"pki": {
			    "TESTKEY1": {
				"name": "Alice Van den Budenmayer",
				"roles": [
				    "employee"
				]
			    },
			    "TESTKEY2": {
				"name": "Bob Wagner",
				"roles": [
				    "manager",
				    "it"
				]
			    }
			},
			"validators": {
			    "auction": [
				{
				    "id": "initFormat",	
				    "type": "init",
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
					"required": [
					    "seller",
					    "lot",
					    "initialPrice"
					]
				    }
				},
				{
					"id": "bidFormat",	
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
				    "id": "messageFormat",	
				    "type": "message",
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
					"required": [
					    "to",
					    "content"
					]
				    }
				},
				{
				    "id": "initSigned",
				    "type": "init",
				    "signatures": ["manager", "it"]
				}
			    ]
			}
		    }
		`

		testFile := createTMPFile(t, validJSONConfig)
		defer os.Remove(testFile)
		validators, err := LoadConfig(testFile)

		assert.NoError(t, err, "LoadConfig()")
		assert.NotNil(t, validators)

		var schemaValidatorCount, pkiValidatorCount int
		for _, v := range validators {
			if _, ok := v.(*pkiValidator); ok {
				pkiValidatorCount++
			} else if _, ok := v.(*schemaValidator); ok {
				schemaValidatorCount++
			}
		}
		assert.Equal(t, 3, schemaValidatorCount)
		assert.Equal(t, 2, pkiValidatorCount)
	})

	t.Run("Null signatures", func(T *testing.T) {

		const validJSONSig = `
		{
			"pki": {
			    "TESTKEY1": {
				"name": "Alice Van den Budenmayer",
				"roles": [
				    "employee"
				]
			    },
			    "TESTKEY2": {
				"name": "Bob Wagner",
				"roles": [
				    "manager",
				    "it"
				]
			    }
			},
			"validators": {
			    "test": [
				{
				    "id": "initSigned",
				    "type": "init",
				    "signatures": null,
				    "schema": {}
				}
			    ]
			}
		    }
	`

		testFile := createTMPFile(t, validJSONSig)
		defer os.Remove(testFile)
		validators, err := LoadConfig(testFile)

		require.NoError(t, err, "LoadConfig()")
		assert.NotNil(t, validators)

		require.Len(t, validators, 1)
		_, ok := validators[0].(*schemaValidator)
		assert.True(t, ok)
	})

	t.Run("Empty signatures", func(T *testing.T) {

		const validJSONSig = `
		{
			"pki": {
			    "TESTKEY1": {
				"name": "Alice Van den Budenmayer",
				"roles": [
				    "employee"
				]
			    },
			    "TESTKEY2": {
				"name": "Bob Wagner",
				"roles": [
				    "manager",
				    "it"
				]
			    }
			},
			"validators": {
			    "test": [
				{
				    "id": "initSigned",
				    "type": "init",
				    "signatures": [],
				    "schema": {}
				}
			    ]
			}
		    }
	`

		testFile := createTMPFile(t, validJSONSig)
		defer os.Remove(testFile)
		validators, err := LoadConfig(testFile)

		require.NoError(t, err, "LoadConfig()")
		assert.NotNil(t, validators)

		require.Len(t, validators, 1)
		_, ok := validators[0].(*schemaValidator)
		assert.True(t, ok)
	})

	t.Run("No PKI", func(T *testing.T) {

		const validJSONSig = `
		{
			"validators": {
			    "test": [
				{
				    "id": "initSigned",
				    "type": "init",
				    "schema": {
					"type": "object"
				    }
				}
			    ]
			}
		    }
		`

		testFile := createTMPFile(t, validJSONSig)
		defer os.Remove(testFile)
		validators, err := LoadConfig(testFile)

		require.NoError(t, err, "LoadConfig()")
		assert.NotNil(t, validators)

		assert.Len(t, validators, 1)
		require.Len(t, validators, 1)
		_, ok := validators[0].(*schemaValidator)
		assert.True(t, ok)
	})

}

func TestLoadValidators_Error(t *testing.T) {
	t.Run("Missing schema", func(T *testing.T) {
		const invalidValidatorConfig = `
		{
			"pki": {},
			"validators": {
			    "auction": [
				{
				    "id": "wrongValidator",
				    "type": "init"
				}
			    ]
			}
		    }
		`
		testFile := createTMPFile(t, invalidValidatorConfig)
		validators, err := LoadConfig(testFile)

		assert.Nil(t, validators)
		assert.EqualError(t, err, ErrInvalidValidator.Error())
	})

	t.Run("Missing identifier", func(T *testing.T) {
		const invalidValidatorConfig = `
		{
			"pki": {},
			"validators": {
			    "auction": [
				{
				    "type": "init",
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
					}
				    }
				}
			    ]
			}
		    }
		`
		testFile := createTMPFile(t, invalidValidatorConfig)
		defer os.Remove(testFile)
		validators, err := LoadConfig(testFile)

		assert.Nil(t, validators)
		assert.EqualError(t, err, ErrMissingIdentifier.Error())
	})

	t.Run("Missing type", func(T *testing.T) {
		const invalidValidatorConfig = `
		{
			"pki": {},
			"validators": {
			    "auction": [
				{
				    "id": "missingType",
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
					}
				    }
				}
			    ]
			}
		    }
		`
		testFile := createTMPFile(t, invalidValidatorConfig)
		defer os.Remove(testFile)
		validators, err := LoadConfig(testFile)

		assert.Nil(t, validators)
		assert.EqualError(t, err, ErrMissingLinkType.Error())
	})

	t.Run("Bad signature validator", func(T *testing.T) {
		const invalidValidatorConfig = `
		{
			"pki": {},
			"validators": {
			    "auction": [
				{
				    "id": "missingType",
				    "type": "action",	
				    "signatures": "test"
				}
			    ]
			}
		    }
		`
		testFile := createTMPFile(t, invalidValidatorConfig)
		defer os.Remove(testFile)
		validators, err := LoadConfig(testFile)

		assert.Nil(t, validators)
		assert.Error(t, err)
	})
}

func TestLoadPKI_Error(t *testing.T) {

	t.Run("No PKI", func(T *testing.T) {
		const noPKIConfig = `
		{
			"validators": {
				"auction": [
				    {
					"id": "missingType",
					"type": "action",	
					"signatures": ["test"]
				    }
				]
			    }
		    }
		`
		testFile := createTMPFile(t, noPKIConfig)
		defer os.Remove(testFile)
		validators, err := LoadConfig(testFile)

		assert.Nil(t, validators)
		assert.EqualError(t, err, "rules.json needs a 'pki' field to list authorized public keys")
	})

	t.Run("Bad public key", func(T *testing.T) {
		const invalidPKIConfig = `
		{
			"pki": {
				"": {
				    "name": "Alice Van den Budenmayer",
				    "roles": [
					"employee"
				    ]
				}
			},
			"validators": {}
		}
		`
		testFile := createTMPFile(t, invalidPKIConfig)
		defer os.Remove(testFile)
		validators, err := LoadConfig(testFile)

		assert.Nil(t, validators)
		assert.EqualError(t, err, "Error while parsing PKI: public key must be a non null base64 encoded string")
	})
}

func createTMPFile(t *testing.T, data string) string {
	tmpfile, err := ioutil.TempFile("", "invalid-config")
	require.NoError(t, err, "ioutil.TempFile()")

	_, err = tmpfile.WriteString(data)
	require.NoError(t, err, "tmpfile.WriteString()")
	return tmpfile.Name()
}
