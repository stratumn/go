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

package tmpop

import (
	"github.com/stratumn/sdk/store"
	"github.com/tendermint/tmlibs/db"
)

// DBAdapter implements github.com/tendermint/tmlibs/db/db.DB.
type DBAdapter struct {
	storeAdapter store.Adapter
}

// NewDBAdapter returns a new DB Adapter
func NewDBAdapter(a store.Adapter) *DBAdapter {
	return &DBAdapter{a}
}

// Get implements github.com/tendermint/tmlibs/db/db.DB.Get
func (a *DBAdapter) Get(key []byte) []byte {
	value, err := a.storeAdapter.GetValue(key)
	if err != nil {
		panic(err)
	}
	return value
}

// Set implements github.com/tendermint/tmlibs/db/db.DB.Set
func (a *DBAdapter) Set(key, value []byte) {
	err := a.storeAdapter.SaveValue(key, value)
	if err != nil {
		panic(err)
	}
}

// SetSync implements github.com/tendermint/tmlibs/db/db.DB.SetSync
func (a *DBAdapter) SetSync(key, value []byte) {
	a.Set(key, value)
}

// Delete implements github.com/tendermint/tmlibs/db/db.DB.Delete
func (a *DBAdapter) Delete(key []byte) {
	_, err := a.storeAdapter.DeleteValue(key)
	if err != nil {
		panic(err)
	}
}

// DeleteSync implements github.com/tendermint/tmlibs/db/db.DB.DeleteSync
func (a *DBAdapter) DeleteSync(key []byte) {
	a.Delete(key)
}

// Close implements github.com/tendermint/tmlibs/db/db.DB.Close
func (a *DBAdapter) Close() {

}

// NewBatch implements github.com/tendermint/tmlibs/db/db.DB.NewBatch
func (a *DBAdapter) NewBatch() db.Batch {
	b, err := a.storeAdapter.NewBatch()
	if err != nil {
		panic(err)
	}
	return NewBatchAdapter(b)
}

// Print implements github.com/tendermint/tmlibs/db/db.DB.Print. Print is for debugging
func (a *DBAdapter) Print() {

}

// Iterator implements github.com/tendermint/tmlibs/db/db.DB.Iterator. Iterator is for debugging.
func (a *DBAdapter) Iterator() db.Iterator {
	return nil
}

// Stats implements github.com/tendermint/tmlibs/db/db.DB.Stats. Stats is for debugging.
func (a *DBAdapter) Stats() map[string]string {
	return nil
}

// SetAdapter sets a new adapter on the DB
func (a *DBAdapter) SetAdapter(adapter store.Adapter) {
	a.storeAdapter = adapter
}
