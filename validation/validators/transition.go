// Copyright 2016-2018 Stratumn SAS. All rights reserved.
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
	"fmt"

	"github.com/pkg/errors"
	"github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-core/store"

	"go.opencensus.io/stats"
	"go.opencensus.io/tag"
)

// Errors used by the transition validator.
var (
	ErrInvalidTransition = errors.New("unauthorized process step transition")
)

// TransitionValidator restricts the transitions from a step to another.
// It checks that the parent step was in the list of authorized steps.
type TransitionValidator struct {
	*ProcessStepValidator
	from []string
}

// NewTransitionValidator returns a new TransitionValidator for the given
// process and step.
func NewTransitionValidator(processStepValidator *ProcessStepValidator, from []string) Validator {
	return &TransitionValidator{
		ProcessStepValidator: processStepValidator,
		from:                 from,
	}
}

// Hash the process, step and allowed previous steps.
func (tv TransitionValidator) Hash() ([]byte, error) {
	psh, err := tv.ProcessStepValidator.Hash()
	if err != nil {
		return nil, err
	}

	toHash := psh[:]
	for _, t := range tv.from {
		toHash = append(toHash, []byte(t)...)
	}

	h := sha256.Sum256(toHash)
	return h[:], nil
}

// Validate that the link's new step follows an authorized transition.
// If there is no previous link the allowed transitions must explicitly contain
// an empty string.
func (tv TransitionValidator) Validate(ctx context.Context, store store.SegmentReader, link *chainscript.Link) error {
	error := func(src string) error {
		ctx, _ = tag.New(ctx, tag.Upsert(linkErr, "Transition"))
		stats.Record(ctx, linksErr.M(1))
		return errors.Wrapf(ErrInvalidTransition, "%s --> %s", src, tv.step)
	}

	prevLinkHash := link.PrevLinkHash()
	if len(prevLinkHash) == 0 {
		for _, t := range tv.from {
			if t == "" {
				return nil
			}
		}

		return error("()")
	}

	parent, err := store.GetSegment(ctx, prevLinkHash)
	if err != nil {
		ctx, _ = tag.New(ctx, tag.Upsert(linkErr, "TransitionParentErr"))
		stats.Record(ctx, linksErr.M(1))
		return errors.Wrapf(err, "cannot retrieve previous segment %s", prevLinkHash.String())
	}
	if parent == nil {
		ctx, _ = tag.New(ctx, tag.Upsert(linkErr, "TransitionParentNil"))
		stats.Record(ctx, linksErr.M(1))
		return fmt.Errorf("previous segment not found: %s", prevLinkHash.String())
	}

	for _, t := range tv.from {
		if t == parent.Link.Meta.Step {
			return nil
		}
	}

	return error(parent.Link.Meta.Step)
}
