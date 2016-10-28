package config

import (
	"os"

	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

const (
	configPrefix = "paus"
)

type Config struct {
	AWSRegion     string
	BaseDomain    string `envconfig:"base_domain"`
	ClusterName   string `envconfig:"cluster_name"`
	DockerHost    string `envconfig:"docker_host"    default:"tcp://localhost:2375"`
	EtcdEndpoint  string `envconfig:"etcd_endpoint"  default:"http://localhost:2379"`
	MaxAppDeploy  int64  `envconfig:"max_app_deploy" default:"10"`
	RepositoryDir string `envconfig:"repository_dir" default:"/repos"`
	URIScheme     string `envconfig:"uri_scheme"     default:"http"`
}

// LoadConfig loads config values from environment variables
func LoadConfig() (*Config, error) {
	var config Config

	err := envconfig.Process(configPrefix, &config)

	if err != nil {
		return nil, errors.Wrap(err, "Failed to load config from envs.")
	}

	config.AWSRegion = os.Getenv("AWS_REGION")

	return &config, nil
}
