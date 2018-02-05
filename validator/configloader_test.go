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
		    "id": "initSigned",
		    "type": "init",
		    "signatures": true
		}
	    ]
	}
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
		tmpfile, err := ioutil.TempFile("", "invalid-config")
		require.NoError(t, err, "ioutil.TempFile()")

		defer os.Remove(tmpfile.Name())

		_, err = tmpfile.WriteString(invalidValidatorConfig)
		require.NoError(t, err, "tmpfile.WriteString()")

		cfg, err := LoadConfig(tmpfile.Name())

		assert.Nil(t, cfg)
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
		tmpfile, err := ioutil.TempFile("", "invalid-config")
		require.NoError(t, err, "ioutil.TempFile()")

		defer os.Remove(tmpfile.Name())

		_, err = tmpfile.WriteString(invalidValidatorConfig)
		require.NoError(t, err, "tmpfile.WriteString()")

		cfg, err := LoadConfig(tmpfile.Name())

		assert.Nil(t, cfg)
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
		tmpfile, err := ioutil.TempFile("", "invalid-config")
		require.NoError(t, err, "ioutil.TempFile()")

		defer os.Remove(tmpfile.Name())

		_, err = tmpfile.WriteString(invalidValidatorConfig)
		require.NoError(t, err, "tmpfile.WriteString()")

		cfg, err := LoadConfig(tmpfile.Name())

		assert.Nil(t, cfg)
		assert.EqualError(t, err, ErrMissingLinkType.Error())
	})
}

func TestLoadPKI_Error(t *testing.T) {

	t.Run("No PKI", func(T *testing.T) {
		const NoPKIConfig = `
		{
			"validators": {}
		}
		`
		tmpfile, err := ioutil.TempFile("", "invalid-config")
		require.NoError(t, err, "ioutil.TempFile()")

		defer os.Remove(tmpfile.Name())

		_, err = tmpfile.WriteString(NoPKIConfig)
		require.NoError(t, err, "tmpfile.WriteString()")

		cfg, err := LoadConfig(tmpfile.Name())

		assert.Nil(t, cfg)
		assert.EqualError(t, err, "rules.json needs a 'pki' field to list authorized public keys")
	})

	t.Run("Bad public key", func(T *testing.T) {
		const InvalidPKIConfig = `
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
		tmpfile, err := ioutil.TempFile("", "invalid-config")
		require.NoError(t, err, "ioutil.TempFile()")

		defer os.Remove(tmpfile.Name())

		_, err = tmpfile.WriteString(InvalidPKIConfig)
		require.NoError(t, err, "tmpfile.WriteString()")

		cfg, err := LoadConfig(tmpfile.Name())

		assert.Nil(t, cfg)
		assert.EqualError(t, err, "Error while parsing PKI: Public key must be a non-null base64 encoded string")
	})
}
