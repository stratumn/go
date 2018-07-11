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
	"fmt"
	"strings"

	"github.com/lib/pq"

	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/types"
)

const (
	sqlCreateLink = `
		INSERT INTO links (
			link_hash,
			priority,
			map_id,
			prev_link_hash,
			tags,
			data,
			process
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (link_hash)
		DO UPDATE SET
			priority = $2,
			map_id = $3,
			prev_link_hash = $4,
			tags = $5,
			data = $6,
			process = $7
	`
	sqlGetSegment = `
		SELECT l.link_hash, l.data, e.data FROM links l
		LEFT JOIN evidences e ON l.link_hash = e.link_hash
		WHERE l.link_hash = $1
	`
	sqlDeleteLink = `
		DELETE FROM links
		WHERE link_hash = $1
		RETURNING data
	`
	sqlGetMapIDs = `
		SELECT l.map_id FROM links l
		WHERE (length($3) = 0 OR process = $3)
		GROUP BY l.map_id
		ORDER BY MAX(l.updated_at) DESC
		OFFSET $1 LIMIT $2
	`
	sqlSaveValue = `
		INSERT INTO values (
			key,
			value
		)
		VALUES ($1, $2)
		ON CONFLICT (key)
		DO UPDATE SET
			value = $2
	`
	sqlGetValue = `
		SELECT value FROM values
		WHERE key = $1
	`
	sqlDeleteValue = `
		DELETE FROM values
		WHERE key = $1
		RETURNING value
	`
	sqlGetEvidences = `
		SELECT data FROM evidences
		WHERE link_hash = $1
	`
	sqlAddEvidence = `
		INSERT INTO evidences (
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
	`
		CREATE TABLE links (
			id BIGSERIAL PRIMARY KEY,
			link_hash bytea NOT NULL,
			priority double precision NOT NULL,
			map_id text NOT NULL,
			prev_link_hash bytea DEFAULT NULL,
			tags text[] DEFAULT NULL,
			data jsonb NOT NULL,
			process text NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`,
	`
		CREATE UNIQUE INDEX links_link_hash_idx
		ON links (link_hash)
	`,
	`
		CREATE INDEX links_priority_created_at_idx
		ON links (priority DESC, created_at DESC)
	`,
	`
		CREATE INDEX links_map_id_idx
		ON links (map_id text_pattern_ops)
	`,
	`
		CREATE INDEX links_map_id_priority_created_at_idx
		ON links (map_id, priority DESC, created_at DESC)
	`,
	`
		CREATE INDEX links_prev_link_hash_priority_created_at_idx
		ON links (prev_link_hash, priority DESC, created_at DESC)
	`,
	`
		CREATE INDEX links_tags_idx
		ON links USING gin(tags)
	`,
	`
		CREATE TABLE evidences (
			id BIGSERIAL PRIMARY KEY,
			link_hash bytea NOT NULL,
			provider text NOT NULL,
			data jsonb NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)
	`,
	`
		CREATE UNIQUE INDEX evidences_link_hash_provider_idx
		ON evidences (link_hash, provider)
	`,
	`
		CREATE INDEX evidences_link_hash_idx
		ON evidences (link_hash)
	`,
	`
		CREATE TABLE values (
			id BIGSERIAL PRIMARY KEY,
			key bytea NOT NULL,
			value bytea NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`,
	`
		CREATE UNIQUE INDEX values_key_idx
		ON values (key)
	`,
}

var sqlDrop = []string{
	"DROP TABLE links, evidences, values",
}

type writeStmts struct {
	CreateLink  *sql.Stmt
	DeleteLink  *sql.Stmt
	SaveValue   *sql.Stmt
	DeleteValue *sql.Stmt
	AddEvidence *sql.Stmt
}

type readStmts struct {
	// DB.Query or Tx.Query depending on if we are in batch.
	query func(query string, args ...interface{}) (*sql.Rows, error)

	GetSegment   *sql.Stmt
	GetMapIDs    *sql.Stmt
	GetValue     *sql.Stmt
	GetEvidences *sql.Stmt
}

type stmts struct {
	readStmts
	writeStmts
}

type batchStmts stmts

func newStmts(db *sql.DB) (*stmts, error) {
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

	s.GetSegment = prepare(sqlGetSegment)
	s.GetMapIDs = prepare(sqlGetMapIDs)
	s.GetValue = prepare(sqlGetValue)
	s.GetEvidences = prepare(sqlGetEvidences)

	s.CreateLink = prepare(sqlCreateLink)
	s.DeleteLink = prepare(sqlDeleteLink)
	s.SaveValue = prepare(sqlSaveValue)
	s.DeleteValue = prepare(sqlDeleteValue)
	s.AddEvidence = prepare(sqlAddEvidence)

	if err != nil {
		return nil, err
	}

	s.query = db.Query

	return &s, nil
}

func newBatchStmts(tx *sql.Tx) (*batchStmts, error) {
	var (
		s   batchStmts
		err error
	)

	prepare := func(str string) (stmt *sql.Stmt) {
		if err == nil {
			stmt, err = tx.Prepare(str)
		}
		return
	}

	s.GetSegment = prepare(sqlGetSegment)
	s.GetMapIDs = prepare(sqlGetMapIDs)
	s.GetValue = prepare(sqlGetValue)

	s.CreateLink = prepare(sqlCreateLink)
	s.DeleteLink = prepare(sqlDeleteLink)
	s.SaveValue = prepare(sqlSaveValue)
	s.DeleteValue = prepare(sqlDeleteValue)

	if err != nil {
		return nil, err
	}

	s.query = tx.Query

	return &s, nil
}

// FindSegments formats a read query and retrieves segments according to the filter.
func (s *readStmts) FindSegmentsWithFilters(filter *store.SegmentFilter) (*sql.Rows, error) {
	sqlHead := `SELECT l.link_hash, l.data, e.data FROM links l
	LEFT JOIN evidences e ON l.link_hash = e.link_hash
	`

	sqlTail := fmt.Sprintf(`
		ORDER BY priority DESC, l.created_at DESC
		OFFSET %d LIMIT %d
		`,
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

	if filter.MapIDPrefix != "" {
		filters = append(filters, fmt.Sprintf("map_id LIKE $%d", cnt))
		values = append(values, fmt.Sprintf("%s%%", filter.MapIDPrefix))
		cnt++
	}

	if filter.MapIDSuffix != "" {
		filters = append(filters, fmt.Sprintf("map_id LIKE $%d", cnt))
		values = append(values, fmt.Sprintf("%%%s", filter.MapIDSuffix))
		cnt++
	}

	if filter.Process != "" {
		filters = append(filters, fmt.Sprintf("process = $%d", cnt))
		values = append(values, filter.Process)
		cnt++
	}

	if filter.PrevLinkHash != nil {

		if *filter.PrevLinkHash == "" {
			filters = append(filters, "prev_link_hash = '\\x'")
		} else {
			prevLinkHashBytes, err := types.NewBytes32FromString(*filter.PrevLinkHash)
			if err != nil {
				return nil, err
			}

			filters = append(filters, fmt.Sprintf("prev_link_hash = $%d", cnt))
			values = append(values, prevLinkHashBytes[:])
			cnt++
		}
	}

	if len(filter.LinkHashes) > 0 {
		linkHashes, err := cs.NewLinkHashesFromStrings(filter.LinkHashes)
		if err != nil {
			return nil, err
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

	return s.query(query, values...)
}
