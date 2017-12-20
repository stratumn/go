package agenttestcases

import (
	"testing"

	"github.com/stratumn/sdk/agent"
	"github.com/stretchr/testify/assert"
)

// TestGetInfoOK tests the client's ability to handle a GetInfo request.
func (f Factory) TestGetInfoOK(t *testing.T) {
	actual, err := f.Client.GetInfo()

	expected := agent.Info{
		Processes: agent.ProcessesMap{
			"test": &agent.Process{},
		},
		Stores: []agent.StoreInfo{
			agent.StoreInfo{
				"url": "http://localhost:5000",
			},
		},
		Fossilizers: []agent.FossilizerInfo{},
		Plugins: []agent.PluginInfo{
			agent.PluginInfo{
				Name:        "Agent URL",
				Description: "Saves in segment meta the URL that can be used to retrieve a segment.",
				ID:          "1",
			},
			agent.PluginInfo{
				Name:        "Action arguments",
				Description: "Saves the action and its arguments in link meta information.",
				ID:          "2",
			},
			agent.PluginInfo{
				Name:        "State Hash",
				Description: "Computes and adds the hash of the state in meta.",
				ID:          "3",
			},
		},
	}
	assert.NoError(t, err)
	assert.NotNil(t, actual)
	assert.Equal(t, expected.Stores, actual.Stores)
	assert.Equal(t, expected.Plugins, actual.Plugins)
	assert.Equal(t, expected.Fossilizers, actual.Fossilizers)
	assert.NotNil(t, expected.Processes["test"])
	assert.Equal(t, len(expected.Processes), len(actual.Processes))
}
