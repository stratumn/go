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

package elasticsearchstore

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
	test *testing.T
)

const (
	domain = "0.0.0.0"
	port   = "9200"
)

func TestMain(m *testing.M) {
	flag.Parse()
	// ElasticSearch container configuration.
	imageName := "docker.elastic.co/elasticsearch/elasticsearch:6.2.1"
	containerName := "sdk_elasticsearchstore_test"
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

	// Start elasticsearch container.
	env := []string{"discovery.type=single-node"}
	if err := utils.RunContainerWithEnv(containerName, imageName, env, exposedPorts, portBindings); err != nil {
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
	}, 60); err != nil {
		fmt.Printf(err.Error())
		os.Exit(1)
	}

	// Run tests.
	testResult := m.Run()

	// Stop elasticsearch container.
	if err := utils.KillContainer(containerName); err != nil {
		fmt.Printf(err.Error())
		os.Exit(1)
	}

	os.Exit(testResult)
}

func TestElasticSearchStore(t *testing.T) {
	test = t
	factory := storetestcases.Factory{
		New:               newTestElasticSearchStoreAdapter,
		NewKeyValueStore:  newTestElasticSearchStoreKeyValue,
		Free:              freeTestElasticSearchStoreAdapter,
		FreeKeyValueStore: freeTestElasticSearchStoreKeyValue,
	}

	factory.RunStoreTests(t)
	factory.RunKeyValueStoreTests(t)
}

func TestElasticSearchTMPop(t *testing.T) {
	tmpoptestcases.Factory{
		New:  newTestElasticSearchStoreTMPop,
		Free: freeTestElasticSearchStoreTMPop,
	}.RunTests(t)
}

func newTestElasticSearchStore() (*ESStore, error) {
	config := &Config{
		URL: fmt.Sprintf("http://%s:%s", domain, port),
	}
	return New(config)
}

func newTestElasticSearchStoreAdapter() (store.Adapter, error) {
	return newTestElasticSearchStore()
}

func newTestElasticSearchStoreKeyValue() (store.KeyValueStore, error) {
	return newTestElasticSearchStore()
}

func newTestElasticSearchStoreTMPop() (store.Adapter, store.KeyValueStore, error) {
	a, err := newTestElasticSearchStore()
	return a, a, err
}

func freeTestElasticSearchStore(a *ESStore) {
	if err := a.deleteIndex(linksIndex); err != nil {
		test.Fatal(err)
	}
	if err := a.deleteIndex(evidencesIndex); err != nil {
		test.Fatal(err)
	}
	if err := a.deleteIndex(valuesIndex); err != nil {
		test.Fatal(err)
	}
}

func freeTestElasticSearchStoreAdapter(a store.Adapter) {
	freeTestElasticSearchStore(a.(*ESStore))
}

func freeTestElasticSearchStoreKeyValue(a store.KeyValueStore) {
	freeTestElasticSearchStore(a.(*ESStore))
}

func freeTestElasticSearchStoreTMPop(a store.Adapter, _ store.KeyValueStore) {
	freeTestElasticSearchStoreAdapter(a)
}
