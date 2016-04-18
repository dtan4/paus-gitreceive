package main

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type ComposeFile struct {
	filePath string
	yaml     map[interface{}]interface{}
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

func (c *ComposeFile) IsVersion2() bool {
	return c.yaml["version"] != nil && c.yaml["version"] == "2"
}
