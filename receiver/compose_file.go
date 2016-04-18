package main

import (
	"io/ioutil"
	"strings"

	"gopkg.in/yaml.v2"
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
