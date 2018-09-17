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
//	POST /links
//		Saves then renders a link.
//		Body should be a JSON encoded link.
//
//	POST /evidences/:linkHash
//		Adds evidence to a link.
//		Body should be a JSON encoded evidence.
//
//	GET /segments/:linkHash
//		Renders a segment.
//
//	GET /segments?[offset=offset]&[limit=limit]&[mapIds[]=id1]&[mapIds[]=id2]&[prevLinkHash=prevLinkHash]&[tags[]=tag1]&[tags[]=tag2]
//		Finds and renders segments.
//
//	GET /maps?[offset=offset]&[limit=limit]
//		Finds and renders map IDs.
//
//	GET /websocket
//		A web socket that broadcasts messages from the store:
//			{ "type": "SavedLink", "data": [link] }
//			{ "type": "SavedEvidence", "data": [evidence] }
package storehttp

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"

	"github.com/julienschmidt/httprouter"
	"github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-indigocore/jsonhttp"
	"github.com/stratumn/go-indigocore/jsonws"
	"github.com/stratumn/go-indigocore/monitoring"
	"github.com/stratumn/go-indigocore/store"

	"go.opencensus.io/trace"
)

const (
	// DefaultStoreEventsChanSize is the default size of the store events channel.
	DefaultStoreEventsChanSize = 256

	// DefaultAddress is the default address of the server.
	DefaultAddress = ":5000"
)

// Server is an HTTP server for stores.
type Server struct {
	*jsonhttp.Server
	adapter         store.Adapter
	ws              *jsonws.Basic
	storeEventsChan chan *store.Event
}

// Config contains configuration options for the server.
type Config struct {
	// The size of the store event channel.
	StoreEventsChanSize int
}

// Info is the info returned by the root route.
type Info struct {
	Adapter interface{} `json:"adapter"`
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
		Server:          jsonhttp.New(httpConfig),
		adapter:         a,
		ws:              jsonws.NewBasic(basicConfig, bufConnConfig),
		storeEventsChan: make(chan *store.Event, config.StoreEventsChanSize),
	}

	s.Get("/", s.root)
	s.Post("/links", s.createLink)
	s.Post("/evidences/:linkHash", s.addEvidence)
	s.Get("/segments/:linkHash", s.getSegment)
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
	close(s.storeEventsChan)
	return s.Server.Shutdown(ctx)
}

// Start starts the main loops. You do not need to call this if you call
// ListenAndServe().
func (s *Server) Start() {
	s.adapter.AddStoreEventChannel(s.storeEventsChan)

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
	for event := range s.storeEventsChan {
		s.ws.Broadcast(&jsonws.Message{
			Type: string(event.EventType),
			Data: event.Data,
		}, nil)
	}
}

func (s *Server) root(w http.ResponseWriter, r *http.Request, _ httprouter.Params) (interface{}, error) {
	ctx, span := trace.StartSpan(r.Context(), "storehttp/root")
	defer span.End()

	adapterInfo, err := s.adapter.GetInfo(ctx)
	if err != nil {
		span.SetStatus(trace.Status{Code: monitoring.Unknown, Message: err.Error()})
		return nil, err
	}

	return &Info{
		Adapter: adapterInfo,
	}, nil
}

func (s *Server) createLink(w http.ResponseWriter, r *http.Request, _ httprouter.Params) (interface{}, error) {
	ctx, span := trace.StartSpan(r.Context(), "storehttp/createLink")
	defer span.End()

	decoder := json.NewDecoder(r.Body)

	var link chainscript.Link
	if err := decoder.Decode(&link); err != nil {
		span.SetStatus(trace.Status{Code: monitoring.InvalidArgument, Message: err.Error()})
		return nil, jsonhttp.NewErrBadRequest(err.Error())
	}

	if _, err := s.adapter.CreateLink(ctx, &link); err != nil {
		span.SetStatus(trace.Status{Code: monitoring.Unknown, Message: err.Error()})
		return nil, err
	}

	return link.Segmentify()
}

func (s *Server) addEvidence(w http.ResponseWriter, r *http.Request, p httprouter.Params) (interface{}, error) {
	ctx, span := trace.StartSpan(r.Context(), "storehttp/addEvidence")
	defer span.End()

	linkHash, err := chainscript.NewLinkHashFromString(p.ByName("linkHash"))
	if err != nil {
		span.SetStatus(trace.Status{Code: monitoring.InvalidArgument, Message: err.Error()})
		return nil, err
	}

	decoder := json.NewDecoder(r.Body)

	var evidence chainscript.Evidence
	if err := decoder.Decode(&evidence); err != nil {
		span.SetStatus(trace.Status{Code: monitoring.InvalidArgument, Message: err.Error()})
		return nil, jsonhttp.NewErrBadRequest(err.Error())
	}

	if err := s.adapter.AddEvidence(ctx, linkHash, &evidence); err != nil {
		span.SetStatus(trace.Status{Code: monitoring.Unknown, Message: err.Error()})
		return nil, err
	}

	return nil, nil
}

func (s *Server) getSegment(w http.ResponseWriter, r *http.Request, p httprouter.Params) (interface{}, error) {
	ctx, span := trace.StartSpan(r.Context(), "storehttp/getSegment")
	defer span.End()

	linkHash, err := chainscript.NewLinkHashFromString(p.ByName("linkHash"))
	if err != nil {
		span.SetStatus(trace.Status{Code: monitoring.InvalidArgument, Message: err.Error()})
		return nil, jsonhttp.NewErrBadRequest(err.Error())
	}

	seg, err := s.adapter.GetSegment(ctx, linkHash)
	if err != nil {
		span.SetStatus(trace.Status{Code: monitoring.Unknown, Message: err.Error()})
		return nil, err
	}
	if seg == nil {
		span.SetStatus(trace.Status{Code: monitoring.NotFound})
		return nil, jsonhttp.NewErrNotFound("")
	}

	return seg, nil
}

func (s *Server) findSegments(w http.ResponseWriter, r *http.Request, _ httprouter.Params) (interface{}, error) {
	ctx, span := trace.StartSpan(r.Context(), "storehttp/findSegments")
	defer span.End()

	filter, e := parseSegmentFilter(r)
	if e != nil {
		span.SetStatus(trace.Status{Code: monitoring.InvalidArgument, Message: e.Error()})
		return nil, jsonhttp.NewErrBadRequest(e.Error())
	}

	slice, err := s.adapter.FindSegments(ctx, filter)
	if err != nil {
		span.SetStatus(trace.Status{Code: monitoring.Unknown, Message: err.Error()})
		return nil, err
	}

	return slice, nil
}

func (s *Server) getMapIDs(w http.ResponseWriter, r *http.Request, _ httprouter.Params) (interface{}, error) {
	ctx, span := trace.StartSpan(r.Context(), "storehttp/getMapIDs")
	defer span.End()

	filter, e := parseMapFilter(r)
	if e != nil {
		span.SetStatus(trace.Status{Code: monitoring.InvalidArgument, Message: e.Error()})
		return nil, jsonhttp.NewErrBadRequest(e.Error())
	}

	slice, err := s.adapter.GetMapIDs(ctx, filter)
	if err != nil {
		span.SetStatus(trace.Status{Code: monitoring.Unknown, Message: err.Error()})
		return nil, err
	}

	return slice, nil
}

func (s *Server) getWebSocket(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	s.ws.Handle(w, r)
}
