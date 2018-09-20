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
	"database/sql"

	log "github.com/sirupsen/logrus"
)

// TxFactory creates and manages transactions.
type TxFactory interface {
	NewTx() (*sql.Tx, error)
	CommitTx(tx *sql.Tx) error
	RollbackTx(tx *sql.Tx, err error)
}

// StandardTxFactory is just a wrapper around SQL transactions.
type StandardTxFactory struct {
	db *sql.DB
}

// NewStandardTxFactory creates a transaction factory using the underlying DB.
func NewStandardTxFactory(db *sql.DB) TxFactory {
	return &StandardTxFactory{db: db}
}

// NewTx creates a new DB transaction.
func (f *StandardTxFactory) NewTx() (*sql.Tx, error) {
	return f.db.Begin()
}

// CommitTx commits the transaction to the DB.
func (f *StandardTxFactory) CommitTx(tx *sql.Tx) error {
	return tx.Commit()
}

// RollbackTx rolls back the transaction and logs errors.
func (f *StandardTxFactory) RollbackTx(tx *sql.Tx, _ error) {
	err := tx.Rollback()
	if err != nil {
		log.Warnf("Error during transaction rollback: %s", err.Error())
	}
}

// SingletonTxFactory uses a single transaction under the hood.
// It can be used for batches since a transaction cannot contain
// sub-transactions.
type SingletonTxFactory struct {
	tx       *sql.Tx
	rollback bool
	err      error
}

// NewSingletonTxFactory creates a transaction factory using the underlying
// transaction.
// It makes sure all the work is done in a single transaction instead of many
// (since a transaction cannot contain nested sub-transactions).
func NewSingletonTxFactory(tx *sql.Tx) TxFactory {
	return &SingletonTxFactory{tx: tx}
}

// NewTx return the DB transaction.
// Clients should not directly call tx.Commit() or tx.Rollback().
func (f *SingletonTxFactory) NewTx() (*sql.Tx, error) {
	return f.tx, nil
}

// CommitTx does nothing. The owner of the SingletonTxFactory will commit
// directly on the tx object when ready (batch.Write()).
func (f *SingletonTxFactory) CommitTx(tx *sql.Tx) error {
	return nil
}

// RollbackTx marks the transaction as rolled back.
// The owner of the SingletonTxFactory should rollback the transaction.
func (f *SingletonTxFactory) RollbackTx(tx *sql.Tx, err error) {
	f.rollback = true
	f.err = err
}
