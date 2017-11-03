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

package bufferedbatch

import (
	"bytes"
	"fmt"

	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/store"
	"github.com/stratumn/sdk/types"
)

// Batch can be used as a base class for types
// that want to implement github.com/stratumn/sdk/store.Batch.
// All operations are stored in arrays and can be replayed.
// Only the Write method must be implemented.
type Batch struct {
	originalStore store.Adapter
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

// NewBatch creates a new Batch.
func NewBatch(a store.Adapter) *Batch {
	return &Batch{originalStore: a}
}

// SaveValue implements github.com/stratumn/sdk/store.Adapter.SaveValue.
func (b *Batch) SaveValue(key, value []byte) error {
	b.ValueOps = append(b.ValueOps, ValueOperation{OpTypeSet, key, value})
	return nil
}

// DeleteValue implements github.com/stratumn/sdk/store.Adapter.DeleteValue.
func (b *Batch) DeleteValue(key []byte) (value []byte, err error) {
	// remove all existing save operations and get the last saved value.
	for i := len(b.ValueOps) - 1; i >= 0; i-- {
		sOp := b.ValueOps[i]
		if bytes.Compare(sOp.Key, key) == 0 {
			if value == nil && sOp.OpType == OpTypeSet {
				value = sOp.Value
			}
			b.ValueOps = append(b.ValueOps[:i], b.ValueOps[i+1:]...)
		}
	}

	b.ValueOps = append(b.ValueOps, ValueOperation{OpTypeDelete, key, value})

	if value != nil {
		return value, nil
	}
	return b.originalStore.GetValue(key)
}

// SaveSegment implements github.com/stratumn/sdk/store.Adapter.SaveSegment.
func (b *Batch) SaveSegment(segment *cs.Segment) error {
	if err := segment.Validate(b.GetSegment); err != nil {
		return err
	}
	b.SegmentOps = append(b.SegmentOps, SegmentOperation{OpTypeSet, segment.GetLinkHash(), segment})
	return nil
}

// DeleteSegment implements github.com/stratumn/sdk/store.Adapter.DeleteSegment.
func (b *Batch) DeleteSegment(linkHash *types.Bytes32) (segment *cs.Segment, err error) {
	// remove all existing save operations and get the last saved value.
	for i := len(b.SegmentOps) - 1; i >= 0; i-- {
		sOp := b.SegmentOps[i]
		if sOp.LinkHash != nil && linkHash != nil && *sOp.LinkHash == *linkHash {
			if segment == nil && sOp.OpType == OpTypeSet {
				segment = sOp.Segment
			}
			b.SegmentOps = append(b.SegmentOps[:i], b.SegmentOps[i+1:]...)
		}
	}

	b.SegmentOps = append(b.SegmentOps, SegmentOperation{OpTypeDelete, linkHash, segment})

	if segment != nil {
		return segment, nil
	}
	return b.originalStore.GetSegment(linkHash)
}

// GetSegment returns a segment from the cache or delegates the call to the store
func (b *Batch) GetSegment(linkHash *types.Bytes32) (segment *cs.Segment, err error) {
	deleted := false
	for _, sOp := range b.SegmentOps {
		if *sOp.LinkHash == *linkHash {
			switch sOp.OpType {
			case OpTypeSet:
				segment = sOp.Segment
				deleted = false
			case OpTypeDelete:
				deleted = true
			}
		}
	}
	if deleted {
		return nil, nil
	}
	if segment != nil {
		return segment, nil
	}

	return b.originalStore.GetSegment(linkHash)
}

// FindSegments returns the union of segments in the store and not commited yet
func (b *Batch) FindSegments(filter *store.SegmentFilter) (cs.SegmentSlice, error) {
	segments, err := b.originalStore.FindSegments(filter)
	if err != nil {
		return segments, err
	}
	for _, sOp := range b.SegmentOps {
		if sOp.Segment == nil || filter.Match(sOp.Segment) {
			switch sOp.OpType {
			case OpTypeSet:
				segments = append(segments, sOp.Segment)
			case OpTypeDelete:
				for i := len(segments) - 1; i >= 0; i-- {
					s := segments[i]
					if *s.GetLinkHash() == *sOp.LinkHash {
						segments = append(segments[:i], segments[i+1:]...)
					}
				}
			}
		}
	}
	return filter.Pagination.PaginateSegments(segments), nil
}

func (b *Batch) filterMapIDsBySegmentToDelete(pagination store.Pagination, mapIDs map[string]int, linkHashesToDel []*types.Bytes32) (map[string]int, error) {
	if len(linkHashesToDel) == 0 {
		return mapIDs, nil
	}

	// Group segment to delete per mapId
	segToDelMap := make(map[string]int)
	for _, l := range linkHashesToDel {
		s, err := b.originalStore.GetSegment(l)
		if err == nil && s != nil {
			segToDelMap[s.Link.Meta["mapId"].(string)]++
		}
	}

	// Retrieve segments per mapId and delete from results
	for mapID, nbSegsToDel := range segToDelMap {
		segs, err := b.originalStore.FindSegments(&store.SegmentFilter{
			Pagination: store.Pagination{Limit: nbSegsToDel + 1},
			MapIDs:     []string{mapID},
		})
		if err != nil {
			return nil, fmt.Errorf("cannot find segments from mapId '%s' to delete (%s)", mapID, err)
		}
		if nbSegsToDel >= len(segs) {
			delete(mapIDs, mapID)
		}
	}
	return mapIDs, nil
}

// GetMapIDs returns the union of mapIds in the store and not commited yet
func (b *Batch) GetMapIDs(filter *store.MapFilter) ([]string, error) {
	tmpMapIDs, err := b.originalStore.GetMapIDs(filter)
	if err != nil {
		return tmpMapIDs, err
	}
	mapIDs := make(map[string]int, len(tmpMapIDs))
	for _, m := range tmpMapIDs {
		mapIDs[m] = 0
	}

	// Apply uncommited segments
	var linkHashesToDel []*types.Bytes32
	for _, sOp := range b.SegmentOps {
		switch sOp.OpType {
		case OpTypeSet:
			if sOp.Segment != nil {
				mapID := sOp.Segment.Link.Meta["mapId"].(string)
				mapIDs[mapID]++
			}
		case OpTypeDelete:
			linkHashesToDel = append(linkHashesToDel, sOp.LinkHash)
		}
	}

	mapIDs, err = b.filterMapIDsBySegmentToDelete(filter.Pagination, mapIDs, linkHashesToDel)

	ids := make([]string, 0, len(mapIDs))
	for k := range mapIDs {
		ids = append(ids, k)
	}
	return filter.Pagination.PaginateStrings(ids), err
}

// GetValue returns a segment from the cache or delegates the call to the store
func (b *Batch) GetValue(key []byte) (value []byte, err error) {
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

// Write implements github.com/stratumn/sdk/store.Batch.Write
func (b *Batch) Write() (err error) {
	for _, op := range b.ValueOps {
		switch op.OpType {
		case OpTypeSet:
			err = b.originalStore.SaveValue(op.Key, op.Value)
		case OpTypeDelete:
			_, err = b.originalStore.DeleteValue(op.Key)
		default:
			err = fmt.Errorf("Invalid Batch operation type: %v", op.OpType)
		}
		if err != nil {
			break
		}
	}

	if err != nil {
		return
	}

	for _, op := range b.SegmentOps {
		switch op.OpType {
		case OpTypeSet:
			err = b.originalStore.SaveSegment(op.Segment)
		case OpTypeDelete:
			_, err = b.originalStore.DeleteSegment(op.LinkHash)
		default:
			err = fmt.Errorf("Invalid Batch operation type: %v", op.OpType)
		}
		if err != nil {
			break
		}
	}

	return
}
