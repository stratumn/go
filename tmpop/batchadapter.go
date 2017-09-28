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

package tmpop

import "github.com/stratumn/sdk/store"

// BatchAdapter implements github.com/tendermint/tmlibs/db/db.Batch.
type BatchAdapter struct {
	batch store.Batch
}

// NewBatchAdapter returns a new Batch Adapter
func NewBatchAdapter(batch store.Batch) *BatchAdapter {
	return &BatchAdapter{batch}
}

// Set implements github.com/tendermint/tmlibs/db/db.Batch.Set
func (b *BatchAdapter) Set(key, value []byte) {
	saveError := b.batch.SaveValue(key, value)

	if saveError != nil {
		panic(saveError)
	}
}

// Delete implements github.com/tendermint/tmlibs/db/db.Batch.Delete
func (b *BatchAdapter) Delete(key []byte) {
	_, saveError := b.batch.DeleteValue(key)

	if saveError != nil {
		panic(saveError)
	}
}

// Write implements github.com/tendermint/tmlibs/db/db.Batch.Write
func (b *BatchAdapter) Write() {
	err := b.batch.Write()
	if err != nil {
		panic(err)
	}
}
