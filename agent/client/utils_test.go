package client_test

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stratumn/sdk/agent"
	"github.com/stratumn/sdk/agent/client"
	"github.com/stratumn/sdk/cs"
)

type mockHTTPServer struct{}

func (a *mockHTTPServer) decodePostParams(r *http.Request) ([]interface{}, error) {
	decoder := json.NewDecoder(r.Body)
	params := []interface{}{}
	if err := decoder.Decode(&params); err != nil {
		return nil, err
	}
	return params, nil
}

func (m *mockHTTPServer) sendError(w http.ResponseWriter, status int, message string) {
	errorData := client.ErrorData{
		Status:  status,
		Message: message,
	}
	bytes, err := json.Marshal(errorData)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}

func (m *mockHTTPServer) sendResponse(w http.ResponseWriter, status int, data interface{}) {
	bytes, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}

func (m *mockHTTPServer) mockCreateLink(w http.ResponseWriter, r *http.Request) {
	params, err := m.decodePostParams(r)
	if err != nil {
		m.sendError(w, http.StatusOK, err.Error())
		return
	}
	if len(params) < 2 {
		m.sendError(w, http.StatusOK, "a title is required")
		return
	}
	refs := params[0]
	arg := params[1].(string)

	vars := mux.Vars(r)
	if vars["linkHash"] == "0000000000000000000000000000000000000000000000000000000000000000" {
		m.sendError(w, http.StatusOK, "Not Found")
		return
	} else if vars["action"] == "wrong" {
		m.sendError(w, http.StatusOK, "not found")
		return
	} else if vars["process"] == "wrong" {
		m.sendError(w, http.StatusOK, "process 'wrong' does not exist")
		return
	} else if vars["process"] == "wrongref" {
		m.sendError(w, http.StatusOK, "missing segment or (process and linkHash)")
		return
	}

	s := cs.Segment{
		Link: cs.Link{
			State: map[string]interface{}{
				"title": arg,
			},
			Meta: map[string]interface{}{
				"mapId": "mapId",
				"refs":  refs,
			},
		},
	}
	m.sendResponse(w, 200, &s)
}

func (m *mockHTTPServer) mockCreateMap(w http.ResponseWriter, r *http.Request) {
	params, err := m.decodePostParams(r)
	if err != nil {
		m.sendError(w, http.StatusOK, err.Error())
		return
	}
	if len(params) < 2 {
		m.sendError(w, http.StatusOK, "a title is required")
		return
	}
	refs := params[0]
	arg := params[1].(string)

	vars := mux.Vars(r)
	if vars["process"] == "wrongref" {
		m.sendError(w, http.StatusOK, "missing segment or (process and linkHash)")
		return
	}
	s := cs.Segment{
		Link: cs.Link{
			State: map[string]interface{}{
				"title": arg,
			},
			Meta: map[string]interface{}{
				"mapId": "mapId",
				"refs":  refs,
			},
		},
	}
	m.sendResponse(w, 200, s)
}

func (m *mockHTTPServer) mockFindSegments(w http.ResponseWriter, r *http.Request) {
	var (
		q             = r.URL.Query()
		limit, _      = strconv.Atoi(q.Get("limit"))
		linkHashesStr = append(q["linkHashes[]"], q["linkHashes%5B%5D"]...)
		mapIDs        = append(q["mapIds[]"], q["mapIds%5B%5D"]...)
	)
	s := cs.SegmentSlice{}
	vars := mux.Vars(r)
	if vars["process"] == "wrong" {
		m.sendError(w, http.StatusOK, "process 'wrong' does not exist")
		return
	}
	if len(linkHashesStr) > 0 || len(mapIDs) > 0 {
		s = append(s, &cs.Segment{})
	} else {
		for i := 0; i < limit; i++ {
			s = append(s, &cs.Segment{})
		}
	}
	m.sendResponse(w, 200, s)
}

