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
	"bytes"
	"context"
	"database/sql"

	"github.com/lib/pq"
	"github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-core/monitoring"
	"github.com/stratumn/go-core/store"
	"github.com/stratumn/go-core/types"
)

// CreateLink implements github.com/stratumn/go-core/store.Adapter.CreateLink.
func (s *scopedStore) CreateLink(ctx context.Context, link *chainscript.Link) (chainscript.LinkHash, error) {
	var (
		priority     = link.Meta.Priority
		mapID        = link.Meta.MapId
		prevLinkHash = link.Meta.GetPrevLinkHash()
		tags         = link.Meta.Tags
		process      = link.Meta.Process.Name
		step         = link.Meta.Step
	)

	linkHash, err := link.Hash()
	if err != nil {
		return linkHash, types.WrapError(err, monitoring.InvalidArgument, store.Component, "could not hash link")
	}

	data, err := chainscript.MarshalLink(link)
	if err != nil {
		return linkHash, types.WrapError(err, monitoring.InvalidArgument, store.Component, "could not marshal link")
	}

	if len(prevLinkHash) == 0 {
		err = s.createLink(linkHash, priority, mapID, []byte{}, tags, data, process, step)
		return linkHash, err
	}

	parent, err := s.GetSegment(ctx, prevLinkHash)
	if err != nil {
		return linkHash, err
	}

	parentDegree := parent.Link.Meta.OutDegree
	if parentDegree < 0 {
		err = s.createLink(linkHash, priority, mapID, prevLinkHash, tags, data, process, step)
		return linkHash, err
	}

	if parentDegree == 0 {
		return linkHash, types.WrapError(chainscript.ErrOutDegree, monitoring.FailedPrecondition, store.Component, "could not create link")
	}

	// Inserting the link and updating its parent's current degree needs to be
	// done in a transaction to protect the DB from race conditions.

	tx, err := s.txFactory.NewTx()
	if err != nil {
		return linkHash, err
	}

	currentDegree, err := s.getLinkDegree(tx, prevLinkHash)
	if err != nil {
		s.txFactory.RollbackTx(tx, err)
		return linkHash, err
	}

	if int(parentDegree) <= currentDegree {
		s.txFactory.RollbackTx(tx, chainscript.ErrOutDegree)
		return linkHash, types.WrapError(chainscript.ErrOutDegree, monitoring.FailedPrecondition, store.Component, "could not create link")
	}

	err = s.createLinkInTx(tx, linkHash, priority, mapID, prevLinkHash, tags, data, process, step)
	if err != nil {
		s.txFactory.RollbackTx(tx, err)
		return linkHash, err
	}

	err = s.incrementLinkDegree(tx, prevLinkHash, currentDegree)
	if err != nil {
		s.txFactory.RollbackTx(tx, err)
		return linkHash, err
	}

	return linkHash, s.txFactory.CommitTx(tx)
}

// getLinkDegree reads the current degree of the given link.
// It locks the associated row until the transaction completes.
func (s *scopedStore) getLinkDegree(tx *sql.Tx, linkHash chainscript.LinkHash) (int, error) {
	degreeLock, err := tx.Prepare(SQLLockLinkDegree)
	if err != nil {
		return 0, types.WrapError(err, monitoring.InvalidArgument, store.Component, "could not lock link degree")
	}

	row := degreeLock.QueryRow(linkHash)
	currentDegree := 0
	err = row.Scan(&currentDegree)

	// If the link doesn't have children yet, no rows will be found.
	// That should not be considered an error.
	if err == sql.ErrNoRows {
		return 0, nil
	}

	if err != nil {
		return 0, types.WrapError(err, monitoring.Internal, store.Component, "could not get link degree")
	}

	return currentDegree, nil
}

// incrementLinkDegree increments the degree of the given link.
// A lock should have been acquired previously by the transaction to ensure
// consistency.
func (s *scopedStore) incrementLinkDegree(tx *sql.Tx, linkHash chainscript.LinkHash, currentDegree int) error {
	updateDegree, err := tx.Prepare(SQLUpdateLinkDegree)
	if err != nil {
		return types.WrapError(err, monitoring.InvalidArgument, store.Component, "could not increment link degree")
	}

	_, err = updateDegree.Exec(linkHash, currentDegree+1)
	if err != nil {
		return types.WrapError(err, monitoring.Internal, store.Component, "could not increment link degree")
	}

	return nil
}

// createLink adds the given link to the DB.
func (s *scopedStore) createLink(
	linkHash chainscript.LinkHash,
	priority float64,
	mapID string,
	prevLinkHash chainscript.LinkHash,
	tags []string,
	data []byte,
	process string,
	step string,
) error {
	tx, err := s.txFactory.NewTx()
	if err != nil {
		return err
	}

	err = s.createLinkInTx(tx, linkHash, priority, mapID, prevLinkHash, tags, data, process, step)
	if err != nil {
		s.txFactory.RollbackTx(tx, err)
		return err
	}

	return s.txFactory.CommitTx(tx)
}

