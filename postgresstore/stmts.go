// Copyright 2016-2018 Stratumn SAS. All rights reserved.
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
	"fmt"
	"strings"

	"github.com/lib/pq"
	"github.com/stratumn/go-core/monitoring/errorcode"
	"github.com/stratumn/go-core/store"
	"github.com/stratumn/go-core/types"
)

// Plain SQL statements.
// They need to be prepared before they can be used.
const (
	SQLCreateLink = `
		INSERT INTO store.links (
			link_hash,
			priority,
			map_id,
			prev_link_hash,
			tags,
			data,
			process,
			step
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	SQLCreateLinkDegree = `
		INSERT INTO store_private.links_degree (
			link_hash,
			out_degree
		)
		VALUES ($1, 0)
	`
	SQLLockLinkDegree = `
		SELECT out_degree FROM store_private.links_degree 
		WHERE link_hash = $1 FOR UPDATE
	`
	SQLUpdateLinkDegree = `
		UPDATE store_private.links_degree SET out_degree = $2 
		WHERE link_hash = $1
	`
	SQLInitMap = `
		INSERT INTO store_private.process_maps (
			process,
			map_id
		)
		VALUES ($1, $2)
	`
	SQLGetSegment = `
		SELECT l.link_hash, l.data, e.data FROM store.links l
		LEFT JOIN store.evidences e ON l.link_hash = e.link_hash
		WHERE l.link_hash = $1
	`
	SQLSaveValue = `
		INSERT INTO store.values (
			key,
			value
		)
		VALUES ($1, $2)
		ON CONFLICT (key)
		DO UPDATE SET
			value = $2
	`
	SQLGetValue = `
		SELECT value FROM store.values
		WHERE key = $1
	`
	SQLDeleteValue = `
		DELETE FROM store.values
		WHERE key = $1
		RETURNING value
	`
	SQLGetEvidences = `
		SELECT data FROM store.evidences
		WHERE link_hash = $1
	`
	SQLAddEvidence = `
		INSERT INTO store.evidences (
			link_hash,
			provider,
			data
		)
		VALUES ($1, $2, $3)
		ON CONFLICT (link_hash, provider)
		DO NOTHING
	`
)

var sqlCreate = []string{
	`CREATE SCHEMA IF NOT EXISTS store`,
	`CREATE SCHEMA IF NOT EXISTS store_private`,
	`
		CREATE TABLE IF NOT EXISTS store.links (
			id BIGSERIAL PRIMARY KEY,
			link_hash bytea NOT NULL UNIQUE,
			priority double precision NOT NULL,
			map_id text NOT NULL,
			prev_link_hash bytea DEFAULT NULL,
			tags text[] DEFAULT NULL,
			data bytea NOT NULL,
			process text NOT NULL,
			step text NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`,
	`
		CREATE INDEX IF NOT EXISTS links_priority_created_at_idx
		ON store.links (priority DESC, created_at DESC)
	`,
	`
		CREATE INDEX IF NOT EXISTS links_map_id_idx
		ON store.links (map_id text_pattern_ops)
	`,
	`
		CREATE INDEX IF NOT EXISTS links_map_id_priority_created_at_idx
		ON store.links (map_id, priority DESC, created_at DESC)
	`,
	`
		CREATE INDEX IF NOT EXISTS links_prev_link_hash_priority_created_at_idx
		ON store.links (prev_link_hash, priority DESC, created_at DESC)
	`,
	`
		CREATE INDEX IF NOT EXISTS links_tags_idx
		ON store.links USING gin(tags)
	`,
	`
		CREATE TABLE IF NOT EXISTS store_private.links_degree (
			id BIGSERIAL PRIMARY KEY,
			link_hash bytea references store.links(link_hash) UNIQUE,
			out_degree integer
		)
	`,
	`
		CREATE TABLE IF NOT EXISTS store.evidences (
			id BIGSERIAL PRIMARY KEY,
			link_hash bytea references store.links(link_hash),
			provider text NOT NULL,
			data bytea NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(link_hash, provider)
		)
	`,
	`
		CREATE INDEX IF NOT EXISTS evidences_link_hash_idx
		ON store.evidences (link_hash)
	`,
	`
		CREATE TABLE IF NOT EXISTS store.values (
			id BIGSERIAL PRIMARY KEY,
			key bytea NOT NULL UNIQUE,
			value bytea NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`,
	`
		CREATE TABLE IF NOT EXISTS store_private.process_maps (
			id BIGSERIAL PRIMARY KEY,
			process text NOT NULL,
			map_id text NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(process, map_id)
		)
	`,
}

var sqlDrop = []string{
	"DROP SCHEMA store CASCADE",
	"DROP SCHEMA store_private CASCADE",
}

// SQLPreparer prepares statements.
type SQLPreparer interface {
	Prepare(query string) (*sql.Stmt, error)
}

// SQLQuerier executes queries.
type SQLQuerier interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

// SQLPreparerQuerier prepares statements and executes queries.
type SQLPreparerQuerier interface {
	SQLPreparer
	SQLQuerier
}

// stmts exposes prepared SQL queries in a DB or Tx scope.
type stmts struct {
	CreateLink       *sql.Stmt
	CreateLinkDegree *sql.Stmt
	GetSegment       *sql.Stmt
	LockLinkDegree   *sql.Stmt
	UpdateLinkDegree *sql.Stmt

	DeleteValue *sql.Stmt
	GetValue    *sql.Stmt
	SaveValue   *sql.Stmt

	AddEvidence  *sql.Stmt
	GetEvidences *sql.Stmt

	// DB.Query or Tx.Query depending on if we are in batch.
	query func(query string, args ...interface{}) (*sql.Rows, error)
}

func newStmts(db SQLPreparerQuerier) (*stmts, error) {
	var (
		s   stmts
		err error
	)

	prepare := func(str string) (stmt *sql.Stmt) {
		if err == nil {
			stmt, err = db.Prepare(str)
		}

		return
	}

	s.CreateLink = prepare(SQLCreateLink)
	s.CreateLinkDegree = prepare(SQLCreateLinkDegree)
	s.GetSegment = prepare(SQLGetSegment)
	s.LockLinkDegree = prepare(SQLLockLinkDegree)
	s.UpdateLinkDegree = prepare(SQLUpdateLinkDegree)

	s.DeleteValue = prepare(SQLDeleteValue)
	s.GetValue = prepare(SQLGetValue)
	s.SaveValue = prepare(SQLSaveValue)

	s.AddEvidence = prepare(SQLAddEvidence)
	s.GetEvidences = prepare(SQLGetEvidences)

	if err != nil {
		return nil, types.WrapError(err, errorcode.InvalidArgument, store.Component, "could not prepare statements")
	}

	s.query = db.Query

	return &s, nil
}

// GetMapIDsWithFilters retrieves maps ids from the store given some filters.
func (s *stmts) GetMapIDsWithFilters(filter *store.MapFilter) (*sql.Rows, error) {
	sqlHead := `
		SELECT l.map_id FROM store.links l
	`
	sqlTail := fmt.Sprintf(`
		GROUP BY l.map_id
		ORDER BY MAX(l.updated_at) DESC
		OFFSET %d LIMIT %d
	`,
		filter.Pagination.Offset,
		filter.Pagination.Limit,
	)

	filters := []string{}
	values := []interface{}{}
	cnt := 1

	if filter.Prefix != "" {
		filters = append(filters, fmt.Sprintf("map_id LIKE $%d", cnt))
		values = append(values, fmt.Sprintf("%s%%", filter.Prefix))
		cnt++
	}

	if filter.Suffix != "" {
		filters = append(filters, fmt.Sprintf("map_id LIKE $%d", cnt))
		values = append(values, fmt.Sprintf("%%%s", filter.Suffix))
		cnt++
	}

	if filter.Process != "" {
		filters = append(filters, fmt.Sprintf("process = $%d", cnt))
		values = append(values, filter.Process)
	}

	sqlBody := ""
	if len(filters) > 0 {
		sqlBody = "\nWHERE "
		sqlBody += strings.Join(filters, "\n AND ")
	}

	query := sqlHead + sqlBody + sqlTail

	rows, err := s.query(query, values...)
	if err != nil {
		return nil, types.WrapError(err, errorcode.InvalidArgument, store.Component, "could not get map ids")
	}

	return rows, nil
}

func getOrderingWay(reverse bool) string {
	if reverse {
		return "ASC"
	}
	return "DESC"
}

// FindSegments formats a read query and retrieves segments according to the filter.
func (s *stmts) FindSegmentsWithFilters(filter *store.SegmentFilter) (*sql.Rows, error) {
	// Method to count distinct over: https://www.sqlservercentral.com/Forums/FindPost1824788.aspx
	sqlTotalCount := `DENSE_RANK() OVER (ORDER BY l.link_hash ASC) +
	DENSE_RANK() OVER (ORDER BY l.link_hash DESC) - 1 AS total_count
	`

	sqlHead := fmt.Sprintf(`SELECT l.link_hash,
	l.data,
	e.data,
	%s
	FROM store.links l
	LEFT JOIN store.evidences e ON l.link_hash = e.link_hash
	`,
		sqlTotalCount,
	)

	sqlTail := fmt.Sprintf(`
		ORDER BY l.priority %[1]s, l.created_at %[1]s
		OFFSET %[2]d LIMIT %[3]d
		`,
		getOrderingWay(filter.Reverse),
		filter.Pagination.Offset,
		filter.Pagination.Limit,
	)

	filters := []string{}
	values := []interface{}{}
	cnt := 1

	if len(filter.MapIDs) > 0 {
		filters = append(filters, fmt.Sprintf("map_id = ANY($%d::text[])", cnt))
		values = append(values, pq.Array(filter.MapIDs))
		cnt++
	}

	if filter.Process != "" {
		filters = append(filters, fmt.Sprintf("process = $%d", cnt))
		values = append(values, filter.Process)
		cnt++
	}

	if filter.Step != "" {
		filters = append(filters, fmt.Sprintf("step = $%d", cnt))
		values = append(values, filter.Step)
		cnt++
	}

	if filter.WithoutParent {
		filters = append(filters, "prev_link_hash = '\\x'")
	} else if len(filter.PrevLinkHash) > 0 {
		filters = append(filters, fmt.Sprintf("prev_link_hash = $%d", cnt))
		values = append(values, filter.PrevLinkHash)
		cnt++
	}

	if len(filter.LinkHashes) > 0 {
		var linkHashes []*types.Bytes32
		for _, lh := range filter.LinkHashes {
			linkHashes = append(linkHashes, types.NewBytes32FromBytes(lh))
		}

		filters = append(filters, fmt.Sprintf("l.link_hash = ANY($%d::bytea[])", cnt))
		values = append(values, pq.Array(linkHashes))
		cnt++
	}

	if len(filter.Tags) > 0 {
		filters = append(filters, fmt.Sprintf("tags @>  $%d", cnt))
		values = append(values, pq.Array(filter.Tags))
	}

	sqlBody := ""
	if len(filters) > 0 {
		sqlBody = "\nWHERE "
		sqlBody += strings.Join(filters, "\n AND ")
	}

	query := sqlHead + sqlBody + sqlTail

	rows, err := s.query(query, values...)
	if err != nil {
		return nil, types.WrapError(err, errorcode.InvalidArgument, store.Component, "could not find segments")
	}

	return rows, nil
}
