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
	WillCreate(*cs.Link)
	DidCreateLink(*cs.Link)
	DidCreateSegment(*cs.Segment)
	FilterSegment(*cs.Segment)
}

// Plugins is a list of Plugin
type Plugins []Plugin
