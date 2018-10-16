// Copyright 2016-2018 Stratumn SAS. All rights reserved.
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

// Package fossilizerhttp is used to create an HTTP server from a fossilizer
// adapter.
//
// It serves the following routes:
//	GET /
//		Renders information about the fossilizer.
//
//	POST /fossils
//		Requests data to be fossilized.
//		Form.data should be a hex encoded buffer.
//		Form.process should be the name of the process that generated
// 		the data
package fossilizerhttp

import (
	"context"
	"encoding/hex"
	"net/http"
	"sync"

	"github.com/julienschmidt/httprouter"
	"github.com/stratumn/go-core/fossilizer"
	"github.com/stratumn/go-core/jsonhttp"
	"github.com/stratumn/go-core/jsonws"
	"github.com/stratumn/go-core/monitoring"
	"github.com/stratumn/go-core/monitoring/errorcode"
	"github.com/stratumn/go-core/types"

	"go.opencensus.io/trace"
)

const (
	// DefaultAddress is the default address of the server.
	DefaultAddress = ":6000"

	// DefaultMinDataLen is the default minimum fossilize data length.
	DefaultMinDataLen = 32

	// DefaultMaxDataLen is the default maximum fossilize data length.
	DefaultMaxDataLen = 64

	// DefaultFossilizerEventChanSize is the default size of the fossilizer event channel.
	DefaultFossilizerEventChanSize = 256
)

// Config contains configuration options for the server.
type Config struct {
	// The minimum fossilize data length.
	MinDataLen int

	// The maximum fossilize data length.
	MaxDataLen int

	// The size of the EventChan channel.
	FossilizerEventChanSize int
}

// Info is the info returned by the root route.
type Info struct {
	Adapter interface{} `json:"adapter"`
}

// Server is an HTTP server for fossilizers.
type Server struct {
	*jsonhttp.Server
	adapter             fossilizer.Adapter
	config              *Config
	ws                  *jsonws.Basic
	fossilizerEventChan chan *fossilizer.Event
}

// New create an instance of a server.
func New(
	a fossilizer.Adapter,
	config *Config,
	httpConfig *jsonhttp.Config,
	basicConfig *jsonws.BasicConfig,
	bufConnConfig *jsonws.BufferedConnConfig,
) *Server {
	s := Server{
		Server:              jsonhttp.New(httpConfig),
		adapter:             a,
		config:              config,
		ws:                  jsonws.NewBasic(basicConfig, bufConnConfig),
		fossilizerEventChan: make(chan *fossilizer.Event, config.FossilizerEventChanSize),
	}

	s.Get("/", s.root)
	s.Post("/fossils", s.fossilize)
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
	close(s.fossilizerEventChan)
	return s.Server.Shutdown(ctx)
}

// Start starts the main loops. You do not need to call this if you call
// ListenAndServe().
func (s *Server) Start() {
	s.adapter.AddFossilizerEventChan(s.fossilizerEventChan)

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		s.ws.Start()
		wg.Done()
	}()

	go func() {
		s.handleEvents()
		wg.Done()
	}()

	wg.Wait()
}

// Forward events to websocket
func (s *Server) handleEvents() {
	for event := range s.fossilizerEventChan {
		s.ws.Broadcast(&jsonws.Message{
			Type: string(event.EventType),
			Data: event.Data,
		}, nil)
	}
}

func (s *Server) root(w http.ResponseWriter, r *http.Request, _ httprouter.Params) (interface{}, error) {
	ctx, span := trace.StartSpan(r.Context(), "fossilizerhttp/root")
	defer span.End()

	adapterInfo, err := s.adapter.GetInfo(ctx)
	if err != nil {
		return nil, jsonhttp.NewErrHTTP(err)
	}

	return &Info{
		Adapter: adapterInfo,
	}, nil
}

func (s *Server) fossilize(w http.ResponseWriter, r *http.Request, p httprouter.Params) (_ interface{}, err error) {
	ctx, span := trace.StartSpan(r.Context(), "fossilizerhttp/fossilize")
	defer func() {
		monitoring.SetSpanStatusAndEnd(span, err)
	}()

	data, process, err := s.parseFossilizeValues(r)
	if err != nil {
		return nil, jsonhttp.NewErrHTTP(err)
	}

	if err := s.adapter.Fossilize(ctx, data, []byte(process)); err != nil {
		return nil, jsonhttp.NewErrHTTP(err)
	}

	return "ok", nil
}

func (s *Server) parseFossilizeValues(r *http.Request) ([]byte, string, error) {
	if err := r.ParseForm(); err != nil {
		return nil, "", types.WrapError(err, errorcode.InvalidArgument, Component, "form data required")
	}

	datastr := r.Form.Get("data")
	if datastr == "" {
		return nil, "", types.NewError(errorcode.InvalidArgument, Component, "data required")
	}

	l := len(datastr)
	if l < s.config.MinDataLen {
		return nil, "", types.NewError(errorcode.InvalidArgument, Component, "invalid data length (too short)")
	}
	if s.config.MaxDataLen > 0 && l > s.config.MaxDataLen {
		return nil, "", types.NewError(errorcode.InvalidArgument, Component, "invalid data length (too big)")
	}

	data, err := hex.DecodeString(datastr)
	if err != nil {
		return nil, "", types.WrapError(err, errorcode.InvalidArgument, Component, "could not decode data")
	}

	process := r.Form.Get("process")
	if process == "" {
		return nil, "", types.NewError(errorcode.InvalidArgument, Component, "process is required")
	}

	return data, process, nil
}

func (s *Server) getWebSocket(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	s.ws.Handle(w, r)
}
