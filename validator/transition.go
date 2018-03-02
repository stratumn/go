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

package validator

import (
	"crypto/sha256"

	cj "github.com/gibson042/canonicaljson-go"
	"github.com/pkg/errors"

	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/types"
)

type transitionSources = []string

type allowedTransitions = map[string]transitionSources

// transitionValidator defines the source state where a transition can be applied and its destination state.
type transitionValidator struct {
	Config      *validatorBaseConfig
	Transitions allowedTransitions
}

func newTransitionValidator(baseConfig *validatorBaseConfig, transitions allowedTransitions) Validator {
	return &transitionValidator{
		Config:      baseConfig,
		Transitions: transitions,
	}
}

func (tv transitionValidator) Hash() (*types.Bytes32, error) {
	b, err := cj.Marshal(tv)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	validationsHash := types.Bytes32(sha256.Sum256(b))
	return &validationsHash, nil
}

func (tv transitionValidator) ShouldValidate(link *cs.Link) bool {
	return tv.Config.ShouldValidate(link)
}

// Validate checks that the provided dignatures match the required ones.
// a requirement can either be: a public key, a name defined in PKI, a role defined in PKI.
func (tv transitionValidator) Validate(store store.SegmentReader, link *cs.Link) error {
	error := func(src string) error {
		return errors.Errorf("no transition found %s --%s--> %s", src, link.Meta.Action, tv.Config.LinkType)
	}

	transitions, found := tv.Transitions[link.Meta.Action]
	if !found {
		return error("###")
	}

	prevLinkHash := link.Meta.GetPrevLinkHash()
	if prevLinkHash == nil {
		for _, t := range transitions {
			if t == "" {
				return nil
			}
		}
		return error("()")
	}

	seg, err := store.GetSegment(prevLinkHash)
	if err != nil {
		return errors.Wrapf(err, "cannot retrieve previous segment %s", prevLinkHash.String())
	}
	if seg == nil {
		return errors.Errorf("previous segment not found: %s", prevLinkHash.String())
	}

	for _, t := range transitions {
		if t == seg.Link.Meta.Type {
			return nil
		}
	}
	return error(seg.Link.Meta.Type)
}
