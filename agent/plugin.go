// Copyright 2017 Stratumn SAS. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package agent

import "github.com/stratumn/sdk/cs"

// PluginInfo is the data structure used by the agent when returning
// informations about a process' plugins
type PluginInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ID          string `json:"id"`
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
type Plugins []PluginInfo
