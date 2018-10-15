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

	"github.com/pkg/errors"
	"github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-core/monitoring/errorcode"
	"github.com/stratumn/go-core/store"
	"github.com/stratumn/go-core/types"

	"go.opencensus.io/stats"
	"go.opencensus.io/tag"
)

const (
	// StepValidatorName for monitoring.
	StepValidatorName = "step-validator"
)

// Errors used by the ProcessStepValidator.
var (
	ErrMissingProcess       = errors.New("validator requires a process")
	ErrMissingLinkStep      = errors.New("validator requires a link step")
	ErrInvalidProcessOrStep = errors.New("link process or step doesn't match expectations")
)

// ProcessStepValidator validates a link's process and step.
type ProcessStepValidator struct {
	process string
	step    string
}

// NewProcessStepValidator creates a new ProcessStepValidator for the given
// process and step.
func NewProcessStepValidator(process, step string) (*ProcessStepValidator, error) {
	if len(process) == 0 {
		return nil, types.WrapError(ErrMissingProcess, errorcode.InvalidArgument, StepValidatorName, "could not create step validator")
	}

	if len(step) == 0 {
		return nil, types.WrapError(ErrMissingLinkStep, errorcode.InvalidArgument, StepValidatorName, "could not create step validator")
	}

	return &ProcessStepValidator{process: process, step: step}, nil
}

// ShouldValidate returns true if the link matches the validator's process
// and type. Otherwise the link is considered valid because this validator
// doesn't apply to it.
func (v *ProcessStepValidator) ShouldValidate(link *chainscript.Link) bool {
	if link == nil || link.Meta == nil || link.Meta.Process == nil {
		return false
	}

	if link.Meta.Process.Name != v.process {
		return false
	}

	if link.Meta.Step != v.step {
		return false
	}

	return true
}

// Validate that the process and step match the configured values.
func (v *ProcessStepValidator) Validate(ctx context.Context, _ store.SegmentReader, link *chainscript.Link) error {
	if !v.ShouldValidate(link) {
		ctx, _ = tag.New(ctx, tag.Upsert(linkErr, StepValidatorName))
		stats.Record(ctx, linksErr.M(1))
		return types.WrapError(ErrInvalidProcessOrStep, errorcode.InvalidArgument, StepValidatorName, "step validation failed")
	}

	return nil
}

// Hash the process and step.
func (v *ProcessStepValidator) Hash() ([]byte, error) {
	h := sha256.Sum256(append([]byte(v.process), []byte(v.step)...))
	return h[:], nil
}
