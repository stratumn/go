package postgresstore

import "database/sql"

const (
	sqlSaveSegment = `
		INSERT INTO segments (data)
		VALUES ($1)
		ON CONFLICT ((data#>>'{meta,linkHash}'))
		DO UPDATE SET data = $1
	`
	sqlGetSegment = `
		SELECT data FROM segments
		WHERE data#>>'{meta,linkHash}' = $1
	`
	sqlDeleteSegment = `
		DELETE FROM segments
		WHERE data#>>'{meta,linkHash}' = $1
		RETURNING data
	`
	sqlFindSegments = `
		SELECT data FROM segments
		ORDER BY data#>>'{link,meta,priority}' DESC NULLS LAST, created_at DESC
		OFFSET $1 LIMIT $2
	`
	sqlFindSegmentsWithMapID = `
		SELECT data FROM segments
		WHERE data#>>'{link,meta,mapId}' = $1
		ORDER BY data#>>'{link,meta,priority}' DESC NULLS LAST, created_at DESC
		OFFSET $2 LIMIT $3
	`
	sqlFindSegmentsWithPrevLinkHash = `
		SELECT data FROM segments
		WHERE data#>>'{link,meta,prevLinkHash}' = $1
		ORDER BY data#>>'{link,meta,priority}' DESC NULLS LAST, created_at DESC
		OFFSET $2 LIMIT $3
	`
	sqlFindSegmentsWithTags = `
		SELECT data FROM segments
		WHERE data#>'{link,meta,tags}' ?& $1
		ORDER BY data#>>'{link,meta,priority}' DESC NULLS LAST, created_at DESC
		OFFSET $2 LIMIT $3
	`
	sqlFindSegmentsWithMapIDAndTags = `
		SELECT data FROM segments
		WHERE data#>>'{link,meta,mapId}' = $1 AND data#>'{link,meta,tags}' ?& $2
		ORDER BY data#>>'{link,meta,priority}' DESC NULLS LAST, created_at DESC
		OFFSET $3 LIMIT $4
	`
	sqlFindSegmentsWithPrevLinkHashAndTags = `
		SELECT data FROM segments
		WHERE data#>>'{link,meta,prevLinkHash}' = $1 AND data#>'{link,meta,tags}' ?& $2
		ORDER BY data#>>'{link,meta,priority}' DESC NULLS LAST, created_at DESC
		OFFSET $3 LIMIT $4
	`
	sqlGetMapIDs = `
		SELECT DISTINCT data#>>'{link,meta,mapId}' FROM segments
		ORDER BY data#>>'{link,meta,mapId}'
		OFFSET $1 LIMIT $2
	`
)

var sqlCreate = []string{
	`
		CREATE TABLE IF NOT EXISTS segments (
			id BIGSERIAL PRIMARY KEY,
			data jsonb NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`,
	`
		CREATE UNIQUE INDEX IF NOT EXISTS segments_link_hash_idx
		ON segments ((data#>>'{meta,linkHash}'))
	`,
	`
		CREATE INDEX IF NOT EXISTS segments_priority_created_at_idx
		ON segments ((data#>>'{link,meta,priority}') NULLS LAST, created_at)
	`,
	`
		CREATE INDEX IF NOT EXISTS segments_map_id_idx
		ON segments ((data#>>'{link,meta,mapId}'))
	`,
	`
		CREATE INDEX IF NOT EXISTS segments_map_id_priority_created_at_idx
		ON segments ((data#>>'{link,meta,mapId}'), (data#>>'{link,meta,priority}') NULLS LAST, created_at)
	`,
	`
		CREATE INDEX IF NOT EXISTS segments_prev_link_hash_priority_created_at_idx
		ON segments ((data#>>'{link,meta,prevLinkHash}'), (data#>>'{link,meta,priority}') NULLS LAST, created_at)
	`,
	`
		CREATE INDEX IF NOT EXISTS segments_tags_idx
		ON segments USING gin((data#>'{link,meta,tags}'))
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

type stmts struct {
	SaveSegment                         *sql.Stmt
	GetSegment                          *sql.Stmt
	DeleteSegment                       *sql.Stmt
	FindSegments                        *sql.Stmt
	FindSegmentsWithMapID               *sql.Stmt
	FindSegmentsWithPrevLinkHash        *sql.Stmt
	FindSegmentsWithTags                *sql.Stmt
	FindSegmentsWithTagsAndLimit        *sql.Stmt
	FindSegmentsWithMapIDAndTags        *sql.Stmt
	FindSegmentsWithPrevLinkHashAndTags *sql.Stmt
	GetMapIDs                           *sql.Stmt
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

	if err != nil {
		return nil, err
	}

	return &s, nil
}
