// Copyright 2017 Stratumn SAS. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package tmpop

import (
	"github.com/stratumn/sdk/store"
	merkle "github.com/tendermint/go-merkle"
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
func NewState(tree merkle.Tree, a store.Adapter) State {
	return State{
		committed:         tree,
		deliverTx:         tree.Copy(),
		checkTx:           tree.Copy(),
		segments:          a,
		deliveredSegments: a.NewBatch(),
		checkedSegments:   a.NewBatch(),
	}
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
func (s *State) Commit() []byte {
	err := s.deliveredSegments.Write()
	if err != nil {
		panic(err)
	}
	s.deliveredSegments = s.segments.NewBatch()
	s.checkedSegments = s.segments.NewBatch()

	var hash []byte
	hash = s.deliverTx.Save()

	s.committed = s.deliverTx
	s.deliverTx = s.committed.Copy()
	s.checkTx = s.committed.Copy()
	return hash
}
