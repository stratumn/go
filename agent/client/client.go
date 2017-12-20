package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/google/go-querystring/query"
	"github.com/stratumn/sdk/agent"
	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/jsonhttp"
	"github.com/stratumn/sdk/store"
	"github.com/stratumn/sdk/types"
)

// ErrorData is the format used by an agent to format errors.
type ErrorData struct {
	Status  int    `json:"status"`
	Message string `json:"error"`
}

//SegmentRef defines a format for a valid reference.
type SegmentRef struct {
	LinkHash *types.Bytes32 `json:"linkHash"`
	Process  string         `json:"process"`
	Segment  *cs.Segment    `json:"segment"`
	Meta     interface{}    `json:"meta"`
}

// AgentClient is the interface for an agent client
// It can be used to access an agent's http endpoints.
type AgentClient interface {
	CreateMap(process string, refs []SegmentRef, args ...string) (*cs.Segment, error)
	CreateSegment(process string, linkHash *types.Bytes32, action string, refs []SegmentRef, args ...string) (*cs.Segment, error)
	FindSegments(filter *store.SegmentFilter) (cs.SegmentSlice, error)
	GetInfo() (*agent.Info, error)
	GetMapIds(filter *store.MapFilter) ([]string, error)
	GetProcess(name string) (*agent.Process, error)
	GetProcesses() (agent.Processes, error)
	GetSegment(process string, linkHash *types.Bytes32) (*cs.Segment, error)
	URL() string
}

// agentClient wraps an http.Client used to send request to the agent's server.
type agentClient struct {
	c         *http.Client
	agentURL  *url.URL
	agentInfo agent.Info
}

