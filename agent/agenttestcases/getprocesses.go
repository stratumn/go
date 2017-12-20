package agenttestcases

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGetProcessesOK tests the client's ability to handle a GetProcesses request.
func (f Factory) TestGetProcessesOK(t *testing.T) {
	processes, err := f.Client.GetProcesses()
	assert.NoError(t, err)
	assert.Equal(t, len(processes), 1)
	assert.Equal(t, processes[0].Name, "test")
}
