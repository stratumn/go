package client_test

import (
	"testing"

	"github.com/stratumn/sdk/agent/client"
)

var agentUrl = "http://localhost:3000"

func TestNewAgentClient(t *testing.T) {
	client, err := client.NewAgentClient(agentUrl)
	if err != nil {
		t.Error(err)
	}
	if got, want := client.URL(), agentUrl; got != want {
		t.Errorf("TestNewAgentClient : wrong default url, got %s, want %s\n", got, want)
	}
}

func TestNewAgentClient_DefaultURL(t *testing.T) {
	client, err := client.NewAgentClient("")
	if err != nil {
		t.Error(err)
	}
	want := "http://agent:3000"
	if got := client.URL(); got != want {
		t.Errorf("TestNewAgentClient : wrong default url, got %s, want %s\n", got, want)
	}
}

func TestNewAgentClient_WrongURL(t *testing.T) {
	agentUrl := "//http:\\"
	_, err := client.NewAgentClient(agentUrl)
	if err == nil {
		t.Errorf("TestNewAgentClient_WrongURL should have failed, got err = %s, want %s", err, "parse //http:\\: invalid character \"\\\" in host name")
	}
}
