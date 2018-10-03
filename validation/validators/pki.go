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
	"github.com/stratumn/go-core/store"
)

// Errors used by the PKI validator.
var (
	ErrMissingSignature = errors.New("missing mandatory signature")
)

// PKI maps a public key to an identity.
// It lists all legimate keys, assigns real names to public keys and
// establishes n-to-n relationships between users and roles.
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

// Identity represents an actor in a network.
type Identity struct {
	Keys  []string
	Roles []string
}

// PKIValidator validates the json signature requirements of a link's data.
type PKIValidator struct {
	*ProcessStepValidator

	RequiredSignatures []string
	PKI                *PKI
}

// NewPKIValidator returns a new PKIValidator.
func NewPKIValidator(processStepValidator *ProcessStepValidator, required []string, pki *PKI) Validator {
	return &PKIValidator{
		ProcessStepValidator: processStepValidator,
		RequiredSignatures:   required,
		PKI:                  pki,
	}
}

// Hash the signature requirements.
func (pv PKIValidator) Hash() ([]byte, error) {
	b, err := cj.Marshal(pv)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	h := sha256.Sum256(b)
	return h[:], nil
}

// Validate that the provided signatures match the required ones.
// A requirement can be:
//	* a public key
//	* a name defined in PKI
//	* a role defined in PKI
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
			return errors.Wrapf(ErrMissingSignature, "%s.%s requires a signature from %s", pv.process, pv.step, required)
		}
	}

	return nil
}
