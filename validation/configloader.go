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

package validation

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	"github.com/stratumn/go-crypto/keys"
	"github.com/stratumn/go-indigocore/validation/validators"
)

var (
	// ErrInvalidValidator is returned when the schema and the signatures are both missing in a validator.
	ErrInvalidValidator = errors.New("a validator requires a JSON schema, a signature or a transition criteria to be valid")

	// ErrBadPublicKey is returned when a public key is empty or not base64-encoded
	ErrBadPublicKey = errors.New("public key must be a non null base64 encoded string")

	// ErrNoPKI is returned when rules.json doesn't contain a `pki` field
	ErrNoPKI = errors.New("rules.json needs a 'pki' field to list authorized public keys")
)

type processesRules map[string]RulesSchema

// RulesSchema represents the structure of validation rules.
type RulesSchema struct {
	PKI   *validators.PKI       `json:"pki"`
	Types map[string]TypeSchema `json:"types"`
}

type rulesListener func(process string, schema *RulesSchema, validators validators.Validators)

// LoadConfig loads the validators configuration from a json file.
// The configuration returned can then be used in NewMultiValidator().
func LoadConfig(validationCfg *Config, listener rulesListener) (validators.ProcessesValidators, error) {
	f, err := os.Open(validationCfg.RulesPath)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return LoadConfigContent(data, validationCfg.PluginsPath, listener)
}

// LoadConfigContent loads the validators configuration from json data.
// The configuration returned can then be used in NewMultiValidator().
func LoadConfigContent(data []byte, pluginsPath string, listener rulesListener) (validators.ProcessesValidators, error) {
	var rules processesRules
	err := json.Unmarshal(data, &rules)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	validators := make(validators.ProcessesValidators, 0)
	for process, schema := range rules {
		validators[process], err = LoadProcessRules(&schema, process, pluginsPath, listener)
		if err != nil {
			return nil, err
		}
	}

	return validators, nil
}

// LoadProcessRules loads the validators configuration for a single process.
// The configuration returned can then be used in NewMultiValidator().
func LoadProcessRules(schema *RulesSchema, process, pluginsPath string, listener rulesListener) (validators.Validators, error) {
	if err := checkPKIConfig(schema.PKI); err != nil {
		return nil, err
	}

	processValidators, err := loadValidatorsConfig(process, pluginsPath, schema.Types, schema.PKI)
	if err != nil {
		return nil, err
	}
	if listener != nil {
		listener(process, schema, processValidators)
	}
	return processValidators, nil
}

// checkPKIConfig checks that public keys are base64 encoded.
func checkPKIConfig(data *validators.PKI) error {
	if data == nil {
		return nil
	}

	for _, id := range *data {
		for _, key := range id.Keys {
			if _, err := keys.ParsePublicKey([]byte(key)); err != nil {
				return errors.Wrapf(err, "error while parsing public key [%s]", key)
			}
		}
	}
	return nil
}

// TypeSchema represent the structure of validation rules.
type TypeSchema struct {
	Signatures  []string                 `json:"signatures"`
	Schema      map[string]interface{}   `json:"schema"`
	Transitions []string                 `json:"transitions"`
	Script      *validators.ScriptConfig `json:"script"`
}

func loadValidatorsConfig(process, pluginsPath string, jsonStruct map[string]TypeSchema, pki *validators.PKI) (validators.Validators, error) {
	missingTransitionValidation := make([]string, 0)
	var validatorList validators.Validators
	for linkType, val := range jsonStruct {
		if linkType == "" {
			return nil, validators.ErrMissingLinkType
		}
		if len(val.Signatures) == 0 && val.Schema == nil && val.Transitions == nil {
			return nil, ErrInvalidValidator
		}

		baseConfig, err := validators.NewValidatorBaseConfig(process, linkType)
		if err != nil {
			return nil, err
		}
		if len(val.Signatures) > 0 {
			// if no PKI was provided, one cannot require signatures.
			if pki == nil {
				return nil, ErrNoPKI
			}
			validatorList = append(validatorList, validators.NewPKIValidator(baseConfig, val.Signatures, pki))
		}

		if val.Schema != nil {
			schemaData, err := json.Marshal(val.Schema)
			if err != nil {
				return nil, err
			}
			schemaValidator, err := validators.NewSchemaValidator(baseConfig, schemaData)
			if err != nil {
				return nil, err
			}
			validatorList = append(validatorList, schemaValidator)
		}

		if val.Script != nil {
			scriptValidator, err := validators.NewScriptValidator(baseConfig, val.Script, pluginsPath)
			if err != nil {
				return nil, err
			}
			validatorList = append(validatorList, scriptValidator)
		}

		if len(val.Transitions) > 0 {
			validatorList = append(validatorList, validators.NewTransitionValidator(baseConfig, val.Transitions))
		} else {
			missingTransitionValidation = append(missingTransitionValidation, linkType)
		}
	}

	if len(missingTransitionValidation) > 0 && len(missingTransitionValidation) != len(jsonStruct) {
		return nil, errors.Errorf("missing transition definition for process %s and linkTypes %v", process, missingTransitionValidation)
	}

	return validatorList, nil
}
