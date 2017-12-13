package agenttestcases

import (
	"net/http"
	"testing"

	log "github.com/sirupsen/logrus"

	"github.com/stratumn/sdk/agent/client"
)

var agentURL = "http://localhost:3000"

// Factory wraps functions to mock an HTTP server for an agent and run
// the tests for the agent and its client
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

	t.Run("TestCreateMap", f.TestCreateMap)
	t.Run("TestCreateMapWithRefs", f.TestCreateMapWithRefs)
	t.Run("TestCreateMapWithBadRefs", f.TestCreateMapWithBadRefs)
	t.Run("TestCreateMapHandlesWrongInitArgs", f.TestCreateMapHandlesWrongInitArgs)

	t.Run("TestCreateLink", f.TestCreateLink)
	t.Run("TestCreateLinkWithRefs", f.TestCreateLinkWithRefs)
	t.Run("TestCreateLinkWithBadRefs", f.TestCreateLinkWithBadRefs)
	t.Run("TestCreateLinkHandlesWrongAction", f.TestCreateLinkHandlesWrongAction)
	t.Run("TestCreateLinkHandlesWrongActionArgs", f.TestCreateLinkHandlesWrongActionArgs)
	t.Run("TestCreateLinkHandlesWrongLinkHash", f.TestCreateLinkHandlesWrongLinkHash)
	t.Run("TestCreateLinkHandlesWrongProcess", f.TestCreateLinkHandlesWrongProcess)

	t.Run("TestFindSegments", f.TestFindSegments)
	t.Run("TestFindSegmentsLimit", f.TestFindSegmentsLimit)
	t.Run("TestFindSegmentsLinkHashes", f.TestFindSegmentsLinkHashes)
	t.Run("TestFindSegmentsMapIDs", f.TestFindSegmentsMapIDs)
	t.Run("TestFindSegmentsTags", f.TestFindSegmentsTags)
	t.Run("TestFindSegmentsNoMatch", f.TestFindSegmentsNoMatch)

	t.Run("TestGetInfo", f.TestGetInfo)

	t.Run("TestGetMapIds", f.TestGetMapIds)
	t.Run("TestGetMapIdsLimit", f.TestGetMapIdsLimit)
	t.Run("TestGetMapIdsNoLimit", f.TestGetMapIdsNoLimit)
	t.Run("TestGetMapIdsNoMatch", f.TestGetMapIdsNoMatch)

	t.Run("TestGetProcess", f.TestGetProcess)
	t.Run("TestGetProcessNotFound", f.TestGetProcessNotFound)

	t.Run("TestGetProcesses", f.TestGetProcesses)

	t.Run("TestGetSegment", f.TestGetSegment)
	t.Run("TestGetSegmentNotFound", f.TestGetSegmentNotFound)
}
