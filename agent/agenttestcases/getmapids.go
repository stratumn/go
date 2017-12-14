package agenttestcases

import (
	"testing"

	"github.com/stratumn/sdk/store"
	"github.com/stretchr/testify/assert"
)

// TestGetMapIdsOK tests the client's ability to handle a GetMapIds request
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
// when a limit is set in the filter
func (f Factory) TestGetMapIdsLimit(t *testing.T) {
	process := "test"
	created := 10
	expected := 0
	for i := 0; i != created; i++ {
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

// TestGetMapIdsNoLimit tests the client's ability to handle a GetMapIds request
// when the limit is set to -1 to retrieve all map IDs
func (f Factory) TestGetMapIdsNoLimit(t *testing.T) {
	process := "test"
	created := 40
	limit := -1
	for i := 0; i != created; i++ {
		f.Client.CreateMap(process, nil, "test")
	}

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
}

// TestGetMapIdsNoMatch tests the client's ability to handle a GetMapIds request
// when no mapID is found
func (f Factory) TestGetMapIdsNoMatch(t *testing.T) {
	process := "wrong"
	filter := store.MapFilter{
		Process: process,
	}
	ids, err := f.Client.GetMapIds(&filter)
	assert.EqualError(t, err, "process 'wrong' does not exist")
	assert.Nil(t, ids)
}
