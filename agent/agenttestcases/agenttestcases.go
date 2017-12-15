package agenttestcases

import (
	"net/http"
	"testing"

	log "github.com/sirupsen/logrus"

	"github.com/stratumn/sdk/agent/client"
)

var agentURL = "http://localhost:3000"

// Factory wraps functions to create a client and a mock agent.
// After its creation, the client is stored in the factory to avoid
// re-creating it in every test.
type Factory struct {
	NewMock   func(t *testing.T, agentURL string) *http.Server
	NewClient func(agentURL string) (client.AgentClient, error)

	Client client.AgentClient
}

// RunAgentClientTests runs the test suite for an agent client
func (f Factory) RunAgentClientTests(t *testing.T) {
	if f.NewMock != nil {
		srv := f.NewMock(t, agentURL)
		defer func() {
			if err := srv.Shutdown(nil); err != nil {
				log.WithField("error", err).Fatal("Failed to shutdown HTTP server")
			}
		}()

	}
	f.Client, _ = f.NewClient(agentURL)

	t.Run("Test creating map", f.TestCreateMap)
	t.Run("Test creating link", f.TestCreateLink)
	t.Run("Test find segments", f.TestFindSegments)
	t.Run("Test getting infos", f.TestGetInfo)
	t.Run("Test getting a process", f.TestGetProcess)
	t.Run("Test getting all the processes", f.TestGetProcesses)
	t.Run("Test getting a segment", f.TestGetSegment)
}

// TestCreateMap tests what happens when creating a map with various inputs
func (f Factory) TestCreateMap(t *testing.T) {
	t.Run("TestCreateMap", f.TestCreateMapOK)
	t.Run("TestCreateMapWithRefs", f.TestCreateMapWithRefs)
	t.Run("TestCreateMapWithBadRefs", f.TestCreateMapWithBadRefs)
	t.Run("TestCreateMapHandlesWrongInitArgs", f.TestCreateMapHandlesWrongInitArgs)
}

// TestCreateLink tests what happens when creating a link with various inputs
func (f Factory) TestCreateLink(t *testing.T) {
	t.Run("TestCreateLink", f.TestCreateLinkOK)
	t.Run("TestCreateLinkWithRefs", f.TestCreateLinkWithRefs)
	t.Run("TestCreateLinkWithBadRefs", f.TestCreateLinkWithBadRefs)
	t.Run("TestCreateLinkHandlesWrongAction", f.TestCreateLinkHandlesWrongAction)
	t.Run("TestCreateLinkHandlesWrongActionArgs", f.TestCreateLinkHandlesWrongActionArgs)
	t.Run("TestCreateLinkHandlesWrongLinkHash", f.TestCreateLinkHandlesWrongLinkHash)
	t.Run("TestCreateLinkHandlesWrongProcess", f.TestCreateLinkHandlesWrongProcess)
}

// TestFindSegments tests what happens when finding segments with various filters
func (f Factory) TestFindSegments(t *testing.T) {
	t.Run("TestFindSegments", f.TestFindSegmentsOK)
	t.Run("TestFindSegmentsLimit", f.TestFindSegmentsLimit)
	t.Run("TestFindSegmentsLinkHashes", f.TestFindSegmentsLinkHashes)
	t.Run("TestFindSegmentsMapIDs", f.TestFindSegmentsMapIDs)
	t.Run("TestFindSegmentsTags", f.TestFindSegmentsTags)
	t.Run("TestFindSegmentsNoMatch", f.TestFindSegmentsNoMatch)
}

// TestGetInfo tests what happens when getting an agent's infos
func (f Factory) TestGetInfo(t *testing.T) {
	t.Run("TestGetInfo", f.TestGetInfoOK)
}

// TestGetMapIds tests what happens when finding map ids with varisou filters
func (f Factory) TestGetMapIds(t *testing.T) {
	t.Run("TestGetMapIds", f.TestGetMapIdsOK)
	t.Run("TestGetMapIdsLimit", f.TestGetMapIdsLimit)
	t.Run("TestGetMapIdsNoLimit", f.TestGetMapIdsNoLimit)
	t.Run("TestGetMapIdsNoMatch", f.TestGetMapIdsNoMatch)
}

// TestGetProcess tests what happens when getting informations about a process
func (f Factory) TestGetProcess(t *testing.T) {
	t.Run("TestGetProcess", f.TestGetProcessOK)
	t.Run("TestGetProcessNotFound", f.TestGetProcessNotFound)
}

// TestGetProcesses tests what happens whengetting all the processes from an agent
func (f Factory) TestGetProcesses(t *testing.T) {
	t.Run("TestGetProcesses", f.TestGetProcessesOK)
}

// TestGetSegment tests what happens when getting a segment
func (f Factory) TestGetSegment(t *testing.T) {
	t.Run("TestGetSegment", f.TestGetSegmentOK)
	t.Run("TestGetSegmentNotFound", f.TestGetSegmentNotFound)
}
