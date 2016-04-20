package main

import (
	"github.com/kelseyhightower/envconfig"
)

const (
	ConfigPrefix = "paus"
)

type Config struct {
	BaseDomain    string `envconfig:"base_domain"    required:"true"`
	DockerHost    string `envconfig:"docker_host"    default:"tcp://localhost:2375"`
	EtcdEndpoint  string `envconfig:"etcd_endpoint"  default:"http://localhost:2379"`
	RepositoryDir string `envconfig:"repository_dir" default:"/repos"`
}

func LoadConfig() (*Config, error) {
	var config Config

	err := envconfig.Process(ConfigPrefix, &config)

	if err != nil {
		return nil, err
	}

	return &config, nil
}
