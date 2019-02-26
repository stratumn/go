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
	"encoding/json"

	"github.com/lib/pq"
	"github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-core/monitoring/errorcode"
	"github.com/stratumn/go-core/store"
	"github.com/stratumn/go-core/types"
)

// CreateLink implements github.com/stratumn/go-core/store.Adapter.CreateLink.
func (s *scopedStore) CreateLink(ctx context.Context, link *chainscript.Link) (chainscript.LinkHash, error) {
	linkHash, err := link.Hash()
	if err != nil {
		return linkHash, types.WrapError(err, errorcode.InvalidArgument, store.Component, "could not hash link")
	}

	data, err := json.Marshal(link)
	if err != nil {
		return linkHash, types.WrapError(err, errorcode.InvalidArgument, store.Component, "could not marshal link")
	}

	if len(link.PrevLinkHash()) == 0 {
		err = s.createLink(ctx, linkHash, data, link)
		return linkHash, err
	}

	parent, err := s.GetSegment(ctx, link.PrevLinkHash())
	if err != nil {
		return linkHash, err
	}

	parentDegree := parent.Link.Meta.OutDegree
	if parentDegree < 0 {
		err = s.createLink(ctx, linkHash, data, link)
		return linkHash, err
	}

	if parentDegree == 0 {
		return linkHash, types.WrapError(chainscript.ErrOutDegree, errorcode.FailedPrecondition, store.Component, "could not create link")
	}

	// Inserting the link and updating its parent's current degree needs to be
	// done in a transaction to protect the DB from race conditions.

	tx, err := s.txFactory.NewTx()
	if err != nil {
		return linkHash, err
	}

	currentDegree, err := s.getLinkDegree(ctx, tx, link.PrevLinkHash())
	if err != nil {
		s.txFactory.RollbackTx(tx, err)
		return linkHash, err
	}

	if int(parentDegree) <= currentDegree {
		s.txFactory.RollbackTx(tx, chainscript.ErrOutDegree)
		return linkHash, types.WrapError(chainscript.ErrOutDegree, errorcode.FailedPrecondition, store.Component, "could not create link")
	}

	err = s.createLinkInTx(ctx, tx, linkHash, data, link)
	if err != nil {
		s.txFactory.RollbackTx(tx, err)
		return linkHash, err
	}

	err = s.incrementLinkDegree(ctx, tx, link.PrevLinkHash(), currentDegree)
	if err != nil {
		s.txFactory.RollbackTx(tx, err)
		return linkHash, err
	}

	return linkHash, s.txFactory.CommitTx(tx)
}

// getLinkDegree reads the current degree of the given link.
// It locks the associated row until the transaction completes.
func (s *scopedStore) getLinkDegree(ctx context.Context, tx *sql.Tx, linkHash chainscript.LinkHash) (int, error) {
	degreeLock, err := tx.Prepare(SQLLockLinkDegree)
	if err != nil {
		return 0, types.WrapError(err, errorcode.InvalidArgument, store.Component, "could not lock link degree")
	}

	row := degreeLock.QueryRowContext(ctx, linkHash)
	currentDegree := 0
	err = row.Scan(&currentDegree)

	// If the link doesn't have children yet, no rows will be found.
	// That should not be considered an error.
	if err == sql.ErrNoRows {
		return 0, nil
	}

	if err != nil {
		return 0, types.WrapError(err, errorcode.Internal, store.Component, "could not get link degree")
	}

	return currentDegree, nil
}

// incrementLinkDegree increments the degree of the given link.
// A lock should have been acquired previously by the transaction to ensure
// consistency.
func (s *scopedStore) incrementLinkDegree(ctx context.Context, tx *sql.Tx, linkHash chainscript.LinkHash, currentDegree int) error {
	updateDegree, err := tx.Prepare(SQLUpdateLinkDegree)
	if err != nil {
		return types.WrapError(err, errorcode.InvalidArgument, store.Component, "could not increment link degree")
	}

	_, err = updateDegree.ExecContext(ctx, linkHash, currentDegree+1)
	if err != nil {
		return types.WrapError(err, errorcode.Internal, store.Component, "could not increment link degree")
	}

	return nil
}

// createLink adds the given link to the DB.
func (s *scopedStore) createLink(
	ctx context.Context,
	linkHash chainscript.LinkHash,
	data []byte,
	link *chainscript.Link,
) error {
	tx, err := s.txFactory.NewTx()
	if err != nil {
		return err
	}

	err = s.createLinkInTx(ctx, tx, linkHash, data, link)
	if err != nil {
		s.txFactory.RollbackTx(tx, err)
		return err
	}

	return s.txFactory.CommitTx(tx)
}

