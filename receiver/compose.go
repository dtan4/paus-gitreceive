package main

// TODO: Use github.com/docker/libcompose

import (
	"bufio"
	"fmt"
	"os/exec"
)

type Compose struct {
	composeFilePath string
	projectName     string
}

func runCommand(command *exec.Cmd) error {
	stdout, err := command.StdoutPipe()

	if err != nil {
		return err
	}

	command.Start()
	scanner := bufio.NewScanner(stdout)

	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}

	command.Wait()

	return nil
}

func NewCompose(composeFilePath, projectName string) *Compose {
	return &Compose{
		composeFilePath,
		projectName,
	}
}

func (c *Compose) Build() error {
	cmd := exec.Command("docker-compose", "-p", c.projectName, "build")

	if err := runCommand(cmd); err != nil {
		return err
	}

	return nil
}

func (c *Compose) Pull() error {
	cmd := exec.Command("docker-compose", "-p", c.projectName, "pull")

	if err := runCommand(cmd); err != nil {
		return err
	}

	return nil
}

func (c *Compose) Up() error {
	cmd := exec.Command("docker-compose", "-p", c.projectName, "up", "-d")

	if err := runCommand(cmd); err != nil {
		return err
	}

	return nil
}
