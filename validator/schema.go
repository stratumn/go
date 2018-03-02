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
	"crypto/sha256"
	"encoding/json"
	"fmt"

	cj "github.com/gibson042/canonicaljson-go"
	"github.com/pkg/errors"
	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/types"

	"github.com/xeipuuv/gojsonschema"
)

// schemaValidator validates the json schema of a link's state.
type schemaValidator struct {
	Config     *validatorBaseConfig
	schema     *gojsonschema.Schema
	SchemaHash types.Bytes32
}

func newSchemaValidator(baseConfig *validatorBaseConfig, schemaData []byte) (Validator, error) {
	schema, err := gojsonschema.NewSchema(gojsonschema.NewBytesLoader(schemaData))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &schemaValidator{
		Config:     baseConfig,
		schema:     schema,
		SchemaHash: types.Bytes32(sha256.Sum256(schemaData)),
	}, nil
}

func (sv schemaValidator) Hash() (*types.Bytes32, error) {
	b, err := cj.Marshal(sv)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	validationsHash := types.Bytes32(sha256.Sum256(b))
	return &validationsHash, nil
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
	result, err := sv.schema.Validate(stateData)
	if err != nil {
		return errors.WithStack(err)
	}

	if !result.Valid() {
		return fmt.Errorf("link validation failed: %s", result.Errors())
	}

	return nil
}
