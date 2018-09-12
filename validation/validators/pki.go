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

package validators

import (
	"context"
	"crypto/sha256"
	"strings"

	cj "github.com/gibson042/canonicaljson-go"
	"github.com/pkg/errors"
	"github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/types"
)

// PKI maps a public key to an identity.
// It lists all legimate keys, assign real names to public keys
// and establishes n-to-n relationships between users and roles.
type PKI map[string]*Identity

func (p PKI) getIdentityByPublicKey(publicKey string) *Identity {
	for _, identity := range p {
		for _, key := range identity.Keys {
			if key == publicKey {
				return identity
			}
		}
	}
	return nil
}

func (p PKI) matchRequirement(requirement, publicKey string) bool {
	if requirement == publicKey {
		return true
	}

	identity := p.getIdentityByPublicKey(publicKey)
	if identity == nil {
		return false
	}

	if required, ok := p[requirement]; ok && identity == required {
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
	Keys  []string
	Roles []string
}

// PKIValidator validates the json signature of a link's state.
type PKIValidator struct {
	Config             *ValidatorBaseConfig
	RequiredSignatures []string
	PKI                *PKI
}

// NewPKIValidator returns a new PKIValidator
func NewPKIValidator(baseConfig *ValidatorBaseConfig, required []string, pki *PKI) Validator {
	return &PKIValidator{
		Config:             baseConfig,
		RequiredSignatures: required,
		PKI:                pki,
	}
}

// Hash implements github.com/stratumn/go-indigocore/validation/validators.Validator.Hash.
func (pv PKIValidator) Hash() (*types.Bytes32, error) {
	b, err := cj.Marshal(pv)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	validationsHash := types.Bytes32(sha256.Sum256(b))
	return &validationsHash, nil
}

// ShouldValidate implements github.com/stratumn/go-indigocore/validation/validators.Validator.ShouldValidate.
func (pv PKIValidator) ShouldValidate(link *chainscript.Link) bool {
	return pv.Config.ShouldValidate(link)
}

// Validate implements github.com/stratumn/go-indigocore/validation/validators.Validator.Validate.
// it checks that the provided signatures match the required ones.
// a requirement can either be: a public key, a name defined in PKI, a role defined in PKI.
func (pv PKIValidator) Validate(_ context.Context, _ store.SegmentReader, link *chainscript.Link) error {
	for _, required := range pv.RequiredSignatures {
		fulfilled := false
		for _, sig := range link.Signatures {
			if pv.PKI.matchRequirement(required, string(sig.PublicKey)) {
				fulfilled = true
				break
			}
		}

		if !fulfilled {
			return errors.Errorf("Missing signatory for validator %s of process %s: signature from %s is required", pv.Config.LinkStep, pv.Config.Process, required)
		}
	}

	return nil
}
