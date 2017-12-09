package client_test

import (
	"testing"

	"github.com/stratumn/sdk/agent/client"
	"github.com/stretchr/testify/assert"
)

var agentURL = "http://localhost:3000"

func TestNewAgentClient(t *testing.T) {
	client, err := client.NewAgentClient(agentURL)
	assert.NoError(t, err)
	assert.Equal(t, agentURL, client.URL())
}

func TestNewAgentClient_DefaultURL(t *testing.T) {
	client, err := client.NewAgentClient("")
	assert.NoError(t, err)
	assert.Equal(t, "http://agent:3000", client.URL())
}

func TestNewAgentClient_WrongURL(t *testing.T) {
	agentURL := "//http:\\"
	_, err := client.NewAgentClient(agentURL)
	assert.EqualError(t, err, "parse //http:\\: invalid character \"\\\\\" in host name")
}
