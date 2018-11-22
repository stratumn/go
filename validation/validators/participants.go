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

	"github.com/pkg/errors"
	"github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-core/monitoring/errorcode"
	"github.com/stratumn/go-core/store"
	"github.com/stratumn/go-core/types"
)

const (
	// ParticipantsValidatorName name for monitoring.
	ParticipantsValidatorName = "participants-validator"

	// ParticipantsMap is the ID of the map containing participants in the
	// governance process.
	ParticipantsMap = "_participants"
)

// Allowed steps in the participants map.
const (
	ParticipantsAcceptStep = "accept"
	ParticipantsUpdateStep = "update"
	ParticipantsVoteStep   = "vote"
)

// Errors used by the participants validator.
var (
	ErrInvalidParticipantStep = errors.New("invalid step in network participants update")
)

// ParticipantsValidator validates changes to the governance participants list.
type ParticipantsValidator struct{}

// NewParticipantsValidator creates a new participants validator for the
// network.
// A participants validator is needed for decentralized networks that leverage
// governance to update processes' validation rules.
func NewParticipantsValidator() Validator {
	return &ParticipantsValidator{}
}

// Validate a participants update.
func (v *ParticipantsValidator) Validate(ctx context.Context, r store.SegmentReader, l *chainscript.Link) error {
	switch l.Meta.Step {
	case ParticipantsAcceptStep:
		panic("not implemented")
	case ParticipantsUpdateStep:
		panic("not implemented")
	case ParticipantsVoteStep:
		panic("not implemented")
	default:
		return types.WrapError(ErrInvalidParticipantStep, errorcode.InvalidArgument, ParticipantsValidatorName, "participants validation failed")
	}
}

// ShouldValidate returns true if the segment belongs to the participants map
// in the governance process.
func (v *ParticipantsValidator) ShouldValidate(l *chainscript.Link) bool {
	return l.Meta.Process.Name == GovernanceProcess && l.Meta.MapId == ParticipantsMap
}

// Hash returns an empty hash since ParticipantsValidator doesn't have any
// configuration (it works the same for every decentralized network).
func (v *ParticipantsValidator) Hash() ([]byte, error) {
	return nil, nil
}
