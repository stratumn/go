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

// SetValue implements github.com/stratumn/sdk/store.KeyValueStore.SetValue.
func (a *writer) SetValue(key []byte, value []byte) error {
	_, err := a.stmts.SaveValue.Exec(key, value)
	if err != nil {
		return err
	}

	return nil
}

// DeleteValue implements github.com/stratumn/sdk/store.KeyValueStore.DeleteValue.
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

// CreateLink implements github.com/stratumn/sdk/store.Adapter.CreateLink.
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
