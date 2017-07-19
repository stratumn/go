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

package fossilizerhttp

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stratumn/sdk/fossilizer"
	"github.com/stratumn/sdk/fossilizer/fossilizertesting"
	"github.com/stratumn/sdk/jsonhttp"
	"github.com/stratumn/sdk/testutil"
)

func TestRoot(t *testing.T) {
	s, a := createServer()
	a.MockGetInfo.Fn = func() (interface{}, error) { return "test", nil }

	var body map[string]interface{}
	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/", nil, &body)
	if err != nil {
		t.Fatalf("testutil.RequestJSON(): err: %s", err)
	}

	if got, want := w.Code, http.StatusOK; got != want {
		t.Errorf("w.StatusCode = %d want %d", got, want)
	}
	if got, want := body["adapter"].(string), "test"; got != want {
		t.Errorf(`body["adapter"] = %q want %q`, got, want)
	}
	if got, want := a.MockGetInfo.CalledCount, 1; got != want {
		t.Errorf("a.MockGetInfo.CalledCount = %d want %d", got, want)
	}
}

func TestRoot_err(t *testing.T) {
	s, a := createServer()
	a.MockGetInfo.Fn = func() (interface{}, error) { return "test", errors.New("error") }

	var body map[string]interface{}
	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/", nil, &body)
	if err != nil {
		t.Fatalf("testutil.RequestJSON(): err: %s", err)
	}

	if got, want := w.Code, jsonhttp.NewErrInternalServer("").Status(); got != want {
		t.Errorf("w.Code = %d want %d", got, want)
	}
	if got, want := body["error"].(string), jsonhttp.NewErrInternalServer("").Error(); got != want {
		t.Errorf(`body["error"] = %q want %q`, got, want)
	}
	if got, want := a.MockGetInfo.CalledCount, 1; got != want {
		t.Errorf("a.MockGetInfo.CalledCount = %d want %d", got, want)
	}
}

func TestFossilize(t *testing.T) {
	a := &fossilizertesting.MockAdapter{}
	ready := make(chan struct{})

	// Capture result channel.
	var rc chan *fossilizer.Result
	a.MockAddResultChan.Fn = func(c chan *fossilizer.Result) {
		rc = c
		ready <- struct{}{}
	}

	// Mock fossilize to publish result to channel.
	a.MockFossilize.Fn = func(data []byte, meta []byte) error {
		rc <- &fossilizer.Result{
			Evidence: "it is known",
			Data:     data,
			Meta:     meta,
		}
		return nil
	}

	s := New(a, &Config{
		MinDataLen: 2,
		MaxDataLen: 16,
	}, &jsonhttp.Config{})

	go s.Start()
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer s.Shutdown(ctx)
	defer cancel()

	<-ready

	// Mock result endpoint.
	l, err := net.Listen("tcp", ":8888")
	if err != nil {
		t.Fatalf("net.Listen(): err: %s", err)
	}
	defer l.Close()
	h := &resultHandler{
		t: t, listener: l,
		want: "\"it is known\"",
		done: make(chan struct{}),
	}
	go http.Serve(l, h)

	// Make request.
	req := httptest.NewRequest("POST", "/fossils", nil)
	req.Form = url.Values{}
	req.Form.Set("data", "1234567890")
	req.Form.Set("callbackUrl", "http://localhost:8888")

	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)

	if got, want := w.Code, http.StatusOK; got != want {
		t.Errorf("w.Code = %d want %d", got, want)
	}

	select {
	case <-h.done:
	case <-time.After(time.Second):
		t.Fatal("callback URL not called after 1s")
	}
}

func TestFossilize_noData(t *testing.T) {
	s, _ := createServer()

	req := httptest.NewRequest("POST", "/fossils", nil)
	req.Form = url.Values{}
	req.Form.Set("callbackUrl", "http://localhost:6666")

	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)

	if got, want := w.Code, newErrData("").Status(); got != want {
		t.Errorf("w.Code = %d want %d", got, want)
	}
}

func TestFossilize_dataTooShort(t *testing.T) {
	s, _ := createServer()

	req := httptest.NewRequest("POST", "/fossils", nil)
	req.Form = url.Values{}
	req.Form.Set("callbackUrl", "http://localhost:6666")
	req.Form.Set("data", "1")

	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)

	if got, want := w.Code, newErrData("").Status(); got != want {
		t.Errorf("w.Code = %d want %d", got, want)
	}
}

func TestFossilize_dataTooLong(t *testing.T) {
	s, _ := createServer()

	req := httptest.NewRequest("POST", "/fossils", nil)
	req.Form = url.Values{}
	req.Form.Set("callbackUrl", "http://localhost:6666")
	req.Form.Set("data", "12345678901234567890")

	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)

	if got, want := w.Code, newErrData("").Status(); got != want {
		t.Errorf("w.Code = %d want %d", got, want)
	}
}

func TestFossilize_dataNotHex(t *testing.T) {
	s, _ := createServer()

	req := httptest.NewRequest("POST", "/fossils", nil)
	req.Form = url.Values{}
	req.Form.Set("callbackUrl", "http://localhost:6666")
	req.Form.Set("data", "azertyuiop")

	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)

	if got, want := w.Code, newErrData("").Status(); got != want {
		t.Errorf("w.Code = %d want %d", got, want)
	}
}

func TestFossilize_noCallback(t *testing.T) {
	s, _ := createServer()

	req := httptest.NewRequest("POST", "/fossils", nil)
	req.Form = url.Values{}
	req.Form.Set("data", "1234567890")

	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)

	if got, want := w.Code, http.StatusBadRequest; got != want {
		t.Errorf("w.Code = %d want %d", got, want)
	}
}

func TestFossilize_noBody(t *testing.T) {
	s, _ := createServer()

	req := httptest.NewRequest("POST", "/fossils", nil)

	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)

	if got, want := w.Code, http.StatusBadRequest; got != want {
		t.Errorf("w.Code = %d want %d", got, want)
	}
}

func TestNotFound(t *testing.T) {
	s, _ := createServer()

	var body map[string]interface{}
	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/azerty", nil, &body)
	if err != nil {
		t.Fatalf("testutil.RequestJSON(): err: %s", err)
	}

	if got, want := w.Code, jsonhttp.NewErrNotFound("").Status(); got != want {
		t.Errorf("w.Code = %d want %d", got, want)
	}
	if got, want := body["error"].(string), jsonhttp.NewErrNotFound("").Error(); got != want {
		t.Errorf(`body["error"] = %q want %q`, got, want)
	}
}
