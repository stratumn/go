package postgresstore

import "database/sql"

const (
	sqlSaveSegment = `
		INSERT INTO segments (
			link_hash,
			priority,
			map_id,
			prev_link_hash,
			tags,
			data
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (link_hash)
		DO UPDATE SET
			priority = $2,
			map_id = $3,
			prev_link_hash = $4,
			tags = $5,
			data = $6
	`
	sqlGetSegment = `
		SELECT data FROM segments
		WHERE link_hash = $1
	`
	sqlDeleteSegment = `
		DELETE FROM segments
		WHERE link_hash = $1
		RETURNING data
	`
	sqlFindSegments = `
		SELECT data FROM segments
		ORDER BY priority DESC, created_at DESC
		OFFSET $1 LIMIT $2
	`
	sqlFindSegmentsWithMapID = `
		SELECT data FROM segments
		WHERE map_id = $1
		ORDER BY priority DESC, created_at DESC
		OFFSET $2 LIMIT $3
	`
	sqlFindSegmentsWithPrevLinkHash = `
		SELECT data FROM segments
		WHERE prev_link_hash = $1
		ORDER BY priority DESC, created_at DESC
		OFFSET $2 LIMIT $3
	`
	sqlFindSegmentsWithTags = `
		SELECT data FROM segments
		WHERE tags @> $1
		ORDER BY priority DESC, created_at DESC
		OFFSET $2 LIMIT $3
	`
	sqlFindSegmentsWithMapIDAndTags = `
		SELECT data FROM segments
		WHERE map_id = $1 AND tags @> $2
		ORDER BY priority DESC, created_at DESC
		OFFSET $3 LIMIT $4
	`
	sqlFindSegmentsWithPrevLinkHashAndTags = `
		SELECT data FROM segments
		WHERE prev_link_hash = $1 AND tags @> $2
		ORDER BY priority DESC, created_at DESC
		OFFSET $3 LIMIT $4
	`
	sqlGetMapIDs = `
		SELECT DISTINCT map_id FROM segments
		ORDER BY map_id
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
)

var sqlCreate = []string{
	`
		CREATE TABLE segments (
			id BIGSERIAL PRIMARY KEY,
			link_hash bytea NOT NULL,
			priority double precision NOT NULL,
			map_id text NOT NULL,
			prev_link_hash bytea DEFAULT NULL,
			tags text[] DEFAULT NULL,
			data jsonb NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`,
	`
		CREATE UNIQUE INDEX segments_link_hash_idx
		ON segments (link_hash)
	`,
	`
		CREATE INDEX segments_priority_created_at_idx
		ON segments (priority DESC, created_at DESC)
	`,
	`
		CREATE INDEX segments_map_id_idx
		ON segments (map_id)
	`,
	`
		CREATE INDEX segments_map_id_priority_created_at_idx
		ON segments (map_id, priority DESC, created_at DESC)
	`,
	`
		CREATE INDEX segments_prev_link_hash_priority_created_at_idx
		ON segments (prev_link_hash, priority DESC, created_at DESC)
	`,
	`
		CREATE INDEX segments_tags_idx
		ON segments USING gin(tags)
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
	"DROP TABLE segments, values",
}

type writeStmts struct {
	SaveSegment   *sql.Stmt
	DeleteSegment *sql.Stmt
	SaveValue     *sql.Stmt
	DeleteValue   *sql.Stmt
}

type stmts struct {
	writeStmts

	GetSegment                          *sql.Stmt
	FindSegments                        *sql.Stmt
	FindSegmentsWithMapID               *sql.Stmt
	FindSegmentsWithPrevLinkHash        *sql.Stmt
	FindSegmentsWithTags                *sql.Stmt
	FindSegmentsWithTagsAndLimit        *sql.Stmt
	FindSegmentsWithMapIDAndTags        *sql.Stmt
	FindSegmentsWithPrevLinkHashAndTags *sql.Stmt
	GetMapIDs                           *sql.Stmt
	GetValue                            *sql.Stmt
}

type batchStmts struct {
	writeStmts
}

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

	s.SaveSegment = prepare(sqlSaveSegment)
	s.GetSegment = prepare(sqlGetSegment)
	s.DeleteSegment = prepare(sqlDeleteSegment)
	s.FindSegments = prepare(sqlFindSegments)
	s.FindSegmentsWithMapID = prepare(sqlFindSegmentsWithMapID)
	s.FindSegmentsWithPrevLinkHash = prepare(sqlFindSegmentsWithPrevLinkHash)
	s.FindSegmentsWithTags = prepare(sqlFindSegmentsWithTags)
	s.FindSegmentsWithMapIDAndTags = prepare(sqlFindSegmentsWithMapIDAndTags)
	s.FindSegmentsWithPrevLinkHashAndTags = prepare(sqlFindSegmentsWithPrevLinkHashAndTags)
	s.GetMapIDs = prepare(sqlGetMapIDs)
	s.SaveValue = prepare(sqlSaveValue)
	s.DeleteValue = prepare(sqlDeleteValue)
	s.GetValue = prepare(sqlGetValue)

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

	s.SaveSegment = prepare(sqlSaveSegment)
	s.DeleteSegment = prepare(sqlDeleteSegment)
	s.SaveValue = prepare(sqlSaveValue)
	s.DeleteValue = prepare(sqlDeleteValue)

	if err != nil {
		return nil, err
	}

	return &s, nil
}
