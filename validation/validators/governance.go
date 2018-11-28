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

const (
	// GovernanceProcess is the name of the process containing validation rules
	// in a decentralized network.
	// Special rules apply to segments inside this process.
	GovernanceProcess = "_governance"
)

// GovernanceRulesValidator validates the evolution of custom rules in a p2p
// decentralized network.
// The governance process will contain one map per business process. This map
// will be responsible for updating the custom validation rules that apply to
// the process' segments.
// A governance process map should have the following structure:
//
// ,------------,                                         ,--------------,                                            ,----------------,
// |   accept   | <====================================== |    accept    | <========================================= |     accept     |
// | (rules v1) | <-.                                     |  (rules v2)  | <-.                                        |   (rules v3)   |
// `------------'   |    ,----------,      ,---------,    `--------------'   |                                        `----------------'
//                  |`-- |  update  |      |  vote   |           |           |                        ,---------,            |  |
//                  |    | rules v2 | <=== | (alice) | <---------'           |   ,-------------,      |  vote   |            |  |
//                  |    `----------'      `---------'                       |   |             | <=== | (alice) | <----------'  |
//                  |    ,----------,                                        |   |   update    |      `---------'               |
//                   `-- |  update  |                                         `--|  rules v3   |      ,---------,               |
//                       | rules v3 | ---- X                                     |             | <=== |  vote   |               |
//                       `----------'                                            `-------------'      |  (bob)  | <-------------'
//                                                                                                    `---------'
//
// where:
// <=== represents a parent relationship
// <--- represents a reference
//
// Each accept link should also reference the latest link in the participants
// map and check votes against these network participants.
type GovernanceRulesValidator struct{}

// NewGovernanceRulesValidator creates a validator for custom validation rules
// updates.
func NewGovernanceRulesValidator() Validator {
	return &GovernanceRulesValidator{}
}

// Validate an update to a process' validation rules.
func (v *GovernanceRulesValidator) Validate(context.Context, store.SegmentReader, *chainscript.Link) error {
	return errors.New("not implemented")
}

// ShouldValidate returns true if the segment is a process governance segment.
func (v *GovernanceRulesValidator) ShouldValidate(l *chainscript.Link) bool {
	return l.Meta.Process.Name == GovernanceProcess &&
		l.Meta.MapId != ParticipantsMap
}

// Hash returns an empty hash since the validator doesn't have any
// configuration (it works the same for every decentralized network).
func (v *GovernanceRulesValidator) Hash() ([]byte, error) {
	return nil, nil
}
