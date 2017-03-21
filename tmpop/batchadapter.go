// Copyright 2017 Stratumn SAS. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package tmpop

import "github.com/stratumn/sdk/store"

// BatchAdapter implements github.com/tendermint/go-db/db.Batch.
type BatchAdapter struct {
	batch store.Batch
}

// NewBatchAdapter returns a new Batch Adapter
func NewBatchAdapter(batch store.Batch) *BatchAdapter {
	return &BatchAdapter{batch}
}

// Set implements github.com/tendermint/go-db/db.Batch.Set
func (b *BatchAdapter) Set(key, value []byte) {
	saveError := b.batch.SaveValue(key, value)

	if saveError != nil {
		panic(saveError)
	}
}

// Delete implements github.com/tendermint/go-db/db.Batch.Delete
func (b *BatchAdapter) Delete(key []byte) {
	_, saveError := b.batch.DeleteValue(key)

	if saveError != nil {
		panic(saveError)
	}
}

// Write implements github.com/tendermint/go-db/db.Batch.Write
func (b *BatchAdapter) Write() {
	err := b.batch.Write()
	if err != nil {
		panic(err)
	}
}
