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
	sqlFindSegments = `
		SELECT l.link_hash, l.data, e.data FROM links l
		LEFT JOIN evidences e ON l.link_hash = e.link_hash
		WHERE (length($3) = 0 OR process = $3)
		ORDER BY priority DESC, l.created_at DESC
		OFFSET $1 LIMIT $2
	`
	sqlFindSegmentsWithLinkHashes = `
		SELECT l.link_hash, l.data, e.data FROM links l
		LEFT JOIN evidences e ON l.link_hash = e.link_hash
		WHERE l.link_hash = any($1::bytea[])
		AND (length($4) = 0 OR process = $4)
		ORDER BY priority DESC, l.created_at DESC
		OFFSET $2 LIMIT $3
	`
	sqlFindSegmentsWithLinkHashesAndMapIDs = `
		SELECT l.link_hash, l.data, e.data FROM links l
		LEFT JOIN evidences e ON l.link_hash = e.link_hash
		WHERE l.link_hash = any($1::bytea[])
		AND map_id = any($2::text[])
		AND (length($5) = 0 OR process = $5)
		ORDER BY priority DESC, l.created_at DESC
		OFFSET $3 LIMIT $4
	`
	sqlFindSegmentsWithLinkHashesAndTags = `
		SELECT l.link_hash, l.data, e.data FROM links l
		LEFT JOIN evidences e ON l.link_hash = e.link_hash
		WHERE l.link_hash = any($1::bytea[])
		AND tags @> $2
		AND (length($5) = 0 OR process = $5)
		ORDER BY priority DESC, l.created_at DESC
		OFFSET $3 LIMIT $4
	`
	sqlFindSegmentsWithLinkHashesAndMapIDsAndTags = `
		SELECT l.link_hash, l.data, e.data FROM links l
		LEFT JOIN evidences e ON l.link_hash = e.link_hash
		WHERE l.link_hash = any($1::bytea[])
		AND map_id = any($2::text[]) AND tags @> $3
		AND (length($6) = 0 OR process = $6)
		ORDER BY priority DESC, l.created_at DESC
		OFFSET $4 LIMIT $5
	`
	sqlFindSegmentsWithMapIDs = `
		SELECT l.link_hash, l.data, e.data FROM links l
		LEFT JOIN evidences e ON l.link_hash = e.link_hash
		WHERE map_id = any($1::text[])
		AND (length($4) = 0 OR process = $4)
		ORDER BY priority DESC, l.created_at DESC
		OFFSET $2 LIMIT $3
	`
	sqlFindSegmentsWithPrevLinkHash = `
		SELECT l.link_hash, l.data, e.data FROM links l
		LEFT JOIN evidences e ON l.link_hash = e.link_hash
		WHERE prev_link_hash = $1
		AND (length($4) = 0 OR process = $4)
		ORDER BY priority DESC, l.created_at DESC
		OFFSET $2 LIMIT $3
	`
	sqlFindSegmentsWithTags = `
		SELECT l.link_hash, l.data, e.data FROM links l
		LEFT JOIN evidences e ON l.link_hash = e.link_hash
		WHERE tags @> $1
		AND (length($4) = 0 OR process = $4)
		ORDER BY priority DESC, l.created_at DESC
		OFFSET $2 LIMIT $3
	`
	sqlFindSegmentsWithMapIDsAndTags = `
		SELECT l.link_hash, l.data, e.data FROM links l
		LEFT JOIN evidences e ON l.link_hash = e.link_hash
		WHERE map_id = any($1::text[]) AND tags @> $2
		AND (length($5) = 0 OR process = $5)
		ORDER BY priority DESC, l.created_at DESC
		OFFSET $3 LIMIT $4
	`
	sqlFindSegmentsWithPrevLinkHashAndTags = `
		SELECT l.link_hash, l.data, e.data FROM links l
		LEFT JOIN evidences e ON l.link_hash = e.link_hash
		WHERE prev_link_hash = $1 AND tags @> $2
		AND (length($5) = 0 OR process = $5)
		ORDER BY priority DESC, l.created_at DESC
		OFFSET $3 LIMIT $4
	`
	sqlFindSegmentsWithPrevLinkHashAndMapIDs = `
		SELECT l.link_hash, l.data, e.data FROM links l
		LEFT JOIN evidences e ON l.link_hash = e.link_hash
		WHERE prev_link_hash = $1
		AND map_id = any($2::text[])
		AND (length($5) = 0 OR process = $5)
		ORDER BY priority DESC, l.created_at DESC
		OFFSET $3 LIMIT $4
	`
	sqlFindSegmentsWithPrevLinkHashAndMapIDsAndTags = `
		SELECT l.link_hash, l.data, e.data FROM links l
		LEFT JOIN evidences e ON l.link_hash = e.link_hash
		WHERE prev_link_hash = $1
		AND map_id = any($2::text[]) AND tags @> $3
		AND (length($6) = 0 OR process = $6)
		ORDER BY priority DESC, l.created_at DESC
		OFFSET $4 LIMIT $5
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
		ON links (map_id)
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
	GetSegment   *sql.Stmt
	FindSegments *sql.Stmt
	GetMapIDs    *sql.Stmt
	GetValue     *sql.Stmt
	GetEvidences *sql.Stmt

	FindSegmentsWithMapIDs                       *sql.Stmt
	FindSegmentsWithPrevLinkHash                 *sql.Stmt
	FindSegmentsWithTags                         *sql.Stmt
	FindSegmentsWithLinkHashes                   *sql.Stmt
	FindSegmentsWithLinkHashesAndMapIDs          *sql.Stmt
	FindSegmentsWithLinkHashesAndTags            *sql.Stmt
	FindSegmentsWithLinkHashesAndMapIDsAndTags   *sql.Stmt
	FindSegmentsWithTagsAndLimit                 *sql.Stmt
	FindSegmentsWithMapIDsAndTags                *sql.Stmt
	FindSegmentsWithPrevLinkHashAndTags          *sql.Stmt
	FindSegmentsWithPrevLinkHashAndMapIDs        *sql.Stmt
	FindSegmentsWithPrevLinkHashAndMapIDsAndTags *sql.Stmt
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
	s.FindSegments = prepare(sqlFindSegments)
	s.GetMapIDs = prepare(sqlGetMapIDs)
	s.GetValue = prepare(sqlGetValue)
	s.GetEvidences = prepare(sqlGetEvidences)

	s.FindSegmentsWithMapIDs = prepare(sqlFindSegmentsWithMapIDs)
	s.FindSegmentsWithPrevLinkHash = prepare(sqlFindSegmentsWithPrevLinkHash)
	s.FindSegmentsWithTags = prepare(sqlFindSegmentsWithTags)

	s.FindSegmentsWithLinkHashes = prepare(sqlFindSegmentsWithLinkHashes)
	s.FindSegmentsWithLinkHashesAndMapIDs = prepare(sqlFindSegmentsWithLinkHashesAndMapIDs)
	s.FindSegmentsWithLinkHashesAndTags = prepare(sqlFindSegmentsWithLinkHashesAndTags)
	s.FindSegmentsWithLinkHashesAndMapIDsAndTags = prepare(sqlFindSegmentsWithLinkHashesAndMapIDsAndTags)

	s.FindSegmentsWithMapIDsAndTags = prepare(sqlFindSegmentsWithMapIDsAndTags)
	s.FindSegmentsWithPrevLinkHashAndTags = prepare(sqlFindSegmentsWithPrevLinkHashAndTags)
	s.FindSegmentsWithPrevLinkHashAndMapIDs = prepare(sqlFindSegmentsWithPrevLinkHashAndMapIDs)
	s.FindSegmentsWithPrevLinkHashAndMapIDsAndTags = prepare(sqlFindSegmentsWithPrevLinkHashAndMapIDsAndTags)

	s.CreateLink = prepare(sqlCreateLink)
	s.DeleteLink = prepare(sqlDeleteLink)
	s.SaveValue = prepare(sqlSaveValue)
	s.DeleteValue = prepare(sqlDeleteValue)
	s.AddEvidence = prepare(sqlAddEvidence)

	if err != nil {
		return nil, err
	}

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
	s.FindSegments = prepare(sqlFindSegments)
	s.GetMapIDs = prepare(sqlGetMapIDs)
	s.GetValue = prepare(sqlGetValue)

	s.CreateLink = prepare(sqlCreateLink)
	s.DeleteLink = prepare(sqlDeleteLink)
	s.SaveValue = prepare(sqlSaveValue)
	s.DeleteValue = prepare(sqlDeleteValue)

	if err != nil {
		return nil, err
	}

	return &s, nil
}
