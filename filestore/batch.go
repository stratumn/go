// Copyright 2017-2018 Stratumn SAS. All rights reserved.
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
	"context"

	"github.com/stratumn/go-core/bufferedbatch"
	"github.com/stratumn/go-core/monitoring"
	"go.opencensus.io/trace"
)

// Batch is the type that implements github.com/stratumn/go-core/store.Batch.
type Batch struct {
	*bufferedbatch.Batch

	originalFileStore *FileStore
}

// NewBatch creates a new Batch
func NewBatch(ctx context.Context, a *FileStore) *Batch {
	return &Batch{
		Batch:             bufferedbatch.NewBatch(ctx, a),
		originalFileStore: a,
	}
}

// Write implements github.com/stratumn/go-core/store.Batch.Write
func (b *Batch) Write(ctx context.Context) (err error) {
	_, span := trace.StartSpan(ctx, "filestore/batch/Write")
	defer monitoring.SetSpanStatusAndEnd(span, err)

	b.originalFileStore.mutex.Lock()
	defer b.originalFileStore.mutex.Unlock()

	for _, link := range b.Links {
		if _, err := b.originalFileStore.createLink(link); err != nil {
			return err
		}
	}

	return nil
}
