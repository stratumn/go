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

	cj "github.com/gibson042/canonicaljson-go"
	"github.com/pkg/errors"

	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/types"
)

// TransitionValidator defines the source state where a transition can be applied and its destination state.
type TransitionValidator struct {
	Config      *ValidatorBaseConfig
	Transitions []string
}

// NewTransitionValidator returns a new TransitionValidator.
func NewTransitionValidator(baseConfig *ValidatorBaseConfig, transitions []string) Validator {
	return &TransitionValidator{
		Config:      baseConfig,
		Transitions: transitions,
	}
}

// Hash implements github.com/stratumn/go-indigocore/validation/validators.Validator.Hash.
func (tv TransitionValidator) Hash() (*types.Bytes32, error) {
	b, err := cj.Marshal(tv)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	validationsHash := types.Bytes32(sha256.Sum256(b))
	return &validationsHash, nil
}

// ShouldValidate implements github.com/stratumn/go-indigocore/validation/validators.Validator.ShouldValidate.
func (tv TransitionValidator) ShouldValidate(link *cs.Link) bool {
	return tv.Config.ShouldValidate(link)
}

// Validate implements github.com/stratumn/go-indigocore/validation/validators.Validator.Validate.
// It checks that the link follows a valid transition.
// If there is no previous link, an empty link has to be allowed,
// Otherwise the meta.type of the prevLink must exist in authorized previous statement.
func (tv TransitionValidator) Validate(ctx context.Context, store store.SegmentReader, link *cs.Link) error {
	error := func(src string) error {
		return errors.Errorf("no transition found %s --> %s", src, tv.Config.LinkType)
	}

	prevLinkHash := link.Meta.GetPrevLinkHash()
	if prevLinkHash == nil {
		for _, t := range tv.Transitions {
			if t == "" {
				return nil
			}
		}
		return error("()")
	}

	seg, err := store.GetSegment(ctx, prevLinkHash)
	if err != nil {
		return errors.Wrapf(err, "cannot retrieve previous segment %s", prevLinkHash.String())
	}
	if seg == nil {
		return errors.Errorf("previous segment not found: %s", prevLinkHash.String())
	}

	for _, t := range tv.Transitions {
		if t == seg.Link.Meta.Type {
			return nil
		}
	}
	return error(seg.Link.Meta.Type)
}
