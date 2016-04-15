package main

import (
	"fmt"
	"os"
)

const (
	DefaultDockerHost   = "tcp://localhost:2375"
	DefaultEtcdEndpoint = "http://localhost:2379"
)

func main() {
	baseDir, _ := os.Getwd()
	dockerHost := os.Getenv("DOCKER_HOST")
	etcdEndpoint := os.Getenv("ETCD_ENDPOINT")

	if dockerHost == "" {
		dockerHost = DefaultDockerHost
	}

	if etcdEndpoint == "" {
		etcdEndpoint = DefaultEtcdEndpoint
	}

	commitMetadata := NewCommitMetadataFromArgs(os.Args[1:])

	fmt.Println(baseDir)
	fmt.Println(commitMetadata.Repository)
	fmt.Println(commitMetadata.Revision)
	fmt.Println(commitMetadata.Username)
	fmt.Println(commitMetadata.AppName)
	fmt.Println(commitMetadata.ProjectName)
}
