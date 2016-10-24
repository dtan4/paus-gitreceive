package model

// TODO: Use github.com/docker/libcompose

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/docker/libcompose/config"
	"github.com/docker/libcompose/lookup"
	"github.com/docker/libcompose/project"
	"github.com/dtan4/paus-gitreceive/receiver/msg"
	"github.com/dtan4/paus-gitreceive/receiver/service"
	"github.com/dtan4/paus-gitreceive/receiver/util"
	"github.com/fsouza/go-dockerclient"
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
	RegistryDomain  string

	dockerHost string
	project    *project.Project
}

type ComposeConfig struct {
	Version  string                           `yaml:"version,omitempty"`
	Services map[string]*config.ServiceConfig `yaml:"services,omitempty"`
	Volumes  map[string]*config.VolumeConfig  `yaml:"volumes,omitempty"`
	Networks map[string]*config.NetworkConfig `yaml:"networks,omitempty"`
}

func NewCompose(dockerHost, composeFilePath, projectName, awsRegion string) (*Compose, error) {
	var registryDomain string

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

	if awsRegion != "" {
		accountID, err := service.GetAWSAccountID()
		if err != nil {
			return nil, err
		}

		registryDomain = service.GetRegistryDomain(accountID, awsRegion)
	}

	return &Compose{
		ComposeFilePath: composeFilePath,
		ProjectName:     projectName,
		RegistryDomain:  registryDomain,
		dockerHost:      dockerHost,
		project:         prj,
	}, nil
}

func (c *Compose) Build(deployment *Deployment) ([]*Image, error) {
	var (
		buildArgs []docker.BuildArg
		opts      docker.BuildImageOptions
		svc       *config.ServiceConfig
		image     *Image
	)

	reader, outputBuf := io.Pipe()
	go func() {
		sc := bufio.NewScanner(reader)

		for sc.Scan() {
			msg.Println(sc.Text())
		}
	}()

	client, _ := docker.NewClient(c.dockerHost)
	images := []*Image{}

	for _, name := range c.project.ServiceConfigs.Keys() {
		svc, _ = c.project.ServiceConfigs.Get(name)

		if svc.Build.Context == "" {
			continue
		}

		for k, v := range svc.Build.Args {
			buildArgs = append(buildArgs, docker.BuildArg{Name: k, Value: v})
		}

		if svc.Image == "" {
			n := deployment.App.Username + "-" + deployment.App.AppName + "-" + name
			image = NewImage(c.RegistryDomain, n, deployment.Revision)
		} else {
			var err error

			image, err = ImageFromString(svc.Image)
			if err != nil {
				return []*Image{}, err
			}
		}

		opts = docker.BuildImageOptions{
			BuildArgs:      buildArgs,
			ContextDir:     svc.Build.Context,
			Dockerfile:     svc.Build.Dockerfile,
			Name:           image.String(),
			OutputStream:   outputBuf,
			Pull:           true,
			SuppressOutput: false,
		}

		if err := client.BuildImage(opts); err != nil {
			return []*Image{}, errors.Wrapf(err, "Failed to build image. service: %s", name)
		}

		images = append(images, image)
	}

	return images, nil
}

func convertToContainerDef(name string, svc *config.ServiceConfig) (*ecs.ContainerDefinition, error) {
	containerDef := &ecs.ContainerDefinition{
		Name: aws.String(name),
	}

	// memory

	// environment variables

	// volumes from

	// mount points

	// extra hosts

	// logs

	// ulimits

	// popular container definition

	return containerDef, nil
}

// ConvertToTaskDefinition converts docker-compose.yml to ECS TaskDefinition
func (c *Compose) ConvertToTaskDefinition() (*ecs.TaskDefinition, error) {
	containerDefinitions := []*ecs.ContainerDefinition{}

	for _, name := range c.project.ServiceConfigs.Keys() {
		svc, _ := c.project.ServiceConfigs.Get(name)

		containerDef, err := convertToContainerDef(name, svc)
		if err != nil {
			return nil, err
		}

		containerDefinitions = append(containerDefinitions, containerDef)
	}

	taskDefinition := &ecs.TaskDefinition{
		ContainerDefinitions: containerDefinitions,
	}

	return taskDefinition, nil
}

func (c *Compose) Push(images []*Image) error {
	var (
		opts docker.PushImageOptions
	)

	client, _ := docker.NewClient(c.dockerHost)

	accountID, err := service.GetAWSAccountID()
	if err != nil {
		return err
	}

	// default registry ID is equivalent to AWS account ID
	authConf, err := service.GetECRAuthConf(accountID)
	if err != nil {
		return err
	}

	for _, image := range images {
		if !service.RepositoryExists(image.Registry, image.Name) {
			if err := service.CreateRepository(image.Registry, image.Name); err != nil {
				return err
			}
		}

		opts = docker.PushImageOptions{
			Registry: image.Registry,
			Name:     image.Registry + "/" + image.Name,
			Tag:      image.Tag,
		}

		if err := client.PushImage(opts, authConf); err != nil {
			return err
		}
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
