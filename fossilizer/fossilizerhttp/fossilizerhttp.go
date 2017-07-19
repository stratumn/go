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
//		Form.callbackUrl should be a URL to be called when the evidence
//		is ready.
package fossilizerhttp

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"

	"github.com/stratumn/sdk/fossilizer"
	"github.com/stratumn/sdk/jsonhttp"
)

const (
	// DefaultAddress is the default address of the server.
	DefaultAddress = ":6000"

	// DefaultNumResultWorkers is the default number of goroutines that will
	// be used to handle fossilizer results.
	DefaultNumResultWorkers = 8

	// DefaultMinDataLen is the default minimum fossilize data length.
	DefaultMinDataLen = 32

	// DefaultMaxDataLen is the default maximum fossilize data length.
	DefaultMaxDataLen = 64

	// DefaultCallbackTimeout is the default timeout of requests to the
	// callback URLs.
	DefaultCallbackTimeout = 10 * time.Second
)

// Config contains configuration options for the server.
type Config struct {
	// The number of goroutines that will be used to handle
	// fossilizer results.
	NumResultWorkers int

	// The minimum fossilize data length.
	MinDataLen int

	// The maximum fossilize data length.
	MaxDataLen int

	// The timeout of requests to the callback URLs.
	CallbackTimeout time.Duration
}

// Info is the info returned by the root route.
type Info struct {
	Adapter interface{} `json:"adapter"`
}

// Server is an HTTP server for fossilizers.
type Server struct {
	*jsonhttp.Server
	adapter    fossilizer.Adapter
	config     *Config
	resultChan chan *fossilizer.Result
}

// New create an instance of a server.
func New(a fossilizer.Adapter, config *Config, httpConfig *jsonhttp.Config) *Server {
	if config.NumResultWorkers < 1 {
		config.NumResultWorkers = DefaultNumResultWorkers
	}

	s := Server{
		Server:     jsonhttp.New(httpConfig),
		adapter:    a,
		config:     config,
		resultChan: make(chan *fossilizer.Result, config.NumResultWorkers),
	}

	s.Get("/", s.root)
	s.Post("/fossils", s.fossilize)

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
	close(s.resultChan)
	return s.Server.Shutdown(ctx)
}

// Start starts the main loops. You do not need to call this if you call
// ListenAndServe().
func (s *Server) Start() {
	s.adapter.AddResultChan(s.resultChan)
	client := http.Client{Timeout: s.config.CallbackTimeout}

	for i := 0; i < s.config.NumResultWorkers; i++ {
		go handleResults(s.resultChan, &client)
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

func (s *Server) fossilize(w http.ResponseWriter, r *http.Request, p httprouter.Params) (interface{}, error) {
	data, url, err := s.parseFossilizeValues(r)
	if err != nil {
		return nil, err
	}

	if err := s.adapter.Fossilize(data, []byte(url)); err != nil {
		return nil, err
	}

	return "ok", nil
}

func (s *Server) parseFossilizeValues(r *http.Request) ([]byte, string, error) {
	if err := r.ParseForm(); err != nil {
		return nil, "", err
	}

	datastr := r.Form.Get("data")
	if datastr == "" {
		return nil, "", newErrData("")
	}

	l := len(datastr)
	if l < s.config.MinDataLen {
		return nil, "", newErrDataLen("")
	}
	if s.config.MaxDataLen > 0 && l > s.config.MaxDataLen {
		return nil, "", newErrDataLen("")
	}

	data, err := hex.DecodeString(datastr)
	if err != nil {
		return nil, "", jsonhttp.NewErrHTTP(err.Error(), http.StatusBadRequest)
	}

	url := r.Form.Get("callbackUrl")
	if url == "" {
		return nil, "", newErrCallbackURL("")
	}

	return data, url, nil
}

func handleResults(resultChan chan *fossilizer.Result, client *http.Client) {
	for r := range resultChan {
		body, err := json.Marshal(r.Evidence)
		if err != nil {
			log.WithField("error", err).Error("Failed to marshal evidence")
			continue
		}

		url := string(r.Meta)
		req, err := http.NewRequest("POST", string(r.Meta), bytes.NewReader(body))
		if err != nil {
			log.WithFields(log.Fields{
				"url":   url,
				"error": err,
			}).Error("Failed to create callback request")
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		res, err := client.Do(req)
		if err != nil {
			log.WithFields(log.Fields{
				"url":   url,
				"error": err,
			}).Error("Failed to execute callback request")
			continue
		} else if res.StatusCode >= 300 {
			log.WithFields(log.Fields{
				"url":    url,
				"status": res.StatusCode,
				"error":  err,
			}).Error("Invalid callback status code")
		}
		if err := res.Body.Close(); err != nil {
			log.WithFields(log.Fields{
				"url":   url,
				"error": err,
			}).Error("Failed to close callback request")
		}
	}
}