func (m *mockHTTPServer) mockGetInfo(w http.ResponseWriter, r *http.Request) {
	info := agent.Info{
		Processes: agent.ProcessesMap{
			"test": &agent.Process{},
		},
		Stores: []agent.StoreInfo{
			{
				"url": "http://localhost:5000",
			},
		},
		Fossilizers: []agent.FossilizerInfo{},
		Plugins: []agent.PluginInfo{
			{
				Name:        "Agent URL",
				Description: "Saves in segment meta the URL that can be used to retrieve a segment.",
				ID:          "1",
			},
			{
				Name:        "Action arguments",
				Description: "Saves the action and its arguments in link meta information.",
				ID:          "2",
			},
			{
				Name:        "State Hash",
				Description: "Computes and adds the hash of the state in meta.",
				ID:          "3",
			},
		},
	}
	m.sendResponse(w, 200, info)
}

func (m *mockHTTPServer) mockGetMapIds(w http.ResponseWriter, r *http.Request) {
	var (
		q         = r.URL.Query()
		limit, _  = strconv.Atoi(q.Get("limit"))
		offset, _ = strconv.Atoi(q.Get("offset"))
	)
	s := make([]string, 0)
	vars := mux.Vars(r)
	if vars["process"] == "wrong" {
		m.sendError(w, http.StatusOK, "process 'wrong' does not exist")
		return
	}
	if offset > limit {
		m.sendResponse(w, 200, s)
		return
	}
	for i := 0; i < limit; i++ {
		s = append(s, "mapID")
	}
	m.sendResponse(w, 200, s)
}

func (m *mockHTTPServer) mockGetProcesses(w http.ResponseWriter, r *http.Request) {
	p := agent.Processes{}
	p = append(p, &agent.Process{
		Name: "test",
		ProcessInfo: agent.ProcessInfo{
			Actions: map[string]agent.ActionInfo{
				"one": {},
				"two": {},
			},
		}})
	m.sendResponse(w, 200, p)
}

func (m *mockHTTPServer) mockGetSegment(w http.ResponseWriter, r *http.Request) {
	s := cs.Segment{}
	vars := mux.Vars(r)
	if vars["linkHash"] == "0000000000000000000000000000000000000000000000000000000000000000" {
		m.sendError(w, http.StatusNotFound, "Not Found")
		return
	}
	m.sendResponse(w, 200, s)
}

func mockAgent(t *testing.T, address string) *agent.MockAgent {
	url := cleanURL(address)
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockObj := agent.NewMockAgent(mockCtrl)
	mockObj.EXPECT().HttpServer().DoAndReturn(func() *http.Server {
		mux := mux.NewRouter()
		handler := mockHTTPServer{}
		mux.HandleFunc("/", handler.mockGetInfo).Methods("GET")
		mux.HandleFunc("/processes", handler.mockGetProcesses).Methods("GET")
		mux.HandleFunc("/{process}/segments", handler.mockCreateMap).Methods("POST")
		mux.HandleFunc("/{process}/segments/{linkHash}/{action}", handler.mockCreateLink).Methods("POST")
		mux.HandleFunc("/{process}/segments/{linkHash}", handler.mockGetSegment).Methods("GET")
		mux.HandleFunc("/{process}/maps", handler.mockGetMapIds).Methods("GET")
		mux.HandleFunc("/{process}/segments", handler.mockFindSegments).Methods("GET")
		return &http.Server{
			Addr:    url,
			Handler: mux,
		}
	}).AnyTimes()
	return mockObj
}

func cleanURL(address string) string {
	if strings.Contains(address, "http://") {
		address = address[7:]
	}
	return filepath.Clean(address)

}

func mockAgentServer(t *testing.T, agentURL string) *http.Server {
	server := mockAgent(t, agentURL).HttpServer()

	go func() {
		log.Info("Listening on ", agentURL, "...")
		if err := server.ListenAndServe(); err != nil {
			log.WithField("info", err).Info("Server stopped")
			return
		}
	}()
	return server
}
