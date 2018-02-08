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
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/store/storetestcases"
	"github.com/stratumn/go-indigocore/tmpop/tmpoptestcases"
	"github.com/stratumn/go-indigocore/utils"
)

var (
	test        *testing.T
	integration = flag.Bool("integration", false, "Run integration tests")
)

const (
	domain = "0.0.0.0"
	port   = "5984"
)

func TestMain(m *testing.M) {
	flag.Parse()
	if *integration {
		// Couch container configuration.
		imageName := "couchdb:2.1"
		containerName := "sdk_couchstore_integration_test"
		p, _ := nat.NewPort("tcp", port)
		exposedPorts := map[nat.Port]struct{}{p: {}}
		portBindings := nat.PortMap{
			p: []nat.PortBinding{
				{
					HostIP:   domain,
					HostPort: port,
				},
			},
		}

		// Stop container if it is already running, swallow error.
		utils.KillContainer(containerName)

		// Start couchdb container
		if err := utils.RunContainer(containerName, imageName, exposedPorts, portBindings); err != nil {
			fmt.Printf(err.Error())
			os.Exit(1)
		}

		// Retry until container is ready.
		if err := utils.Retry(func(attempt int) (bool, error) {
			_, err := http.Get(fmt.Sprintf("http://%s:%s", domain, port))
			if err != nil {
				time.Sleep(1 * time.Second)
				return true, err
			}
			return false, err
		}, 10); err != nil {
			fmt.Printf(err.Error())
			os.Exit(1)
		}

		// Run tests.
		testResult := m.Run()

		// Stop couchdb container.
		if err := utils.KillContainer(containerName); err != nil {
			fmt.Printf(err.Error())
			os.Exit(1)
		}

		os.Exit(testResult)
	}
}

func TestCouchStore(t *testing.T) {
	test = t
	if *integration {
		factory := storetestcases.Factory{
			New:               newTestCouchStoreAdapter,
			NewKeyValueStore:  newTestCouchStoreKeyValue,
			Free:              freeTestCouchStoreAdapter,
			FreeKeyValueStore: freeTestCouchStoreKeyValue,
		}

		factory.RunStoreTests(t)
		factory.RunKeyValueStoreTests(t)
	}
}

func TestCouchTMPop(t *testing.T) {
	if *integration {
		tmpoptestcases.Factory{
			New:  newTestCouchStoreTMPop,
			Free: freeTestCouchStoreTMPop,
		}.RunTests(t)
	}
}

func newTestCouchStore() (*CouchStore, error) {
	config := &Config{
		Address: fmt.Sprintf("http://%s:%s", domain, port),
	}
	return New(config)
}

func newTestCouchStoreAdapter() (store.Adapter, error) {
	return newTestCouchStore()
}

func newTestCouchStoreKeyValue() (store.KeyValueStore, error) {
	return newTestCouchStore()
}

func newTestCouchStoreTMPop() (store.Adapter, store.KeyValueStore, error) {
	a, err := newTestCouchStore()
	return a, a, err
}

func freeTestCouchStore(a *CouchStore) {
	if err := a.deleteDatabase(dbLink); err != nil {
		test.Fatal(err)
	}
	if err := a.deleteDatabase(dbEvidences); err != nil {
		test.Fatal(err)
	}
}

func freeTestCouchStoreAdapter(a store.Adapter) {
	freeTestCouchStore(a.(*CouchStore))
}

func freeTestCouchStoreKeyValue(a store.KeyValueStore) {
	freeTestCouchStore(a.(*CouchStore))
}

func freeTestCouchStoreTMPop(a store.Adapter, _ store.KeyValueStore) {
	freeTestCouchStoreAdapter(a)
}
