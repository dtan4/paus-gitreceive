package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"

	"github.com/kelseyhightower/envconfig"
)

const (
	ConfigPrefix    = "paus"
	ConfigDirectory = "/root/paus/config"
)

var (
	ConfigNames = []string{
		"BaseDomain",
		"DockerHost",
		"EtcdEndpoint",
		"RepositoryDir",
	}
)

type Config struct {
	BaseDomain    string `envconfig:"base_domain"`
	DockerHost    string `envconfig:"docker_host"    default:"tcp://localhost:2375"`
	EtcdEndpoint  string `envconfig:"etcd_endpoint"  default:"http://localhost:2379"`
	RepositoryDir string `envconfig:"repository_dir" default:"/repos"`
}

func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)

	return err == nil
}

func loadConfigFromFile(filePath string) (string, error) {
	buf, err := ioutil.ReadFile(filePath)

	if err != nil {
		return "", err
	}

	return string(buf), nil
}

func LoadConfig() (*Config, error) {
	var config Config

	err := envconfig.Process(ConfigPrefix, &config)

	if err != nil {
		return nil, err
	}

	for _, configName := range ConfigNames {
		filePath := filepath.Join(ConfigDirectory, configName)

		if !fileExists(filePath) {
			continue
		}

		configValue, err := loadConfigFromFile(filePath)

		if err != nil {
			return nil, err
		}

		reflect.ValueOf(&config).Elem().FieldByName(configName).SetString(configValue)
	}

	return &config, nil
}
