// Copyright 2017 Stratumn SAS. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package tmstore

import (
	"fmt"

	"github.com/stratumn/sdk/store"
)

// Batch is the type that implements github.com/stratumn/sdk/store.Batch.
type Batch struct {
	*store.BufferedBatch
	originalTMStore *TMStore
}

// NewBatch creates a new Batch.
func NewBatch(a *TMStore) *Batch {
	return &Batch{store.NewBufferedBatch(a), a}
}

func (b *Batch) Write() error {
	for _, op := range b.ValueOps {
		switch op.OpType {
		case store.OpTypeSet:
			b.originalTMStore.SaveValue(op.Key, op.Value)
		case store.OpTypeDelete:
			b.originalTMStore.DeleteValue(op.Key)
		default:
			return fmt.Errorf("Invalid Batch operation type: %v", op.OpType)
		}
	}
	for _, op := range b.SegmentOps {
		switch op.OpType {
		case store.OpTypeSet:
			b.originalTMStore.SaveSegment(op.Segment)
		case store.OpTypeDelete:
			b.originalTMStore.DeleteSegment(op.LinkHash)
		default:
			return fmt.Errorf("Invalid Batch operation type: %v", op.OpType)
		}
	}

	return nil
}
