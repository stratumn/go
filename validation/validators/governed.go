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
	"github.com/stratumn/go-core/store"
)

// GovernedProcessValidator validates a segment against the latest voted rules
// in a p2p decentralized network.
type GovernedProcessValidator struct{}

// NewGovernedProcessValidator creates a segment validator.
func NewGovernedProcessValidator() Validator {
	return &GovernedProcessValidator{}
}

// Validate the segment against the latest voted rules.
func (v *GovernedProcessValidator) Validate(context.Context, store.SegmentReader, *chainscript.Link) error {
	// TODO:
	//	- get latest validation rules for the segment process
	//	- if none, reject segment
	//	- otherwise build children validators based on the validation rules and run them
	return errors.New("not implemented")
}

// ShouldValidate returns true if the segment is a process segment (i.e. not
// a segment from special administrative processes like governance).
func (v *GovernedProcessValidator) ShouldValidate(l *chainscript.Link) bool {
	return l.Meta.Process.Name != GovernanceProcess
}

// Hash returns an empty hash since the validator doesn't have any
// configuration (it works the same for every decentralized network).
func (v *GovernedProcessValidator) Hash() ([]byte, error) {
	return nil, nil
}
