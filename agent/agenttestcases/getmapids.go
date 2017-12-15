package agenttestcases

import (
	"testing"
)

// TestGetMapIdsOK tests the client's ability to handle a GetMapIds request
func (f Factory) TestGetMapIdsOK(t *testing.T) {
}

// TestGetMapIdsLimit tests the client's ability to handle a GetMapIds request
// when a limit is set in the filter
func (f Factory) TestGetMapIdsLimit(t *testing.T) {

}

// TestGetMapIdsNoLimit tests the client's ability to handle a GetMapIds request
// when the limit is set to -1 to retrieve all map IDs
func (f Factory) TestGetMapIdsNoLimit(t *testing.T) {

}

// TestGetMapIdsNoMatch tests the client's ability to handle a GetMapIds request
// when no mapID is found
func (f Factory) TestGetMapIdsNoMatch(t *testing.T) {

}
