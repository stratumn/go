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

package jsonhttp_test

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"

	"github.com/julienschmidt/httprouter"

	"github.com/stratumn/sdk/jsonhttp"
)

// This example shows how to create a server and add a route with a named param.
// It also tests the route using net/http/httptest.
func ExampleServer() {
	// Create the server.
	s := jsonhttp.New(&jsonhttp.Config{Address: ":3333"})

	// Add a route with a named param.
	s.Get("/items/:id", func(r http.ResponseWriter, _ *http.Request, p httprouter.Params) (interface{}, error) {
		// Return a map containing the ID.
		result := map[string]string{
			"id": p.ByName("id"),
		}

		return result, nil
	})

	// Create a test server.
	ts := httptest.NewServer(s)
	defer ts.Close()

	// Test our route.
	res, err := http.Get(ts.URL + "/items/one")
	if err != nil {
		log.Fatal(err)
	}

	item, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s", item)
	// Output: {"id":"one"}
}
