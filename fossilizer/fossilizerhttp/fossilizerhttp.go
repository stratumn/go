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
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"

	"github.com/stratumn/go/fossilizer"
	"github.com/stratumn/go/jsonhttp"
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
	// The default number of goroutines that will be used to handle
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

type context struct {
	adapter fossilizer.Adapter
	config  *Config
}

type handle func(http.ResponseWriter, *http.Request, httprouter.Params, *context) (interface{}, error)

type handler struct {
	context *context
	handle  handle
}

func (h handler) serve(w http.ResponseWriter, r *http.Request, p httprouter.Params) (interface{}, error) {
	return h.handle(w, r, p, h.context)
}

// New create an instance of a server.
func New(a fossilizer.Adapter, config *Config, httpConfig *jsonhttp.Config) *jsonhttp.Server {
	if config.NumResultWorkers < 1 {
		config.NumResultWorkers = DefaultNumResultWorkers
	}

	s := jsonhttp.New(httpConfig)
	ctx := &context{a, config}

	s.Get("/", handler{ctx, root}.serve)
	s.Post("/fossils", handler{ctx, fossilize}.serve)

	// Launch result workers.
	rc := make(chan *fossilizer.Result, config.NumResultWorkers)
	a.AddResultChan(rc)
	client := http.Client{Timeout: config.CallbackTimeout}
	for i := 0; i < config.NumResultWorkers; i++ {
		go handleResults(rc, &client)
	}

	return s
}

func root(w http.ResponseWriter, r *http.Request, _ httprouter.Params, c *context) (interface{}, error) {
	adapterInfo, err := c.adapter.GetInfo()
	if err != nil {
		return nil, err
	}

	return &Info{
		Adapter: adapterInfo,
	}, nil
}

func fossilize(w http.ResponseWriter, r *http.Request, p httprouter.Params, c *context) (interface{}, error) {
	data, url, err := parseFossilizeValues(r, c)
	if err != nil {
		return nil, err
	}

	if err := c.adapter.Fossilize(data, []byte(url)); err != nil {
		return nil, err
	}

	return "ok", nil
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

func parseFossilizeValues(r *http.Request, c *context) ([]byte, string, error) {
	if err := r.ParseForm(); err != nil {
		return nil, "", err
	}

	datastr := r.Form.Get("data")
	if datastr == "" {
		return nil, "", newErrData("")
	}

	l := len(datastr)
	if l < c.config.MinDataLen {
		return nil, "", newErrDataLen("")
	}
	if c.config.MaxDataLen > 0 && l > c.config.MaxDataLen {
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
