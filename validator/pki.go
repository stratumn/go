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
	"strings"

	cj "github.com/gibson042/canonicaljson-go"
	"github.com/pkg/errors"

	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/types"
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

// pkiValidator validates the json signature of a link's state.
type pkiValidator struct {
	Config             *validatorBaseConfig
	requiredSignatures []string
	pki                *PKI
}

func newPkiValidator(baseConfig *validatorBaseConfig, required []string, pki *PKI) Validator {
	return &pkiValidator{
		Config:             baseConfig,
		requiredSignatures: required,
		pki:                pki,
	}
}

func (pv pkiValidator) Hash() (*types.Bytes32, error) {
	b, err := cj.Marshal(pv)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	validationsHash := types.Bytes32(sha256.Sum256(b))
	return &validationsHash, nil
}
func (pv pkiValidator) ShouldValidate(link *cs.Link) bool {
	return pv.Config.ShouldValidate(link)
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
			return errors.Errorf("Missing signatory for validator %s: signature from %s is required", pv.Config.ID, required)
		}
	}
	return nil
}
