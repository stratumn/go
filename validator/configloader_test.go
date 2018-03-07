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
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/stratumn/go-indigocore/utils"
)

func TestLoadConfig_Success(t *testing.T) {

	t.Run("schema & signatures & transitions", func(T *testing.T) {
		testFile := utils.CreateTempFile(t, ValidJSONConfig)
		defer os.Remove(testFile)
		validators, err := LoadConfig(testFile, nil)

		assert.NoError(t, err, "LoadConfig()")
		assert.NotNil(t, validators)

		var schemaValidatorCount, pkiValidatorCount, transitionValidatorCount int
		for _, v := range validators {
			if _, ok := v.(*pkiValidator); ok {
				pkiValidatorCount++
			} else if _, ok := v.(*schemaValidator); ok {
				schemaValidatorCount++
			} else if _, ok := v.(*transitionValidator); ok {
				transitionValidatorCount++
			}
		}
		assert.Equal(t, 3, schemaValidatorCount)
		assert.Equal(t, 2, pkiValidatorCount)
		assert.Equal(t, 4, transitionValidatorCount)
	})

	t.Run("schema & signatures & transitions with listener", func(T *testing.T) {
		testFile := utils.CreateTempFile(t, ValidJSONConfig)
		defer os.Remove(testFile)
		validatorProcessCount := 0
		validatorCount := 0
		validators, err := LoadConfig(testFile, func(process string, schema rulesSchema, validators []Validator) {
			validatorProcessCount++
			validatorCount = validatorCount + len(validators)
		})
		assert.NoError(t, err, "LoadConfig()")
		assert.NotNil(t, validators)
		assert.Equal(t, 2, validatorProcessCount)
		assert.Equal(t, 9, validatorCount)
		assert.Len(t, validators, validatorCount)
	})

	t.Run("Null signatures", func(T *testing.T) {

		const validJSONSig = `
		{
			"testProcess": {
			  "pki": {
			    "alice.vandenbudenmayer@stratumn.com": {
					"keys": ["TESTKEY1"],
					"name": "Alice Van den Budenmayer",
					"roles": ["employee"]
			    }
			  },
			  "types": {
					"init": {
						"signatures": null,      
						"schema": {}
					}
			  	}
			}
		}`

		testFile := utils.CreateTempFile(t, validJSONSig)
		defer os.Remove(testFile)
		validators, err := LoadConfig(testFile, nil)

		require.NoError(t, err, "LoadConfig()")
		assert.NotNil(t, validators)

		require.Len(t, validators, 1)
		assert.IsType(t, &schemaValidator{}, validators[0])
	})

	t.Run("Empty signatures", func(T *testing.T) {

		const validJSONSig = `
		{
			"test": {
			    "pki": {
					"alice.vandenbudenmayer@stratumn.com": {
						"keys": ["TESTKEY1"],
						"name": "Alice Van den Budenmayer",
						"roles": ["employee"]
					}
			    },
			    "types": {
					"init": {
						"signatures": [],
						"schema": {}
					}
			    }
			}
		}`

		testFile := utils.CreateTempFile(t, validJSONSig)
		defer os.Remove(testFile)
		validators, err := LoadConfig(testFile, nil)

		require.NoError(t, err, "LoadConfig()")
		assert.NotNil(t, validators)

		require.Len(t, validators, 1)
		assert.IsType(t, &schemaValidator{}, validators[0])
	})

	t.Run("No PKI", func(T *testing.T) {

		const validJSONSig = `
		{
			"test": {
			    "types": {
				"init": {
				    "schema": {
					"type": "object"
				    }
				}
			    }
			}
		}`

		testFile := utils.CreateTempFile(t, validJSONSig)
		defer os.Remove(testFile)
		validators, err := LoadConfig(testFile, nil)

		require.NoError(t, err, "LoadConfig()")
		assert.NotNil(t, validators)

		require.Len(t, validators, 1)
		assert.IsType(t, &schemaValidator{}, validators[0])
	})

}

