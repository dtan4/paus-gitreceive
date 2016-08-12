package config

import (
	"bufio"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

const (
	configPrefix   = "paus"
	configFilePath = "/paus/config"
)

var (
	configNames = []string{
		"BaseDomain",
		"DockerHost",
		"EtcdEndpoint",
		"MaxAppDeploy",
		"RegistryDomain",
		"RepositoryDir",
		"URIScheme",
	}
)

type Config struct {
	BaseDomain     string `envconfig:"base_domain"`
	DockerHost     string `envconfig:"docker_host"    default:"tcp://localhost:2375"`
	EtcdEndpoint   string `envconfig:"etcd_endpoint"  default:"http://localhost:2379"`
	MaxAppDeploy   int64  `envconfig:"max_app_deploy" default:"10"`
	RegistryDomain string `envconfig:"registry_domain" default:""`
	RepositoryDir  string `envconfig:"repository_dir" default:"/repos"`
	URIScheme      string `envconfig:"uri_scheme"     default:"http"`
}

func loadConfigFromFile(filePath string) (map[string]string, error) {
	config := map[string]string{}

	fp, err := os.Open(filePath)

	if err != nil {
		return nil, errors.Wrapf(err, "Failed to open %s.", filePath)
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

	err := envconfig.Process(configPrefix, &config)

	if err != nil {
		return nil, errors.Wrap(err, "Failed to load config from envs.")
	}

	if _, err := os.Stat(configFilePath); err != nil {
		return &config, nil
	}

	configFromFile, err := loadConfigFromFile(configFilePath)

	if err != nil {
		return nil, err
	}

	for _, configName := range configNames {
		if configName == "MaxAppDeploy" {
			n, err := strconv.ParseInt(configFromFile[configName], 10, 64)

			if err != nil {
				return nil, errors.Wrapf(err, "Failed to parse %s as integer. value: %s", configName, configFromFile[configName])
			}

			reflect.ValueOf(&config).Elem().FieldByName(configName).SetInt(n)
		} else {
			reflect.ValueOf(&config).Elem().FieldByName(configName).SetString(configFromFile[configName])
		}
	}

	return &config, nil
}
