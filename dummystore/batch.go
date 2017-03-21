// Copyright 2017 Stratumn SAS. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package dummystore

import (
	"fmt"

	"github.com/stratumn/sdk/store"
)

// Batch is the type that implements github.com/stratumn/sdk/store.Batch.
type Batch struct {
	*store.BufferedBatch
	originalDummyStore *DummyStore
}

// NewBatch creates a new Batch
func NewBatch(a *DummyStore) *Batch {
	return &Batch{store.NewBufferedBatch(a), a}
}

// Write implements github.com/stratumn/sdk/store.Adapter.Write.
func (b *Batch) Write() error {
	b.originalDummyStore.mutex.Lock()
	defer b.originalDummyStore.mutex.Unlock()

	for _, op := range b.ValueOps {
		switch op.OpType {
		case store.OpTypeSet:
			b.originalDummyStore.saveValue(op.Key, op.Value)
		case store.OpTypeDelete:
			b.originalDummyStore.deleteValue(op.Key)
		default:
			return fmt.Errorf("Invalid Batch operation type: %v", op.OpType)
		}
	}
	for _, op := range b.SegmentOps {
		switch op.OpType {
		case store.OpTypeSet:
			b.originalDummyStore.saveSegment(op.Segment)
		case store.OpTypeDelete:
			b.originalDummyStore.deleteSegment(op.LinkHash)
		default:
			return fmt.Errorf("Invalid Batch operation type: %v", op.OpType)
		}
	}
	return nil
}
