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

package couchstore

import (
	"github.com/pkg/errors"

	"github.com/stratumn/sdk/bufferedbatch"
)

// Batch is the type that implements github.com/stratumn/sdk/store.Batch.
type Batch struct {
	*bufferedbatch.Batch
	originalCouchStore *CouchStore
}

// NewBatch creates a new Batch.
func NewBatch(a *CouchStore) *Batch {
	return &Batch{bufferedbatch.NewBatch(a), a}
}

func (b *Batch) Write() (err error) {
	for _, op := range b.ValueOps {
		switch op.OpType {
		case bufferedbatch.OpTypeSet:
			err = b.originalCouchStore.SaveValue(op.Key, op.Value)
		case bufferedbatch.OpTypeDelete:
			_, err = b.originalCouchStore.DeleteValue(op.Key)
		default:
			err = errors.Errorf("Invalid Batch operation type: %v", op.OpType)
		}
	}

	if err != nil {
		return
	}

	for _, op := range b.SegmentOps {
		switch op.OpType {
		case bufferedbatch.OpTypeSet:
			err = b.originalCouchStore.SaveSegment(op.Segment)
		case bufferedbatch.OpTypeDelete:
			_, err = b.originalCouchStore.DeleteSegment(op.LinkHash)
		default:
			err = errors.Errorf("Invalid Batch operation type: %v", op.OpType)
		}
	}

	return
}
