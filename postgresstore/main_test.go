// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by the license that can be found in the
// LICENSE file.

package postgresstore

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/stratumn/sdk/utils"
)

func TestMain(m *testing.M) {
	const (
		domain = "0.0.0.0"
		port   = "5432"
	)

	seed := int64(time.Now().Nanosecond())
	fmt.Printf("using seed %d\n", seed)
	rand.Seed(seed)
	flag.Parse()

	// Postgres container configuration.
	imageName := "postgres:latest"
	containerName := "indigo_postgresstore_test"
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

	// Start postgres container
	if err := utils.RunContainer(containerName, imageName, exposedPorts, portBindings); err != nil {
		fmt.Printf(err.Error())
		os.Exit(1)
	}
	// Retry until container is ready.
	if err := utils.Retry(createDatabase, 10); err != nil {
		fmt.Printf(err.Error())
		os.Exit(1)
	}

	// Run tests.
	testResult := m.Run()

	// Stop postgres container.
	if err := utils.KillContainer(containerName); err != nil {
		fmt.Printf(err.Error())
		os.Exit(1)
	}

	os.Exit(testResult)

}

func createDatabase(attempt int) (bool, error) {
	cmd := exec.Command("psql", "-h", "localhost", "-c", "create database goprivate_test;", "-U", "postgres")
	err := cmd.Run()
	if err != nil {
		time.Sleep(1 * time.Second)
		return true, err
	}
	return false, err
}
