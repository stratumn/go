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

	cj "github.com/gibson042/canonicaljson-go"
	"github.com/pkg/errors"
	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/store"
	"github.com/stratumn/sdk/types"
)

// MultiValidatorConfig sets the behavior of the validator.
// Its hash can be used to know which validations were applied to a block.
type MultiValidatorConfig struct {
	SchemaConfigs []*schemaValidatorConfig
	PkiConfigs    []*pkiValidatorConfig
}

type multiValidator struct {
	config           *MultiValidatorConfig
	validators       []childValidator
	defaultValidator childValidator
}

// NewMultiValidator creates a validator that will simply be a collection
// of single-purpose validators.
// The configuration should be loaded from a JSON file via validator.LoadConfig().
func NewMultiValidator(config *MultiValidatorConfig) Validator {
	if config == nil {
		return &multiValidator{}
	}

	var v []childValidator
	for _, schemaCfg := range config.SchemaConfigs {
		v = append(v, newSchemaValidator(schemaCfg))
	}
	for _, pkiCfg := range config.PkiConfigs {
		v = append(v, newPkiValidator(pkiCfg))
	}
	defaultValidator := newSignatureValidator(&signatureValidatorConfig{})

	return &multiValidator{
		config:           config,
		validators:       v,
		defaultValidator: defaultValidator,
	}
}

func (v multiValidator) Hash() (*types.Bytes32, error) {
	b, err := cj.Marshal(v.config)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	validationsHash := types.Bytes32(sha256.Sum256(b))
	return &validationsHash, nil
}

func (v multiValidator) matchValidators(l *cs.Link) (linkValidators []childValidator) {
	for _, child := range v.validators {
		if child.ShouldValidate(l) {
			linkValidators = append(linkValidators, child)
		}
	}
	return
}

// Validate runs the validation on every child validator matching the provided link.
// It is the multiValidator's responsability to call child.ShouldValidate() before running the validation.
func (v multiValidator) Validate(r store.SegmentReader, l *cs.Link) error {
	linkValidators := v.matchValidators(l)
	for _, child := range linkValidators {
		err := child.Validate(r, l)
		if err != nil {
			return err
		}
	}

	return v.defaultValidator.Validate(r, l)
}
