package model

// TODO: Use github.com/docker/libcompose

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/docker/libcompose/config"
	"github.com/docker/libcompose/project"
	"github.com/dtan4/paus-gitreceive/receiver/util"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

const (
	portBindingRegexp = `"?\d+:(\d+)"?`
)

var (
	portBinding = regexp.MustCompile(portBindingRegexp)
)

type Compose struct {
	dockerHost      string
	composeFilePath string
	projectName     string
	project         *project.Project
}

type ComposeConfig struct {
	version  string                           `yaml:"version,omitempty"`
	services map[string]*config.ServiceConfig `yaml:"services,omitempty"`
	volumes  map[string]*config.VolumeConfig  `yaml:"volumes,omitempty"`
	networks map[string]*config.NetworkConfig `yaml:"networks,omitempty"`
}

func NewCompose(dockerHost, composeFilePath, projectName string) (*Compose, error) {
	prj := project.NewProject(&project.Context{
		ComposeFiles: []string{composeFilePath},
		ProjectName:  projectName,
	}, nil, nil)

	if err := prj.Parse(); err != nil {
		return nil, errors.Wrap(err, "Failed to parse docker-compose.yml.")
	}

	return &Compose{
		dockerHost,
		composeFilePath,
		projectName,
		prj,
	}, nil
}

func (c *Compose) Build() error {
	cmd := exec.Command("docker-compose", "-f", c.composeFilePath, "-p", c.projectName, "build")
	cmd.Env = append(os.Environ(), "DOCKER_HOST="+c.dockerHost)

	if err := util.RunCommand(cmd); err != nil {
		return errors.Wrap(err, fmt.Sprintf("Failed to build Docker Compose project. projectName: %s", c.projectName))
	}

	return nil
}

func (c *Compose) GetContainerID(service string) (string, error) {
	cmd := exec.Command("docker-compose", "-f", c.composeFilePath, "-p", c.projectName, "ps", "-q", service)
	cmd.Env = append(os.Environ(), "DOCKER_HOST="+c.dockerHost)
	out, err := cmd.Output()

	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("Failed to get container ID. projectName: %s, service: %s", c.projectName, service))
	}

	return strings.Replace(string(out), "\n", "", -1), nil
}

func (c *Compose) InjectBuildArgs(buildArgs map[string]string) {
	webService := c.webService()

	if webService == nil {
		return
	}

	for k, v := range buildArgs {
		webService.Build.Args[k] = v
	}
}

func (c *Compose) InjectEnvironmentVariables(envs map[string]string) {
	webService := c.webService()

	if webService == nil {
		return
	}

	for k, v := range envs {
		webService.Environment = append(webService.Environment, fmt.Sprintf("%s=\"%s\"", k, v))
	}
}

func (c *Compose) Pull() error {
	cmd := exec.Command("docker-compose", "-f", c.composeFilePath, "-p", c.projectName, "pull")
	cmd.Env = append(os.Environ(), "DOCKER_HOST="+c.dockerHost)

	if err := util.RunCommand(cmd); err != nil {
		return errors.Wrap(err, fmt.Sprintf("Failed to pull images in Docker Compose project. projectName: %s", c.projectName))
	}

	return nil
}

func (c *Compose) RewritePortBindings() {
	var newPorts []string

	for _, key := range c.project.ServiceConfigs.Keys() {
		if svc, ok := c.project.ServiceConfigs.Get(key); ok {
			if len(svc.Ports) == 0 {
				continue
			}

			newPorts = []string{}

			for _, port := range svc.Ports {
				matchResult := portBinding.FindStringSubmatch(port)

				if len(matchResult) == 2 {
					newPorts = append(newPorts, matchResult[1])
				} else {
					newPorts = append(newPorts, port)
				}
			}

			svc.Ports = newPorts
		}
	}
}

func (c *Compose) SaveAs(filePath string) error {
	fmt.Println(filePath)

	services := map[string]*config.ServiceConfig{}

	for _, key := range c.project.ServiceConfigs.Keys() {
		if svc, ok := c.project.ServiceConfigs.Get(key); ok {
			services[key] = svc
		}
	}

	cfg := &ComposeConfig{
		version:  "2",
		services: services,
		volumes:  c.project.VolumeConfigs,
		networks: c.project.NetworkConfigs,
	}

	data, err := yaml.Marshal(cfg)

	if err != nil {
		return errors.Wrap(err, "Failed to generate YAML file.")
	}

	if err = ioutil.WriteFile(filePath, data, 0644); err != nil {
		return errors.Wrap(err, fmt.Sprintf("Failed to save as YAML file. path: %s", filePath))
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

func (c *Compose) webService() *config.ServiceConfig {
	if svc, ok := c.project.ServiceConfigs.Get("web"); ok {
		return svc
	}

	return nil
}
