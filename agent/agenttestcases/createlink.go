package agenttestcases

import (
	"testing"

	cj "github.com/gibson042/canonicaljson-go"

	"github.com/stratumn/sdk/agent/client"
	"github.com/stratumn/sdk/testutil"
	"github.com/stratumn/sdk/types"
	"github.com/stretchr/testify/assert"
)

// TestCreateLinkOK tests the client's ability to handle a CreateLink request
func (f Factory) TestCreateLinkOK(t *testing.T) {
	process, action := "test", "test"
	parent, err := f.Client.CreateMap(process, nil, "test")

	segment, err := f.Client.CreateLink(process, parent.GetLinkHash(), action, nil, "test")
	assert.NoError(t, err)
	assert.NotNil(t, segment)
	assert.Equal(t, "test", segment.Link.State["title"])
}

// TestCreateLinkWithRefs tests the client's ability to handle a CreateLink request
// when a reference is passed
func (f Factory) TestCreateLinkWithRefs(t *testing.T) {
	process, action := "test", "test"
	parent, err := f.Client.CreateMap(process, nil, "test")
	ref := client.SegmentRef{Process: "other", LinkHash: testutil.RandomHash()}
	refs := make([]client.SegmentRef, 0)
	refs = append(refs, ref)

	segment, err := f.Client.CreateLink(process, parent.GetLinkHash(), action, refs, "one")
	assert.NoError(t, err)
	assert.NotNil(t, segment)
	assert.NotNil(t, segment.Link.Meta["refs"])
	want, _ := cj.Marshal(refs)
	got, _ := cj.Marshal(segment.Link.Meta["refs"])
	assert.Equal(t, want, got)
}

// TestCreateLinkWithBadRefs tests the client's ability to handle a CreateLink request
// when a reference is passed
func (f Factory) TestCreateLinkWithBadRefs(t *testing.T) {
	process, action, arg := "test", "test", "wrongref"
	parent, err := f.Client.CreateMap(process, nil, "test")
	ref := client.SegmentRef{Process: "wrong"}
	refs := make([]client.SegmentRef, 0)
	refs = append(refs, ref)

	segment, err := f.Client.CreateLink(process, parent.GetLinkHash(), action, refs, arg)
	assert.EqualError(t, err, "missing segment or (process and linkHash)")
	assert.Nil(t, segment)
}

// TestCreateLinkHandlesWrongProcess tests the client's ability to handle a CreateLink request
// when the provided process does not exist
func (f Factory) TestCreateLinkHandlesWrongProcess(t *testing.T) {
	process, linkHash, action := "wrong", testutil.RandomHash(), "test"
	segment, err := f.Client.CreateLink(process, linkHash, action, nil, "test")
	assert.EqualError(t, err, "process 'wrong' does not exist")
	assert.Nil(t, segment)
}

// TestCreateLinkHandlesWrongLinkHash tests the client's ability to handle a CreateLink request
// when the provided parent's linkHash does not exist
func (f Factory) TestCreateLinkHandlesWrongLinkHash(t *testing.T) {
	linkHash, _ := types.NewBytes32FromString("0000000000000000000000000000000000000000000000000000000000000000")
	process, action := "test", "test"
	segment, err := f.Client.CreateLink(process, linkHash, action, nil, "test")
	assert.EqualError(t, err, "Not Found")
	assert.Nil(t, segment)
}

// TestCreateLinkHandlesWrongAction tests the client's ability to handle a CreateLink request
// when the provided action does not exist
func (f Factory) TestCreateLinkHandlesWrongAction(t *testing.T) {
	process, action := "test", "wrong"
	parent, err := f.Client.CreateMap(process, nil, "test")

	segment, err := f.Client.CreateLink(process, parent.GetLinkHash(), action, nil, "test")
	assert.EqualError(t, err, "not found")
	assert.Nil(t, segment)
}

// TestCreateLinkHandlesWrongActionArgs tests the client's ability to handle a CreateLink request
// when the provided action's arguments do not match the actual ones
func (f Factory) TestCreateLinkHandlesWrongActionArgs(t *testing.T) {
	process, action := "test", "test"
	parent, err := f.Client.CreateMap(process, nil, "test")

	segment, err := f.Client.CreateLink(process, parent.GetLinkHash(), action, nil)
	assert.EqualError(t, err, "a title is required")
	assert.Nil(t, segment)
}
