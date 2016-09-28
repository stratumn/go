// Copyright 2016 Stratumn SAS. All rights reserved.
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

// Package jsonhttp defines a simple HTTP server that renders JSON.
//
// Routes can be added by passing a handle that should return JSON serializable data or an error.
package jsonhttp

import (
	"encoding/json"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"
)

const (
	// DefaultAddress is the default address of the server.
	DefaultAddress = ":5000"
)

// Config contains configuration options for the server.
type Config struct {
	// The address of the server.
	Address string

	// Optionally, the path to a TLS certificate.
	CertFile string

	// Optionally, the path to a TLS private key.
	KeyFile string
}

// Server is the type that implements net/http.Handler.
type Server struct {
	router *httprouter.Router
	config *Config
}

// Handle is the function type for a route handle.
type Handle func(http.ResponseWriter, *http.Request, httprouter.Params, *Config) (interface{}, error)

// NotFound is a handle for a route that is not found.
func NotFound(w http.ResponseWriter, r *http.Request, _ httprouter.Params, _ *Config) (interface{}, error) {
	return nil, NewErrNotFound("")
}

// New creates an instance of Server.
func New(config *Config) *Server {
	r := httprouter.New()
	r.NotFound = notFoundHandler{config, NotFound}
	return &Server{r, config}
}

// ServeHTTP implements net/http.Handler.ServeHTTP.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// Get adds a GET route.
func (s *Server) Get(path string, handle Handle) {
	s.router.GET(path, handler{s.config, handle}.ServeHTTP)
}

// Post adds a POST route.
func (s *Server) Post(path string, handle Handle) {
	s.router.POST(path, handler{s.config, handle}.ServeHTTP)
}

// Put adds a PUT route.
func (s *Server) Put(path string, handle Handle) {
	s.router.PUT(path, handler{s.config, handle}.ServeHTTP)
}

// Delete adds a DELETE route.
func (s *Server) Delete(path string, handle Handle) {
	s.router.DELETE(path, handler{s.config, handle}.ServeHTTP)
}

// Patch adds a PATCH route.
func (s *Server) Patch(path string, handle Handle) {
	s.router.PATCH(path, handler{s.config, handle}.ServeHTTP)
}

// Options adds an OPTIONS route.
func (s *Server) Options(path string, handle Handle) {
	s.router.OPTIONS(path, handler{s.config, handle}.ServeHTTP)
}

// ListenAndServe starts the server.
func (s *Server) ListenAndServe() error {
	addr := s.config.Address
	if addr == "" {
		addr = DefaultAddress
	}

	if s.config.CertFile != "" && s.config.KeyFile != "" {
		return http.ListenAndServeTLS(addr, s.config.CertFile, s.config.KeyFile, s)
	}

	return http.ListenAndServe(addr, s)
}

type handler struct {
	config *Config
	serve  Handle
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var err error

	data, err := h.serve(w, r, p, h.config)
	if err != nil {
		renderErr(w, r, err, h.config)
		return
	}

	js, err := json.Marshal(data)
	if err != nil {
		renderErr(w, r, err, h.config)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func renderErr(w http.ResponseWriter, r *http.Request, err error, c *Config) {
	e, ok := err.(ErrHTTP)
	if ok {
		log.WithFields(log.Fields{
			"status": e.Status(),
			"method": r.Method,
			"url":    r.RequestURI,
			"origin": r.RemoteAddr,
			"error":  err,
		}).Warn("Failed to handle request")
	} else {
		log.WithFields(log.Fields{
			"status": 500,
			"method": r.Method,
			"url":    r.RequestURI,
			"origin": r.RemoteAddr,
			"error":  err,
		}).Error("Failed to handle request")
		e = NewErrInternalServer("")
	}

	w.Header().Set("Content-Type", "application/json")
	http.Error(w, string(e.JSONMarshal()), e.Status())
}

type notFoundHandler handler

func (h notFoundHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler(h).ServeHTTP(w, r, nil)
}
