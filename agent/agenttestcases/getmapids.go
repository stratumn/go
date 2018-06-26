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

package agenttestcases

import (
	"testing"

	"github.com/stratumn/go-indigocore/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetMapIdsOK tests the client's ability to handle a GetMapIds request.
func (f Factory) TestGetMapIdsOK(t *testing.T) {
	process := "test"
	expected := 20
	for i := 0; i != expected; i++ {
		_, err := f.Client.CreateMap(process, nil, "test")
		require.NoError(t, err)
	}

	filter := store.MapFilter{
		Process: process,
		Pagination: store.Pagination{
			Limit: expected,
		},
	}
	ids, err := f.Client.GetMapIds(&filter)
	assert.NoError(t, err)
	assert.NotNil(t, ids)
	assert.Equal(t, expected, len(ids))
}

// TestGetMapIdsLimit tests the client's ability to handle a GetMapIds request
// when a limit is set in the filter and when the limit is set to -1
func (f Factory) TestGetMapIdsLimit(t *testing.T) {
	process := "test"
	created := 30
	for i := 0; i != created; i++ {
		_, err := f.Client.CreateMap(process, nil, "test")
		require.NoError(t, err)
	}

	t.Run("With a limit", func(t *testing.T) {
		limit := 5
		filter := store.MapFilter{
			Process: process,
			Pagination: store.Pagination{
				Limit: limit,
			},
		}
		ids, err := f.Client.GetMapIds(&filter)
		assert.NoError(t, err)
		assert.NotNil(t, ids)
		assert.Equal(t, limit, len(ids))
	})

	t.Run("Without a limit", func(t *testing.T) {
		limit := -1
		filter := store.MapFilter{
			Process: process,
			Pagination: store.Pagination{
				Limit: limit,
			},
		}
		ids, err := f.Client.GetMapIds(&filter)
		assert.NoError(t, err)
		assert.NotNil(t, ids)
		assert.True(t, len(ids) > created)
	})
}

// TestGetMapIdsNoMatch tests the client's ability to handle a GetMapIds request
// when no mapID is found.
func (f Factory) TestGetMapIdsNoMatch(t *testing.T) {
	process := "wrong"
	filter := store.MapFilter{
		Process: process,
	}
	ids, err := f.Client.GetMapIds(&filter)
	assert.EqualError(t, err, "process 'wrong' does not exist")
	assert.Nil(t, ids)
}
