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
	"strings"

	cj "github.com/gibson042/canonicaljson-go"
	jmespath "github.com/jmespath/go-jmespath"
	"github.com/pkg/errors"

	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/store"
	"github.com/stratumn/sdk/validator/signature"
)

var (

	// ErrMissingSignature is returned when there are no signatures in the link.
	ErrMissingSignature = errors.New("signature validation requires link.signatures to contain at least one element")

	// ErrEmptyPayload is returned when the JMESPATH query didn't match any element of the link.
	ErrEmptyPayload = errors.New("JMESPATH query does not match any link data")
)

// signatureValidatorConfig contains everything a signatureValidator needs to
// validate links.
type signatureValidatorConfig struct {
	*validatorBaseConfig
	requiredSignatures []string
	pki                *PKI
}

// newSignatureValidatorConfig creates a signatureValidatorConfig for a given process and type.
func newSignatureValidatorConfig(process, id, linkType string, required []string, pki *PKI) (*signatureValidatorConfig, error) {
	baseConfig, err := newValidatorBaseConfig(process, id, linkType)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &signatureValidatorConfig{
		validatorBaseConfig: baseConfig,
		requiredSignatures:  required,
		pki:                 pki,
	}, nil
}

// signatureValidator validates the json signature of a link's state.
type signatureValidator struct {
	*signatureValidatorConfig
}

func newSignatureValidator(config *signatureValidatorConfig) validator {
	return &signatureValidator{config}
}

func (sv signatureValidator) matchRequirement(requirement, publicKey string) bool {
	if requirement == publicKey {
		return true
	}

	identity, ok := (*sv.pki)[publicKey]
	if !ok {
		return false
	}
	if strings.EqualFold(identity.Name, requirement) {
		return true
	}
	for _, role := range identity.Roles {
		if strings.EqualFold(role, requirement) {
			return true
		}
	}
	return false
}

func (sv signatureValidator) isSignatureRequired(publicKey string) bool {
	for _, required := range sv.requiredSignatures {
		if sv.matchRequirement(required, publicKey) {
			return true
		}
	}
	return false
}

// missingSignatory checks that the provided dignatures match the required ones.
// a requirement can either be : a public key, a name defined in PKI, a role defined in PKI
func (sv signatureValidator) missingSignatory(signatures cs.Signatures) error {
	for _, required := range sv.requiredSignatures {
		fulfilled := false
		for _, sig := range signatures {
			if sv.matchRequirement(required, sig.PublicKey) {
				fulfilled = true
				break
			}
		}
		if !fulfilled {
			return errors.Errorf("A signature from %s is required", required)
		}
	}
	return nil
}

// Validate validates the signature of a link's state.
func (sv signatureValidator) Validate(_ store.SegmentReader, link *cs.Link) error {
	if !sv.shouldValidate(link) {
		return nil
	}

	if len(link.Signatures) == 0 {
		return ErrMissingSignature
	}

	if err := sv.missingSignatory(link.Signatures); err != nil {
		return errors.Wrapf(err, "Missing signatory for validator %s", sv.ID)
	}

	for _, sig := range link.Signatures {

		// don't check decoding errors here, this is done in link.Validate() beforehand
		keyBytes, _ := base64.StdEncoding.DecodeString(sig.PublicKey)
		sigBytes, _ := base64.StdEncoding.DecodeString(sig.Signature)

		payload, err := jmespath.Search(sig.Payload, link)
		if err != nil {
			return errors.Wrapf(err, "failed to execute jmespath query")
		}
		if payload == nil {
			return ErrEmptyPayload
		}

		payloadBytes, err := cj.Marshal(payload)
		if err != nil {
			return errors.WithStack(err)
		}

		if err := signature.Verify(sig.Type, keyBytes, sigBytes, payloadBytes); err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}
