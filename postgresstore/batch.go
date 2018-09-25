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

package postgresstore

import (
	"context"
	"database/sql"

	log "github.com/sirupsen/logrus"
	"github.com/stratumn/go-core/monitoring"

	"go.opencensus.io/trace"
)

// Batch is the type that implements github.com/stratumn/go-core/store.Batch.
type Batch struct {
	*reader
	*writer
	done      bool
	tx        *sql.Tx
	txFactory *SingletonTxFactory
}

// NewBatch creates a new instance of a Postgres Batch.
func NewBatch(tx *sql.Tx) (*Batch, error) {
	stmts, err := newBatchStmts(tx)
	if err != nil {
		return nil, err
	}

	txFactory := NewSingletonTxFactory(tx).(*SingletonTxFactory)
	r := newReader(readStmts(stmts.readStmts))
	w := newWriter(txFactory, r, writeStmts(stmts.writeStmts))

	return &Batch{
		reader:    r,
		writer:    w,
		tx:        tx,
		txFactory: txFactory,
	}, nil
}

// Write implements github.com/stratumn/go-core/store.Batch.Write.
func (b *Batch) Write(ctx context.Context) (err error) {
	_, span := trace.StartSpan(ctx, "postgresstore/batch/Write")
	defer monitoring.SetSpanStatusAndEnd(span, err)

	b.done = true
	if b.txFactory.rollback {
		err := b.tx.Rollback()
		if err != nil {
			log.Warnf("Error during transaction rollback: %s", err.Error())
		}

		return b.txFactory.err
	}

	return b.tx.Commit()
}
