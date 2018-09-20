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

	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/stratumn/go-chainscript"
)

// Errors used by the write statements.
var (
	ErrLinkAlreadyExists = errors.New("link already exists")
)

type writer struct {
	txFactory TxFactory
	reader    *reader
	stmts     writeStmts
}

func newWriter(txFactory TxFactory, r *reader, stmts writeStmts) *writer {
	return &writer{
		txFactory: txFactory,
		reader:    r,
		stmts:     stmts,
	}
}

// SetValue implements github.com/stratumn/go-indigocore/store.KeyValueStore.SetValue.
func (a *writer) SetValue(ctx context.Context, key []byte, value []byte) error {
	_, err := a.stmts.SaveValue.Exec(key, value)
	return err
}

// DeleteValue implements github.com/stratumn/go-indigocore/store.KeyValueStore.DeleteValue.
func (a *writer) DeleteValue(ctx context.Context, key []byte) ([]byte, error) {
	var data []byte

	if err := a.stmts.DeleteValue.QueryRow(key).Scan(&data); err != nil {
		if err.Error() == notFoundError {
			return nil, nil
		}
		return nil, err
	}

	return data, nil
}

// CreateLink implements github.com/stratumn/go-indigocore/store.Adapter.CreateLink.
func (a *writer) CreateLink(ctx context.Context, link *chainscript.Link) (chainscript.LinkHash, error) {
	var (
		priority     = link.Meta.Priority
		mapID        = link.Meta.MapId
		prevLinkHash = link.Meta.GetPrevLinkHash()
		tags         = link.Meta.Tags
		process      = link.Meta.Process.Name
	)

	linkHash, err := link.Hash()
	if err != nil {
		return linkHash, err
	}

	data, err := chainscript.MarshalLink(link)
	if err != nil {
		return linkHash, err
	}

	if len(prevLinkHash) == 0 {
		err = a.createLink(linkHash, priority, mapID, []byte{}, tags, data, process)
		return linkHash, err
	}

	parent, err := a.reader.GetSegment(ctx, prevLinkHash)
	if err != nil {
		return linkHash, err
	}

	parentDegree := parent.Link.Meta.OutDegree
	if parentDegree < 0 {
		err = a.createLink(linkHash, priority, mapID, prevLinkHash, tags, data, process)
		return linkHash, err
	}

	if parentDegree == 0 {
		return linkHash, chainscript.ErrOutDegree
	}

	// Inserting the link and updating its parent's current degree needs to be
	// done in a transaction to protect the DB from race conditions.

	tx, err := a.txFactory.NewTx()
	if err != nil {
		return linkHash, err
	}

	currentDegree, err := a.getLinkDegree(tx, prevLinkHash)
	if err != nil {
		a.txFactory.RollbackTx(tx, err)
		return linkHash, err
	}

	if int(parentDegree) <= currentDegree {
		a.txFactory.RollbackTx(tx, chainscript.ErrOutDegree)
		return linkHash, chainscript.ErrOutDegree
	}

	err = a.createLinkInTx(tx, linkHash, priority, mapID, prevLinkHash, tags, data, process)
	if err != nil {
		a.txFactory.RollbackTx(tx, err)
		return linkHash, err
	}

	err = a.incrementLinkDegree(tx, prevLinkHash, currentDegree)
	if err != nil {
		a.txFactory.RollbackTx(tx, err)
		return linkHash, err
	}

	return linkHash, a.txFactory.CommitTx(tx)
}

// getLinkDegree reads the current degree of the given link.
// It locks the associated row until the transaction completes.
func (a *writer) getLinkDegree(tx *sql.Tx, linkHash chainscript.LinkHash) (int, error) {
	degreeLock, err := tx.Prepare(SQLLockLinkDegree)
	if err != nil {
		return 0, err
	}

	row := degreeLock.QueryRow(linkHash)
	currentDegree := 0
	err = row.Scan(&currentDegree)

	// If the link doesn't have children yet, no rows will be found.
	// That should not be considered an error.
	if err == sql.ErrNoRows {
		return 0, nil
	}

	return currentDegree, err
}

// incrementLinkDegree increments the degree of the given link.
// A lock should have been acquired previously by the transaction to ensure
// consistency.
func (a *writer) incrementLinkDegree(tx *sql.Tx, linkHash chainscript.LinkHash, currentDegree int) error {
	updateDegree, err := tx.Prepare(SQLUpdateLinkDegree)
	if err != nil {
		return err
	}

	_, err = updateDegree.Exec(linkHash, currentDegree+1)
	return err
}

// createLink adds the given link to the DB.
func (a *writer) createLink(
	linkHash chainscript.LinkHash,
	priority float64,
	mapID string,
	prevLinkHash chainscript.LinkHash,
	tags []string,
	data []byte,
	process string,
) error {
	tx, err := a.txFactory.NewTx()
	if err != nil {
		return err
	}

	err = a.createLinkInTx(tx, linkHash, priority, mapID, prevLinkHash, tags, data, process)
	if err != nil {
		a.txFactory.RollbackTx(tx, err)
		return err
	}

	return a.txFactory.CommitTx(tx)
}

// createLink inserts the given link in a transaction context.
// If the link already exists it will return an error.
func (a *writer) createLinkInTx(
	tx *sql.Tx,
	linkHash chainscript.LinkHash,
	priority float64,
	mapID string,
	prevLinkHash chainscript.LinkHash,
	tags []string,
	data []byte,
	process string,
) error {
	createLink, err := tx.Prepare(SQLCreateLink)
	if err != nil {
		return err
	}

	res, err := createLink.Exec(linkHash, priority, mapID, prevLinkHash, pq.Array(tags), data, process)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrLinkAlreadyExists
	}

	initDegree, err := tx.Prepare(SQLCreateLinkDegree)
	if err != nil {
		return err
	}

	_, err = initDegree.Exec(linkHash)
	return err
}
