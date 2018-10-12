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
	// RefsValidatorName for monitoring.
	RefsValidatorName = "refs-validator"
)

// Errors returned by the RefsValidator.
var (
	ErrMapIDMismatch   = errors.New("mapID doesn't match previous link")
	ErrParentNotFound  = errors.New("parent is missing from store")
	ErrProcessMismatch = errors.New("process doesn't match referenced link")
	ErrRefNotFound     = errors.New("reference is missing from store")
)

// RefsValidator validates link references (parent and refs).
type RefsValidator struct{}

// NewRefsValidator creates a new RefsValidator.
func NewRefsValidator() Validator {
	return &RefsValidator{}
}

// Validate all references (parent and refs).
func (v *RefsValidator) Validate(ctx context.Context, r store.SegmentReader, l *chainscript.Link) error {
	if err := v.validateParent(ctx, r, l); err != nil {
		ctx, _ = tag.New(ctx, tag.Upsert(linkErr, "Parent"))
		stats.Record(ctx, linksErr.M(1))
		return err
	}

	if err := v.validateReferences(ctx, r, l); err != nil {
		ctx, _ = tag.New(ctx, tag.Upsert(linkErr, "Refs"))
		stats.Record(ctx, linksErr.M(1))
		return err
	}

	return nil
}

func (v *RefsValidator) validateParent(ctx context.Context, r store.SegmentReader, l *chainscript.Link) error {
	if len(l.PrevLinkHash()) == 0 {
		return nil
	}

	parent, err := r.GetSegment(ctx, l.PrevLinkHash())
	if err != nil {
		return types.WrapError(err, errorcode.NotFound, RefsValidatorName, "parent validation failed")
	}

	if parent == nil || parent.Link == nil {
		return types.WrapError(ErrParentNotFound, errorcode.NotFound, RefsValidatorName, "parent validation failed")
	}

	if parent.Link.Meta.Process.Name != l.Meta.Process.Name {
		return types.WrapError(ErrProcessMismatch, errorcode.InvalidArgument, RefsValidatorName, "parent validation failed")
	}

	if parent.Link.Meta.MapId != l.Meta.MapId {
		return types.WrapError(ErrMapIDMismatch, errorcode.InvalidArgument, RefsValidatorName, "parent validation failed")
	}

	if parent.Link.Meta.OutDegree == 0 {
		return types.WrapError(chainscript.ErrOutDegree, errorcode.FailedPrecondition, RefsValidatorName, "parent validation failed")
	}

	if parent.Link.Meta.OutDegree > 0 {
		children, err := r.FindSegments(ctx, &store.SegmentFilter{
			Pagination:   store.Pagination{Limit: 1},
			PrevLinkHash: l.PrevLinkHash(),
		})
		if err != nil {
			return types.WrapError(err, errorcode.NotFound, RefsValidatorName, "parent validation failed")
		}

		if int(parent.Link.Meta.OutDegree) <= children.TotalCount {
			return types.WrapError(chainscript.ErrOutDegree, errorcode.FailedPrecondition, RefsValidatorName, "parent validation failed")
		}
	}

	return nil
}

func (v *RefsValidator) validateReferences(ctx context.Context, r store.SegmentReader, l *chainscript.Link) error {
	if len(l.Meta.Refs) == 0 {
		return nil
	}

	var lhs []chainscript.LinkHash
	for _, ref := range l.Meta.Refs {
		lhs = append(lhs, ref.LinkHash)
	}

	segments, err := r.FindSegments(ctx, &store.SegmentFilter{
		Pagination: store.Pagination{Limit: len(lhs)},
		LinkHashes: lhs,
	})
	if err != nil {
		return types.WrapError(err, errorcode.NotFound, RefsValidatorName, "refs validation failed")
	}

	if len(segments.Segments) != len(l.Meta.Refs) {
		return types.WrapError(ErrRefNotFound, errorcode.NotFound, RefsValidatorName, "refs validation failed")
	}

	for _, ref := range l.Meta.Refs {
		found := false

		for _, s := range segments.Segments {
			if bytes.Equal(ref.LinkHash, s.LinkHash()) {
				found = true
				if ref.Process != s.Link.Meta.Process.Name {
					return types.WrapError(ErrProcessMismatch, errorcode.InvalidArgument, RefsValidatorName, "refs validation failed")
				}

				break
			}
		}

		if !found {
			return types.WrapError(ErrRefNotFound, errorcode.NotFound, RefsValidatorName, "refs validation failed")
		}
	}

	return nil
}

// ShouldValidate always evaluates to true, as all links should validate their
// references.
func (v *RefsValidator) ShouldValidate(*chainscript.Link) bool {
	return true
}

// Hash returns an empty hash since RefsValidator doesn't have any state.
func (v *RefsValidator) Hash() ([]byte, error) {
	return nil, nil
}
