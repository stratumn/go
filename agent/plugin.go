package agent

import "github.com/stratumn/sdk/cs"

// PluginInfo is the data structure used by the agent when returning
// informations about a process' plugins
type PluginInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Plugin is the interface describing the handlers that a plugin can implement
type Plugin interface {
	// WillCreate is called right before a transition function from the agent's actions.
	// It takes the existing link as an argument. It should be updated in-place.
	WillCreate(*cs.Link)

	//is called whenever a link has been created by a transition function.
	// It takes the new link as an argument. It should be updated in-place.
	DidCreateLink(*cs.Link)

	// is called when segments are retrieved by the agent from the underlying storage.
	// It should return true if the plugins accepts the segment, false otherwise.
	FilterSegment(*cs.Segment) bool
}

// Plugins is a list of Plugin
type Plugins []Plugin
