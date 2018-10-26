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

package fossilizerhttp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stratumn/go-core/fossilizer"
	"github.com/stratumn/go-core/fossilizer/fossilizertesting"
	"github.com/stratumn/go-core/jsonhttp"
	"github.com/stratumn/go-core/jsonws"
	"github.com/stratumn/go-core/jsonws/jsonwstesting"
	"github.com/stratumn/go-core/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoot(t *testing.T) {
	testCases := []struct {
		name       string
		stub       func() (interface{}, error)
		statusCode int
		validate   func(*testing.T, map[string]interface{})
	}{{
		"success",
		func() (interface{}, error) { return "test", nil },
		http.StatusOK,
		func(t *testing.T, body map[string]interface{}) {
			assert.Equal(t, "test", body["adapter"])
		},
	}, {
		"failure",
		func() (interface{}, error) { return "test", errors.New("error") },
		http.StatusInternalServerError,
		func(t *testing.T, body map[string]interface{}) {
			assert.Equal(t, "error", body["error"])
		},
	}}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			s, a := createServer()
			a.MockGetInfo.Fn = tt.stub

			var body map[string]interface{}
			w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/", nil, &body)
			require.NoError(t, err)

			assert.Equal(t, tt.statusCode, w.Code)
			assert.Equal(t, 1, a.MockGetInfo.CalledCount)
			tt.validate(t, body)
		})
	}
}

func TestFossilize(t *testing.T) {
	testCases := []struct {
		name          string
		data          map[string]string
		fossilizerErr error
		statusCode    int
	}{{
		"success",
		map[string]string{
			"data": "42",
			"meta": "zou",
		},
		nil,
		http.StatusOK,
	}, {
		"fossilizer error",
		map[string]string{
			"data": "42",
			"meta": "zou",
		},
		errors.New("fatal"),
		http.StatusInternalServerError,
	}, {
		"missing data",
		map[string]string{
			"meta": "zou",
		},
		nil,
		http.StatusBadRequest,
	}, {
		"data too short",
		map[string]string{
			"data": "1",
			"meta": "zou",
		},
		nil,
		http.StatusBadRequest,
	}, {
		"data too long",
		map[string]string{
			"data": "42424242424242424242424242424242",
			"meta": "zou",
		},
		nil,
		http.StatusBadRequest,
	}, {
		"data not hex",
		map[string]string{
			"data": "spongebob",
			"meta": "zou",
		},
		nil,
		http.StatusBadRequest,
	}, {
		"missing meta",
		map[string]string{
			"data": "42",
		},
		nil,
		http.StatusOK,
	}, {
		"missing request body",
		nil,
		nil,
		http.StatusBadRequest,
	}}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			s, a := createServer()
			a.MockFossilize.Fn = func(data []byte, meta []byte) error {
				return tt.fossilizerErr
			}

			reqBytes, err := json.Marshal(tt.data)
			require.NoError(t, err)

			req := httptest.NewRequest("POST", "/fossils", bytes.NewBuffer(reqBytes))
			req.Header.Add("Content-Type", "application/json")

			w := httptest.NewRecorder()
			s.ServeHTTP(w, req)

			assert.Equal(t, tt.statusCode, w.Code)
		})
	}
}

func TestNotFound(t *testing.T) {
	s, _ := createServer()

	var body map[string]interface{}
	w, err := testutil.RequestJSON(s.ServeHTTP, "GET", "/azerty", nil, &body)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, "Not Found", body["error"])
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
		assert.Equal(t, expected, got)
	case <-time.After(2 * time.Second):
		t.Fatalf("fossilized segment not broadcasted")
	}
}
