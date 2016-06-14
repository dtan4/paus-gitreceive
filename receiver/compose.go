package main

// TODO: Use github.com/docker/libcompose

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/dtan4/paus-gitreceive/receiver/util"
	"github.com/pkg/errors"
)

type Compose struct {
	dockerHost      string
	composeFilePath string
	projectName     string
}

func NewCompose(dockerHost, composeFilePath, projectName string) *Compose {
	return &Compose{
		dockerHost,
		composeFilePath,
		projectName,
	}
}

func (c *Compose) Build() error {
	cmd := exec.Command("docker-compose", "-f", c.composeFilePath, "-p", c.projectName, "build")
	cmd.Env = append(os.Environ(), "DOCKER_HOST="+c.dockerHost)

	if err := util.RunCommand(cmd); err != nil {
		return errors.Wrap(err, fmt.Sprintf("Failed to build Docker Compose project. projectName: %s", c.projectName))
	}

	return nil
}

func (c *Compose) GetContainerId(service string) (string, error) {
	cmd := exec.Command("docker-compose", "-f", c.composeFilePath, "-p", c.projectName, "ps", "-q", service)
	cmd.Env = append(os.Environ(), "DOCKER_HOST="+c.dockerHost)
	out, err := cmd.Output()

	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("Failed to get container ID. projectName: %s, service: %s", c.projectName, service))
	}

	return strings.Replace(string(out), "\n", "", -1), nil
}

func (c *Compose) Pull() error {
	cmd := exec.Command("docker-compose", "-f", c.composeFilePath, "-p", c.projectName, "pull")
	cmd.Env = append(os.Environ(), "DOCKER_HOST="+c.dockerHost)

	if err := util.RunCommand(cmd); err != nil {
		return errors.Wrap(err, fmt.Sprintf("Failed to pull images in Docker Compose project. projectName: %s", c.projectName))
	}

	return nil
}

func (c *Compose) Up() error {
	cmd := exec.Command("docker-compose", "-f", c.composeFilePath, "-p", c.projectName, "up", "-d")
	cmd.Env = append(os.Environ(), "DOCKER_HOST="+c.dockerHost)

	if err := util.RunCommand(cmd); err != nil {
		return errors.Wrap(err, fmt.Sprintf("Failed to start Docker Compose project. projectName: %s", c.projectName))
	}

	return nil
}
