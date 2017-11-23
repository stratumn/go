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
	"github.com/stratumn/sdk/store"
	"github.com/stratumn/sdk/store/storetestcases"
	"github.com/stratumn/sdk/tmpop/tmpoptestcases"
	"github.com/stratumn/sdk/utils"
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
		imageName := "couchdb:latest"
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
		if err := utils.Retry(pingCouchContainer, 10); err != nil {
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
		storetestcases.Factory{
			New:  newTestCouchStore,
			Free: freeTestCouchStore,
		}.RunTests(t)
	}
}

func TestCouchTMPop(t *testing.T) {
	if *integration {
		tmpoptestcases.Factory{
			New:  newTestCouchStore,
			Free: freeTestCouchStore,
		}.RunTests(t)
	}
}

func newTestCouchStore() (store.Adapter, error) {
	config := &Config{
		Address: fmt.Sprintf("http://%s:%s", domain, port),
	}
	return New(config)
}

func freeTestCouchStore(a store.Adapter) {
	if err := a.(*CouchStore).deleteDatabase(dbSegment); err != nil {
		test.Fatal(err)
	}
}

func pingCouchContainer(attempt int) (bool, error) {
	_, err := http.Get(fmt.Sprintf("http://%s:%s", domain, port))
	if err != nil {
		time.Sleep(1 * time.Second)
		return true, err
	}
	return false, err
}
