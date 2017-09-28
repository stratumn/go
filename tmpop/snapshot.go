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
	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/store"
	"github.com/stratumn/sdk/types"
	"github.com/tendermint/tmlibs/merkle"
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
