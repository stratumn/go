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

package client_test

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/stratumn/go-indigocore/agent"
	"github.com/stratumn/go-indigocore/agent/agenttestcases"
	"github.com/stratumn/go-indigocore/agent/client"
	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/utils"
)

type mockHTTPServer struct{}

func (m *mockHTTPServer) decodePostParams(r *http.Request) ([]interface{}, error) {
	decoder := json.NewDecoder(r.Body)
	params := []interface{}{}
	if err := decoder.Decode(&params); err != nil {
		return nil, err
	}
	return params, nil
}

func (m *mockHTTPServer) decodeRefs(input interface{}) ([]cs.SegmentReference, error) {
	if input == nil {
		return []cs.SegmentReference{}, nil
	}
	inputSlice, ok := input.([]interface{})
	if !ok {
		return nil, errors.New("refs should be a slice")
	}
	var refs []cs.SegmentReference
	for refIdx, in := range inputSlice {
		ref, ok := in.(map[string]interface{})
		if !ok {
			return nil, errors.Errorf("refs[%d] should be a map", refIdx)
		}

		process, ok := ref["process"].(string)
		if !ok || process == "" {
			return nil, errors.Errorf("refs[%d].process should be a non empty string", refIdx)
		}
		linkHashStr, ok := ref["linkHash"].(string)
		if !ok || linkHashStr == "" {
			return nil, errors.Errorf("refs[%d].linkHash should be a non empty string", refIdx)
		}
		refs = append(refs, cs.SegmentReference{Process: process, LinkHash: linkHashStr})
	}
	return refs, nil
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

func (m *mockHTTPServer) mockCreateSegment(w http.ResponseWriter, r *http.Request) {
	params, err := m.decodePostParams(r)
	if err != nil {
		m.sendError(w, http.StatusBadRequest, err.Error())
		return
	}
	if len(params) < 2 {
		m.sendError(w, http.StatusBadRequest, "a title is required")
		return
	}
	refs, err := m.decodeRefs(params[0])
	if err != nil {
		m.sendError(w, http.StatusBadRequest, errors.Wrap(err, "cannot decode references").Error())
		return
	}
	arg := params[1].(string)

	vars := mux.Vars(r)
	if vars["linkHash"] == "0000000000000000000000000000000000000000000000000000000000000000" {
		m.sendError(w, http.StatusBadRequest, "Not Found")
		return
	}
	if vars["action"] == "wrong" {
		m.sendError(w, http.StatusBadRequest, "not found")
		return
	}
	if vars["process"] == "wrong" {
		m.sendError(w, http.StatusBadRequest, "process 'wrong' does not exist")
		return
	}
	if arg == "wrongref" {
		m.sendError(w, http.StatusBadRequest, "missing segment or (process and linkHash)")
		return
	}

	s := cs.Segment{
		Link: cs.Link{
			State: map[string]interface{}{
				"title": arg,
			},
			Meta: cs.LinkMeta{
				MapID: "mapId",
				Refs:  refs,
			},
		},
	}
	m.sendResponse(w, http.StatusOK, &s)
}

func (m *mockHTTPServer) mockUploadProcess(w http.ResponseWriter, r *http.Request) {
	uploadProcessBody := &client.UploadProcessBody{}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		m.sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := json.Unmarshal(body, uploadProcessBody); err != nil {
		m.sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	m.sendResponse(w, http.StatusOK, &agent.Processes{
		&agent.Process{
			Name: "test",
		},
	})
}

func (m *mockHTTPServer) mockCreateMap(w http.ResponseWriter, r *http.Request) {
	params, err := m.decodePostParams(r)
	if err != nil {
		m.sendError(w, http.StatusBadRequest, err.Error())
		return
	}
	if len(params) < 2 {
		m.sendError(w, http.StatusBadRequest, "a title is required")
		return
	}
	refs, err := m.decodeRefs(params[0])
	if err != nil {
		m.sendError(w, http.StatusBadRequest, errors.Wrap(err, "cannot decode references").Error())
		return
	}
	arg := params[1].(string)

	if arg == "wrongref" {
		m.sendError(w, http.StatusBadRequest, "missing segment or (process and linkHash)")
		return
	}
	s := cs.Segment{
		Link: cs.Link{
			State: map[string]interface{}{
				"title": arg,
			},
			Meta: cs.LinkMeta{
				MapID: "mapId",
				Refs:  refs,
			},
		},
	}
	m.sendResponse(w, http.StatusOK, s)
}

func (m *mockHTTPServer) mockFindSegments(w http.ResponseWriter, r *http.Request) {
	var (
		q             = r.URL.Query()
		limit, _      = strconv.Atoi(q.Get("limit"))
		offset, _     = strconv.Atoi(q.Get("offset"))
		linkHashesStr = append(q["linkHashes[]"], q["linkHashes%5B%5D"]...)
		mapIDs        = append(q["mapIds[]"], q["mapIds%5B%5D"]...)
		tags          = append(q["tags[]"], q["tags%5B%5D"]...)
	)
	s := cs.SegmentSlice{}
	vars := mux.Vars(r)
	if vars["process"] == "wrong" {
		m.sendError(w, http.StatusBadRequest, "process 'wrong' does not exist")
		return
	}
	if len(linkHashesStr) > 0 || len(mapIDs) > 0 {
		s = append(s, &cs.Segment{})
	} else if offset > limit {
		m.sendResponse(w, http.StatusOK, s)
		return
	} else if len(tags) > 0 {

		s = append(s, &cs.Segment{Link: cs.Link{
			Meta: cs.LinkMeta{
				Tags: tags,
			},
		}})
	} else {
		for i := 0; i < limit; i++ {
			s = append(s, &cs.Segment{})
		}
	}
	m.sendResponse(w, http.StatusOK, s)
}

func (m *mockHTTPServer) mockGetInfo(w http.ResponseWriter, r *http.Request) {
	info := agent.Info{
		Processes: agent.ProcessesMap{
			"test": &agent.Process{},
		},
		Stores: []agent.StoreInfo{
			{
				"url": agenttestcases.StoreURL,
			},
		},
		Fossilizers: []agent.FossilizerInfo{},
		Plugins:     []agent.PluginInfo{},
	}
	m.sendResponse(w, http.StatusOK, info)
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
		m.sendError(w, http.StatusBadRequest, "process 'wrong' does not exist")
		return
	}
	if offset > limit {
		m.sendResponse(w, http.StatusOK, s)
		return
	}
	for i := 0; i < limit; i++ {
		s = append(s, "mapID")
	}
	m.sendResponse(w, http.StatusOK, s)
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
	m.sendResponse(w, http.StatusOK, p)
}

func (m *mockHTTPServer) mockGetSegment(w http.ResponseWriter, r *http.Request) {
	s := cs.Segment{}
	vars := mux.Vars(r)
	if vars["linkHash"] == "0000000000000000000000000000000000000000000000000000000000000000" {
		m.sendError(w, http.StatusNotFound, "Not Found")
		return
	}
	m.sendResponse(w, http.StatusOK, s)
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
		mux.HandleFunc("/{process}/segments/{linkHash}/{action}", handler.mockCreateSegment).Methods("POST")
		mux.HandleFunc("/{process}/segments/{linkHash}", handler.mockGetSegment).Methods("GET")
		mux.HandleFunc("/{process}/maps", handler.mockGetMapIds).Methods("GET")
		mux.HandleFunc("/{process}/segments", handler.mockFindSegments).Methods("GET")
		mux.HandleFunc("/{process}/upload", handler.mockUploadProcess).Methods("POST")
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
	started := make(chan bool)
	go func() {
		if err := utils.Retry(func(attempt int) (bool, error) {
			listener, err := net.Listen("tcp", server.Addr)
			if err != nil {
				log.Error("Error starting server : ", err)
				time.Sleep(1 * time.Second)
				return true, err
			}
			log.Info("Listening on ", server.Addr, "...")
			started <- true
			if err := server.Serve(listener); err != nil {
				log.WithField("info", err).Info("Server stopped")
				return false, nil
			}
			return false, nil
		}, 10); err != nil {
			t.Error(err)
		}
	}()
	<-started
	return server
}
