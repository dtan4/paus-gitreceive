package main

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

const (
	ConfigPrefix   = "paus"
	ConfigFilePath = "/paus/config"
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

func loadConfigFromFile(filePath string) (map[string]string, error) {
	config := map[string]string{}

	fp, err := os.Open(filePath)

	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Failed to open %s.", filePath))
	}

	defer fp.Close()

	scanner := bufio.NewScanner(fp)

	for scanner.Scan() {
		line := scanner.Text()
		keyValue := strings.Split(line, "=")

		if len(keyValue) < 2 {
			continue
		}

		key, value := keyValue[0], strings.Join(keyValue[1:], "=")
		config[key] = value
	}

	return config, nil
}

func LoadConfig() (*Config, error) {
	var config Config

	err := envconfig.Process(ConfigPrefix, &config)

	if err != nil {
		return nil, errors.Wrap(err, "Failed to load config from envs.")
	}

	if _, err := os.Stat(ConfigFilePath); err != nil {
		return &config, nil
	}

	configFromFile, err := loadConfigFromFile(ConfigFilePath)

	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Failed to load config from file. path: %s", ConfigFilePath))
	}

	for _, configName := range ConfigNames {
		reflect.ValueOf(&config).Elem().FieldByName(configName).SetString(configFromFile[configName])
	}

	return &config, nil
}
