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

package utils

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	docker "docker.io/go-docker"
	types "docker.io/go-docker/api/types"
	"docker.io/go-docker/api/types/container"
	"docker.io/go-docker/api/types/network"
	"github.com/docker/go-connections/nat"
)

// PullImage pulls an image from docker hub if image is not available or not up to date.
func PullImage(imageName string) error {
	cli, err := docker.NewEnvClient()
	if err != nil {
		return err
	}

	out, err := cli.ImagePull(context.Background(), imageName, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(os.Stdout, out)
	return err
}

// KillContainer stops and removes specified container.
func KillContainer(containerName string) error {
	cli, err := docker.NewEnvClient()
	if err != nil {
		return err
	}

	timeout := 1 * time.Millisecond

	if err := cli.ContainerStop(context.Background(), containerName, &timeout); err != nil {
		return err
	}

	return cli.ContainerRemove(context.Background(), containerName, types.ContainerRemoveOptions{})
}

// RunContainerWithEnv reproduces docker run command.
func RunContainerWithEnv(containerName string, imageName string, envVariables []string, exposedPorts nat.PortSet, portBindings nat.PortMap) error {
	cli, err := docker.NewEnvClient()
	if err != nil {
		return err
	}

	err = PullImage(imageName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot pull docker image %s: %s\n", imageName, err)
	}

	_, err = cli.ContainerCreate(
		context.Background(),
		&container.Config{
			Image:        imageName,
			ExposedPorts: exposedPorts,
			Env:          envVariables,
		},
		&container.HostConfig{
			PortBindings: portBindings,
		},
		&network.NetworkingConfig{},
		containerName,
	)
	if err != nil {
		return err
	}

	return cli.ContainerStart(
		context.Background(),
		containerName,
		types.ContainerStartOptions{},
	)
}

// RunContainer reproduces docker run command.
func RunContainer(containerName string, imageName string, exposedPorts nat.PortSet, portBindings nat.PortMap) error {
	return RunContainerWithEnv(containerName, imageName, []string{}, exposedPorts, portBindings)
}
