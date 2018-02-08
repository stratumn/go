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

// schemaValidator validates the json schema of a link's state.
type schemaValidator struct {
	Config *validatorBaseConfig
	Schema *gojsonschema.Schema
}

func newSchemaValidator(baseConfig *validatorBaseConfig, schemaData []byte) (ChildValidator, error) {
	schema, err := gojsonschema.NewSchema(gojsonschema.NewBytesLoader(schemaData))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &schemaValidator{
		Config: baseConfig,
		Schema: schema,
	}, nil
}
func (sv schemaValidator) ShouldValidate(link *cs.Link) bool {
	return sv.Config.ShouldValidate(link)
}

// Validate validates the schema of a link's state.
func (sv schemaValidator) Validate(_ store.SegmentReader, link *cs.Link) error {
	stateBytes, err := json.Marshal(link.State)
	if err != nil {
		return errors.WithStack(err)
	}

	stateData := gojsonschema.NewBytesLoader(stateBytes)
	result, err := sv.Schema.Validate(stateData)
	if err != nil {
		return errors.WithStack(err)
	}

	if !result.Valid() {
		return fmt.Errorf("link validation failed: %s", result.Errors())
	}

	return nil
}
