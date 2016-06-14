package model

import (
	"fmt"

	"github.com/fsouza/go-dockerclient"
	"github.com/pkg/errors"
)

type Container struct {
	ContainerId   string
	client        *docker.Client
	containerInfo *docker.Container
	exposedPort   docker.PortBinding
}

func firstExposedPort(ports map[docker.Port][]docker.PortBinding) docker.PortBinding {
	ary := []docker.PortBinding{}

	for _, p := range ports {
		for _, port := range p {
			ary = append(ary, port)
		}
	}

	return ary[0]
}

func ContainerFromID(dockerHost, containerId string) (*Container, error) {
	client, _ := docker.NewClient(dockerHost)
	containerInfo, err := client.InspectContainer(containerId)

	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Failed to get container info. containerID %s", containerId))
	}

	exposedPort := firstExposedPort(containerInfo.NetworkSettings.Ports)

	return &Container{containerId, client, containerInfo, exposedPort}, nil
}

func (c *Container) HostIP() string {
	return c.exposedPort.HostIP
}

func (c *Container) HostPort() string {
	return c.exposedPort.HostPort
}
