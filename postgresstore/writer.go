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
	"context"
	"encoding/json"

	"github.com/lib/pq"
	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/monitoring"
	"github.com/stratumn/go-indigocore/types"

	"go.opencensus.io/trace"
)

type writer struct {
	stmts writeStmts
}

// SetValue implements github.com/stratumn/go-indigocore/store.KeyValueStore.SetValue.
func (a *writer) SetValue(ctx context.Context, key []byte, value []byte) (err error) {
	ctx, span := trace.StartSpan(ctx, "postgresstore/SetValue")
	defer monitoring.SetSpanStatusAndEnd(span, err)

	_, err = a.stmts.SaveValue.Exec(key, value)
	return
}

// DeleteValue implements github.com/stratumn/go-indigocore/store.KeyValueStore.DeleteValue.
func (a *writer) DeleteValue(ctx context.Context, key []byte) (_ []byte, err error) {
	ctx, span := trace.StartSpan(ctx, "postgresstore/DeleteValue")
	defer monitoring.SetSpanStatusAndEnd(span, err)

	var data []byte

	if err := a.stmts.DeleteValue.QueryRow(key).Scan(&data); err != nil {
		if err.Error() == notFoundError {
			return nil, nil
		}
		return nil, err
	}

	return data, nil
}

// CreateLink implements github.com/stratumn/go-indigocore/store.Adapter.CreateLink.
func (a *writer) CreateLink(ctx context.Context, link *cs.Link) (_ *types.Bytes32, err error) {
	ctx, span := trace.StartSpan(ctx, "postgresstore/CreateLink")
	defer monitoring.SetSpanStatusAndEnd(span, err)

	var (
		priority     = link.Meta.Priority
		mapID        = link.Meta.MapID
		prevLinkHash = link.Meta.GetPrevLinkHash()
		tags         = link.Meta.Tags
		process      = link.Meta.Process
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
