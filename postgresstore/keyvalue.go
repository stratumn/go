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
	"context"
	"database/sql"

	"github.com/stratumn/go-core/monitoring/errorcode"
	"github.com/stratumn/go-core/store"
	"github.com/stratumn/go-core/types"
)

// GetValue implements github.com/stratumn/go-core/store.KeyValueStore.GetValue.
func (s *scopedStore) GetValue(ctx context.Context, key []byte) ([]byte, error) {
	var data []byte

	if err := s.stmts.GetValue.QueryRowContext(ctx, key).Scan(&data); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, types.WrapError(err, errorcode.Internal, store.Component, "could not get value")
	}

	return data, nil
}

// SetValue implements github.com/stratumn/go-core/store.KeyValueStore.SetValue.
func (s *scopedStore) SetValue(ctx context.Context, key []byte, value []byte) error {
	_, err := s.stmts.SaveValue.ExecContext(ctx, key, value)
	if err != nil {
		return types.WrapError(err, errorcode.Unavailable, store.Component, "could not set value")
	}

	return nil
}

// DeleteValue implements github.com/stratumn/go-core/store.KeyValueStore.DeleteValue.
func (s *scopedStore) DeleteValue(ctx context.Context, key []byte) ([]byte, error) {
	var data []byte

	if err := s.stmts.DeleteValue.QueryRowContext(ctx, key).Scan(&data); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, types.WrapError(err, errorcode.Internal, store.Component, "could not delete value")
	}

	return data, nil
}
