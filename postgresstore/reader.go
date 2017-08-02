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
	"encoding/json"

	"github.com/lib/pq"
	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/store"
	"github.com/stratumn/sdk/types"
)

type reader struct {
	stmts readStmts
}

// GetSegment implements github.com/stratumn/sdk/store.Adapter.GetSegment.
func (a *reader) GetSegment(linkHash *types.Bytes32) (*cs.Segment, error) {
	var (
		data    string
		segment cs.Segment
	)

	if err := a.stmts.GetSegment.QueryRow(linkHash[:]).Scan(&data); err != nil {
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

// FindSegments implements github.com/stratumn/sdk/store.Adapter.FindSegments.
func (a *reader) FindSegments(filter *store.Filter) (cs.SegmentSlice, error) {
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

// GetMapIDs implements github.com/stratumn/sdk/store.Adapter.GetMapIDs.
func (a *reader) GetMapIDs(pagination *store.Pagination) ([]string, error) {
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

// GetValue implements github.com/stratumn/sdk/store.Adapter.GetValue.
func (a *reader) GetValue(key []byte) ([]byte, error) {
	var data []byte

	if err := a.stmts.GetValue.QueryRow(key).Scan(&data); err != nil {
		if err.Error() == notFoundError {
			return nil, nil
		}
		return nil, err
	}

	return data, nil
}
