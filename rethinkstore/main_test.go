// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by the license that can be found in the
// LICENSE file.

package rethinkstore

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/stratumn/sdk/utils"
)

var (
	test *testing.T
)

const (
	domain = "0.0.0.0"
	port   = "28015"
	dbName = "test"
)

func TestMain(m *testing.M) {
	seed := int64(time.Now().Nanosecond())
	fmt.Printf("using seed %d\n", seed)
	rand.Seed(seed)
	flag.Parse()

	// Rethinkdb container configuration.
	imageName := "rethinkdb:latest"
	containerName := "indigo_rethinkstore_test"
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
