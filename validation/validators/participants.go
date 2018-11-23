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
	ErrInvalidParticipantStep         = errors.New("invalid step in network participants update")
	ErrInvalidParticipantData         = errors.New("invalid participant data")
	ErrParticipantsAlreadyInitialized = errors.New("participants map already initialized")
	ErrInvalidAcceptParticipant       = errors.New("invalid accept participant link")
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
		return v.validateAccept(ctx, r, l)
	case ParticipantsUpdateStep:
		panic("not implemented")
	case ParticipantsVoteStep:
		panic("not implemented")
	default:
		return types.WrapError(ErrInvalidParticipantStep, errorcode.InvalidArgument, ParticipantsValidatorName, "participants validation failed")
	}
}

func (v *ParticipantsValidator) validateAccept(ctx context.Context, r store.SegmentReader, l *chainscript.Link) error {
	if l.Meta.OutDegree != 1 {
		return types.WrapError(ErrInvalidAcceptParticipant, errorcode.InvalidArgument, ParticipantsValidatorName, "link out degree should be 1")
	}

	var participants []*Participant
	err := l.StructurizeData(&participants)
	if err != nil {
		return types.WrapError(ErrInvalidParticipantData, errorcode.InvalidArgument, ParticipantsValidatorName, "link should contain a list of participants")
	}

	if len(participants) == 0 {
		return types.WrapError(ErrInvalidParticipantData, errorcode.InvalidArgument, ParticipantsValidatorName, "link should contain at least one participant")
	}

	for _, p := range participants {
		if err := p.Validate(); err != nil {
			return types.WrapError(ErrInvalidParticipantData, errorcode.InvalidArgument, ParticipantsValidatorName, err.Error())
		}
	}

	// If this is the first participants link, verify that the map has not
	// already been initialized.
	if len(l.PrevLinkHash()) == 0 {
		s, err := r.FindSegments(ctx, &store.SegmentFilter{
			MapIDs:     []string{ParticipantsMap},
			Pagination: store.Pagination{Limit: 1},
		})
		if err != nil {
			return types.WrapError(err, errorcode.Unknown, ParticipantsValidatorName, "could not get participants map")
		}

		if s.TotalCount > 0 {
			return types.WrapError(ErrParticipantsAlreadyInitialized, errorcode.FailedPrecondition, ParticipantsValidatorName, "cannot add accept link")
		}

		return nil
	}

	// Otherwise verify that the voting policy is enforced.
	panic("votes verification not implemented")
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
