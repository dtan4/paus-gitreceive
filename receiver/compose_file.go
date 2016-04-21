package main

import (
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

const (
	PortBindingRegexp = `"?\d+:(\d+)"?`
)

var (
	PortBinding = regexp.MustCompile(PortBindingRegexp)
)

type ComposeFile struct {
	filePath string
	Yaml     map[interface{}]interface{}
}

func NewComposeFile(composeFilePath string) (*ComposeFile, error) {
	buf, err := ioutil.ReadFile(composeFilePath)

	if err != nil {
		return nil, err
	}

	m := make(map[interface{}]interface{})
	err = yaml.Unmarshal(buf, &m)

	if err != nil {
		return nil, err
	}

	return &ComposeFile{composeFilePath, m}, nil
}

func (c *ComposeFile) environment(service map[interface{}]interface{}) []interface{} {
	return service["environment"].([]interface{})
}

func (c *ComposeFile) environmentMap(environment []interface{}) map[string]string {
	result := map[string]string{}

	for _, envString := range environment {
		splited := strings.Split(envString.(string), "=")
		key, value := splited[0], strings.Join(splited[1:], "=")
		result[key] = value
	}

	return result
}

func (c *ComposeFile) isVersion2() bool {
	return c.Yaml["version"] != nil && c.Yaml["version"] == "2"
}

func (c *ComposeFile) service(serviceName string) map[interface{}]interface{} {
	if c.isVersion2() {
		return c.Yaml["services"].(map[interface{}]interface{})[serviceName].(map[interface{}]interface{})
	}

	return c.Yaml[serviceName].(map[interface{}]interface{})
}

func (c *ComposeFile) serviceList() []string {
	var (
		services     map[interface{}]interface{}
		serviceNames []string
	)

	if c.isVersion2() {
		services = c.Yaml["services"].(map[interface{}]interface{})
	} else {
		services = c.Yaml
	}

	for name, _ := range services {
		serviceNames = append(serviceNames, name.(string))
	}

	return serviceNames
}

func (c *ComposeFile) InjectEnvironmentVariables(environmentVariables map[string]string) {
	var envString string

	webService := c.service("web")
	environment := c.environment(webService)
	environmentMap := c.environmentMap(environment)

	for key, value := range environmentVariables {
		environmentMap[key] = value
	}

	newEnvironment := []interface{}{}

	for key, value := range environmentMap {
		envString = key + "=" + value
		newEnvironment = append(newEnvironment, envString)
	}

	webService["environment"] = newEnvironment
}

func (c *ComposeFile) RewritePortBindings() {
	var portString string

	for _, serviceName := range c.serviceList() {
		service := c.service(serviceName)
		newPorts := []interface{}{}

		if service["ports"] == nil {
			continue
		}

		for _, port := range service["ports"].([]interface{}) {
			switch p := port.(type) {
			case int:
				portString = strconv.Itoa(p)
			case string:
				portString = p
			}

			matchResult := PortBinding.FindStringSubmatch(portString)

			if len(matchResult) == 2 {
				newPorts = append(newPorts, matchResult[1])
			} else {
				newPorts = append(newPorts, portString)
			}
		}

		service["ports"] = newPorts
	}
}

func (c *ComposeFile) SaveAs(filePath string) error {
	data, err := yaml.Marshal(c.Yaml)

	if err != nil {
		return err
	}

	if err = ioutil.WriteFile(filePath, data, 0644); err != nil {
		return err
	}

	return nil
}
