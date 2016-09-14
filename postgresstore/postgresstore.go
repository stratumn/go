// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package postgresstore implements a store that saves all the segments in a PostgreSQL database.
// It requires PostgreSQL >= 9.5 for "ON CONFLICT DO UPDATE" support.
package postgresstore

import (
	"database/sql"
	"encoding/json"

	"github.com/lib/pq"
	"github.com/stratumn/go/cs"
	"github.com/stratumn/go/store"
	"github.com/stratumn/go/types"
)

const (
	// Name is the name set in the store's information.
	Name = "postgres"

	// Description is the description set in the store's information.
	Description = "Stratumn PostgreSQL Store"

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

// Store is the type that implements github.com/stratumn/go/store.Adapter.
type Store struct {
	config *Config
	db     *sql.DB
	stmts  *stmts
}

// New creates an instance of a Store.
func New(config *Config) (*Store, error) {
	db, err := sql.Open("postgres", config.URL)
	if err != nil {
		return nil, err
	}
	return &Store{config: config, db: db}, nil
}

// GetInfo implements github.com/stratumn/go/store.Adapter.GetInfo.
func (a *Store) GetInfo() (interface{}, error) {
	return &Info{
		Name:        Name,
		Description: Description,
		Version:     a.config.Version,
		Commit:      a.config.Commit,
	}, nil
}

// SaveSegment implements github.com/stratumn/go/store.Adapter.SaveSegment.
func (a *Store) SaveSegment(segment *cs.Segment) error {
	var (
		err          error
		linkHash     = segment.GetLinkHash()
		priority     = segment.Link.GetPriority()
		mapID        = segment.Link.GetMapID()
		prevLinkHash = segment.Link.GetPrevLinkHash()
		tags         = segment.Link.GetTags()
	)

	data, err := json.Marshal(segment)
	if err != nil {
		return err
	}

	if prevLinkHash == nil {
		_, err = a.stmts.SaveSegment.Exec(linkHash[:], priority, mapID, nil, pq.Array(tags), string(data))
	} else {
		_, err = a.stmts.SaveSegment.Exec(linkHash[:], priority, mapID, prevLinkHash[:], pq.Array(tags), string(data))
	}

	return err
}

// GetSegment implements github.com/stratumn/go/store.Adapter.GetSegment.
func (a *Store) GetSegment(linkHash *types.Bytes32) (*cs.Segment, error) {
	var data string

	if err := a.stmts.GetSegment.QueryRow(linkHash[:]).Scan(&data); err != nil {
		if err.Error() == notFoundError {
			return nil, nil
		}
		return nil, err
	}

	var segment cs.Segment
	if err := json.Unmarshal([]byte(data), &segment); err != nil {
		return nil, err
	}

	return &segment, nil
}

// DeleteSegment implements github.com/stratumn/go/store.Adapter.DeleteSegment.
func (a *Store) DeleteSegment(linkHash *types.Bytes32) (*cs.Segment, error) {
	var (
		data    string
		segment cs.Segment
	)

	if err := a.stmts.DeleteSegment.QueryRow(linkHash[:]).Scan(&data); err != nil {
		if err.Error() == notFoundError {
			return nil, nil
		}
		return nil, err
	}

	if err := json.Unmarshal([]byte(data), &segment); err != nil {
		return nil, err
	}

	return &segment, nil
}

// FindSegments implements github.com/stratumn/go/store.Adapter.FindSegments.
func (a *Store) FindSegments(filter *store.Filter) (cs.SegmentSlice, error) {
	var (
		rows     *sql.Rows
		err      error
		limit    = filter.Limit
		offset   = filter.Offset
		segments = make(cs.SegmentSlice, 0, limit)
	)

	if filter.PrevLinkHash != nil {
		prevLinkHash := filter.PrevLinkHash[:]
		if len(filter.Tags) > 0 {
			tags := pq.Array(filter.Tags)
			rows, err = a.stmts.FindSegmentsWithPrevLinkHashAndTags.Query(prevLinkHash, tags, offset, limit)
		} else {
			rows, err = a.stmts.FindSegmentsWithPrevLinkHash.Query(prevLinkHash, offset, limit)
		}
	} else if mapID := filter.MapID; mapID != "" {
		if len(filter.Tags) > 0 {
			tags := pq.Array(filter.Tags)
			rows, err = a.stmts.FindSegmentsWithMapIDAndTags.Query(mapID, tags, offset, limit)
		} else {
			rows, err = a.stmts.FindSegmentsWithMapID.Query(mapID, offset, limit)
		}
	} else if len(filter.Tags) > 0 {
		tags := pq.Array(filter.Tags)
		rows, err = a.stmts.FindSegmentsWithTags.Query(tags, offset, limit)
	} else {
		rows, err = a.stmts.FindSegments.Query(offset, limit)
	}

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var (
			data    string
			segment cs.Segment
		)

		if err = rows.Scan(&data); err != nil {
			return nil, err
		}

		if err = json.Unmarshal([]byte(data), &segment); err != nil {
			return nil, err
		}

		segments = append(segments, &segment)
	}

	return segments, nil
}

// GetMapIDs implements github.com/stratumn/go/store.Adapter.GetMapIDs.
func (a *Store) GetMapIDs(pagination *store.Pagination) ([]string, error) {
	rows, err := a.stmts.GetMapIDs.Query(pagination.Offset, pagination.Limit)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	mapIDs := make([]string, 0, pagination.Limit)

	for rows.Next() {
		var mapID string
		if err = rows.Scan(&mapID); err != nil {
			return nil, err
		}

		mapIDs = append(mapIDs, mapID)
	}

	return mapIDs, nil
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
func (a *Store) Prepare() (err error) {
	a.stmts, err = newStmts(a.db)
	return
}

// Drop drops the database tables and indexes.
func (a *Store) Drop() error {
	for _, query := range sqlDrop {
		if _, err := a.db.Exec(query); err != nil {
			return err
		}
	}
	return nil
}
