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
