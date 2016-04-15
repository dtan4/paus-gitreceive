package main

import (
	"github.com/fsouza/go-dockerclient"
)

const (
	WebContainerPort = "8080/tcp"
)

type Container struct {
	client        *docker.Client
	containerInfo *docker.Container
}

func NewContainer(dockerHost, containerId string) (*Container, error) {
	client, _ := docker.NewClient(dockerHost)
	containerInfo, err := client.InspectContainer(containerId)

	if err != nil {
		return nil, err
	}

	return &Container{client, containerInfo}, nil
}

func (c *Container) HostIP() string {
	return c.containerInfo.NetworkSettings.Ports[WebContainerPort][0].HostIP
}

func (c *Container) HostPort() string {
	return c.containerInfo.NetworkSettings.Ports[WebContainerPort][0].HostPort
}
