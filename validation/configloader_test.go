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

package validation_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/stratumn/go-indigocore/testutil"
	"github.com/stratumn/go-indigocore/utils"
	"github.com/stratumn/go-indigocore/validation"
	"github.com/stratumn/go-indigocore/validation/testutils"
	"github.com/stratumn/go-indigocore/validation/validators"
)

var pluginFile string

const (
	pluginsPath      = "testutils/plugins"
	pluginSourceFile = "custom_validator.go"
)

func TestMain(m *testing.M) {
	var res int
	defer os.Exit(res)

	var err error
	pluginFile, err = testutil.CompileGoPlugin(pluginsPath, pluginSourceFile)
	if err != nil {
		fmt.Println("could not launch validator tests: error while compiling validation plugin")
		os.Exit(2)
	}
	defer os.Remove(pluginFile)

	res = m.Run()
}

func TestLoadConfig_Success(t *testing.T) {

	t.Run("schema & signatures & transitions & plugins", func(T *testing.T) {
		testFile := utils.CreateTempFile(t, testutils.ValidJSONConfig)
		defer os.Remove(testFile)
		validatorMap, err := validation.LoadConfig(&validation.Config{
			RulesPath:   testFile,
			PluginsPath: pluginsPath,
		}, nil)

		assert.NoError(t, err, "LoadConfig()")
		assert.NotNil(t, validatorMap)

		var schemaValidatorCount, pkiValidatorCount, transitionValidatorCount int
		for _, validatorList := range validatorMap {
			for _, v := range validatorList {
				if _, ok := v.(*validators.PKIValidator); ok {
					pkiValidatorCount++
				} else if _, ok := v.(*validators.SchemaValidator); ok {
					schemaValidatorCount++
				} else if _, ok := v.(*validators.TransitionValidator); ok {
					transitionValidatorCount++
				}
			}
		}
		assert.Equal(t, 3, schemaValidatorCount)
		assert.Equal(t, 2, pkiValidatorCount)
		assert.Equal(t, 4, transitionValidatorCount)
	})

	t.Run("schema & signatures & transitions & plugins with listener", func(T *testing.T) {
		testFile := utils.CreateTempFile(t, testutils.ValidJSONConfig)
		defer os.Remove(testFile)
		validatorProcessCount := 0
		validatorCount := 0
		validators, err := validation.LoadConfig(&validation.Config{
			RulesPath:   testFile,
			PluginsPath: pluginsPath,
		}, func(process string, schema *validation.RulesSchema, processValidators validators.Validators) {
			validatorProcessCount++
			validatorCount = validatorCount + len(processValidators)
		})
		assert.NoError(t, err, "LoadConfig()")
		assert.NotNil(t, validators)
		assert.Equal(t, 2, validatorProcessCount)
		assert.Equal(t, 10, validatorCount)
		assert.Len(t, validators, validatorProcessCount)
		assert.Len(t, validators["chat"], 4)
		assert.Len(t, validators["auction"], 6)
	})

	t.Run("Null signatures", func(T *testing.T) {

		var validJSONSig = fmt.Sprintf(`
		{
			"testProcess": {
			  "pki": {
			    "alice.vandenbudenmayer@stratumn.com": {
					"keys": ["%s"],
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
		}`, testutils.AlicePublicKey)

		testFile := utils.CreateTempFile(t, validJSONSig)
		defer os.Remove(testFile)
		validatorMap, err := validation.LoadConfig(&validation.Config{
			RulesPath: testFile,
		}, nil)

		require.NoError(t, err, "LoadConfig()")
		assert.NotNil(t, validatorMap)

		require.Len(t, validatorMap, 1)
		require.Len(t, validatorMap["testProcess"], 1)
		assert.IsType(t, &validators.SchemaValidator{}, validatorMap["testProcess"][0])
	})

	t.Run("Empty signatures", func(T *testing.T) {

		var validJSONSig = fmt.Sprintf(`
		{
			"test": {
			    "pki": {
					"alice.vandenbudenmayer@stratumn.com": {
						"keys": ["%s"],
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
		}`, testutils.AlicePublicKey)

		testFile := utils.CreateTempFile(t, validJSONSig)
		defer os.Remove(testFile)
		validatorMap, err := validation.LoadConfig(&validation.Config{
			RulesPath: testFile,
		}, nil)

		require.NoError(t, err, "LoadConfig()")
		assert.NotNil(t, validatorMap)

		require.Len(t, validatorMap, 1)
		require.Len(t, validatorMap["test"], 1)
		assert.IsType(t, &validators.SchemaValidator{}, validatorMap["test"][0])
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
		validatorMap, err := validation.LoadConfig(&validation.Config{
			RulesPath: testFile,
		}, nil)

		require.NoError(t, err, "LoadConfig()")
		assert.NotNil(t, validatorMap)

		require.Len(t, validatorMap, 1)
		require.Len(t, validatorMap["test"], 1)
		assert.IsType(t, &validators.SchemaValidator{}, validatorMap["test"][0])
	})

	t.Run("Transitions only", func(T *testing.T) {

		const validJSONSig = `
		{
			"test": {
			    "types": {
				"init": {
				    "transitions": ["test"]
				}
			    }
			}
		}`

		testFile := utils.CreateTempFile(t, validJSONSig)
		defer os.Remove(testFile)
		validatorMap, err := validation.LoadConfig(&validation.Config{
			RulesPath: testFile,
		}, nil)

		require.NoError(t, err, "LoadConfig()")
		assert.NotNil(t, validatorMap)

		require.Len(t, validatorMap, 1)
		require.Len(t, validatorMap["test"], 1)
		assert.IsType(t, &validators.TransitionValidator{}, validatorMap["test"][0])
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
		validatorList, err := validation.LoadConfig(&validation.Config{
			RulesPath: testFile,
		}, nil)

		assert.Nil(t, validatorList)
		assert.EqualError(t, err, validators.ErrMissingProcess.Error())
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
		validatorList, err := validation.LoadConfig(&validation.Config{
			RulesPath: testFile,
		}, nil)

		assert.Nil(t, validatorList)
		assert.EqualError(t, err, validators.ErrMissingLinkType.Error())
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
		validators, err := validation.LoadConfig(&validation.Config{
			RulesPath: testFile,
		}, nil)

		assert.Nil(t, validators)
		assert.EqualError(t, err, validation.ErrInvalidValidator.Error())
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
		validators, err := validation.LoadConfig(&validation.Config{
			RulesPath: testFile,
		}, nil)

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
		validators, err := validation.LoadConfig(&validation.Config{
			RulesPath: testFile,
		}, nil)

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
		validators, err := validation.LoadConfig(&validation.Config{
			RulesPath: testFile,
		}, nil)

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
		validators, err := validation.LoadConfig(&validation.Config{
			RulesPath: testFile,
		}, nil)

		assert.Nil(t, validators)
		assert.EqualError(t, err, "rules.json needs a 'pki' field to list authorized public keys")
	})

	t.Run("Bad public key", func(T *testing.T) {
		const invalidPKIConfig = `
		{
			"test": {
			  "pki": {
			    "alice.vandenbudenmayer@stratumn.com": {
			      "keys": ["badPrivateKey"],
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
		validators, err := validation.LoadConfig(&validation.Config{
			RulesPath: testFile,
		}, nil)

		assert.Nil(t, validators)
		assert.EqualError(t, err, "error while parsing public key [badPrivateKey]: failed to decode PEM block")
	})
}
