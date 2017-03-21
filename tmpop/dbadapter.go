// Copyright 2017 Stratumn SAS. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package tmpop

import (
	"github.com/stratumn/sdk/store"
	db "github.com/tendermint/go-db"
)

// DBAdapter implements github.com/tendermint/go-db/db.DB.
type DBAdapter struct {
	storeAdapter store.Adapter
}

// NewDBAdapter returns a new DB Adapter
func NewDBAdapter(a store.Adapter) *DBAdapter {
	return &DBAdapter{a}
}

// Get implements github.com/tendermint/go-db/db.DB.Get
func (a *DBAdapter) Get(key []byte) []byte {
	value, err := a.storeAdapter.GetValue(key)
	if err != nil {
		panic(err)
	}
	return value
}

// Set implements github.com/tendermint/go-db/db.DB.Set
func (a *DBAdapter) Set(key, value []byte) {
	err := a.storeAdapter.SaveValue(key, value)
	if err != nil {
		panic(err)
	}
}

// SetSync implements github.com/tendermint/go-db/db.DB.SetSync
func (a *DBAdapter) SetSync(key, value []byte) {
	a.Set(key, value)
}

// Delete implements github.com/tendermint/go-db/db.DB.Delete
func (a *DBAdapter) Delete(key []byte) {
	_, err := a.storeAdapter.DeleteValue(key)
	if err != nil {
		panic(err)
	}
}

// DeleteSync implements github.com/tendermint/go-db/db.DB.DeleteSync
func (a *DBAdapter) DeleteSync(key []byte) {
	a.Delete(key)
}

// Close implements github.com/tendermint/go-db/db.DB.Close
func (a *DBAdapter) Close() {

}

// NewBatch implements github.com/tendermint/go-db/db.DB.NewBatch
func (a *DBAdapter) NewBatch() db.Batch {
	return NewBatchAdapter(a.storeAdapter.NewBatch())
}

// Print is for debugging
func (a *DBAdapter) Print() {

}

// SetAdapter sets a new adapter on the DB
func (a *DBAdapter) SetAdapter(adapter store.Adapter) {
	a.storeAdapter = adapter
}
