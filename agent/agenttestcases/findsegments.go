package agenttestcases

import (
	"testing"
)

// TestFindSegmentsOK tests the client's ability to handle a FindSegment request
func (f Factory) TestFindSegmentsOK(t *testing.T) {
}

// TestFindSegmentsTags tests the client's ability to handle a FindSegment request
// when tags are set in the filter
func (f Factory) TestFindSegmentsTags(t *testing.T) {
}

// TestFindSegmentsLinkHashes tests the client's ability to handle a FindSegment request
// when LinkHashes are set in the filter
func (f Factory) TestFindSegmentsLinkHashes(t *testing.T) {

}

// TestFindSegmentsMapIDs tests the client's ability to handle a FindSegment request
// when a map ID is set in the filter
func (f Factory) TestFindSegmentsMapIDs(t *testing.T) {

}

// TestFindSegmentsLimit tests the client's ability to handle a FindSegment request
// when a limit is set in the filter
func (f Factory) TestFindSegmentsLimit(t *testing.T) {

}

// TestFindSegmentsNoLimit tests the client's ability to handle a FindSegment request
// when the limit is set to -1 in the filter to retrieve all segments
func (f Factory) TestFindSegmentsNoLimit(t *testing.T) {

}

// TestFindSegmentsNoMatch tests the client's ability to handle a FindSegment request
// when no segment is found
func (f Factory) TestFindSegmentsNoMatch(t *testing.T) {

}
