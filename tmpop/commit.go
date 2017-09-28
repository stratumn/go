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

	log "github.com/Sirupsen/logrus"
)

// Commit represents a committed state of the blockchain.
// It uses a tree to store all linkhashes and have a canonical app for the state of the application.
// It uses a store for indexed search but all results from the store must be checked within the tree.
type Commit struct {
	tree    merkle.Tree
	adapter store.Adapter
}

// FindSegments emulates github.com/stratumn/sdk/store.Adapter.FindSegments. with additional proofs.
func (c *Commit) FindSegments(filter *store.SegmentFilter) ([]*cs.Segment, [][]byte, error) {
	segments, err := c.adapter.FindSegments(filter)
	if err != nil {
		return nil, nil, err
	}
	var values = make([]*cs.Segment, 0)
	var proofs = make([][]byte, 0)

	for _, s := range segments {
		_, proof, exists := c.Proof(s.GetLinkHash()[:])
		if exists == true {
			values = append(values, s)
			proofs = append(proofs, proof)
		}
	}

	return values, proofs, nil
}

// GetSegment emulates github.com/stratumn/sdk/store.Adapter.GetSegments.
// GetSegment returns a segment and its proof.
func (c *Commit) GetSegment(lh *types.Bytes32) (*cs.Segment, []byte, error) {
	s, err := c.adapter.GetSegment(lh)
	if s == nil || err != nil {
		return nil, nil, err
	}

	_, proof, exists := c.Proof(s.GetLinkHash()[:])
	if exists == true {
		return s, proof, nil
	}

	log.Warn("Segment in adapter but not in tree")

	return nil, nil, nil
}

// Size returns the current number of values stored.
func (c *Commit) Size() int {
	return c.tree.Size()
}

// Hash returns the current hash of the Commit (root hash of the tree).
func (c *Commit) Hash() []byte {
	return c.tree.Hash()
}

// Proof returns the Merkle Proof of a given key.
func (c *Commit) Proof(key []byte) ([]byte, []byte, bool) {
	return c.tree.Proof(key)
}

// GetMapIDs implements github.com/stratumn/sdk/store.Adapter.GetMapIDs.
// It might be out of sync with the tree.
func (c *Commit) GetMapIDs(filter *store.MapFilter) ([]string, error) {
	return c.adapter.GetMapIDs(filter)
}
