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

package tmstore

import (
	"net/http"
	"testing"

	"github.com/stratumn/sdk/cs/cstesting"
	"github.com/stratumn/sdk/jsonhttp"
	"github.com/stratumn/sdk/store"
	"github.com/stratumn/sdk/store/storetestcases"
	"github.com/stretchr/testify/assert"
)

var (
	tmstore *TMStore
)

func TestTMStore(t *testing.T) {
	storetestcases.Factory{
		New:  newTestTMStore,
		Free: resetTMPop,
	}.RunStoreTests(t)
}

func newTestTMStore() (store.Adapter, error) {
	tmstore = NewTestClient()
	tmstore.RetryStartWebsocket(DefaultWsRetryInterval)

	return tmstore, nil
}

func resetTMPop(_ store.Adapter) {
	ResetNode()
}

// TestWebSocket tests how the web socket with Tendermint behaves
func TestWebSocket(t *testing.T) {
	tmstore = NewTestClient()

	t.Run("Start and stop websocket", func(t *testing.T) {
		err := tmstore.StartWebsocket()
		assert.NoError(t, err)

		err = tmstore.StopWebsocket()
		assert.NoError(t, err)
	})

	t.Run("Start websocket multiple times", func(t *testing.T) {
		err := tmstore.StartWebsocket()
		assert.NoError(t, err)

		err = tmstore.StartWebsocket()
		assert.NoError(t, err)

		err = tmstore.StopWebsocket()
		assert.NoError(t, err)
	})

	t.Run("Stop already stopped websocket", func(t *testing.T) {
		err := tmstore.StopWebsocket()
		assert.NoError(t, err)
	})
}

// TestValidation tests custom validation rules
func TestValidation(t *testing.T) {
	tmstore, err := newTestTMStore()
	assert.NoError(t, err)

	l := cstesting.RandomLink()
	l.Meta["process"] = "testProcess"
	l.Meta["action"] = "init"
	l.State["string"] = 42

	_, err = tmstore.CreateLink(l)
	assert.Error(t, err, "A validation error is expected")

	errHTTP, ok := err.(jsonhttp.ErrHTTP)
	assert.True(t, ok, "Invalid error received: want ErrHTTP")
	assert.Equal(t, http.StatusBadRequest, errHTTP.Status())
}
