package client_test

import (
	"flag"
	"testing"

	"github.com/stratumn/sdk/agent/agenttestcases"
	"github.com/stratumn/sdk/agent/client"
	"github.com/stretchr/testify/assert"
)

var (
	agentURL    = "http://localhost:3000"
	integration = flag.Bool("integration", false, "Run integration tests")
)

func TestMain(m *testing.M) {
	flag.Parse()
	m.Run()
}

// TestNewAgentClient checks the initialisation of a new client
func TestNewAgentClient(t *testing.T) {
	if *integration == false {
		srv := mockAgentServer(t, agentURL)
		defer srv.Shutdown(nil)
	}
	client, err := client.NewAgentClient(agentURL)
	assert.NoError(t, err)
	assert.Equal(t, agentURL, client.URL())
}

// TestNewAgentClient_ExtraSlash tests if the client handles correctly
// the connection when the url ends with and extra '/'
func TestNewAgentClient_ExtraSlash(t *testing.T) {
	if *integration == false {
		agentURL := "http://localhost:3000/"
		srv := mockAgentServer(t, agentURL)
		defer srv.Shutdown(nil)
	}
	client, err := client.NewAgentClient(agentURL)
	assert.NoError(t, err)
	assert.Equal(t, agentURL, client.URL())
}

// TestNewAgentClient_WrongURL tests the error handling when the
// provided url is ill formatted
func TestNewAgentClient_WrongURL(t *testing.T) {
	agentURL := "//http:\\"
	_, err := client.NewAgentClient(agentURL)
	assert.EqualError(t, err, "parse //http:\\: invalid character \"\\\\\" in host name")
}

// TestAgentClient runs all the tests for the client
func TestAgentClient(t *testing.T) {
	mockServer := mockAgentServer
	if *integration == true {
		mockServer = nil
	}
	agenttestcases.Factory{
		NewClient: func(agentURL string) (client.AgentClient, error) {
			return client.NewAgentClient(agentURL)
		},
		NewMock: mockServer,
	}.RunAgentClientTests(t)
}
