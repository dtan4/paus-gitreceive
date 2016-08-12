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
	"github.com/docker/libcompose/lookup"
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
	ComposeFilePath string
	ProjectName     string

	dockerHost string
	project    *project.Project
}

type ComposeConfig struct {
	Version  string                           `yaml:"version,omitempty"`
	Services map[string]*config.ServiceConfig `yaml:"services,omitempty"`
	Volumes  map[string]*config.VolumeConfig  `yaml:"volumes,omitempty"`
	Networks map[string]*config.NetworkConfig `yaml:"networks,omitempty"`
}

func NewCompose(dockerHost, composeFilePath, projectName string) (*Compose, error) {
	ctx := project.Context{
		ComposeFiles: []string{composeFilePath},
		ProjectName:  projectName,
	}

	ctx.ResourceLookup = &lookup.FileResourceLookup{}
	ctx.EnvironmentLookup = &lookup.ComposableEnvLookup{
		Lookups: []config.EnvironmentLookup{
			&lookup.OsEnvLookup{},
		},
	}

	prj := project.NewProject(&ctx, nil, nil)

	if err := prj.Parse(); err != nil {
		return nil, errors.Wrap(err, "Failed to parse docker-compose.yml.")
	}

	return &Compose{
		ComposeFilePath: composeFilePath,
		ProjectName:     projectName,
		dockerHost:      dockerHost,
		project:         prj,
	}, nil
}

func (c *Compose) Build() error {
	cmd := exec.Command("docker-compose", "-f", c.ComposeFilePath, "-p", c.ProjectName, "build")
	cmd.Env = append(os.Environ(), "DOCKER_HOST="+c.dockerHost)

	if err := util.RunCommand(cmd); err != nil {
		return err
	}

	return nil
}

func (c *Compose) GetContainerID(service string) (string, error) {
	cmd := exec.Command("docker-compose", "-f", c.ComposeFilePath, "-p", c.ProjectName, "ps", "-q", service)
	cmd.Env = append(os.Environ(), "DOCKER_HOST="+c.dockerHost)
	out, err := cmd.Output()

	if err != nil {
		return "", errors.Wrapf(err, "Failed to get container ID. projectName: %s, service: %s", c.ProjectName, service)
	}

	return strings.Replace(string(out), "\n", "", -1), nil
}

func (c *Compose) InjectBuildArgs(buildArgs map[string]string) {
	webService := c.webService()

	if webService == nil {
		return
	}

	if webService.Build.Args == nil {
		webService.Build.Args = map[string]string{}
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

	envmap := make(map[string]string)

	for _, env := range webService.Environment {
		kv := strings.SplitN(env, "=", 2)

		if len(kv) == 2 {
			envmap[kv[0]] = kv[1]
		}
	}

	for k, v := range envs {
		envmap[k] = v
	}

	webService.Environment = []string{}

	for k, v := range envmap {
		webService.Environment = append(webService.Environment, fmt.Sprintf("%s=%s", k, v))
	}
}

func (c *Compose) Pull() error {
	cmd := exec.Command("docker-compose", "-f", c.ComposeFilePath, "-p", c.ProjectName, "pull")
	cmd.Env = append(os.Environ(), "DOCKER_HOST="+c.dockerHost)

	if err := util.RunCommand(cmd); err != nil {
		return err
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
	services := map[string]*config.ServiceConfig{}

	for _, key := range c.project.ServiceConfigs.Keys() {
		if svc, ok := c.project.ServiceConfigs.Get(key); ok {
			services[key] = svc
		}
	}

	cfg := &ComposeConfig{
		Version:  "2",
		Services: services,
		Volumes:  c.project.VolumeConfigs,
		Networks: c.project.NetworkConfigs,
	}

	data, err := yaml.Marshal(cfg)

	if err != nil {
		return errors.Wrap(err, "Failed to generate YAML file.")
	}

	if err = ioutil.WriteFile(filePath, data, 0644); err != nil {
		return errors.Wrapf(err, "Failed to save as YAML file. path: %s", filePath)
	}

	c.ComposeFilePath = filePath

	return nil
}

func (c *Compose) Stop() error {
	cmd := exec.Command("docker-compose", "-f", c.ComposeFilePath, "-p", c.ProjectName, "stop")
	cmd.Env = append(os.Environ(), "DOCKER_HOST="+c.dockerHost)

	if err := util.RunCommand(cmd); err != nil {
		return err
	}

	return nil
}

func (c *Compose) Up() error {
	cmd := exec.Command("docker-compose", "-f", c.ComposeFilePath, "-p", c.ProjectName, "up", "-d")
	cmd.Env = append(os.Environ(), "DOCKER_HOST="+c.dockerHost)

	if err := util.RunCommand(cmd); err != nil {
		return err
	}

	return nil
}

func (c *Compose) webService() *config.ServiceConfig {
	if svc, ok := c.project.ServiceConfigs.Get("web"); ok {
		return svc
	}

	return nil
}
