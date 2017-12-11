package client

import (
	"bytes"
	"net/http"
	"net/url"

	"github.com/stratumn/sdk/agent"
	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/store"
	"github.com/stratumn/sdk/types"
)

// DefaultURL is the default url for an agent
const DefaultURL = "http://agent"

// DefaultPort is the default port for an agent
const DefaultPort = "3000"

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
	CreateLink(process string, linkHash types.Bytes32, action string, refs []SegmentRef, args ...string) (*cs.Segment, error)
	FindSegments(filter *store.SegmentFilter) (cs.SegmentSlice, error)
	GetInfo() (*agent.Info, error)
	GetMapIds(filter *store.MapFilter) (cs.SegmentSlice, error)
	GetProcesses() (agent.Processes, error)
	GetSegment(process string, linkHash types.Bytes32) (*cs.Segment, error)
	URL() string
}

// agentClient wraps an http.Client used to send request to the agent's server
type agentClient struct {
	c        *http.Client
	agentURL *url.URL
}

// NewAgentClient returns an initialized AgentClient
// If the provided url is empty, it will use a default one
func NewAgentClient(agentURL string) (AgentClient, error) {
	if len(agentURL) == 0 {
		agentURL = DefaultURL + ":" + DefaultPort
	}
	url, err := url.Parse(agentURL)
	if err != nil {
		return nil, err
	}
	return &agentClient{
		c:        &http.Client{},
		agentURL: url,
	}, nil
}

func (a *agentClient) GetProcesses() (agent.Processes, error) {
	processes := agent.Processes{}
	return processes, nil
}

func (a *agentClient) GetInfo() (*agent.Info, error) {
	agentInfo := agent.Info{}
	return &agentInfo, nil
}

func (a *agentClient) CreateMap(process string, refs []SegmentRef, args ...string) (*cs.Segment, error) {
	seg := cs.Segment{}
	return &seg, nil
}

func (a *agentClient) GetSegment(process string, linkHash types.Bytes32) (*cs.Segment, error) {
	seg := cs.Segment{}
	return &seg, nil
}

func (a *agentClient) FindSegments(filter *store.SegmentFilter) (cs.SegmentSlice, error) {
	sgmts := cs.SegmentSlice{}
	return sgmts, nil
}

func (a *agentClient) GetMapIds(filter *store.MapFilter) (cs.SegmentSlice, error) {
	sgmts := cs.SegmentSlice{}
	return sgmts, nil
}

func (a *agentClient) CreateLink(process string, linkHash types.Bytes32, action string, refs []SegmentRef, args ...string) (*cs.Segment, error) {
	seg := cs.Segment{}
	return &seg, nil
}

// URL returns the url of the AgentClient
func (a *agentClient) URL() string {
	return a.agentURL.String()
}

func (a *agentClient) get(endpoint string) (*http.Response, error) {
	path, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	url := a.agentURL.ResolveReference(path)
	return a.c.Get(url.String())

}

func (a *agentClient) post(endpoint string, data []byte) (*http.Response, error) {
	path, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	url := a.agentURL.ResolveReference(path)
	return a.c.Post(url.String(), "application/json", bytes.NewBuffer(data))

}
