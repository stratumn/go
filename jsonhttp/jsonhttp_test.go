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

package jsonhttp

import (
	"errors"
	"net/http"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/stratumn/go-core/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGet(t *testing.T) {
	s := New(&Config{})
	s.Get("/test", func(r http.ResponseWriter, _ *http.Request, p httprouter.Params) (interface{}, error) {
		return map[string]bool{"test": true}, nil
	})

	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/test", nil, nil)
	require.NoError(t, err)
	assert.Equal(t, `{"test":true}`, w.Body.String())
}

func TestPost(t *testing.T) {
	s := New(&Config{})
	s.Post("/test", func(r http.ResponseWriter, _ *http.Request, p httprouter.Params) (interface{}, error) {
		return map[string]bool{"test": true}, nil
	})

	w, err := testutil.RequestJSON(s.ServeHTTP, "POST", "/test", nil, nil)
	require.NoError(t, err)
	assert.Equal(t, `{"test":true}`, w.Body.String())
}

func TestPut(t *testing.T) {
	s := New(&Config{})
	s.Put("/test", func(r http.ResponseWriter, _ *http.Request, p httprouter.Params) (interface{}, error) {
		return map[string]bool{"test": true}, nil
	})

	w, err := testutil.RequestJSON(s.ServeHTTP, "PUT", "/test", nil, nil)
	require.NoError(t, err)
	assert.Equal(t, `{"test":true}`, w.Body.String())
}

func TestDelete(t *testing.T) {
	s := New(&Config{})
	s.Delete("/test", func(r http.ResponseWriter, _ *http.Request, p httprouter.Params) (interface{}, error) {
		return map[string]bool{"test": true}, nil
	})

	w, err := testutil.RequestJSON(s.ServeHTTP, "DELETE", "/test", nil, nil)
	require.NoError(t, err)
	assert.Equal(t, `{"test":true}`, w.Body.String())
}

func TestPatch(t *testing.T) {
	s := New(&Config{})
	s.Patch("/test", func(r http.ResponseWriter, _ *http.Request, p httprouter.Params) (interface{}, error) {
		return map[string]bool{"test": true}, nil
	})

	w, err := testutil.RequestJSON(s.ServeHTTP, "PATCH", "/test", nil, nil)
	require.NoError(t, err)
	assert.Equal(t, `{"test":true}`, w.Body.String())
}

func TestOptions(t *testing.T) {
	s := New(&Config{})
	s.Options("/test", func(r http.ResponseWriter, _ *http.Request, p httprouter.Params) (interface{}, error) {
		return map[string]bool{"test": true}, nil
	})

	w, err := testutil.RequestJSON(s.ServeHTTP, "OPTIONS", "/test", nil, nil)
	require.NoError(t, err)
	assert.Equal(t, `{"test":true}`, w.Body.String())
}

func TestNotFound(t *testing.T) {
	s := New(&Config{})

	var body map[string]interface{}
	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/test", nil, &body)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, "Not Found", body["error"])
	assert.Equal(t, http.StatusNotFound, int(body["status"].(float64)))
}

func TestErrHTTP(t *testing.T) {
	s := New(&Config{})

	s.Get("/test", func(r http.ResponseWriter, _ *http.Request, p httprouter.Params) (interface{}, error) {
		return nil, ErrHTTP{msg: "no", status: http.StatusBadRequest}
	})

	var body map[string]interface{}
	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/test", nil, &body)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "no", body["error"])
	assert.Equal(t, http.StatusBadRequest, int(body["status"].(float64)))
}

func TestError(t *testing.T) {
	s := New(&Config{})

	s.Get("/test", func(r http.ResponseWriter, _ *http.Request, p httprouter.Params) (interface{}, error) {
		return nil, errors.New("no")
	})

	var body map[string]interface{}
	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/test", nil, &body)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "Internal Server Error", body["error"])
	assert.Equal(t, http.StatusInternalServerError, int(body["status"].(float64)))
}