// createLink inserts the given link in a transaction context.
// If the link already exists it will return an error.
func (s *scopedStore) createLinkInTx(
	ctx context.Context,
	tx *sql.Tx,
	linkHash chainscript.LinkHash,
	data []byte,
	link *chainscript.Link,
) error {
	prevLinkHash := link.PrevLinkHash()
	if prevLinkHash == nil {
		prevLinkHash = []byte{}
	}

	// Create the link.
	createLink, err := tx.Prepare(SQLCreateLink)
	if err != nil {
		return types.WrapError(err, errorcode.InvalidArgument, store.Component, "could not create link")
	}

	res, err := createLink.ExecContext(
		ctx,
		linkHash,
		link.Meta.Priority,
		link.Meta.MapId,
		prevLinkHash,
		pq.Array(link.Meta.Tags),
		string(data),
		link.Meta.Process.Name,
		link.Meta.Step,
	)
	if err != nil {
		return types.WrapError(err, errorcode.Internal, store.Component, "could not create link")
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return types.WrapError(err, errorcode.Internal, store.Component, "could not check created link")
	}

	if rowsAffected == 0 {
		return types.WrapError(store.ErrLinkAlreadyExists, errorcode.AlreadyExists, store.Component, "could not create link")
	}

	// Update refs table.
	addRef, err := tx.Prepare(SQLAddReference)
	if err != nil {
		return types.WrapError(err, errorcode.InvalidArgument, store.Component, "could not add reference")
	}

	for _, ref := range link.Meta.Refs {
		_, err = addRef.ExecContext(ctx, ref.LinkHash, linkHash)
		if err != nil {
			return types.WrapError(err, errorcode.Internal, store.Component, "could not add reference")
		}
	}

	// Update maps table.
	if s.enforceUniqueMapEntry && len(prevLinkHash) == 0 {
		initMap, err := tx.Prepare(SQLInitMap)
		if err != nil {
			return types.WrapError(store.ErrUniqueMapEntry, errorcode.FailedPrecondition, store.Component, "could not initialize map")
		}

		_, err = initMap.ExecContext(ctx, link.Meta.Process.Name, link.Meta.MapId)
		if err != nil {
			return types.WrapError(store.ErrUniqueMapEntry, errorcode.FailedPrecondition, store.Component, "could not initialize map")
		}
	}

	// Update degree table.
	initDegree, err := tx.Prepare(SQLCreateLinkDegree)
	if err != nil {
		return types.WrapError(err, errorcode.Internal, store.Component, "could not update link degree")
	}

	_, err = initDegree.ExecContext(ctx, linkHash)
	if err != nil {
		return types.WrapError(err, errorcode.Internal, store.Component, "could not update link degree")
	}

	return nil
}

// GetSegment implements github.com/stratumn/go-core/store.SegmentReader.GetSegment.
func (s *scopedStore) GetSegment(ctx context.Context, linkHash chainscript.LinkHash) (*chainscript.Segment, error) {
	var segments = make(types.SegmentSlice, 0, 1)

	rows, err := s.stmts.GetSegment.QueryContext(ctx, linkHash)
	if err != nil {
		return nil, types.WrapError(err, errorcode.Unavailable, store.Component, "could not get segment")
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
	rows, err := s.stmts.FindSegmentsWithFilters(ctx, filter)
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
			linkData     string
			link         *chainscript.Link
			evidenceData sql.NullString
			evidence     *chainscript.Evidence
			err          error
		)

		if totalCount == nil {
			if err := rows.Scan(&linkHash, &linkData, &evidenceData); err != nil {
				return types.WrapError(err, errorcode.Internal, store.Component, "could not scan rows")
			}
		} else {
			if err := rows.Scan(&linkHash, &linkData, &evidenceData, totalCount); err != nil {
				return types.WrapError(err, errorcode.Internal, store.Component, "could not scan rows")
			}
		}

		if !bytes.Equal(currentHash, linkHash) {
			err = json.Unmarshal([]byte(linkData), &link)
			if err != nil {
				return types.WrapError(err, errorcode.InvalidArgument, store.Component, "could not unmarshal link")
			}

			hash, err := link.Hash()
			if err != nil {
				return types.WrapError(err, errorcode.InvalidArgument, store.Component, "could not hash link")
			}
			currentHash = hash

			currentSegment, err = link.Segmentify()
			if err != nil {
				return types.WrapError(err, errorcode.InvalidArgument, store.Component, "could not segmentify")
			}

			*segments = append(*segments, currentSegment)
		}

		if evidenceData.Valid && len(evidenceData.String) > 0 {
			err = json.Unmarshal([]byte(evidenceData.String), &evidence)
			if err != nil {
				return types.WrapError(err, errorcode.InvalidArgument, store.Component, "could not unmarshal evidence")
			}

			if err := currentSegment.AddEvidence(evidence); err != nil {
				return types.WrapError(err, errorcode.InvalidArgument, store.Component, "could not add evidence")
			}
		}
	}

	err := rows.Err()
	if err != nil {
		return types.WrapError(err, errorcode.Internal, store.Component, "could not scan rows")
	}

	return nil
}

// GetMapIDs implements github.com/stratumn/go-core/store.SegmentReader.GetMapIDs.
func (s *scopedStore) GetMapIDs(ctx context.Context, filter *store.MapFilter) ([]string, error) {
	rows, err := s.stmts.GetMapIDsWithFilters(ctx, filter)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	mapIDs := make([]string, 0, filter.Pagination.Limit)

	for rows.Next() {
		var mapID string
		if err = rows.Scan(&mapID); err != nil {
			return nil, types.WrapError(err, errorcode.Internal, store.Component, "could not get map ids")
		}

		mapIDs = append(mapIDs, mapID)
	}

	return mapIDs, nil
}
