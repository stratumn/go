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
	"bytes"
	"context"

	"github.com/pkg/errors"
	"github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-core/monitoring/errorcode"
	"github.com/stratumn/go-core/store"
	"github.com/stratumn/go-core/types"

	"go.opencensus.io/stats"
	"go.opencensus.io/tag"
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
	ErrParticipantsNotInitialized     = errors.New("participants map not initialized")
	ErrInvalidAcceptParticipant       = errors.New("invalid accept participant link")
	ErrInvalidUpdateParticipant       = errors.New("invalid update participant link")
	ErrInvalidVoteParticipant         = errors.New("invalid vote participant link")
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
	var err error
	switch l.Meta.Step {
	case ParticipantsAcceptStep:
		err = v.validateAccept(ctx, r, l)
	case ParticipantsUpdateStep:
		err = v.validateUpdate(ctx, r, l)
	case ParticipantsVoteStep:
		err = v.validateVote(ctx, r, l)
	default:
		err = types.WrapError(ErrInvalidParticipantStep, errorcode.InvalidArgument, ParticipantsValidatorName, "participants validation failed")
	}

	if err != nil {
		ctx, _ = tag.New(ctx, tag.Upsert(linkErr, ParticipantsValidatorName))
		stats.Record(ctx, linksErr.M(1))
	}

	return err
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
			Process:    GovernanceProcess,
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

func (v *ParticipantsValidator) validateUpdate(ctx context.Context, r store.SegmentReader, l *chainscript.Link) error {
	updates, err := v.validateUpdateStructure(l)
	if err != nil {
		return err
	}

	currentHash, current, err := v.getCurrentParticipants(ctx, r)
	if err != nil {
		return err
	}

	if !bytes.Equal(currentHash, l.Meta.Refs[0].LinkHash) {
		return types.WrapError(ErrInvalidUpdateParticipant, errorcode.FailedPrecondition, ParticipantsValidatorName, "update does not reference the latest accepted link")
	}

	for _, p := range updates {
		if err := p.Validate(current); err != nil {
			return types.WrapError(ErrInvalidParticipantData, errorcode.InvalidArgument, ParticipantsValidatorName, err.Error())
		}
	}

	return nil
}

func (v *ParticipantsValidator) validateUpdateStructure(l *chainscript.Link) ([]*ParticipantUpdate, error) {
	if len(l.Meta.PrevLinkHash) > 0 {
		return nil, types.WrapError(ErrInvalidUpdateParticipant, errorcode.InvalidArgument, ParticipantsValidatorName, "update link should have no parent")
	}

	if l.Meta.OutDegree >= 0 {
		return nil, types.WrapError(ErrInvalidUpdateParticipant, errorcode.InvalidArgument, ParticipantsValidatorName, "update link should have unlimited children")
	}

	if len(l.Meta.Refs) != 1 {
		return nil, types.WrapError(ErrInvalidUpdateParticipant, errorcode.InvalidArgument, ParticipantsValidatorName, "update link should reference an accept link")
	}

	var updates []*ParticipantUpdate
	err := l.StructurizeData(&updates)
	if err != nil {
		return nil, types.WrapError(ErrInvalidParticipantData, errorcode.InvalidArgument, ParticipantsValidatorName, "link should contain a list of participant updates")
	}

	if len(updates) == 0 {
		return nil, types.WrapError(ErrInvalidUpdateParticipant, errorcode.InvalidArgument, ParticipantsValidatorName, "link should contain at least one participant update")
	}

	return updates, nil
}

func (v *ParticipantsValidator) validateVote(ctx context.Context, r store.SegmentReader, l *chainscript.Link) error {
	if l.Meta.OutDegree != 0 {
		return types.WrapError(ErrInvalidVoteParticipant, errorcode.InvalidArgument, ParticipantsValidatorName, "link should not accept children")
	}

	if len(l.Meta.PrevLinkHash) == 0 {
		return types.WrapError(ErrInvalidVoteParticipant, errorcode.InvalidArgument, ParticipantsValidatorName, "link should have a parent")
	}

	if len(l.Signatures) == 0 {
		return types.WrapError(ErrInvalidVoteParticipant, errorcode.InvalidArgument, ParticipantsValidatorName, "link should contain a signature")
	}

	for _, s := range l.Signatures {
		if err := s.Validate(l); err != nil {
			return types.WrapError(ErrInvalidVoteParticipant, errorcode.InvalidArgument, ParticipantsValidatorName, err.Error())
		}
	}

	parent, err := r.GetSegment(ctx, l.Meta.PrevLinkHash)
	if err != nil {
		return types.WrapError(ErrInvalidVoteParticipant, errorcode.Unknown, ParticipantsValidatorName, err.Error())
	}

	if parent.Link.Meta.Step != ParticipantsUpdateStep {
		return types.WrapError(ErrInvalidVoteParticipant, errorcode.InvalidArgument, ParticipantsValidatorName, "parent should be an update proposal")
	}

	return nil
}

func (v *ParticipantsValidator) getCurrentParticipants(ctx context.Context, r store.SegmentReader) (chainscript.LinkHash, []*Participant, error) {
	// Since accepted segments have increasing priority, we only need to get
	// the last one.
	s, err := r.FindSegments(ctx, &store.SegmentFilter{
		Process:    GovernanceProcess,
		MapIDs:     []string{ParticipantsMap},
		Step:       ParticipantsAcceptStep,
		Pagination: store.Pagination{Limit: 1},
	})
	if err != nil {
		return nil, nil, types.WrapError(ErrInvalidUpdateParticipant, errorcode.Unknown, ParticipantsValidatorName, err.Error())
	}

	if len(s.Segments) == 0 {
		return nil, nil, types.WrapError(ErrParticipantsNotInitialized, errorcode.FailedPrecondition, ParticipantsValidatorName, "cannot get latest accepted link")
	}

	latest := s.Segments[0]

	var current []*Participant
	err = latest.Link.StructurizeData(&current)
	if err != nil {
		return nil, nil, types.WrapError(ErrInvalidUpdateParticipant, errorcode.Unknown, ParticipantsValidatorName, err.Error())
	}

	return latest.LinkHash(), current, nil
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
