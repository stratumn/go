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

package jsonhttp

import (
	"errors"
	"net/http"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/stratumn/sdk/testutil"
)

func TestGet(t *testing.T) {
	s := New(&Config{})
	s.Get("/test", func(r http.ResponseWriter, _ *http.Request, p httprouter.Params) (interface{}, error) {
		return map[string]bool{"test": true}, nil
	})

	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/test", nil, nil)
	if err != nil {
		t.Fatalf("testutil.RequestJSON(): err: %s", err)
	}

	if got, want := w.Body.String(), `{"test":true}`; got != want {
		t.Errorf("w.Body = %s want %s", got, want)
	}
}

func TestPost(t *testing.T) {
	s := New(&Config{})
	s.Post("/test", func(r http.ResponseWriter, _ *http.Request, p httprouter.Params) (interface{}, error) {
		return map[string]bool{"test": true}, nil
	})

	w, err := testutil.RequestJSON(s.ServeHTTP, "POST", "/test", nil, nil)
	if err != nil {
		t.Fatalf("testutil.RequestJSON(): err: %s", err)
	}

	if got, want := w.Body.String(), `{"test":true}`; got != want {
		t.Errorf("w.Body = %s want %s", got, want)
	}
}

func TestPut(t *testing.T) {
	s := New(&Config{})
	s.Put("/test", func(r http.ResponseWriter, _ *http.Request, p httprouter.Params) (interface{}, error) {
		return map[string]bool{"test": true}, nil
	})

	w, err := testutil.RequestJSON(s.ServeHTTP, "PUT", "/test", nil, nil)
	if err != nil {
		t.Fatalf("testutil.RequestJSON(): err: %s", err)
	}

	if got, want := w.Body.String(), `{"test":true}`; got != want {
		t.Errorf("w.Body = %s want %s", got, want)
	}
}

func TestDelete(t *testing.T) {
	s := New(&Config{})
	s.Delete("/test", func(r http.ResponseWriter, _ *http.Request, p httprouter.Params) (interface{}, error) {
		return map[string]bool{"test": true}, nil
	})

	w, err := testutil.RequestJSON(s.ServeHTTP, "DELETE", "/test", nil, nil)
	if err != nil {
		t.Fatalf("testutil.RequestJSON(): err: %s", err)
	}

	if got, want := w.Body.String(), `{"test":true}`; got != want {
		t.Errorf("w.Body = %s want %s", got, want)
	}
}

func TestPatch(t *testing.T) {
	s := New(&Config{})
	s.Patch("/test", func(r http.ResponseWriter, _ *http.Request, p httprouter.Params) (interface{}, error) {
		return map[string]bool{"test": true}, nil
	})

	w, err := testutil.RequestJSON(s.ServeHTTP, "PATCH", "/test", nil, nil)
	if err != nil {
		t.Fatalf("testutil.RequestJSON(): err: %s", err)
	}

	if got, want := w.Body.String(), `{"test":true}`; got != want {
		t.Errorf("w.Body = %s want %s", got, want)
	}
}

func TestOptions(t *testing.T) {
	s := New(&Config{})
	s.Options("/test", func(r http.ResponseWriter, _ *http.Request, p httprouter.Params) (interface{}, error) {
		return map[string]bool{"test": true}, nil
	})

	w, err := testutil.RequestJSON(s.ServeHTTP, "OPTIONS", "/test", nil, nil)
	if err != nil {
		t.Fatalf("testutil.RequestJSON(): err: %s", err)
	}

	if got, want := w.Body.String(), `{"test":true}`; got != want {
		t.Errorf("w.Body = %s want %s", got, want)
	}
}

func TestNotFound(t *testing.T) {
	s := New(&Config{})

	var body map[string]interface{}
	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/test", nil, &body)
	if err != nil {
		t.Fatalf("testutil.RequestJSON(): err: %s", err)
	}

	if got, want := w.Code, NewErrNotFound("").Status(); got != want {
		t.Errorf("w.Code = %d want %d", got, want)
	}
	if got, want := body["error"].(string), NewErrNotFound("").Error(); got != want {
		t.Errorf(`body["error"] = %q want %q`, got, want)
	}
	if got, want := int(body["status"].(float64)), NewErrNotFound("").Status(); got != want {
		t.Errorf(`body["status"] = %d want %d`, got, want)
	}
}

func TestErrHTTP(t *testing.T) {
	s := New(&Config{})

	s.Get("/test", func(r http.ResponseWriter, _ *http.Request, p httprouter.Params) (interface{}, error) {
		return nil, NewErrBadRequest("no")
	})

	var body map[string]interface{}
	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/test", nil, &body)
	if err != nil {
		t.Fatalf("testutil.RequestJSON(): err: %s", err)
	}

	if got, want := w.Code, NewErrBadRequest("").Status(); got != want {
		t.Errorf("w.Code = %d want %d", got, want)
	}
	if got, want := body["error"].(string), "no"; got != want {
		t.Errorf(`body["error"] = %q want %q`, got, want)
	}
	if got, want := int(body["status"].(float64)), NewErrBadRequest("").Status(); got != want {
		t.Errorf(`body["status"] = %d want %d`, got, want)
	}
}

func TestError(t *testing.T) {
	s := New(&Config{})

	s.Get("/test", func(r http.ResponseWriter, _ *http.Request, p httprouter.Params) (interface{}, error) {
		return nil, errors.New("no")
	})

	var body map[string]interface{}
	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/test", nil, &body)
	if err != nil {
		t.Fatalf("testutil.RequestJSON(): err: %s", err)
	}

	if got, want := w.Code, NewErrInternalServer("").Status(); got != want {
		t.Errorf("w.Code = %d want %d", got, want)
	}
	if got, want := body["error"].(string), NewErrInternalServer("").Error(); got != want {
		t.Errorf(`body["error"] = %q want %q`, got, want)
	}
	if got, want := int(body["status"].(float64)), NewErrInternalServer("").Status(); got != want {
		t.Errorf(`body["status"] = %d want %d`, got, want)
	}
}
