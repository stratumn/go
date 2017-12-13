package client_test

import (
	"flag"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stratumn/sdk/agent/agenttestcases"
	"github.com/stratumn/sdk/agent/client"
	"github.com/stretchr/testify/assert"
)

var (
	agentURL    = "http://localhost:3000"
	integration = flag.Bool("integration", false, "Run integration tests")
)

func TestMain(m *testing.M) {
	seed := int64(time.Now().Nanosecond())
	fmt.Printf("using seed %d\n", seed)
	rand.Seed(seed)
	flag.Parse()
	m.Run()
}
func TestNewAgentClient(t *testing.T) {
	if *integration == false {
		srv := mockAgentServer(t, agentURL)
		defer srv.Shutdown(nil)
	}
	client, err := client.NewAgentClient(agentURL)
	assert.NoError(t, err)
	assert.Equal(t, agentURL, client.URL())
}

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

func TestNewAgentClient_WrongURL(t *testing.T) {
	agentURL := "//http:\\"
	_, err := client.NewAgentClient(agentURL)
	assert.EqualError(t, err, "parse //http:\\: invalid character \"\\\\\" in host name")
}

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
