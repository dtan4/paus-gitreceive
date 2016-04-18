package main

import (
	"io/ioutil"

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

func (c *ComposeFile) isVersion2() bool {
	return c.Yaml["version"] != nil && c.Yaml["version"] == "2"
}

func (c *ComposeFile) webService() map[interface{}]interface{} {
	if c.isVersion2() {
		return c.Yaml["services"].(map[interface{}]interface{})["web"].(map[interface{}]interface{})
	}

	return c.Yaml["web"].(map[interface{}]interface{})
}

func (c *ComposeFile) InjectEnvironmentVariables(environmentVariables map[string]string) {
	var envString string

	webService := c.webService()
	environment := webService["environment"].([]interface{})

	for key, value := range environmentVariables {
		envString = key + "=" + value
		environment = append(environment, envString)
	}

	webService["environment"] = environment
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
