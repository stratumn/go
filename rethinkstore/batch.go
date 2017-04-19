// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by the license that can be found in the
// LICENSE file.

package rethinkstore

import (
	"fmt"

	"github.com/stratumn/sdk/store"
)

// Batch is the type that implements github.com/stratumn/sdk/store.Batch.
type Batch struct {
	*store.BufferedBatch
	originalRethinkStore *Store
}

// NewBatch return a new instance of Batch
func NewBatch(a *Store) *Batch {
	return &Batch{store.NewBufferedBatch(a), a}
}

func (b *Batch) Write() error {
	for _, op := range b.ValueOps {
		switch op.OpType {
		case store.OpTypeSet:
			b.originalRethinkStore.SaveValue(op.Key, op.Value)
		case store.OpTypeDelete:
			b.originalRethinkStore.DeleteValue(op.Key)
		default:
			return fmt.Errorf("Invalid Batch operation type: %v", op.OpType)
		}
	}
	for _, op := range b.SegmentOps {
		switch op.OpType {
		case store.OpTypeSet:
			b.originalRethinkStore.SaveSegment(op.Segment)
		case store.OpTypeDelete:
			b.originalRethinkStore.DeleteSegment(op.LinkHash)
		default:
			return fmt.Errorf("Invalid Batch operation type: %v", op.OpType)
		}
	}
	return nil
}
