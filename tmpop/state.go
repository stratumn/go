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
	"sync"

	"github.com/stratumn/sdk/bufferedbatch"
	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/store"
	"github.com/stratumn/sdk/validator"
	abci "github.com/tendermint/abci/types"
)

// State represents the app states, separating the committed state (for queries)
// from the working state (for CheckTx and DeliverTx).
type State struct {
	previousAppHash []byte
	// The same validator is used for a whole commit
	// When beginning a new block, the validator can
	// be updated.
	validator *validator.Validator

	adapter            store.AdapterV2
	deliveredLinks     store.BatchV2
	deliveredLinksList []*cs.Link
	checkedLinks       store.BatchV2

	notificationLock     sync.Mutex
	pendingNotifications []*store.Event
}

// NewState creates a new State.
func NewState(a store.AdapterV2) (*State, error) {
	deliveredLinks, err := a.NewBatchV2()
	if err != nil {
		return nil, err
	}

	// With transactional databases we cannot really use two transactions as they'd lock each other
	// (more exactly, checkedSegments would lock out deliveredSegments)
	checkedLinks := bufferedbatch.NewBatchV2(a)

	return &State{
		adapter:          a,
		deliveredLinks:   deliveredLinks,
		checkedLinks:     checkedLinks,
		notificationLock: sync.Mutex{},
	}, nil
}

// Check checks if creating this link is a valid operation
func (s *State) Check(link *cs.Link) abci.Result {
	return s.checkLinkAndAddToBatch(link, s.checkedLinks)
}

// Deliver adds a link to the list of links to be committed
func (s *State) Deliver(link *cs.Link) abci.Result {
	res := s.checkLinkAndAddToBatch(link, s.deliveredLinks)
	if res.IsOK() {
		s.deliveredLinksList = append(s.deliveredLinksList, link)
	}
	return res
}

func (s *State) checkLinkAndAddToBatch(link *cs.Link, batch store.BatchV2) abci.Result {
	err := link.Segmentify().Validate(batch.GetSegment)
	if err != nil {
		return abci.NewError(
			CodeTypeValidation,
			fmt.Sprintf("Link validation failed %v: %v", link, err),
		)
	}

	if s.validator != nil {
		err = (*s.validator).ValidateLink(batch, link)
		if err != nil {
			return abci.NewError(
				CodeTypeValidation,
				fmt.Sprintf("Link validation rules failed %v: %v", link, err),
			)
		}
	}

	if _, err := batch.CreateLink(link); err != nil {
		return abci.NewError(abci.CodeType_InternalError, err.Error())
	}

	return abci.OK
}

// Commit commits the delivered links,
// resets delivered and checked state,
// and returns the hash for the commit.
func (s *State) Commit() ([]byte, error) {
	if err := s.deliveredLinks.WriteV2(); err != nil {
		return nil, err
	}

	deliveredLinks, err := s.adapter.NewBatchV2()
	if err != nil {
		return nil, err
	}
	s.deliveredLinks = deliveredLinks
	s.checkedLinks = bufferedbatch.NewBatchV2(s.adapter)

	appHash, err := s.computeAppHash()
	if err != nil {
		return nil, err
	}

	// Store created links for client notification
	s.notificationLock.Lock()
	defer s.notificationLock.Unlock()

	for _, link := range s.deliveredLinksList {
		s.pendingNotifications = append(s.pendingNotifications, &store.Event{
			EventType: store.SavedLink,
			Details:   link,
		})
	}

	s.deliveredLinksList = nil

	return appHash, nil
}

func (s *State) computeAppHash() ([]byte, error) {
	hash := sha256.New()

	if _, err := hash.Write(s.previousAppHash); err != nil {
		return nil, err
	}

	if s.validator != nil {
		if _, err := hash.Write((*s.validator).Hash()[:]); err != nil {
			return nil, err
		}
	}

	// TODO: hash the content of s.deliveredLinksList in a merkle tree
	if len(s.deliveredLinksList) > 0 {
		for _, link := range s.deliveredLinksList {
			linkHash, _ := link.Hash()
			if _, err := hash.Write(linkHash[:]); err != nil {
				return nil, err
			}
		}
	}

	appHash := hash.Sum(nil)
	return appHash, nil
}

// DeliverNotifications delivers pending notifications and flushes them
func (s *State) DeliverNotifications() []*store.Event {
	s.notificationLock.Lock()
	defer s.notificationLock.Unlock()

	notifications := make([]*store.Event, len(s.pendingNotifications))
	copy(notifications, s.pendingNotifications)

	s.pendingNotifications = nil

	return notifications
}
