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

func (b *Batch) Write() (err error) {
	for _, op := range b.ValueOps {
		switch op.OpType {
		case store.OpTypeSet:
			err = b.originalTMStore.SaveValue(op.Key, op.Value)
		case store.OpTypeDelete:
			_, err = b.originalTMStore.DeleteValue(op.Key)
		default:
			err = fmt.Errorf("Invalid Batch operation type: %v", op.OpType)
		}
	}

	if err != nil {
		return
	}

	for _, op := range b.SegmentOps {
		switch op.OpType {
		case store.OpTypeSet:
			err = b.originalTMStore.SaveSegment(op.Segment)
		case store.OpTypeDelete:
			_, err = b.originalTMStore.DeleteSegment(op.LinkHash)
		default:
			err = fmt.Errorf("Invalid Batch operation type: %v", op.OpType)
		}
	}

	return
}
