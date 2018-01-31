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
	"fmt"

	"github.com/pkg/errors"
	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/store"

	"github.com/xeipuuv/gojsonschema"
)

// schemaValidatorConfig contains everything a schemaValidator needs to
// validate links.
type schemaValidatorConfig struct {
	*validatorBaseConfig
	Schema *gojsonschema.Schema
}

// newSchemaValidatorConfig creates a schemaValidatorConfig for a given process and type.
func newSchemaValidatorConfig(process, linkType string, schemaData []byte) (*schemaValidatorConfig, error) {
	baseConfig, err := newValidatorBaseConfig(process, linkType)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	schema, err := gojsonschema.NewSchema(gojsonschema.NewBytesLoader(schemaData))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &schemaValidatorConfig{
		validatorBaseConfig: baseConfig,
		Schema:              schema,
	}, nil
}

// schemaValidator validates the json schema of a link's state.
type schemaValidator struct {
	config *schemaValidatorConfig
}

func newSchemaValidator(config *schemaValidatorConfig) validator {
	return &schemaValidator{config: config}
}

// Validate validates the schema of a link's state.
func (sv schemaValidator) Validate(_ store.SegmentReader, link *cs.Link) error {
	if !sv.config.shouldValidate(link) {
		return nil
	}

	stateBytes, err := json.Marshal(link.State)
	if err != nil {
		return errors.WithStack(err)
	}

	stateData := gojsonschema.NewBytesLoader(stateBytes)
	result, err := sv.config.Schema.Validate(stateData)
	if err != nil {
		return errors.WithStack(err)
	}

	if !result.Valid() {
		return fmt.Errorf("link validation failed: %s", result.Errors())
	}

	return nil
}
