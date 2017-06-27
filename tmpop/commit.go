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
func (c *Commit) FindSegments(filter *store.Filter) ([]*cs.Segment, [][]byte, error) {
	segments, err := c.adapter.FindSegments(filter)
	if err != nil {
		return nil, nil, err
	}
	var values []*cs.Segment
	var proofs [][]byte

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
func (c *Commit) GetMapIDs(pagination *store.Pagination) ([]string, error) {
	return c.adapter.GetMapIDs(pagination)
}
