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

package couchstore

import (
	"flag"
	"testing"

	"github.com/stratumn/sdk/store"

	"github.com/stratumn/sdk/store/storetestcases"
)

var (
	myCouchstore *CouchStore
	test         *testing.T
	integration  = flag.Bool("integration", false, "Run integration tests")
)

func TestCouchStore(t *testing.T) {
	flag.Parse()
	test = t
	if *integration {
		storetestcases.Factory{
			New:  newTestCouchStore,
			Free: freeTestCouchStore,
		}.RunTests(t)
	}
}

func newTestCouchStore() (store.Adapter, error) {
	config := &Config{
		Address: "http://localhost:5984",
	}
	return New(config)
}

func freeTestCouchStore(a store.Adapter) {
	if err := myCouchstore.deleteDatabase(dbSegment); err != nil {
		test.Fatal(err)
	}
}
