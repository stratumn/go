package agent

import (
	"net/http"

	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/store"
	"github.com/stratumn/sdk/types"
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
