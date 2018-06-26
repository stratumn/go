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

// Package leveldbstore implements a simple key-value local store.
// It's efficient and can be used by stores to save key-value pairs.
package leveldbstore

import (
	"context"

	"github.com/tendermint/tmlibs/db"
)

// LevelDBStore implements github.com/stratumn/go-indigocore/store.KeyValueStore.
type LevelDBStore struct {
	config *Config
	kvDB   db.DB
}

// Config contains configuration options for the store.
type Config struct {
	// Path where key-value pairs will be saved.
	Path string
}

// New creates an instance of a LevelDBStore.
func New(config *Config) (*LevelDBStore, error) {
	db, err := db.NewGoLevelDB("keyvalue-store", config.Path)
	if err != nil {
		return nil, err
	}

	return &LevelDBStore{config: config, kvDB: db}, nil
}

// SetValue implements github.com/stratumn/go-indigocore/store.KeyValueStore.SetValue.
func (a *LevelDBStore) SetValue(ctx context.Context, key []byte, value []byte) error {
	a.kvDB.Set(key, value)
	return nil
}

// GetValue implements github.com/stratumn/go-indigocore/store.KeyValueStore.GetValue.
func (a *LevelDBStore) GetValue(ctx context.Context, key []byte) ([]byte, error) {
	return a.kvDB.Get(key), nil
}

// DeleteValue implements github.com/stratumn/go-indigocore/store.KeyValueStore.DeleteValue.
func (a *LevelDBStore) DeleteValue(ctx context.Context, key []byte) ([]byte, error) {
	v := a.kvDB.Get(key)

	if v != nil {
		a.kvDB.Delete(key)
		return v, nil
	}

	return nil, nil
}
