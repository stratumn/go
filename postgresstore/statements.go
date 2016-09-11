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
		OFFSET $1
    `
	sqlFindSegmentsWithLimit = `
		SELECT data FROM segments
		ORDER BY priority DESC, created_at DESC
		OFFSET $1 LIMIT $2
    `
	sqlFindSegmentsWithMapID = `
		SELECT data FROM segments
		WHERE map_id = $1
		ORDER BY priority DESC, created_at DESC
		OFFSET $2
    `
	sqlFindSegmentsWithMapIDAndLimit = `
		SELECT data FROM segments
		WHERE map_id = $1
		ORDER BY priority DESC, created_at DESC
		OFFSET $2 LIMIT $3
    `
	sqlFindSegmentsWithPrevLinkHash = `
		SELECT data FROM segments
		WHERE prev_link_hash = $1
		ORDER BY priority DESC, created_at DESC
		OFFSET $2
    `
	sqlFindSegmentsWithPrevLinkHashAndLimit = `
		SELECT data FROM segments
		WHERE prev_link_hash = $1
		ORDER BY priority DESC, created_at DESC
		OFFSET $2 LIMIT $3
    `
	sqlFindSegmentsWithTags = `
		SELECT data FROM segments
		WHERE tags @> $1
		ORDER BY priority DESC, created_at DESC
		OFFSET $2
    `
	sqlFindSegmentsWithTagsAndLimit = `
		SELECT data FROM segments
		WHERE tags @> $1
		ORDER BY priority DESC, created_at DESC
		OFFSET $2 LIMIT $3
    `
	sqlFindSegmentsWithMapIDAndTags = `
		SELECT data FROM segments
		WHERE map_id = $1 AND tags @> $2
		ORDER BY priority DESC, created_at DESC
		OFFSET $3
    `
	sqlFindSegmentsWithMapIDAndTagsAndLimit = `
		SELECT data FROM segments
		WHERE map_id = $1 AND tags @> $2
		ORDER BY priority DESC, created_at DESC
		OFFSET $3 LIMIT $4
    `
	sqlFindSegmentsWithPrevLinkHashAndTags = `
		SELECT data FROM segments
		WHERE prev_link_hash = $1 AND tags @> $2
		ORDER BY priority DESC, created_at DESC
		OFFSET $3
    `
	sqlFindSegmentsWithPrevLinkHashAndTagsAndLimit = `
		SELECT data FROM segments
		WHERE prev_link_hash = $1 AND tags @> $2
		ORDER BY priority DESC, created_at DESC
		OFFSET $3 LIMIT $4
    `
	sqlGetMapIDs = `
		SELECT DISTINCT map_id FROM segments
		ORDER BY map_id
		OFFSET $1
    `
	sqlGetMapIDsWithLimit = `
		SELECT DISTINCT map_id FROM segments
		ORDER BY map_id
		OFFSET $1 LIMIT $2
    `
)

var sqlCreate = []string{
	`
		CREATE TABLE IF NOT EXISTS segments (
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
		CREATE UNIQUE INDEX IF NOT EXISTS segments_link_hash_idx
		ON segments (link_hash)
	`,
	`
		CREATE INDEX IF NOT EXISTS segments_priority_created_at_idx
		ON segments (priority, created_at)
	`,
	`
		CREATE INDEX IF NOT EXISTS segments_map_id_idx
		ON segments (map_id)
	`,
	`
		CREATE INDEX IF NOT EXISTS segments_map_id_priority_created_at_idx
		ON segments (map_id, priority, created_at)
	`,
	`
		CREATE INDEX IF NOT EXISTS segments_prev_link_hash_priority_created_at_idx
		ON segments (prev_link_hash, priority, created_at)
	`,
	`
		CREATE INDEX IF NOT EXISTS segments_tags_idx
		ON segments USING gin(tags)
	`,
}

var sqlDrop = []string{
	"DROP TABLE IF EXISTS segments",
	"DROP INDEX IF EXISTS segments_priority_created_at_idx",
	"DROP INDEX IF EXISTS segments_map_id_idx",
	"DROP INDEX IF EXISTS segments_map_id_priority_created_at_idx",
	"DROP INDEX IF EXISTS segments_prev_link_hash_priority_created_at_idx",
	"DROP INDEX IF EXISTS segments_tags_idx",
}

type statements struct {
	SaveSegment                                 *sql.Stmt
	GetSegment                                  *sql.Stmt
	DeleteSegment                               *sql.Stmt
	FindSegments                                *sql.Stmt
	FindSegmentsWithLimit                       *sql.Stmt
	FindSegmentsWithMapID                       *sql.Stmt
	FindSegmentsWithMapIDAndLimit               *sql.Stmt
	FindSegmentsWithPrevLinkHash                *sql.Stmt
	FindSegmentsWithPrevLinkHashAndLimit        *sql.Stmt
	FindSegmentsWithTags                        *sql.Stmt
	FindSegmentsWithTagsAndLimit                *sql.Stmt
	FindSegmentsWithMapIDAndTags                *sql.Stmt
	FindSegmentsWithMapIDAndTagsAndLimit        *sql.Stmt
	FindSegmentsWithPrevLinkHashAndTags         *sql.Stmt
	FindSegmentsWithPrevLinkHashAndTagsAndLimit *sql.Stmt
	GetMapIDs                                   *sql.Stmt
	GetMapIDsWithLimit                          *sql.Stmt
}

func newStatements(db *sql.DB) (*statements, error) {
	var (
		s   statements
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
	s.FindSegmentsWithLimit = prepare(sqlFindSegmentsWithLimit)
	s.FindSegmentsWithMapID = prepare(sqlFindSegmentsWithMapID)
	s.FindSegmentsWithMapIDAndLimit = prepare(sqlFindSegmentsWithMapIDAndLimit)
	s.FindSegmentsWithPrevLinkHash = prepare(sqlFindSegmentsWithPrevLinkHash)
	s.FindSegmentsWithPrevLinkHashAndLimit = prepare(sqlFindSegmentsWithPrevLinkHashAndLimit)
	s.FindSegmentsWithTags = prepare(sqlFindSegmentsWithTags)
	s.FindSegmentsWithTagsAndLimit = prepare(sqlFindSegmentsWithTagsAndLimit)
	s.FindSegmentsWithMapIDAndTags = prepare(sqlFindSegmentsWithMapIDAndTags)
	s.FindSegmentsWithMapIDAndTagsAndLimit = prepare(sqlFindSegmentsWithMapIDAndTagsAndLimit)
	s.FindSegmentsWithPrevLinkHashAndTags = prepare(sqlFindSegmentsWithPrevLinkHashAndTags)
	s.FindSegmentsWithPrevLinkHashAndTagsAndLimit = prepare(sqlFindSegmentsWithPrevLinkHashAndTagsAndLimit)
	s.GetMapIDs = prepare(sqlGetMapIDs)
	s.GetMapIDsWithLimit = prepare(sqlGetMapIDsWithLimit)

	if err != nil {
		return nil, err
	}

	return &s, nil
}
