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

import "database/sql"

// Batch is the type that implements github.com/stratumn/go-indigocore/store.Batch.
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

// Write implements github.com/stratumn/go-indigocore/store.Batch.Write.
func (b *Batch) Write() error {
	b.done = true
	return b.tx.Commit()
}
