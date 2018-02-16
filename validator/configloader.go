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
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
)

var (
	// ErrInvalidValidator is returned when the schema and the signatures are both missing in a validator.
	ErrInvalidValidator = errors.New("a validator requires a JSON schema or a signature criteria to be valid")

	// ErrBadPublicKey is returned when a public key is empty or not base64-encoded
	ErrBadPublicKey = errors.New("public key must be a non null base64 encoded string")

	// ErrNoPKI is returned when rules.json doesn't contain a `pki` field
	ErrNoPKI = errors.New("rules.json needs a 'pki' field to list authorized public keys")
)

type processesRules map[string]rulesSchema

type rulesSchema struct {
	PKI   json.RawMessage `json:"pki"`
	Types json.RawMessage `json:"types"`
}

type rulesListener func(process string, schema rulesSchema, validators []Validator)

// LoadConfig loads the validators configuration from a json file.
// The configuration returned can then be used in NewMultiValidator().
func LoadConfig(path string, listener rulesListener) ([]Validator, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return LoadConfigContent(data, listener)
}

// LoadConfigContent loads the validators configuration from json data.
// The configuration returned can then be used in NewMultiValidator().
func LoadConfigContent(data []byte, listener rulesListener) ([]Validator, error) {
	var rules processesRules
	err := json.Unmarshal(data, &rules)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return LoadProcessRules(rules, listener)
}

// LoadProcessRules loads the validators configuration from a slice of processRule.
// The configuration returned can then be used in NewMultiValidator().
func LoadProcessRules(rules processesRules, listener rulesListener) ([]Validator, error) {
	var err error
	var validators []Validator
	for process, schema := range rules {
		var pki *PKI
		if schema.PKI != nil {
			pki, err = loadPKIConfig(schema.PKI)
			if err != nil {
				return nil, errors.WithStack(err)
			}
		}
		processValidators, err := loadValidatorsConfig(process, schema.Types, pki)
		if err != nil {
			return nil, err
		}
		if listener != nil {
			listener(process, schema, processValidators)
		}
		validators = append(validators, processValidators...)
	}
	return validators, nil
}

// loadPKIConfig deserializes json into a PKI struct.
// It checks that public keys are base64 encoded.
func loadPKIConfig(data json.RawMessage) (*PKI, error) {
	var jsonData PKI
	err := json.Unmarshal(data, &jsonData)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	for _, id := range jsonData {
		for _, key := range id.Keys {
			if _, err := base64.StdEncoding.DecodeString(key); key == "" || err != nil {
				return nil, errors.Wrap(ErrBadPublicKey, "Error while parsing PKI")
			}
		}
	}
	return &jsonData, nil
}

type jsonValidatorData struct {
	Signatures []string         `json:"signatures"`
	Schema     *json.RawMessage `json:"schema"`
}

func loadValidatorsConfig(process string, data json.RawMessage, pki *PKI) ([]Validator, error) {
	var jsonStruct map[string]jsonValidatorData
	err := json.Unmarshal(data, &jsonStruct)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var validators []Validator
	for linkType, val := range jsonStruct {
		if linkType == "" {
			return nil, ErrMissingLinkType
		}
		if len(val.Signatures) == 0 && val.Schema == nil {
			return nil, ErrInvalidValidator
		}

		baseConfig, err := newValidatorBaseConfig(process, linkType)
		if err != nil {
			return nil, err
		}
		if len(val.Signatures) > 0 {
			// if no PKI was provided, one cannot require signatures.
			if pki == nil {
				return nil, ErrNoPKI
			}
			validators = append(validators, newPkiValidator(baseConfig, val.Signatures, pki))
		}

		if val.Schema != nil {
			schemaData, _ := val.Schema.MarshalJSON()
			schemaValidator, err := newSchemaValidator(baseConfig, schemaData)
			if err != nil {
				return nil, err
			}
			validators = append(validators, schemaValidator)
		}
	}

	return validators, nil
}
