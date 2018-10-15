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

package fossilizerhttp_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/stratumn/go-core/dummyfossilizer"
	"github.com/stratumn/go-core/fossilizer/fossilizerhttp"
	"github.com/stratumn/go-core/jsonhttp"
	"github.com/stratumn/go-core/jsonws"
)

// This example shows how to create a server from a dummyfossilizer.
// It also tests the root route of the server using net/http/httptest.
func Example() {
	// Create a dummy adapter.
	a := dummyfossilizer.New(&dummyfossilizer.Config{Version: "x.x.x", Commit: "abc"})
	config := &fossilizerhttp.Config{
		MaxDataLen: 64,
	}
	httpConfig := &jsonhttp.Config{
		Address: ":6000",
	}
	basicConfig := &jsonws.BasicConfig{}
	bufConfig := &jsonws.BufferedConnConfig{
		Size:         256,
		WriteTimeout: 10 * time.Second,
		PongTimeout:  70 * time.Second,
		PingInterval: time.Minute,
		MaxMsgSize:   1024,
	}

	// Create a server.
	s := fossilizerhttp.New(a, config, httpConfig, basicConfig, bufConfig)
	go s.Start()
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer s.Shutdown(ctx)
	defer cancel()

	// Create a test server.
	ts := httptest.NewServer(s)
	defer ts.Close()

	// Test the root route.
	res, err := http.Get(ts.URL)
	if err != nil {
		log.Fatal(err)
	}

	info, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s", info)
	// Output: {"adapter":{"name":"dummyfossilizer","description":"Stratumn's Dummy Fossilizer","version":"x.x.x","commit":"abc"}}
}
