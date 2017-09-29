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

// Package storehttp is used to create an HTTP server from a store adapter.
//
// It serves the following routes:
//	GET /
//		Renders information about the store.
//
//	POST /segments
//		Saves then renders a segment.
//		Body should be a JSON encoded segment.
//
//	GET /segments/:linkHash
//		Renders a segment.
//
//	DELETE /segments/:linkHash
//		Deletes then renders a segment.
//
//	GET /segments?[offset=offset]&[limit=limit]&[mapIds[]=id1]&[mapIds[]=id2]&[prevLinkHash=prevLinkHash]&[tags[]=tag1]&[tags[]=tag2]
//		Finds and renders segments.
//
//	GET /maps?[offset=offset]&[limit=limit]
//		Finds and renders map IDs.
//
//	GET /websocket
//		A web socket that broadcasts messages when a segment is saved:
//			{ "type": "didSave", "data": [segment] }
package storehttp

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"

	"time"

	"github.com/julienschmidt/httprouter"

	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/jsonhttp"
	"github.com/stratumn/sdk/jsonws"
	"github.com/stratumn/sdk/store"
	"github.com/stratumn/sdk/types"
)

const (
	// DefaultDidSaveChanSize is the default size of the DidSave channel.
	DefaultDidSaveChanSize = 256

	// DefaultAddress is the default address of the server.
	DefaultAddress = ":5000"

	// DefaultWebSocketReadBufferSize is the default size of the web socket
	// read buffer in bytes.
	DefaultWebSocketReadBufferSize = 1024

	// DefaultWebSocketWriteBufferSize is the default size of the web socket
	// write buffer in bytes.
	DefaultWebSocketWriteBufferSize = 1024

	// DefaultWebSocketWriteChanSize is the default size of a web socket
	// buffered connection channel.
	DefaultWebSocketWriteChanSize = 256

	// DefaultWebSocketWriteTimeout is the default timeout of a web socket
	// write.
	DefaultWebSocketWriteTimeout = 10 * time.Second

	// DefaultWebSocketPongTimeout is the default timeout of a web socket
	// expected pong.
	DefaultWebSocketPongTimeout = time.Minute

	// DefaultWebSocketPingInterval is the default interval between web
	// socket pings.
	DefaultWebSocketPingInterval = (DefaultWebSocketPongTimeout * 9) / 10

	// DefaultWebSocketMaxMsgSize is the default maximum size of a web
	// socke received message in in bytes.
	DefaultWebSocketMaxMsgSize = 32 * 1024
)

// Web socket message types.
const (
	// DidSave means a segment was saved.
	DidSave = "didSave"
)

// Server is an HTTP server for stores.
type Server struct {
	*jsonhttp.Server
	adapter     store.Adapter
	ws          *jsonws.Basic
	didSaveChan chan *cs.Segment
}

// Config contains configuration options for the server.
type Config struct {
	// The size of the DidSave channel.
	DidSaveChanSize int
}

// Info is the info returned by the root route.
type Info struct {
	Adapter interface{} `json:"adapter"`
}

// msg is a web socket message.
type msg struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// New create an instance of a server.
func New(
	a store.Adapter,
	config *Config,
	httpConfig *jsonhttp.Config,
	basicConfig *jsonws.BasicConfig,
	bufConnConfig *jsonws.BufferedConnConfig,
) *Server {
	s := Server{
		Server:      jsonhttp.New(httpConfig),
		adapter:     a,
		ws:          jsonws.NewBasic(basicConfig, bufConnConfig),
		didSaveChan: make(chan *cs.Segment, config.DidSaveChanSize),
	}

	s.Get("/", s.root)
	s.Post("/segments", s.saveSegment)
	s.Get("/segments/:linkHash", s.getSegment)
	s.Delete("/segments/:linkHash", s.deleteSegment)
	s.Get("/segments", s.findSegments)
	s.Get("/maps", s.getMapIDs)
	s.GetRaw("/websocket", s.getWebSocket)

	return &s
}

// ListenAndServe starts the server.
func (s *Server) ListenAndServe() (err error) {
	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		s.Start()
		wg.Done()
	}()

	go func() {
		err = s.Server.ListenAndServe()
		wg.Done()
	}()

	wg.Wait()

	return err
}

// Shutdown stops the server.
func (s *Server) Shutdown(ctx context.Context) error {
	s.ws.Stop()
	close(s.didSaveChan)
	return s.Server.Shutdown(ctx)
}

// Start starts the main loops. You do not need to call this if you call
// ListenAndServe().
func (s *Server) Start() {
	s.adapter.AddDidSaveChannel(s.didSaveChan)

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		s.ws.Start()
		wg.Done()
	}()

	go func() {
		s.loop()
		wg.Done()
	}()

	wg.Wait()
}

// Web socket loop.
func (s *Server) loop() {
	for seg := range s.didSaveChan {
		s.ws.Broadcast(&msg{
			Type: DidSave,
			Data: seg,
		}, nil)
	}
}

func (s *Server) root(w http.ResponseWriter, r *http.Request, _ httprouter.Params) (interface{}, error) {
	adapterInfo, err := s.adapter.GetInfo()
	if err != nil {
		return nil, err
	}

	return &Info{
		Adapter: adapterInfo,
	}, nil
}

func (s *Server) saveSegment(w http.ResponseWriter, r *http.Request, _ httprouter.Params) (interface{}, error) {
	decoder := json.NewDecoder(r.Body)

	var seg cs.Segment
	if err := decoder.Decode(&seg); err != nil {
		return nil, jsonhttp.NewErrBadRequest("")
	}
	if err := seg.Validate(); err != nil {
		return nil, jsonhttp.NewErrBadRequest(err.Error())
	}
	if err := s.adapter.SaveSegment(&seg); err != nil {
		return nil, err
	}

	return seg, nil
}

func (s *Server) getSegment(w http.ResponseWriter, r *http.Request, p httprouter.Params) (interface{}, error) {
	linkHash, err := types.NewBytes32FromString(p.ByName("linkHash"))
	if err != nil {
		return nil, err
	}

	seg, err := s.adapter.GetSegment(linkHash)
	if err != nil {
		return nil, err
	}
	if seg == nil {
		return nil, jsonhttp.NewErrNotFound("")
	}

	return seg, nil
}

func (s *Server) deleteSegment(w http.ResponseWriter, r *http.Request, p httprouter.Params) (interface{}, error) {
	linkHash, err := types.NewBytes32FromString(p.ByName("linkHash"))
	if err != nil {
		return nil, err
	}

	seg, err := s.adapter.DeleteSegment(linkHash)
	if err != nil {
		return nil, err
	}
	if seg == nil {
		return nil, jsonhttp.NewErrNotFound("")
	}

	return seg, nil
}

func (s *Server) findSegments(w http.ResponseWriter, r *http.Request, _ httprouter.Params) (interface{}, error) {
	filter, e := parseSegmentFilter(r)
	if e != nil {
		return nil, e
	}

	slice, err := s.adapter.FindSegments(filter)
	if err != nil {
		return nil, err
	}

	return slice, nil
}

func (s *Server) getMapIDs(w http.ResponseWriter, r *http.Request, _ httprouter.Params) (interface{}, error) {
	filter, e := parseMapFilter(r)
	if e != nil {
		return nil, e
	}

	slice, err := s.adapter.GetMapIDs(filter)
	if err != nil {
		return nil, err
	}

	return slice, nil
}

func (s *Server) getWebSocket(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	s.ws.Handle(w, r)
}
