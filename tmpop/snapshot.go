// Copyright 2017 Stratumn SAS. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package tmpop

import (
	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/store"
	"github.com/stratumn/sdk/types"
	merkle "github.com/tendermint/go-merkle"
)

// Snapshot represents a version of the state.
type Snapshot struct {
	tree     merkle.Tree
	segments store.Batch
}

// SetSegment adds a new Segment in the Snapshot.
func (s *Snapshot) SetSegment(segment *cs.Segment) error {
	err := s.segments.SaveSegment(segment)
	if err != nil {
		return err
	}
	s.tree.Set(segment.GetLinkHash()[:], nil)

	return nil
}

// DeleteSegment removes a Segment from the Snapshot.
func (s *Snapshot) DeleteSegment(lh *types.Bytes32) (*cs.Segment, bool, error) {
	_, found := s.tree.Remove(lh[:])
	if !found {
		return nil, false, nil
	}

	segment, err := s.segments.DeleteSegment(lh)
	if err != nil {
		return nil, found, err
	}
	return segment, found, nil

}

// SaveValue adds a new value in the Snapshot.
func (s *Snapshot) SaveValue(key, value []byte) bool {
	return s.tree.Set(key, value)
}

// DeleteValue removes a value from the Snapshot.
func (s *Snapshot) DeleteValue(key []byte) ([]byte, bool) {
	return s.tree.Remove(key)
}
