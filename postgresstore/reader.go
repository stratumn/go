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
	"encoding/json"
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
		data     string
		segments cs.SegmentSlice
	)

	if err := a.stmts.FindSegments.QueryRow(filter.Offset, filter.Limit).Scan(&data); err != nil {
		if err.Error() == notFoundError {
			return nil, nil
		}
		return nil, err
	}

	if err := json.Unmarshal([]byte(data), &segments); err != nil {
		return nil, err
	}

	return segments, nil
}

// GetMapIDs implements github.com/stratumn/sdk/store.Adapter.GetMapIDs.
func (a *reader) GetMapIDs(pagination *store.Pagination) ([]string, error) {
	var (
		data string
		ids  []string
	)

	if err := a.stmts.GetMapIDs.QueryRow(pagination.Offset, pagination.Limit).Scan(&data); err != nil {
		if err.Error() == notFoundError {
			return nil, nil
		}
		return nil, err
	}

	if err := json.Unmarshal([]byte(data), &ids); err != nil {
		return nil, err
	}

	return ids, nil
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