func TestLoadValidators_Error(t *testing.T) {

	t.Run("Missing process", func(T *testing.T) {
		const invalidValidatorConfig = `
		{
			"": {
			  "types": {
			    "init": {
					"schema": {}
				}
			  },
			  "pki": {}
			}
		}`
		testFile := utils.CreateTempFile(t, invalidValidatorConfig)
		validators, err := LoadConfig(testFile, nil)

		assert.Nil(t, validators)
		assert.EqualError(t, err, ErrMissingProcess.Error())
	})

	t.Run("Missing link type", func(T *testing.T) {
		const invalidValidatorConfig = `
		{
			"test": {
			  "types": {
			    "": {
					"schema": {}
				}
			  },
			  "pki": {}
			}
		}`
		testFile := utils.CreateTempFile(t, invalidValidatorConfig)
		validators, err := LoadConfig(testFile, nil)

		assert.Nil(t, validators)
		assert.EqualError(t, err, ErrMissingLinkType.Error())
	})

	t.Run("Missing schema", func(T *testing.T) {
		const invalidValidatorConfig = `
		{
			"test": {
			  "types": {
			    "init": {}
			  },
			  "pki": {}
			}
		}`
		testFile := utils.CreateTempFile(t, invalidValidatorConfig)
		validators, err := LoadConfig(testFile, nil)

		assert.Nil(t, validators)
		assert.EqualError(t, err, ErrInvalidValidator.Error())
	})

	t.Run("Bad signature validator", func(T *testing.T) {
		const invalidValidatorConfig = `
		{
			"test": {
				"types": {
				    "init": {
					"signatures": "test"
				    }
				}
			}
		}`
		testFile := utils.CreateTempFile(t, invalidValidatorConfig)
		defer os.Remove(testFile)
		validators, err := LoadConfig(testFile, nil)

		assert.Nil(t, validators)
		assert.Error(t, err)
	})

	t.Run("Bad transitions validator", func(T *testing.T) {
		const invalidValidatorConfig = `
		{
			"test": {
				"types": {
				    "init": {
					"transitions": "test"
				    }
				}
			}
		}`
		testFile := utils.CreateTempFile(t, invalidValidatorConfig)
		defer os.Remove(testFile)
		validators, err := LoadConfig(testFile, nil)

		assert.Nil(t, validators)
		assert.Error(t, err)
	})

	t.Run("Missing transitions validator", func(T *testing.T) {
		const invalidValidatorConfig = `
		{
			"test": {
				"types": {
				    "foo": {
						"schema": { "type": "object" },
						"transitions": ["test"]
				    },
				    "bar": {
						"schema": { "type": "object" }
				    }
				}
			}
		}`
		testFile := utils.CreateTempFile(t, invalidValidatorConfig)
		defer os.Remove(testFile)
		validators, err := LoadConfig(testFile, nil)

		assert.Nil(t, validators)
		assert.EqualError(t, err, "missing transition definition for process test and linkTypes [bar]")
	})
}

func TestLoadPKI_Error(t *testing.T) {

	t.Run("No PKI", func(T *testing.T) {
		const noPKIConfig = `
		{
			"test": {
				"types": {
				    "init": {
					"signatures": ["test"]
				    }
				}
			    }
			}
		`
		testFile := utils.CreateTempFile(t, noPKIConfig)
		defer os.Remove(testFile)
		validators, err := LoadConfig(testFile, nil)

		assert.Nil(t, validators)
		assert.EqualError(t, err, "rules.json needs a 'pki' field to list authorized public keys")
	})

	t.Run("Bad public key", func(T *testing.T) {
		const invalidPKIConfig = `
		{
			"test": {
			  "pki": {
			    "alice.vandenbudenmayer@stratumn.com": {
			      "keys": ["t€st"],
			      "name": "Alice Van den Budenmayer",
			      "roles": ["employee"]
			    }
			  },
			  "types": {
			    "init": {
			      "signatures": [],
			      "schema": {}
			    }
			  }
			}
		}
				      `
		testFile := utils.CreateTempFile(t, invalidPKIConfig)
		defer os.Remove(testFile)
		validators, err := LoadConfig(testFile, nil)

		assert.Nil(t, validators)
		assert.EqualError(t, err, "error while parsing PKI: public key must be a non null base64 encoded string")
	})
}
