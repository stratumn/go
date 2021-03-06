// Copyright 2016-2018 Stratumn SAS. All rights reserved.
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

package postgresstore

import (
	"context"
	"database/sql"
	"sync"

	"github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-core/monitoring"
	"github.com/stratumn/go-core/store"
)

// Batch is the type that implements github.com/stratumn/go-core/store.Batch.
type Batch struct {
	*scopedStore

	lock sync.RWMutex
	done bool

	tx        *sql.Tx
	txFactory *SingletonTxFactory
}

// NewBatch creates a new instance of a Postgres Batch.
func NewBatch(tx *sql.Tx) (*Batch, error) {
	stmts, err := newStmts(tx)
	if err != nil {
		return nil, err
	}

	txFactory := NewSingletonTxFactory(tx).(*SingletonTxFactory)
	return &Batch{
		scopedStore: newScopedStore(stmts, txFactory),
		tx:          tx,
		txFactory:   txFactory,
	}, nil
}

// CreateLink wraps the underlying link creation and stops the batch as soon as
// an invalid link is encountered.
func (b *Batch) CreateLink(ctx context.Context, link *chainscript.Link) (chainscript.LinkHash, error) {
	b.lock.RLock()
	done := b.done
	b.lock.RUnlock()

	if done {
		return nil, store.ErrBatchFailed
	}

	lh, err := b.scopedStore.CreateLink(ctx, link)
	if err != nil {
		b.lock.Lock()
		defer b.lock.Unlock()

		b.done = true
		err := b.tx.Rollback()
		if err != nil {
			monitoring.TxLogEntry(ctx).
				WithError(err).
				Warn("Error during transaction rollback")
		}
	}

	return lh, err
}

// Write implements github.com/stratumn/go-core/store.Batch.Write.
func (b *Batch) Write(ctx context.Context) (err error) {
	span, _ := monitoring.StartSpanOutgoingRequest(ctx, "postgresstore/batch/Write")
	defer func() {
		monitoring.SetSpanStatusAndEnd(span, err)
	}()

	b.lock.Lock()
	defer b.lock.Unlock()

	b.done = true

	if b.txFactory.rollback {
		err := b.tx.Rollback()
		if err != nil {
			monitoring.TxLogEntry(ctx).
				WithError(err).
				Warn("Error during transaction rollback")
		}

		return b.txFactory.err
	}

	return b.tx.Commit()
}
