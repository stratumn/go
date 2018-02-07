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
	"strings"

	"github.com/pkg/errors"

	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/store"
)

// PKI maps a public key to an identity.
// It lists all legimate keys, assign real names to public keys
// and establishes n-to-n relationships between users and roles.
type PKI map[string]*Identity

func (p PKI) matchRequirement(requirement, publicKey string) bool {
	if requirement == publicKey {
		return true
	}

	identity, ok := p[publicKey]
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

// Identity represents an actor of an indigo network
type Identity struct {
	Name  string
	Roles []string
}

// pkiValidatorConfig contains everything a pkiValidator needs to
// validate links.
type pkiValidatorConfig struct {
	*validatorBaseConfig
	requiredSignatures []string
	pki                *PKI
}

// newSignatureValidatorConfig creates a pkiValidatorConfig for a given process and type.
func newPkiValidatorConfig(process, id, linkType string, required []string, pki *PKI) (*pkiValidatorConfig, error) {
	baseConfig, err := newValidatorBaseConfig(process, id, linkType)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &pkiValidatorConfig{
		validatorBaseConfig: baseConfig,
		requiredSignatures:  required,
		pki:                 pki,
	}, nil
}

// pkiValidator validates the json signature of a link's state.
type pkiValidator struct {
	*pkiValidatorConfig
}

func newPkiValidator(config *pkiValidatorConfig) childValidator {
	return &pkiValidator{config}
}

func (pv pkiValidator) isSignatureRequired(publicKey string) bool {
	for _, required := range pv.requiredSignatures {
		if pv.pki.matchRequirement(required, publicKey) {
			return true
		}
	}
	return false
}

// Validate checks that the provided dignatures match the required ones.
// a requirement can either be: a public key, a name defined in PKI, a role defined in PKI.
func (pv pkiValidator) Validate(_ store.SegmentReader, link *cs.Link) error {
	for _, required := range pv.requiredSignatures {
		fulfilled := false
		for _, sig := range link.Signatures {
			if pv.pki.matchRequirement(required, sig.PublicKey) {
				fulfilled = true
				break
			}
		}
		if !fulfilled {
			return errors.Errorf("Missing signatory for validator %s: signature from %s is required", pv.ID, required)
		}
	}
	return nil
}
