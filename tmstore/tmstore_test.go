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
)

var (
	tmstore *TMStore
)

func TestTMStore(t *testing.T) {
	storetestcases.Factory{
		NewV2:  newTestTMStore,
		FreeV2: resetTMPop,
	}.RunTestsV2(t)
}

func newTestTMStore() (store.AdapterV2, error) {
	tmstore = NewTestClient()
	tmstore.RetryStartWebsocket(DefaultWsRetryInterval)

	return tmstore, nil
}

func resetTMPop(_ store.AdapterV2) {
	ResetNode()
}

func TestValidation(t *testing.T) {
	tmstore, err := newTestTMStore()
	if err != nil {
		t.Fatalf("newTestTMStore(): err: %s", err)
	}

	l := cstesting.RandomLink()
	l.Meta["process"] = "testProcess"
	l.Meta["action"] = "init"
	l.State["string"] = 42

	_, err = tmstore.CreateLink(l)
	if err == nil {
		t.Error("A validation error is expected")
	}

	errHTTP, ok := err.(jsonhttp.ErrHTTP)
	if !ok {
		t.Error("Invalid error received: want ErrHTTP")
	}

	if got := errHTTP.Status(); got != http.StatusBadRequest {
		t.Errorf("status = %d want %d", got, http.StatusBadRequest)
	}
}
