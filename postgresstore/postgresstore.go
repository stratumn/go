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

// Package postgresstore implements a store that saves all the segments in a
// PostgreSQL database. It requires PostgreSQL >= 9.5 for
// "ON CONFLICT DO UPDATE" support.
package postgresstore

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/monitoring"
	"github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/types"
	"go.opencensus.io/trace"
)

const (
	// Name is the name set in the store's information.
	Name = "postgres"

	// Description is the description set in the store's information.
	Description = "Indigo's PostgreSQL Store"

	// DefaultURL is the default URL of the database.
	DefaultURL = "postgres://postgres@postgres/postgres?sslmode=disable"
)

const notFoundError = "sql: no rows in result set"

// Config contains configuration options for the store.
type Config struct {
	// A version string that will be set in the store's information.
	Version string

	// A git commit hash that will be set in the store's information.
	Commit string

	// The URL of the PostgreSQL database, such as
	// "postgres://postgres@localhost/store?sslmode=disable".
	URL string
}

// Info is the info returned by GetInfo.
type Info struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Commit      string `json:"commit"`
}

// Store is the type that implements github.com/stratumn/go-indigocore/store.Adapter.
type Store struct {
	*reader
	*writer
	config     *Config
	eventChans []chan *store.Event
	db         *sql.DB
	stmts      *stmts

	batches map[*Batch]*sql.Tx
}

// New creates an instance of a Store.
func New(config *Config) (*Store, error) {
	db, err := sql.Open("postgres", config.URL)
	if err != nil {
		return nil, err
	}
	return &Store{config: config, db: db, batches: make(map[*Batch]*sql.Tx)}, nil
}

// GetInfo implements github.com/stratumn/go-indigocore/store.Adapter.GetInfo.
func (a *Store) GetInfo(ctx context.Context) (_ interface{}, err error) {
	ctx, span := trace.StartSpan(ctx, "postgresstore/GetInfo")
	defer monitoring.SetSpanStatusAndEnd(span, err)

	return &Info{
		Name:        Name,
		Description: Description,
		Version:     a.config.Version,
		Commit:      a.config.Commit,
	}, nil
}

// NewBatch implements github.com/stratumn/go-indigocore/store.Adapter.NewBatch.
func (a *Store) NewBatch(ctx context.Context) (_ store.Batch, err error) {
	ctx, span := trace.StartSpan(ctx, "postgresstore/NewBatch")
	defer monitoring.SetSpanStatusAndEnd(span, err)

	for b := range a.batches {
		if b.done {
			delete(a.batches, b)
		}
	}

	tx, err := a.db.Begin()
	if err != nil {
		return nil, err
	}
	b, err := NewBatch(tx)
	if err != nil {
		return nil, err
	}
	a.batches[b] = tx

	return b, nil
}

// AddStoreEventChannel implements github.com/stratumn/go-indigocore/store.Adapter.AddStoreEventChannel
func (a *Store) AddStoreEventChannel(eventChan chan *store.Event) {
	a.eventChans = append(a.eventChans, eventChan)
}

// CreateLink implements github.com/stratumn/go-indigocore/store.LinkWriter.CreateLink.
func (a *Store) CreateLink(ctx context.Context, link *cs.Link) (_ *types.Bytes32, err error) {
	ctx, span := trace.StartSpan(ctx, "postgresstore/CreateLink")
	defer monitoring.SetSpanStatusAndEnd(span, err)

	linkHash, err := a.writer.CreateLink(ctx, link)
	if err != nil {
		return nil, err
	}

	linkEvent := store.NewSavedLinks(link)

	for _, c := range a.eventChans {
		c <- linkEvent
	}
	return linkHash, nil
}

// AddEvidence implements github.com/stratumn/go-indigocore/store.EvidenceWriter.AddEvidence.
func (a *Store) AddEvidence(ctx context.Context, linkHash *types.Bytes32, evidence *cs.Evidence) (err error) {
	ctx, span := trace.StartSpan(ctx, "postgresstore/AddEvidence")
	defer monitoring.SetSpanStatusAndEnd(span, err)

	data, err := json.Marshal(evidence)
	if err != nil {
		return err
	}

	_, err = a.stmts.AddEvidence.Exec(linkHash[:], evidence.Provider, data)
	if err != nil {
		return err
	}

	evidenceEvent := store.NewSavedEvidences()
	evidenceEvent.AddSavedEvidence(linkHash, evidence)

	for _, c := range a.eventChans {
		c <- evidenceEvent
	}

	return nil
}

// Create creates the database tables and indexes.
func (a *Store) Create() error {
	for _, query := range sqlCreate {
		if _, err := a.db.Exec(query); err != nil {
			return err
		}
	}
	return nil
}

// Prepare prepares the database stmts.
// It should be called once before interacting with segments.
// It assumes the tables have been created using Create().
func (a *Store) Prepare() error {
	stmts, err := newStmts(a.db)
	if err != nil {
		return err
	}
	a.stmts = stmts
	a.reader = &reader{stmts: a.stmts.readStmts}
	a.writer = &writer{stmts: a.stmts.writeStmts}

	return nil
}

// Drop drops the database tables and indexes. It also rollbacks started batches.
func (a *Store) Drop() error {
	for b, tx := range a.batches {
		if !b.done {
			err := tx.Rollback()
			if err != nil {
				return err
			}
		}
	}

	for _, query := range sqlDrop {
		if _, err := a.db.Exec(query); err != nil {
			return err
		}
	}
	return nil
}

// Close closes the database connection.
func (a *Store) Close() error {
	return a.db.Close()
}
