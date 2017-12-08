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

	for _, link := range b.Links {
		if _, err := b.originalFileStore.createLink(link); err != nil {
			return err
		}
	}

	return nil
}
