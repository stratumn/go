// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by the license that can be found in the
// LICENSE file.

package rethinkstore

import (
	"fmt"

	"github.com/stratumn/sdk/bufferedbatch"
)

// Batch is the type that implements github.com/stratumn/sdk/store.Batch.
type Batch struct {
	*bufferedbatch.Batch
	originalRethinkStore *Store
}

// NewBatch return a new instance of Batch
func NewBatch(a *Store) *Batch {
	return &Batch{bufferedbatch.NewBatch(a), a}
}

func (b *Batch) Write() error {
	for _, op := range b.ValueOps {
		switch op.OpType {
		case bufferedbatch.OpTypeSet:
			b.originalRethinkStore.SaveValue(op.Key, op.Value)
		case bufferedbatch.OpTypeDelete:
			b.originalRethinkStore.DeleteValue(op.Key)
		default:
			return fmt.Errorf("Invalid Batch operation type: %v", op.OpType)
		}
	}
	for _, op := range b.SegmentOps {
		switch op.OpType {
		case bufferedbatch.OpTypeSet:
			b.originalRethinkStore.SaveSegment(op.Segment)
		case bufferedbatch.OpTypeDelete:
			b.originalRethinkStore.DeleteSegment(op.LinkHash)
		default:
			return fmt.Errorf("Invalid Batch operation type: %v", op.OpType)
		}
	}
	return nil
}

// BatchV2 is the type that implements github.com/stratumn/sdk/store.Batch.
type BatchV2 struct {
	*bufferedbatch.BatchV2
	originalRethinkStore *Store
}

// NewBatchV2 return a new instance of Batch
func NewBatchV2(a *Store) *BatchV2 {
	return &BatchV2{bufferedbatch.NewBatchV2(a), a}
}

func (b *BatchV2) Write() (err error) {
	for _, link := range b.Links {
		_, err = b.originalRethinkStore.CreateLink(link)
		if err != nil {
			break
		}
	}
	return nil
}
