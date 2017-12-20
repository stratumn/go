package agenttestcases

import (
	"testing"

	"github.com/stratumn/sdk/types"
	"github.com/stretchr/testify/assert"
)

// TestGetSegmentOK tests the client's ability to handle a GetSegment request.
func (f Factory) TestGetSegmentOK(t *testing.T) {
	process := "test"
	parent, _ := f.Client.CreateMap(process, nil, "test")

	segment, err := f.Client.GetSegment(process, parent.GetLinkHash())
	assert.NoError(t, err)
	assert.NotNil(t, segment)
}

// TestGetSegmentNotFound tests the client's ability to handle a GetSegment request
// when the queried linkHash does not exist.
func (f Factory) TestGetSegmentNotFound(t *testing.T) {
	process := "test"
	fakeLinkHash, _ := types.NewBytes32FromString("0000000000000000000000000000000000000000000000000000000000000000")
	segment, err := f.Client.GetSegment(process, fakeLinkHash)
	assert.EqualError(t, err, "Not Found")
	assert.Nil(t, segment)
}
