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
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/stratumn/sdk/fossilizer"
	"github.com/stratumn/sdk/fossilizer/fossilizertesting"
	"github.com/stratumn/sdk/jsonhttp"
	"github.com/stratumn/sdk/jsonws"
	"github.com/stratumn/sdk/jsonws/jsonwstesting"
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
	s, a := createServer()
	a.MockFossilize.Fn = func(data []byte, meta []byte) error {
		return nil
	}

	// Make request.
	req := httptest.NewRequest("POST", "/fossils", nil)
	req.Form = url.Values{}
	req.Form.Set("data", "42")
	req.Form.Set("process", "zou")

	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)

	if got, want := w.Code, http.StatusOK; got != want {
		t.Errorf("w.StatusCode = %d want %d", got, want)
	}
}

func TestFossilize_noData(t *testing.T) {
	s, _ := createServer()

	req := httptest.NewRequest("POST", "/fossils", nil)
	req.Form = url.Values{}
	req.Form.Set("process", "zou")

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
	req.Form.Set("data", "1")
	req.Form.Set("process", "zou")

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
	req.Form.Set("data", "12345678901234567890")
	req.Form.Set("process", "zou")

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
	req.Form.Set("data", "azertyuiop")
	req.Form.Set("process", "zou")

	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)

	if got, want := w.Code, newErrData("").Status(); got != want {
		t.Errorf("w.Code = %d want %d", got, want)
	}
}

func TestFossilize_noProcess(t *testing.T) {
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

func TestGetSocket(t *testing.T) {
	// Chan that will receive the event channel.
	sendChan := make(chan chan *fossilizer.Event)

	// Chan used to wait for the connection to be ready.
	readyChan := make(chan struct{})

	// Chan used to wait for web socket message.
	doneChan := make(chan struct{})

	conn := jsonwstesting.MockConn{}
	conn.MockReadJSON.Fn = func(interface{}) error {
		readyChan <- struct{}{}
		return nil
	}
	conn.MockWriteJSON.Fn = func(interface{}) error {
		doneChan <- struct{}{}
		return nil
	}

	upgradeHandle := func(w http.ResponseWriter, r *http.Request, h http.Header) (jsonws.PingableConn, error) {
		return &conn, nil
	}

	// Mock fossilize to publish result to channel.
	a := &fossilizertesting.MockAdapter{}
	a.MockAddFossilizerEventChan.Fn = func(c chan *fossilizer.Event) {
		sendChan <- c
	}

	config := &Config{
		MinDataLen: 2,
		MaxDataLen: 16,
	}

	basicConfig := &jsonws.BasicConfig{UpgradeHandle: upgradeHandle}
	bufConfig := &jsonws.BufferedConnConfig{
		Size:         256,
		WriteTimeout: 10 * time.Second,
		PongTimeout:  70 * time.Second,
		PingInterval: time.Minute,
		MaxMsgSize:   1024,
	}

	s := New(a, config, &jsonhttp.Config{}, basicConfig, bufConfig)

	go s.Start()
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer s.Shutdown(ctx)
	defer cancel()

	// Register web socket connection.
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/websocket", nil)
	go s.getWebSocket(w, r, nil)

	event := &fossilizer.Event{
		EventType: fossilizer.DidFossilizeLink,
		Data:      &fossilizer.Result{},
	}

	// Wait for channel to be added.
	select {
	case eventChan := <-sendChan:
		// Wait for connection to be ready.
		select {
		case <-readyChan:
		case <-time.After(time.Second):
			t.Fatalf("connection ready timeout")
		}
		eventChan <- event
	case <-time.After(time.Second):
		t.Fatalf("save channel not added")
	}

	// Wait for message to be broadcasted.
	expected := &jsonws.Message{
		Type: string(event.EventType),
		Data: event.Data,
	}
	select {
	case <-doneChan:
		got := conn.MockWriteJSON.LastCalledWith.(*jsonws.Message)
		if !reflect.DeepEqual(got, expected) {
			gotjs, _ := json.MarshalIndent(got, "", "  ")
			wantjs, _ := json.MarshalIndent(expected, "", "  ")
			t.Errorf("conn.MockWriteJSON.LastCalledWith = %s\nwant %s", gotjs, wantjs)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("fossilized segment not broadcasted")
	}
}
