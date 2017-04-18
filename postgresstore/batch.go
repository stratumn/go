// Copyright 2017 Stratumn SAS. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package postgresstore

import "database/sql"

// Batch is the type that implements github.com/stratumn/sdk/store.Batch.
type Batch struct {
	*writer
	store *Store
}

// NewBatch creates a new instance of a Postgres Batch
func NewBatch(a *Store, tx *sql.Tx) (*Batch, error) {
	stmts, err := newBatchStmts(tx)
	if err != nil {
		return nil, err
	}

	return &Batch{
		writer: &writer{stmts: stmts.writeStmts},
		store:  a,
	}, nil
}

func (b *Batch) Write() error {
	return b.store.commit(b)
}
