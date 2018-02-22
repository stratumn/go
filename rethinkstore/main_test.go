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

package rethinkstore

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/stratumn/go-indigocore/utils"
)

var (
	test *testing.T
)

const (
	domain           = "0.0.0.0"
	port             = "28015"
	adminPort        = "8080"
	adminExposedPort = "18080"
	dbName           = "test"
)

func TestMain(m *testing.M) {
	seed := int64(time.Now().Nanosecond())
	fmt.Printf("using seed %d\n", seed)
	rand.Seed(seed)
	flag.Parse()

	// Rethinkdb container configuration.
	imageName := "rethinkdb:2.3"
	containerName := "indigo_rethinkstore_test"
	p, _ := nat.NewPort("tcp", port)
	pa, _ := nat.NewPort("tcp", adminPort)
	exposedPorts := map[nat.Port]struct{}{p: {}, pa: {}}
	portBindings := nat.PortMap{
		p: []nat.PortBinding{
			{
				HostIP:   domain,
				HostPort: port,
			},
		},
		pa: []nat.PortBinding{
			{
				HostIP:   domain,
				HostPort: adminExposedPort,
			},
		},
	}

	// Stop container if it is already running, swallow error.
	utils.KillContainer(containerName)

	// Start rethinkdb container
	if err := utils.RunContainer(containerName, imageName, exposedPorts, portBindings); err != nil {
		fmt.Printf(err.Error())
		os.Exit(1)
	}
	// Retry until container is ready.
	if err := utils.Retry(pingRethinkContainer, 10); err != nil {
		fmt.Printf(err.Error())
		os.Exit(1)
	}

	// Run tests.
	testResult := m.Run()

	// Stop rethinkdb container.
	if err := utils.KillContainer(containerName); err != nil {
		fmt.Printf(err.Error())
		os.Exit(1)
	}

	os.Exit(testResult)

}

func pingRethinkContainer(attempt int) (bool, error) {
	_, err := New(&Config{URL: fmt.Sprintf("%s:%s", domain, port), DB: dbName})
	if err != nil {
		time.Sleep(1 * time.Second)
		return true, err
	}
	return false, err
}
