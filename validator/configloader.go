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
	ErrBadPublicKey = errors.New("Public key must be a non-null base64 encoded string")

	// ErrNoPKI is returned when rules.json doesn't contain a `pki` field
	ErrNoPKI = errors.New("rules.json needs a 'pki' field to list authorized public keys")
)

type rulesSchema struct {
	PKI        json.RawMessage `json:"pki"`
	Validators json.RawMessage `json:"validators"`
}

// LoadConfig loads the validators configuration from a json file.
// The configuration returned can be then be used in NewMultiValidator().
func LoadConfig(path string) (*MultiValidatorConfig, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var rules rulesSchema
	err = json.Unmarshal(data, &rules)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	pki, err := loadPKIConfig(rules.PKI)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return loadValidatorsConfig(rules.Validators, pki)
}

// PKI maps a public key to an identity.
// It lists all legimate keys, assign real names to public keys
// and establishes n-to-n relationships between users and roles.
type PKI map[string]*Identity

// Identity represents an actor of an indigo network
type Identity struct {
	Name  string
	Roles []string
}

// loadPKIConfig deserializes json into a PKI struct. It checks that public keys
func loadPKIConfig(data json.RawMessage) (*PKI, error) {
	if len(data) == 0 {
		return nil, ErrNoPKI
	}

	var jsonData PKI
	err := json.Unmarshal(data, &jsonData)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	for key := range jsonData {
		if _, err := base64.StdEncoding.DecodeString(key); key == "" || err != nil {
			return nil, errors.Wrap(ErrBadPublicKey, "Error while parsing PKI")
		}
	}
	return &jsonData, nil
}

type jsonValidatorData []struct {
	ID         string           `json:"id"`
	Type       string           `json:"type"`
	Signatures []string         `json:"signatures"`
	Schema     *json.RawMessage `json:"schema"`
}

func loadValidatorsConfig(data json.RawMessage, pki *PKI) (*MultiValidatorConfig, error) {
	var jsonStruct map[string]jsonValidatorData
	err := json.Unmarshal(data, &jsonStruct)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var validatorConfig MultiValidatorConfig
	for process, jsonSchemaData := range jsonStruct {
		for _, val := range jsonSchemaData {
			if val.ID == "" {
				return nil, ErrMissingIdentifier
			}
			if val.Type == "" {
				return nil, ErrMissingLinkType
			}
			if len(val.Signatures) == 0 && val.Schema == nil {
				return nil, ErrInvalidValidator
			}

			if len(val.Signatures) > 0 {
				cfg, err := newPkiValidatorConfig(process, val.ID, val.Type, val.Signatures, pki)
				if err != nil {
					return nil, err
				}
				validatorConfig.PkiConfigs = append(validatorConfig.PkiConfigs, cfg)
			}

			if val.Schema != nil {
				schemaData, _ := val.Schema.MarshalJSON()
				cfg, err := newSchemaValidatorConfig(process, val.ID, val.Type, schemaData)
				if err != nil {
					return nil, err
				}
				validatorConfig.SchemaConfigs = append(validatorConfig.SchemaConfigs, cfg)
			}

		}
	}

	return &validatorConfig, nil
}
