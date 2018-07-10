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
	"bytes"
	"context"
	"database/sql"
	"encoding/json"

	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/types"
)

type reader struct {
	stmts readStmts
}

// GetSegment implements github.com/stratumn/go-indigocore/store.SegmentReader.GetSegment.
func (a *reader) GetSegment(ctx context.Context, linkHash *types.Bytes32) (*cs.Segment, error) {
	var segments = make(cs.SegmentSlice, 0, 1)

	rows, err := a.stmts.GetSegment.Query(linkHash[:])
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	if err = scanLinkAndEvidences(rows, &segments); err != nil {
		return nil, err
	}

	if len(segments) == 0 {
		return nil, nil
	}
	return segments[0], nil
}

// FindSegments implements github.com/stratumn/go-indigocore/store.SegmentReader.FindSegments.
func (a *reader) FindSegments(ctx context.Context, filter *store.SegmentFilter) (cs.SegmentSlice, error) {

	rows, err := a.stmts.FindSegmentsWithFilters(filter)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	segments := make(cs.SegmentSlice, 0, filter.Limit)
	err = scanLinkAndEvidences(rows, &segments)

	return segments, err
}

func scanLinkAndEvidences(rows *sql.Rows, segments *cs.SegmentSlice) error {
	var currentSegment *cs.Segment
	var currentHash []byte

	for rows.Next() {
		var (
			linkHash     []byte
			linkData     string
			link         cs.Link
			evidenceData sql.NullString
			evidence     cs.Evidence
		)

		if err := rows.Scan(&linkHash, &linkData, &evidenceData); err != nil {
			return err
		}

		if !bytes.Equal(currentHash, linkHash) {
			if err := json.Unmarshal([]byte(linkData), &link); err != nil {
				return err
			}

			hash, err := link.Hash()
			if err != nil {
				return err
			}
			currentHash = hash[:]

			currentSegment = link.Segmentify()

			*segments = append(*segments, currentSegment)
		}

		if evidenceData.Valid {
			if err := json.Unmarshal([]byte(evidenceData.String), &evidence); err != nil {
				return err
			}

			if err := currentSegment.Meta.AddEvidence(evidence); err != nil {
				return err
			}
		}
	}
	return nil
}

// GetMapIDs implements github.com/stratumn/go-indigocore/store.SegmentReader.GetMapIDs.
func (a *reader) GetMapIDs(ctx context.Context, filter *store.MapFilter) ([]string, error) {
	rows, err := a.stmts.GetMapIDs.Query(filter.Pagination.Offset, filter.Pagination.Limit, filter.Process)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	mapIDs := make([]string, 0, filter.Pagination.Limit)

	for rows.Next() {
		var mapID string
		if err = rows.Scan(&mapID); err != nil {
			return nil, err
		}

		mapIDs = append(mapIDs, mapID)
	}

	return mapIDs, nil
}

// GetValue implements github.com/stratumn/go-indigocore/store.KeyValueStore.GetValue.
func (a *reader) GetValue(ctx context.Context, key []byte) ([]byte, error) {
	var data []byte

	if err := a.stmts.GetValue.QueryRow(key).Scan(&data); err != nil {
		if err.Error() == notFoundError {
			return nil, nil
		}
		return nil, err
	}

	return data, nil
}

// GetEvidences implements github.com/stratumn/go-indigocore/store.EvidenceReader.GetEvidences.
func (a *reader) GetEvidences(ctx context.Context, linkHash *types.Bytes32) (*cs.Evidences, error) {
	var evidences cs.Evidences

	rows, err := a.stmts.GetEvidences.Query(linkHash[:])
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var (
			data     string
			evidence cs.Evidence
		)

		if err := rows.Scan(&data); err != nil {
			return nil, err
		}

		if err := json.Unmarshal([]byte(data), &evidence); err != nil {
			return nil, err
		}
		evidences = append(evidences, &evidence)

	}
	return &evidences, nil
}