// createLink inserts the given link in a transaction context.
// If the link already exists it will return an error.
func (s *scopedStore) createLinkInTx(
	tx *sql.Tx,
	linkHash chainscript.LinkHash,
	priority float64,
	mapID string,
	prevLinkHash chainscript.LinkHash,
	tags []string,
	data []byte,
	process string,
	step string,
) error {
	createLink, err := tx.Prepare(SQLCreateLink)
	if err != nil {
		return types.WrapError(err, monitoring.InvalidArgument, store.Component, "could not create link")
	}

	res, err := createLink.Exec(linkHash, priority, mapID, prevLinkHash, pq.Array(tags), data, process, step)
	if err != nil {
		return types.WrapError(err, monitoring.Internal, store.Component, "could not create link")
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return types.WrapError(err, monitoring.Internal, store.Component, "could not create link")
	}

	if rowsAffected == 0 {
		return types.WrapError(store.ErrLinkAlreadyExists, monitoring.AlreadyExists, store.Component, "could not create link")
	}

	if s.enforceUniqueMapEntry && len(prevLinkHash) == 0 {
		initMap, err := tx.Prepare(SQLInitMap)
		if err != nil {
			return types.WrapError(store.ErrUniqueMapEntry, monitoring.FailedPrecondition, store.Component, "could not create link")
		}

		_, err = initMap.Exec(process, mapID)
		if err != nil {
			return types.WrapError(store.ErrUniqueMapEntry, monitoring.FailedPrecondition, store.Component, "could not create link")
		}
	}

	initDegree, err := tx.Prepare(SQLCreateLinkDegree)
	if err != nil {
		return types.WrapError(err, monitoring.Internal, store.Component, "could not create link")
	}

	_, err = initDegree.Exec(linkHash)
	if err != nil {
		return types.WrapError(err, monitoring.Internal, store.Component, "could not create link")
	}

	return nil
}

// GetSegment implements github.com/stratumn/go-core/store.SegmentReader.GetSegment.
func (s *scopedStore) GetSegment(ctx context.Context, linkHash chainscript.LinkHash) (*chainscript.Segment, error) {
	var segments = make(types.SegmentSlice, 0, 1)

	rows, err := s.stmts.GetSegment.Query(linkHash)
	if err != nil {
		return nil, types.WrapError(err, monitoring.Unavailable, store.Component, "could not get segment")
	}

	defer rows.Close()
	if err = scanLinkAndEvidences(rows, &segments, nil); err != nil {
		return nil, err
	}

	if len(segments) == 0 {
		return nil, nil
	}

	return segments[0], nil
}

// FindSegments implements github.com/stratumn/go-core/store.SegmentReader.FindSegments.
func (s *scopedStore) FindSegments(ctx context.Context, filter *store.SegmentFilter) (*types.PaginatedSegments, error) {
	rows, err := s.stmts.FindSegmentsWithFilters(filter)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	segments := &types.PaginatedSegments{Segments: make(types.SegmentSlice, 0, filter.Limit)}
	err = scanLinkAndEvidences(rows, &segments.Segments, &segments.TotalCount)

	return segments, err
}

func scanLinkAndEvidences(rows *sql.Rows, segments *types.SegmentSlice, totalCount *int) error {
	var currentSegment *chainscript.Segment
	var currentHash chainscript.LinkHash

	for rows.Next() {
		var (
			linkHash     chainscript.LinkHash
			linkData     []byte
			link         *chainscript.Link
			evidenceData []byte
			evidence     *chainscript.Evidence
			err          error
		)

		if totalCount == nil {
			if err := rows.Scan(&linkHash, &linkData, &evidenceData); err != nil {
				return types.WrapError(err, monitoring.Internal, store.Component, "could not scan rows")
			}
		} else {
			if err := rows.Scan(&linkHash, &linkData, &evidenceData, totalCount); err != nil {
				return types.WrapError(err, monitoring.Internal, store.Component, "could not scan rows")
			}
		}

		if !bytes.Equal(currentHash, linkHash) {
			link, err = chainscript.UnmarshalLink(linkData)
			if err != nil {
				return types.WrapError(err, monitoring.InvalidArgument, store.Component, "could not unmarshal link")
			}

			hash, err := link.Hash()
			if err != nil {
				return types.WrapError(err, monitoring.InvalidArgument, store.Component, "could not hash link")
			}
			currentHash = hash

			currentSegment, err = link.Segmentify()
			if err != nil {
				return types.WrapError(err, monitoring.InvalidArgument, store.Component, "could not segmentify")
			}

			*segments = append(*segments, currentSegment)
		}

		if len(evidenceData) > 0 {
			evidence, err = chainscript.UnmarshalEvidence(evidenceData)
			if err != nil {
				return types.WrapError(err, monitoring.InvalidArgument, store.Component, "could not unmarshal evidence")
			}

			if err := currentSegment.AddEvidence(evidence); err != nil {
				return types.WrapError(err, monitoring.InvalidArgument, store.Component, "could not add evidence")
			}
		}
	}

	err := rows.Err()
	if err != nil {
		return types.WrapError(err, monitoring.Internal, store.Component, "could not scan rows")
	}

	return nil
}

// GetMapIDs implements github.com/stratumn/go-core/store.SegmentReader.GetMapIDs.
func (s *scopedStore) GetMapIDs(ctx context.Context, filter *store.MapFilter) ([]string, error) {
	rows, err := s.stmts.GetMapIDsWithFilters(filter)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	mapIDs := make([]string, 0, filter.Pagination.Limit)

	for rows.Next() {
		var mapID string
		if err = rows.Scan(&mapID); err != nil {
			return nil, types.WrapError(err, monitoring.Internal, store.Component, "could not get map ids")
		}

		mapIDs = append(mapIDs, mapID)
	}

	return mapIDs, nil
}
