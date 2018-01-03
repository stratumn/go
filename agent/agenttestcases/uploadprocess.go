package agenttestcases

import (
	"fmt"
	"go/build"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Don't forget to set enableProcessUpload to true on the agent to test
// this feature.

// TestUploadProcessOK tests the client's ability to handle a CreateMap request.
func (f Factory) TestUploadProcessOK(t *testing.T) {
	process, err := f.Client.UploadProcess(
		"test",
		fmt.Sprintf("%v/src/github.com/stratumn/sdk/agent/agenttestcases/actions.js", build.Default.GOPATH),
		StoreURL,
		[]string{},
		[]string{},
	)

	assert.NoError(t, err)
	assert.NotNil(t, process)
}
