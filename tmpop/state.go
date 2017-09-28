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
	"github.com/stratumn/sdk/store"
	"github.com/tendermint/tmlibs/merkle"
)

// State represents the app states, separating the commited state (for queries)
// from the working state (for CheckTx and AppendTx).
type State struct {
	committed merkle.Tree
	deliverTx merkle.Tree
	checkTx   merkle.Tree

	segments          store.Adapter
	deliveredSegments store.Batch
	checkedSegments   store.Batch
}

// NewState creates a new State.
func NewState(tree merkle.Tree, a store.Adapter) (*State, error) {
	deliveredSegments, err := a.NewBatch()
	if err != nil {
		return nil, err
	}
	checkedSegments, err := a.NewBatch()
	if err != nil {
		return nil, err
	}

	return &State{
		committed:         tree,
		deliverTx:         tree.Copy(),
		checkTx:           tree.Copy(),
		segments:          a,
		deliveredSegments: deliveredSegments,
		checkedSegments:   checkedSegments,
	}, nil
}

// Committed returns the committed state.
func (s State) Committed() *Commit {
	return &Commit{s.committed, s.segments}
}

// Append returns the version of the state affected by appended transaction from the current block.
func (s State) Append() *Snapshot {
	return &Snapshot{s.deliverTx, s.deliveredSegments}
}

// Check returns the version of the state affected by checked transaction from the memory pool.
func (s State) Check() *Snapshot {
	return &Snapshot{s.checkTx, s.checkedSegments}
}

// Commit stores the current Append() state as committed
// starts new Append/Check state, and
// returns the hash for the commit.
func (s *State) Commit() ([]byte, error) {
	err := s.deliveredSegments.Write()
	if err != nil {
		return nil, err
	}
	deliveredSegments, err := s.segments.NewBatch()
	if err != nil {
		return nil, err
	}
	s.deliveredSegments = deliveredSegments

	checkedSegments, err := s.segments.NewBatch()
	if err != nil {
		return nil, err
	}
	s.checkedSegments = checkedSegments

	var hash []byte
	hash = s.deliverTx.Save()

	s.committed = s.deliverTx
	s.deliverTx = s.committed.Copy()
	s.checkTx = s.committed.Copy()
	return hash, nil
}
