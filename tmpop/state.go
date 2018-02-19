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

package tmpop

import (
	"crypto/sha256"
	"fmt"

	"github.com/stratumn/go-indigocore/bufferedbatch"
	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/merkle"
	"github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/types"
	"github.com/stratumn/go-indigocore/validator"
)

// State represents the app states, separating the committed state (for queries)
// from the working state (for CheckTx and DeliverTx).
type State struct {
	previousAppHash *types.Bytes32
	// The same validator is used for a whole commit
	// When beginning a new block, the validator can
	// be updated.
	validator validator.Validator

	adapter            store.Adapter
	deliveredLinks     store.Batch
	deliveredLinksList []*cs.Link
	checkedLinks       store.Batch

	governance *validator.GovernanceManager
}

// NewState creates a new State.
func NewState(a store.Adapter, config *Config) (*State, error) {
	deliveredLinks, err := a.NewBatch()
	if err != nil {
		return nil, err
	}

	// With transactional databases we cannot really use two transactions as they'd lock each other
	// (more exactly, checked links would lock out delivered links)
	checkedLinks := bufferedbatch.NewBatch(a)

	state := &State{
		adapter:        a,
		deliveredLinks: deliveredLinks,
		checkedLinks:   checkedLinks,
	}

	state.governance, err = validator.NewGovernanceManager(a, config.ValidatorFilename)
	if err != nil {
		return nil, err
	}

	return state, nil
}

// UpdateValidators updates validators if a new version is available
func (s *State) UpdateValidators() {
	s.governance.UpdateValidators(&s.validator)
}

// Check checks if creating this link is a valid operation
func (s *State) Check(link *cs.Link) *ABCIError {
	return s.checkLinkAndAddToBatch(link, s.checkedLinks)
}

// Deliver adds a link to the list of links to be committed
func (s *State) Deliver(link *cs.Link) *ABCIError {
	res := s.checkLinkAndAddToBatch(link, s.deliveredLinks)
	if res.IsOK() {
		s.deliveredLinksList = append(s.deliveredLinksList, link)
	}
	return res
}

// checkLinkAndAddToBatch validates the link's format and runs the validations (signatures, schema)
func (s *State) checkLinkAndAddToBatch(link *cs.Link, batch store.Batch) *ABCIError {
	err := link.Validate(batch.GetSegment)
	if err != nil {
		return &ABCIError{
			CodeTypeValidation,
			fmt.Sprintf("Link validation failed %v: %v", link, err),
		}
	}

	if s.validator != nil {
		err = s.validator.Validate(batch, link)
		if err != nil {
			return &ABCIError{
				CodeTypeValidation,
				fmt.Sprintf("Link validation rules failed %v: %v", link, err),
			}
		}
	}

	if _, err := batch.CreateLink(link); err != nil {
		return &ABCIError{
			CodeTypeInternalError,
			err.Error(),
		}
	}

	return nil
}

// Commit commits the delivered links,
// resets delivered and checked state,
// and returns the hash for the commit
// and the list of committed links.
func (s *State) Commit() (*types.Bytes32, []*cs.Link, error) {
	appHash, err := s.computeAppHash()
	if err != nil {
		return nil, nil, err
	}

	if err := s.deliveredLinks.Write(); err != nil {
		return nil, nil, err
	}

	if s.deliveredLinks, err = s.adapter.NewBatch(); err != nil {
		return nil, nil, err
	}
	s.checkedLinks = bufferedbatch.NewBatch(s.adapter)

	committedLinks := make([]*cs.Link, len(s.deliveredLinksList))
	copy(committedLinks, s.deliveredLinksList)
	s.deliveredLinksList = nil

	return appHash, committedLinks, nil
}

func (s *State) computeAppHash() (*types.Bytes32, error) {
	var validatorHash *types.Bytes32
	if s.validator != nil {
		h, err := s.validator.Hash()
		if err != nil {
			return nil, err
		}
		validatorHash = h
	}

	var merkleRoot *types.Bytes32
	if len(s.deliveredLinksList) > 0 {
		var treeLeaves []types.Bytes32
		for _, link := range s.deliveredLinksList {
			linkHash, _ := link.Hash()
			treeLeaves = append(treeLeaves, *linkHash)
		}

		merkle, err := merkle.NewStaticTree(treeLeaves)
		if err != nil {
			return nil, err
		}

		merkleRoot = merkle.Root()
	}

	return ComputeAppHash(s.previousAppHash, validatorHash, merkleRoot)
}

// ComputeAppHash computes the app hash from its required parts
// If one of the parts is nil or empty, we'll pad with 0s so that
// we always hash a 96-bytes array
func ComputeAppHash(previous *types.Bytes32, validator *types.Bytes32, root *types.Bytes32) (*types.Bytes32, error) {
	hash := sha256.New()

	if previous == nil {
		previous = &types.Bytes32{}
	}
	if _, err := hash.Write(previous[:]); err != nil {
		return nil, err
	}

	if validator == nil {
		validator = &types.Bytes32{}
	}
	if _, err := hash.Write(validator[:]); err != nil {
		return nil, err
	}

	if root == nil {
		root = &types.Bytes32{}
	}
	if _, err := hash.Write(root[:]); err != nil {
		return nil, err
	}

	appHash := hash.Sum(nil)
	var appHash32 types.Bytes32
	copy(appHash32[:], appHash)

	return &appHash32, nil
}
