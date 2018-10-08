// Copyright 2017-2018 Stratumn SAS. All rights reserved.
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

package validators

import (
	"context"
	"crypto/sha256"

	"github.com/pkg/errors"
	"github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-core/store"
	"github.com/xeipuuv/gojsonschema"
)

// Errors used by the schema validator.
var (
	ErrInvalidLinkSchema = errors.New("invalid link schema")
)

// SchemaValidator validates the json schema of a link's data.
type SchemaValidator struct {
	*ProcessStepValidator

	schema     *gojsonschema.Schema
	schemaHash []byte
}

// NewSchemaValidator returns a new SchemaValidator.
func NewSchemaValidator(processStepValidator *ProcessStepValidator, schemaData []byte) (Validator, error) {
	schema, err := gojsonschema.NewSchema(gojsonschema.NewBytesLoader(schemaData))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	schemaHash := sha256.Sum256(schemaData)
	return &SchemaValidator{
		ProcessStepValidator: processStepValidator,
		schema:               schema,
		schemaHash:           schemaHash[:],
	}, nil
}

// Hash the process, step and expected schema.
func (sv SchemaValidator) Hash() ([]byte, error) {
	psh, err := sv.ProcessStepValidator.Hash()
	if err != nil {
		return nil, err
	}

	h := sha256.Sum256(append(psh[:], sv.schemaHash...))
	return h[:], nil
}

// Validate the schema of a link's data.
func (sv SchemaValidator) Validate(_ context.Context, _ store.SegmentReader, link *chainscript.Link) error {
	linkData := gojsonschema.NewBytesLoader(link.Data)
	result, err := sv.schema.Validate(linkData)
	if err != nil {
		return errors.WithStack(err)
	}

	if !result.Valid() {
		return errors.Wrapf(ErrInvalidLinkSchema, "%s", result.Errors())
	}

	return nil
}
