package main

// TODO: Use github.com/docker/libcompose

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

type Compose struct {
	dockerHost      string
	composeFilePath string
	projectName     string
}

func printLine(r io.Reader) {
	sc := bufio.NewScanner(r)

	for sc.Scan() {
		fmt.Println(sc.Text())
	}
}

func runCommand(command *exec.Cmd) error {
	stdout, err := command.StdoutPipe()

	if err != nil {
		return err
	}

	stderr, err := command.StderrPipe()

	if err != nil {
		return err
	}

	command.Start()

	go printLine(stdout)
	go printLine(stderr)

	if err = command.Wait(); err != nil {
		return err
	}

	return nil
}

func NewCompose(dockerHost, composeFilePath, projectName string) *Compose {
	return &Compose{
		dockerHost,
		composeFilePath,
		projectName,
	}
}

func (c *Compose) Build() error {
	cmd := exec.Command("docker-compose", "-f", c.composeFilePath, "-p", c.projectName, "--x-networking", "--x-network-driver", "overlay", "build")
	cmd.Env = append(os.Environ(), "DOCKER_HOST="+c.dockerHost)

	if err := runCommand(cmd); err != nil {
		return err
	}

	return nil
}

func (c *Compose) GetContainerId(service string) (string, error) {
	cmd := exec.Command("docker-compose", "-f", c.composeFilePath, "-p", c.projectName, "--x-networking", "--x-network-driver", "overlay", "ps", "-q", service)
	cmd.Env = append(os.Environ(), "DOCKER_HOST="+c.dockerHost)
	out, err := cmd.Output()

	if err != nil {
		return "", err
	}

	return strings.Replace(string(out), "\n", "", -1), nil
}

func (c *Compose) Pull() error {
	cmd := exec.Command("docker-compose", "-f", c.composeFilePath, "-p", c.projectName, "--x-networking", "--x-network-driver", "overlay", "pull")
	cmd.Env = append(os.Environ(), "DOCKER_HOST="+c.dockerHost)

	if err := runCommand(cmd); err != nil {
		return err
	}

	return nil
}

func (c *Compose) Up() error {
	cmd := exec.Command("docker-compose", "-f", c.composeFilePath, "-p", c.projectName, "--x-networking", "--x-network-driver", "overlay", "up", "-d")
	cmd.Env = append(os.Environ(), "DOCKER_HOST="+c.dockerHost)

	if err := runCommand(cmd); err != nil {
		return err
	}

	return nil
}
