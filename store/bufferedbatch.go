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

package store

import (
	"bytes"

	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/types"
)

// BufferedBatch can be used as a base class for types
// that want to implement github.com/stratumn/sdk/store.Batch.
// All operations are stored in arrays and can be replayed.
// Only the Write method must be implemented.
type BufferedBatch struct {
	originalStore Adapter
	ValueOps      []ValueOperation
	SegmentOps    []SegmentOperation
}

// OpType represents a operation type on the Batch.
type OpType int

const (
	// OpTypeSet set represents a save operation.
	OpTypeSet = iota

	// OpTypeDelete set represents a delete operation.
	OpTypeDelete
)

// ValueOperation represents a operation on a value.
type ValueOperation struct {
	OpType
	Key   []byte
	Value []byte
}

// SegmentOperation represents a operation on a segment.
type SegmentOperation struct {
	OpType
	LinkHash *types.Bytes32
	Segment  *cs.Segment
}

// NewBufferedBatch creates a new Batch.
func NewBufferedBatch(a Adapter) *BufferedBatch {
	return &BufferedBatch{originalStore: a}
}

// SaveValue implements github.com/stratumn/sdk/store.Adapter.SaveValue.
func (b *BufferedBatch) SaveValue(key, value []byte) error {
	b.ValueOps = append(b.ValueOps, ValueOperation{OpTypeSet, key, value})
	return nil
}

// DeleteValue implements github.com/stratumn/sdk/store.Adapter.DeleteValue.
func (b *BufferedBatch) DeleteValue(key []byte) (value []byte, err error) {
	ops := make([]ValueOperation, len(b.ValueOps))
	copy(ops, b.ValueOps)

	// remove all existing save operations and get the last saved value.
	for i, sOp := range ops {
		if bytes.Compare(sOp.Key, key) == 0 {
			value = sOp.Value
			b.ValueOps = append(b.ValueOps[:i], b.ValueOps[i+1:]...)
		}
	}

	b.ValueOps = append(b.ValueOps, ValueOperation{OpTypeDelete, key, nil})

	if value != nil {
		return value, nil
	}
	return b.originalStore.GetValue(key)
}

// SaveSegment implements github.com/stratumn/sdk/store.Adapter.SaveSegment.
func (b *BufferedBatch) SaveSegment(segment *cs.Segment) error {
	b.SegmentOps = append(b.SegmentOps, SegmentOperation{OpTypeSet, nil, segment})
	return nil
}

// DeleteSegment implements github.com/stratumn/sdk/store.Adapter.DeleteSegment.
func (b *BufferedBatch) DeleteSegment(linkHash *types.Bytes32) (segment *cs.Segment, err error) {
	ops := make([]SegmentOperation, len(b.SegmentOps))
	copy(ops, b.SegmentOps)
	// remove all existing save operations and get the last saved value.
	for i, sOp := range ops {
		if sOp.LinkHash == linkHash {
			segment = sOp.Segment
			b.SegmentOps = append(b.SegmentOps[:i], b.SegmentOps[i+1:]...)
		}
	}

	b.SegmentOps = append(b.SegmentOps, SegmentOperation{OpTypeDelete, linkHash, nil})

	if segment != nil {
		return segment, nil
	}
	return b.originalStore.GetSegment(linkHash)
}

// GetSegment returns a segment from the cache or delegates the call to the store
func (b *BufferedBatch) GetSegment(linkHash *types.Bytes32) (segment *cs.Segment, err error) {
	for _, sOp := range b.SegmentOps {
		if sOp.LinkHash == linkHash && sOp.OpType == OpTypeSet {
			segment = sOp.Segment
		}
	}
	if segment != nil {
		return segment, nil
	}

	return b.originalStore.GetSegment(linkHash)
}

// FindSegments delegates the call to the store
func (b *BufferedBatch) FindSegments(filter *Filter) (cs.SegmentSlice, error) {
	return b.originalStore.FindSegments(filter)
}

// GetMapIDs delegates the call to the store
func (b *BufferedBatch) GetMapIDs(pagination *Pagination) ([]string, error) {
	return b.originalStore.GetMapIDs(pagination)
}

// GetValue returns a segment from the cache or delegates the call to the store
func (b *BufferedBatch) GetValue(key []byte) (value []byte, err error) {
	for _, sOp := range b.ValueOps {
		if bytes.Compare(sOp.Key, key) == 0 && sOp.OpType == OpTypeSet {
			value = sOp.Value
		}
	}
	if value != nil {
		return value, nil
	}

	return b.originalStore.GetValue(key)
}
