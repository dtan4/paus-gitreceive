package main

import (
	"os"
	"path/filepath"

	"github.com/dtan4/paus-gitreceive/receiver/config"
	"github.com/dtan4/paus-gitreceive/receiver/model"
	"github.com/dtan4/paus-gitreceive/receiver/msg"
	"github.com/dtan4/paus-gitreceive/receiver/store"
	"github.com/dtan4/paus-gitreceive/receiver/util"
	"github.com/dtan4/paus-gitreceive/receiver/vulcand"
)

func initialize() (*config.Config, *store.Etcd, error) {
	config, err := config.LoadConfig()
	if err != nil {
		return nil, nil, err
	}

	etcd, err := store.NewEtcd(config.EtcdEndpoint)
	if err != nil {
		return nil, nil, err
	}

	return config, etcd, nil
}

func main() {
	if len(os.Args) > 1 {
		if os.Args[1] == "-v" || os.Args[1] == "--version" {
			printVersion()
			os.Exit(0)
		}
	}

	config, etcd, err := initialize()
	if err != nil {
		printErrorAndExit(err)
	}

	application, err := model.ApplicationFromArgs(os.Args[1:])
	if err != nil {
		printErrorAndExit(err)
	}

	deployment, err := model.DeploymentFromArgs(application, os.Args[1:], "", config.RepositoryDir)
	if err != nil {
		printErrorAndExit(err)
	}

	repositoryPath, err := util.UnpackReceivedFiles(config.RepositoryDir, application.Username, deployment.ProjectName, os.Stdin)
	if err != nil {
		printErrorAndExit(err)
	}

	if err = os.Chdir(repositoryPath); err != nil {
		printErrorAndExit(err)
	}

	msg.PrintTitle("Getting submodules...")

	if err = util.GetSubmodules(repositoryPath); err != nil {
		printErrorAndExit(err)
	}

	composeFilePath := filepath.Join(repositoryPath, "docker-compose.yml")

	if _, err := os.Stat(composeFilePath); err != nil {
		printErrorAndExit(err)
	}

	msg.PrintTitle("docker-compose.yml was found")

	// TODO: rotateDeployments

	compose, err := model.NewCompose(config.DockerHost, composeFilePath, deployment.ProjectName, config.AWSRegion)
	if err != nil {
		printErrorAndExit(err)
	}

	if err := prepareComposeFile(application, deployment, compose); err != nil {
		printErrorAndExit(err)
	}

	serviceAddress, err := deploy(application, compose, deployment, config.ClusterName, config.AWSRegion)
	if err != nil {
		printErrorAndExit(err)
	}

	msg.PrintTitle("Application container is launched!")

	msg.PrintTitle("Registering metadata...")

	identifiers, err := vulcand.RegisterInformation(etcd, deployment, config.BaseDomain, serviceAddress)
	if err != nil {
		printErrorAndExit(err)
	}

	printDeployedURLs(application.Repository, config, identifiers)

	if err = util.RemoveUnpackedFiles(repositoryPath, deployment.ComposeFilePath); err != nil {
		printErrorAndExit(err)
	}
}

func printErrorAndExit(err error) {
	msg.PrintErrorf("%+v\n", err)
	os.Exit(1)
}
