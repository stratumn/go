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
			if err := b.originalTMStore.SaveValue(op.Key, op.Value); err != nil {
				return err
			}
		case store.OpTypeDelete:
			if _, err := b.originalTMStore.DeleteValue(op.Key); err != nil {
				return err
			}
		default:
			return fmt.Errorf("invalid Batch operation type: %v", op.OpType)
		}
	}
	for _, op := range b.SegmentOps {
		switch op.OpType {
		case store.OpTypeSet:
			if err := b.originalTMStore.SaveSegment(op.Segment); err != nil {
				return err
			}
		case store.OpTypeDelete:
			if _, err := b.originalTMStore.DeleteSegment(op.LinkHash); err != nil {
				return err
			}
		default:
			return fmt.Errorf("invalid Batch operation type: %v", op.OpType)
		}
	}

	return nil
}
