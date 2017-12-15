package agenttestcases

import (
	"testing"
)

// TestCreateLinkOK tests the client's ability to handle a CreateLink request
func (f Factory) TestCreateLinkOK(t *testing.T) {
}

// TestCreateLinkWithRefs tests the client's ability to handle a CreateLink request
// when a reference is passed
func (f Factory) TestCreateLinkWithRefs(t *testing.T) {
}

// TestCreateLinkWithBadRefs tests the client's ability to handle a CreateLink request
// when a reference is passed
func (f Factory) TestCreateLinkWithBadRefs(t *testing.T) {
}

// TestCreateLinkHandlesWrongProcess tests the client's ability to handle a CreateLink request
// when the provided process does not exist
func (f Factory) TestCreateLinkHandlesWrongProcess(t *testing.T) {
}

// TestCreateLinkHandlesWrongLinkHash tests the client's ability to handle a CreateLink request
// when the provided parent's linkHash does not exist
func (f Factory) TestCreateLinkHandlesWrongLinkHash(t *testing.T) {
}

// TestCreateLinkHandlesWrongAction tests the client's ability to handle a CreateLink request
// when the provided action does not exist
func (f Factory) TestCreateLinkHandlesWrongAction(t *testing.T) {
}

// TestCreateLinkHandlesWrongActionArgs tests the client's ability to handle a CreateLink request
// when the provided action's arguments do not match the actual ones
func (f Factory) TestCreateLinkHandlesWrongActionArgs(t *testing.T) {
}
