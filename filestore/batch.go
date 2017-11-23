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

package filestore

import (
	"fmt"

	"github.com/stratumn/sdk/bufferedbatch"
)

// Batch is the type that implements github.com/stratumn/sdk/store.Batch.
type Batch struct {
	*bufferedbatch.Batch

	originalFileStore *FileStore
}

// NewBatch creates a new Batch
func NewBatch(a *FileStore) *Batch {
	return &Batch{
		Batch:             bufferedbatch.NewBatch(a),
		originalFileStore: a,
	}
}

// Write implements github.com/stratumn/sdk/store.Batch.Write
func (b *Batch) Write() (err error) {
	b.originalFileStore.mutex.Lock()
	defer b.originalFileStore.mutex.Unlock()

	for _, op := range b.SegmentOps {
		switch op.OpType {
		case bufferedbatch.OpTypeSet:
			err = b.originalFileStore.saveSegment(op.Segment)
		case bufferedbatch.OpTypeDelete:
			_, err = b.originalFileStore.deleteSegment(op.LinkHash)
		default:
			err = fmt.Errorf("Invalid Batch operation type: %v", op.OpType)
		}
		if err != nil {
			return
		}
	}

	for _, op := range b.ValueOps {
		switch op.OpType {
		case bufferedbatch.OpTypeSet:
			err = b.originalFileStore.SaveValue(op.Key, op.Value)
		case bufferedbatch.OpTypeDelete:
			_, err = b.originalFileStore.DeleteValue(op.Key)
		default:
			err = fmt.Errorf("Invalid Batch operation type: %v", op.OpType)
		}
		if err != nil {
			return
		}
	}

	return nil
}
