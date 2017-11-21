package utils

import (
	"context"
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

// RunContainer reproduces docker run command.
func RunContainer(containerName string, imageName string, exposedPorts nat.PortSet, portBindings nat.PortMap) error {
	cli, err := docker.NewEnvClient()
	if err != nil {
		return err
	}

	err = PullImage(imageName)
	if err != nil {
		return err
	}

	_, err = cli.ContainerCreate(
		context.Background(),
		&container.Config{
			Image:        imageName,
			ExposedPorts: exposedPorts,
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
