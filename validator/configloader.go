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
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
)

var (
	// ErrMissingSchema is returned when the schema is missing for schema validation.
	ErrMissingSchema = errors.New("schema validation requires a schema")
)

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

	schemaValidators, err := loadSchemaValidatorsConfig(data)
	if err != nil {
		return nil, err
	}

	return &MultiValidatorConfig{SchemaConfigs: schemaValidators}, nil
}

type jsonSchemaData []struct {
	Type   string           `json:"type"`
	Schema *json.RawMessage `json:"schema"`
}

func loadSchemaValidatorsConfig(data []byte) ([]*schemaValidatorConfig, error) {
	var jsonStruct map[string]jsonSchemaData
	err := json.Unmarshal(data, &jsonStruct)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var schemaValidators []*schemaValidatorConfig
	for process, jsonSchemaData := range jsonStruct {
		for _, val := range jsonSchemaData {
			if val.Schema == nil {
				return nil, ErrMissingSchema
			}

			if val.Type == "" {
				return nil, ErrMissingLinkType
			}

			schemaData, _ := val.Schema.MarshalJSON()
			cfg, err := newSchemaValidatorConfig(process, val.Type, schemaData)
			if err != nil {
				return nil, err
			}

			schemaValidators = append(schemaValidators, cfg)
		}
	}

	return schemaValidators, nil
}
