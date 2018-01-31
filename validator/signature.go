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
	"github.com/pkg/errors"
	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/store"
)

var (
	// ErrMissingSignature is returned when there are no signatures in the link
	ErrMissingSignature = errors.New("signature validation requires link.signatures to contain at least one element")
)

// signatureValidatorConfig contains everything a signatureValidator needs to
// validate links.
type signatureValidatorConfig struct {
	*validatorBaseConfig
}

// newSignatureValidatorConfig creates a signatureValidatorConfig for a given process and type.
func newSignatureValidatorConfig(process, linkType string) (*signatureValidatorConfig, error) {
	baseConfig, err := newValidatorBaseConfig(process, linkType)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &signatureValidatorConfig{baseConfig}, nil
}

// signatureValidator validates the json signature of a link's state.
type signatureValidator struct {
	config *signatureValidatorConfig
}

func newSignatureValidator(config *signatureValidatorConfig) validator {
	return &signatureValidator{config: config}
}

// Validate validates the signature of a link's state.
func (sv signatureValidator) Validate(_ store.SegmentReader, link *cs.Link) error {
	if !sv.config.shouldValidate(link) {
		return nil
	}

	if len(link.Signatures) == 0 {
		return ErrMissingSignature
	}

	// TODO: check that
	// - signature type is supported
	// - signature is correct
	// - required signatures for this action are present/valid

	return nil
}
