// Copyright 2017 Stratumn SAS. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package filestore

import (
	"fmt"

	"github.com/stratumn/sdk/store"
	db "github.com/tendermint/go-db"
)

// Batch is the type that implements github.com/stratumn/sdk/store.Batch.
type Batch struct {
	*store.BufferedBatch

	originalFileStore *FileStore
	originalBatch     db.Batch
}

// NewBatch creates a new Batch
func NewBatch(a *FileStore) *Batch {
	return &Batch{
		BufferedBatch:     store.NewBufferedBatch(a),
		originalFileStore: a,
		originalBatch:     a.kvDB.NewBatch(),
	}
}

// Write implements github.com/stratumn/sdk/store.Batch.Write
func (b *Batch) Write() error {
	b.originalBatch.Write()

	b.originalFileStore.mutex.Lock()
	defer b.originalFileStore.mutex.Unlock()

	for _, op := range b.SegmentOps {
		switch op.OpType {
		case store.OpTypeSet:
			b.originalFileStore.saveSegment(op.Segment)
		case store.OpTypeDelete:
			b.originalFileStore.deleteSegment(op.LinkHash)
		default:
			return fmt.Errorf("Invalid Batch operation type: %v", op.OpType)
		}
	}

	return nil
}

// SaveValue implements github.com/stratumn/sdk/store.Batch.SaveValue
func (b *Batch) SaveValue(key, value []byte) error {
	b.originalBatch.Set(key, value)

	return nil
}

// DeleteValue implements github.com/stratumn/sdk/store.Batch.DeleteValue
func (b *Batch) DeleteValue(key []byte) ([]byte, error) {
	v, err := b.originalFileStore.GetValue(key)
	if err != nil {
		return nil, err
	}

	b.originalBatch.Delete(key)

	return v, nil
}
