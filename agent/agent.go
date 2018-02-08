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

import (
	"net/http"

	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/types"
)

// Info is the data structure returned by Agent.GetInfo()
type Info struct {
	Processes   ProcessesMap     `json:"processes"`
	Stores      []StoreInfo      `json:"stores"`
	Fossilizers []FossilizerInfo `json:"fossilizers"`
	Plugins     Plugins          `json:"plugins"`
}

// StoreInfo is the generic data structured returned by Store.GetInfo()
type StoreInfo map[string]interface{}

// FossilizerInfo is the generic data structure returned by Fossilizer.GetInfo()
type FossilizerInfo map[string]interface{}

// Actions is a map indexing an action function by its name
type Actions map[string]func(...interface{}) interface{}

// ProcessOptions can be used to configure a process when creating a new one
type ProcessOptions struct {
	ReconnectTimeout int     `json:"reconnectTimeout"`
	Plugins          Plugins `json:"plugins"`
}

// Agent is the interface of an agent
type Agent interface {
	AddProcess(process string, actions Actions, storeClient interface{}, fossilizerClients []interface{}, opts *ProcessOptions) (Process, error)
	UploadProcess(processName string, actionsPath string, storeURL string, fossilizerURLs []string, pluginIDs []string) (*Process, error)
	FindSegments(filter store.SegmentFilter) (cs.SegmentSlice, error)
	GetInfo() (*Info, error)
	GetMapIds(filter store.MapFilter) ([]string, error)
	GetProcesses() (Processes, error)
	GetProcess(process string) (*Process, error)
	GetSegment(process string, linkHash types.Bytes32)
	HttpServer() *http.Server
	RemoveProcess(process string) (Processes, error)
	Url() string
}
