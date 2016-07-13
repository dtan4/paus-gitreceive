package model

import (
	"fmt"
	"net/http"
	"time"

	"github.com/fsouza/go-dockerclient"
	"github.com/pkg/errors"
)

type Container struct {
	ContainerId   string
	client        *docker.Client
	containerInfo *docker.Container
	exposedPort   docker.PortBinding
}

type HealthCheckFunc func(path string, try int)

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
		return nil, errors.Wrapf(err, "Failed to get container info. containerID %s", containerId)
	}

	exposedPort := firstExposedPort(containerInfo.NetworkSettings.Ports)

	return &Container{containerId, client, containerInfo, exposedPort}, nil
}

func (c *Container) ExecuteHealthCheck(path string, interval, maxTry int, callback HealthCheckFunc) bool {
	url := fmt.Sprintf("http://%s:%s%s", c.HostIP(), c.HostPort(), path)

	for i := 1; i <= maxTry; i++ {
		callback(path, i)

		resp, err := http.Get(url)
		if err == nil && resp.StatusCode == http.StatusOK {
			return true
		}

		time.Sleep(time.Duration(interval) * time.Second)
	}

	return false
}

func (c *Container) HostIP() string {
	return c.exposedPort.HostIP
}

func (c *Container) HostPort() string {
	return c.exposedPort.HostPort
}
