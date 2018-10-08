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
)

// Errors used by the multi-validator.
var (
	ErrNoMatchingValidator = errors.New("link does not match any validator")
)

// MultiValidator is a collection of validators.
type MultiValidator struct {
	validators Validators
}

// NewMultiValidator creates a validator that will simply be a collection of
// single-purpose validators.
func NewMultiValidator(validators Validators) Validator {
	return &MultiValidator{validators: validators}
}

// ShouldValidate returns true if at least one of the children matches.
func (v MultiValidator) ShouldValidate(link *chainscript.Link) bool {
	for _, child := range v.validators {
		if child.ShouldValidate(link) {
			return true
		}
	}

	return false
}

// Hash concatenates the hashes from its children and hashes the result.
func (v MultiValidator) Hash() ([]byte, error) {
	var toHash []byte
	for _, child := range v.validators {
		childHash, err := child.Hash()
		if err != nil {
			return nil, err
		}

		toHash = append(toHash, childHash...)
	}

	h := sha256.Sum256(toHash)
	return h[:], nil
}

// Validate forwards the link to every child validator that matches.
func (v MultiValidator) Validate(ctx context.Context, r store.SegmentReader, l *chainscript.Link) error {
	validated := false

	for _, child := range v.validators {
		if child.ShouldValidate(l) {
			validated = true
			err := child.Validate(ctx, r, l)
			if err != nil {
				return err
			}
		}
	}

	if !validated {
		return ErrNoMatchingValidator
	}

	return nil
}
