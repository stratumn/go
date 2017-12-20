package agenttestcases

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGetProcessOK tests the client's ability to handle a GetProcess request.
func (f Factory) TestGetProcessOK(t *testing.T) {
	process, err := f.Client.GetProcess("test")
	assert.NoError(t, err)
	assert.NotNil(t, process)
	assert.Equal(t, "test", process.Name)
	assert.Equal(t, 2, len(process.ProcessInfo.Actions))
}

// TestGetProcessNotFound tests the client's ability to handle a FindSegment request
// when no process is found.
func (f Factory) TestGetProcessNotFound(t *testing.T) {
	process, err := f.Client.GetProcess("wrong")
	assert.EqualError(t, err, "Not Found")
	assert.Nil(t, process)
}
