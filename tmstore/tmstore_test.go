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
	test    *testing.T
)

func TestTMStore(t *testing.T) {
	test = t
	storetestcases.Factory{
		New:  newTestTMStore,
		Free: freeTestTMStore,
	}.RunTests(t)
}

func newTestTMStore() (store.Adapter, error) {
	tmstore = NewTestClient()
	tmstore.RetryStartWebsocket(DefaultWsRetryInterval)

	return tmstore, nil
}

func freeTestTMStore(s store.Adapter) {
	mapIDs, err := tmstore.GetMapIDs(&store.MapFilter{Pagination: store.Pagination{Limit: 100}})
	if err != nil {
		test.Fatal(err)
	}
	segments, err := tmstore.FindSegments(&store.SegmentFilter{MapIDs: mapIDs, Pagination: store.Pagination{Limit: 100}})
	if err != nil {
		test.Fatal(err)
	}
	for _, s := range segments {
		tmstore.DeleteSegment(s.GetLinkHash())
	}
}

func TestValidation(t *testing.T) {
	tmstore, err := newTestTMStore()
	if err != nil {
		t.Fatalf("newTestTMStore(): err: %s", err)
	}

	s := cstesting.RandomSegment()
	s.Link.Meta["process"] = "testProcess"
	s.Link.Meta["action"] = "init"
	s.Link.State["string"] = 42

	err = tmstore.SaveSegment(s)
	if err == nil {
		t.Error("a.DeliverTx(): want error")
	}

	errHTTP, ok := err.(jsonhttp.ErrHTTP)
	if !ok {
		t.Error("a.DeliverTx(): want ErrHTTP")
	}

	if got := errHTTP.Status(); got != http.StatusBadRequest {
		t.Errorf("status = %d want %d", got, http.StatusBadRequest)
	}
}
