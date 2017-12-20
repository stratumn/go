package agenttestcases

import (
	"testing"

	"github.com/stratumn/sdk/store"
	"github.com/stretchr/testify/assert"
)

// TestGetMapIdsOK tests the client's ability to handle a GetMapIds request.
func (f Factory) TestGetMapIdsOK(t *testing.T) {
	process := "test"
	expected := 20
	for i := 0; i != expected; i++ {
		f.Client.CreateMap(process, nil, "test")
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
		f.Client.CreateMap(process, nil, "test")
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
