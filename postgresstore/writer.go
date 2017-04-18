// Copyright 2017 Stratumn SAS. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package postgresstore

import (
	"encoding/json"

	"github.com/lib/pq"
	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/types"
)

type writer struct {
	stmts writeStmts
}

// SaveSegment implements github.com/stratumn/sdk/store.Adapter.SaveSegment.
func (a *writer) SaveSegment(segment *cs.Segment) error {
	var (
		err          error
		linkHash     = segment.GetLinkHash()
		priority     = segment.Link.GetPriority()
		mapID        = segment.Link.GetMapID()
		prevLinkHash = segment.Link.GetPrevLinkHash()
		tags         = segment.Link.GetTags()
	)

	data, err := json.Marshal(segment)
	if err != nil {
		return err
	}

	if prevLinkHash == nil {
		_, err = a.stmts.SaveSegment.Exec(linkHash[:], priority, mapID, nil, pq.Array(tags), string(data))
	} else {
		_, err = a.stmts.SaveSegment.Exec(linkHash[:], priority, mapID, prevLinkHash[:], pq.Array(tags), string(data))
	}

	if err != nil {
		return err
	}

	return nil
}

// DeleteSegment implements github.com/stratumn/sdk/store.Adapter.DeleteSegment.
func (a *writer) DeleteSegment(linkHash *types.Bytes32) (*cs.Segment, error) {
	var (
		data    string
		segment cs.Segment
	)

	if err := a.stmts.DeleteSegment.QueryRow(linkHash[:]).Scan(&data); err != nil {
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

// SaveValue implements github.com/stratumn/sdk/store.Adapter.SaveValue.
func (a *writer) SaveValue(key []byte, value []byte) error {
	_, err := a.stmts.SaveValue.Exec(key, value)
	if err != nil {
		return err
	}

	return nil
}

// DeleteValue implements github.com/stratumn/sdk/store.Adapter.DeleteValue.
func (a *writer) DeleteValue(key []byte) ([]byte, error) {
	var data []byte

	if err := a.stmts.DeleteValue.QueryRow(key).Scan(&data); err != nil {
		if err.Error() == notFoundError {
			return nil, nil
		}
		return nil, err
	}

	return data, nil
}
