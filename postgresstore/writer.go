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

	"github.com/lib/pq"
	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/types"
)

type writer struct {
	stmts writeStmts
}

// SaveSegment implements github.com/stratumn/sdk/store.Adapter.SaveSegment.
func (a *writer) SaveSegment(segment *cs.Segment) error {
	linkHash, err := a.CreateLink(&segment.Link)
	if err != nil {
		return err
	}

	for _, e := range segment.Meta.Evidences {
		evidenceData, err := json.Marshal(e)
		if err != nil {
			return err
		}
		_, err = a.stmts.AddEvidence.Exec(linkHash[:], e.Provider, string(evidenceData))

	}

	return nil
}

// DeleteSegment implements github.com/stratumn/sdk/store.Adapter.DeleteSegment.
func (a *writer) DeleteSegment(linkHash *types.Bytes32) (*cs.Segment, error) {
	var (
		data string
		link cs.Link
	)

	if err := a.stmts.DeleteLink.QueryRow(linkHash[:]).Scan(&data); err != nil {
		if err.Error() == notFoundError {
			return nil, nil
		}
		return nil, err
	}

	if err := json.Unmarshal([]byte(data), &link); err != nil {
		return nil, err
	}

	return link.Segmentify(), nil
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

// CreateLink implements github.com/stratumn/sdk/store.AdapterV2.CreateLink.
func (a *writer) CreateLink(link *cs.Link) (*types.Bytes32, error) {
	var (
		priority     = link.GetPriority()
		mapID        = link.GetMapID()
		prevLinkHash = link.GetPrevLinkHash()
		tags         = link.GetTags()
		process      = link.GetProcess()
	)

	linkHash, err := link.Hash()
	if err != nil {
		return linkHash, err
	}

	data, err := json.Marshal(link)
	if err != nil {
		return linkHash, err
	}

	if prevLinkHash == nil {
		_, err = a.stmts.CreateLink.Exec(linkHash[:], priority, mapID, []byte{}, pq.Array(tags), string(data), process)
	} else {
		_, err = a.stmts.CreateLink.Exec(linkHash[:], priority, mapID, prevLinkHash[:], pq.Array(tags), string(data), process)
	}

	return linkHash, err
}
