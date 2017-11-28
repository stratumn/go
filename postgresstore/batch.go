// Copyright 2017 Stratumn SAS. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package postgresstore

import "database/sql"

// Batch is the type that implements github.com/stratumn/sdk/store.Batch.
type Batch struct {
	*reader
	*writer
	done bool
	tx   *sql.Tx
}

// NewBatch creates a new instance of a Postgres Batch.
func NewBatch(tx *sql.Tx) (*Batch, error) {
	stmts, err := newBatchStmts(tx)
	if err != nil {
		return nil, err
	}

	return &Batch{
		reader: &reader{stmts: readStmts(stmts.readStmts)},
		writer: &writer{stmts: writeStmts(stmts.writeStmts)},
		tx:     tx,
	}, nil
}

// Write implements github.com/stratumn/sdk/store.Batch.Write.
func (b *Batch) Write() error {
	b.done = true
	return b.tx.Commit()
}

// WriteV2 implements github.com/stratumn/sdk/store.BatchV2.Write.
func (b *Batch) WriteV2() error {
	b.done = true
	return b.tx.Commit()
}