// NewAgentClient returns an initialized AgentClient
// If the provided url is empty, it will use a default one.
func NewAgentClient(agentURL string) (AgentClient, error) {
	if len(agentURL) == 0 {
		return nil, errors.New(URLRequiredError)
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

// CreateSegment sends a CreateSegment request to the agent and returns
// the newly created segment.
func (a *agentClient) CreateSegment(process string, linkHash *types.Bytes32, action string, refs []SegmentRef, args ...string) (*cs.Segment, error) {
	queryURL := fmt.Sprintf("/%s/segments/%s/%s", process, linkHash, action)
	postParams, err := a.makeActionPostParams(refs, args...)
	if err != nil {
		return nil, err
	}
	resp, err := a.post(queryURL, postParams)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(resp.Body)
	seg := cs.Segment{}
	if err := decoder.Decode(&seg); err != nil {
		return nil, jsonhttp.NewErrBadRequest(err.Error())
	}
	return &seg, nil
}

// CreateMap sends a CreateMap request to the agent and returns
// the first segment of the newly created map.
func (a *agentClient) CreateMap(process string, refs []SegmentRef, args ...string) (*cs.Segment, error) {
	queryURL := fmt.Sprintf("/%s/segments", process)
	postParams, err := a.makeActionPostParams(refs, args...)
	resp, err := a.post(queryURL, postParams)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(resp.Body)
	seg := cs.Segment{}
	if err := decoder.Decode(&seg); err != nil {
		return nil, jsonhttp.NewErrBadRequest(err.Error())
	}
	return &seg, nil
}

// FindSegments sends a FindSegments request to the agent and returns
// the list of found segments.
func (a *agentClient) FindSegments(filter *store.SegmentFilter) (sgmts cs.SegmentSlice, err error) {
	if filter.Limit == -1 {
		filter.Limit = store.DefaultLimit
		batch, err := a.findSegments(filter)
		for ; len(batch) == filter.Limit && err == nil; batch, err = a.findSegments(filter) {
			sgmts = append(sgmts, batch...)
			filter.Offset += filter.Limit
		}
		if err != nil {
			return nil, err
		}
		sgmts = append(sgmts, batch...)
		return sgmts, err
	}
	return a.findSegments(filter)
}

func (a *agentClient) findSegments(filter *store.SegmentFilter) (cs.SegmentSlice, error) {
	queryURL := fmt.Sprintf("/%s/segments", filter.Process)
	resp, err := a.get(queryURL, filter)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(resp.Body)
	sgmts := cs.SegmentSlice{}
	if err := decoder.Decode(&sgmts); err != nil {
		return nil, jsonhttp.NewErrBadRequest(err.Error())
	}
	return sgmts, nil

}

// GetInfo sends a GetInfo request to the agent and returns the result.
func (a *agentClient) GetInfo() (*agent.Info, error) {
	resp, err := a.get("/", nil)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(resp.Body)
	agentInfo := agent.Info{}
	if err := decoder.Decode(&agentInfo); err != nil {
		return nil, jsonhttp.NewErrBadRequest(err.Error())
	}
	a.agentInfo = agentInfo
	return &agentInfo, nil
}

// GetMapIds sends a GetMapIds request to the agent and returns
// a list of found map IDs for a process.
func (a *agentClient) GetMapIds(filter *store.MapFilter) (IDs []string, err error) {
	if filter.Limit == -1 {
		filter.Limit = store.DefaultLimit
		batch, err := a.getMapIds(filter)
		for ; len(batch) == filter.Limit && err == nil; batch, err = a.getMapIds(filter) {
			IDs = append(IDs, batch...)
			filter.Offset += filter.Limit
		}
		if err != nil {
			return nil, err
		}
		IDs = append(IDs, batch...)
		return IDs, err
	}
	return a.getMapIds(filter)
}

func (a *agentClient) getMapIds(filter *store.MapFilter) ([]string, error) {
	queryURL := fmt.Sprintf("/%s/maps", filter.Process)
	resp, err := a.get(queryURL, filter)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(resp.Body)
	mapIDs := make([]string, 0)
	if err := decoder.Decode(&mapIDs); err != nil {
		return nil, jsonhttp.NewErrBadRequest(err.Error())
	}
	return mapIDs, nil
}

// GetProcess returns a process given its name.
func (a *agentClient) GetProcess(name string) (*agent.Process, error) {
	processes, err := a.GetProcesses()
	if err != nil {
		return nil, err
	}
	process := processes.FindProcess(name)
	if process == nil {
		return nil, errors.New(ProcessNotFoundError)
	}
	return process, nil
}

// GetProcesses sends a GetProcesses request to the agent and returns
// a list of all the processes.
func (a *agentClient) GetProcesses() (agent.Processes, error) {
	resp, err := a.get("/processes", nil)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(resp.Body)
	processes := agent.Processes{}
	if err := decoder.Decode(&processes); err != nil {
		return nil, jsonhttp.NewErrBadRequest(err.Error())
	}

	return processes, nil
}

// GetSegment sends a GetSegment request to the agent and returns a segment
// given its link hash.
func (a *agentClient) GetSegment(process string, linkHash *types.Bytes32) (*cs.Segment, error) {
	queryURL := fmt.Sprintf("/%s/segments/%s", process, linkHash)
	resp, err := a.get(queryURL, nil)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(resp.Body)
	seg := cs.Segment{}
	if err := decoder.Decode(&seg); err != nil {
		return nil, jsonhttp.NewErrBadRequest(err.Error())
	}
	return &seg, nil
}

// URL returns the url of the agent.
func (a *agentClient) URL() string {
	return a.agentURL.String()
}

func (a *agentClient) decodeError(resp *http.Response) error {
	decoder := json.NewDecoder(resp.Body)
	errorData := ErrorData{}
	if err := decoder.Decode(&errorData); err != nil {
		return jsonhttp.NewErrBadRequest(err.Error())
	}

	return errors.New(errorData.Message)
}

func (a *agentClient) makeActionPostParams(refs []SegmentRef, args ...string) ([]byte, error) {
	var rawParams []interface{}
	rawParams = append(rawParams, refs)
	for _, a := range args {
		rawParams = append(rawParams, a)
	}
	return json.Marshal(rawParams)
}

// get sends an HTTP GET request and checks the status of the response.
func (a *agentClient) get(endpoint string, params interface{}) (*http.Response, error) {
	path, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	if params != nil {
		queryParams, err := query.Values(params)
		if err != nil {
			return nil, err
		}
		path.RawQuery = queryParams.Encode()
	}

	url := a.agentURL.ResolveReference(path)
	resp, err := a.c.Get(url.String())
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return resp, a.decodeError(resp)
	}
	return resp, nil
}

// post sends an HTTP POST request and checks the status of the response.
func (a *agentClient) post(endpoint string, data []byte) (*http.Response, error) {
	path, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	url := a.agentURL.ResolveReference(path)
	resp, err := a.c.Post(url.String(), "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return resp, a.decodeError(resp)
	}
	return resp, nil

}
