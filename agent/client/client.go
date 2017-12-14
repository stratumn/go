package client

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/stratumn/sdk/agent"
	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/store"
	"github.com/stratumn/sdk/types"
)

// ErrorData is the format used by an agent to format errors
type ErrorData struct {
	Status  int    `json:"status"`
	Message string `json:"error"`
}

//SegmentRef defines a format for a valid reference
type SegmentRef struct {
	LinkHash *types.Bytes32 `json:"linkHash"`
	Process  string         `json:"process"`
	Segment  *cs.Segment    `json:"segment"`
	Meta     interface{}    `json:"meta"`
}

// AgentClient is the interface for an agent client
// It can be used to access an agent's http endpoints
type AgentClient interface {
	CreateMap(process string, refs []SegmentRef, args ...string) (*cs.Segment, error)
	CreateLink(process string, linkHash *types.Bytes32, action string, refs []SegmentRef, args ...string) (*cs.Segment, error)
	FindSegments(filter *store.SegmentFilter) (cs.SegmentSlice, error)
	GetInfo() (*agent.Info, error)
	GetMapIds(filter *store.MapFilter) ([]string, error)
	GetProcess(name string) (*agent.Process, error)
	GetProcesses() (agent.Processes, error)
	GetSegment(process string, linkHash *types.Bytes32) (*cs.Segment, error)
	URL() string
}

// agentClient wraps an http.Client used to send request to the agent's server
type agentClient struct {
	c         *http.Client
	agentURL  *url.URL
	agentInfo agent.Info
}

// NewAgentClient returns an initialized AgentClient
// If the provided url is empty, it will use a default one
func NewAgentClient(agentURL string) (AgentClient, error) {
	if len(agentURL) == 0 {
		return nil, errors.New("An URL must be provided to initialize a client")
	}
	url, err := url.Parse(agentURL)
	if err != nil {
		return nil, err
	}
	client := &agentClient{
		c:        &http.Client{},
		agentURL: url,
	}
	if _, err := client.GetInfo(); err != nil {
		return client, err
	}

	return client, nil
}

func (a *agentClient) CreateLink(process string, linkHash *types.Bytes32, action string, refs []SegmentRef, args ...string) (*cs.Segment, error) {
	seg := cs.Segment{}
	return &seg, nil
}

func (a *agentClient) CreateMap(process string, refs []SegmentRef, args ...string) (*cs.Segment, error) {
	seg := cs.Segment{}
	return &seg, nil
}

func (a *agentClient) FindSegments(filter *store.SegmentFilter) (sgmts cs.SegmentSlice, err error) {
	return a.findSegments(filter)
}

func (a *agentClient) findSegments(filter *store.SegmentFilter) (cs.SegmentSlice, error) {
	sgmts := cs.SegmentSlice{}
	return sgmts, nil

}

func (a *agentClient) GetInfo() (*agent.Info, error) {
	agentInfo := agent.Info{}
	return &agentInfo, nil
}

func (a *agentClient) GetMapIds(filter *store.MapFilter) (IDs []string, err error) {
	return a.getMapIds(filter)
}

func (a *agentClient) getMapIds(filter *store.MapFilter) ([]string, error) {
	return nil, nil
}

func (a *agentClient) GetProcess(name string) (*agent.Process, error) {
	return nil, nil
}

func (a *agentClient) GetProcesses() (agent.Processes, error) {
	processes := agent.Processes{}
	return processes, nil
}
func (a *agentClient) GetSegment(process string, linkHash *types.Bytes32) (*cs.Segment, error) {
	seg := cs.Segment{}
	return &seg, nil
}

// URL returns the url of the AgentClient
func (a *agentClient) URL() string {
	return a.agentURL.String()
}

func (a *agentClient) get(endpoint string, params interface{}) (*http.Response, error) {
	return nil, nil
}

func (a *agentClient) post(endpoint string, data []byte) (*http.Response, error) {
	return nil, nil
}
