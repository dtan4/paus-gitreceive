package main

import (
	"fmt"
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
		msg.PrintErrorf("%+v\n", err)
		os.Exit(1)
	}

	application, err := model.ApplicationFromArgs(os.Args[1:], etcd)

	if err != nil {
		msg.PrintErrorf("%+v\n", err)
		os.Exit(1)
	}

	if !application.DirExists() {
		msg.PrintError("=====> Application not found: " + application.AppName)
		os.Exit(1)
	}

	deployment, err := model.DeploymentFromArgs(application, os.Args[1:], "", config.RepositoryDir)

	if err != nil {
		msg.PrintErrorf("%+v\n", err)
		os.Exit(1)
	}

	repositoryPath, err := util.UnpackReceivedFiles(config.RepositoryDir, application.Username, deployment.ProjectName, os.Stdin)

	if err != nil {
		msg.PrintErrorf("%+v\n", err)
		os.Exit(1)
	}

	if err = os.Chdir(repositoryPath); err != nil {
		msg.PrintErrorf("%+v\n", err)
		os.Exit(1)
	}

	msg.PrintTitle("Getting submodules ...")

	if err = util.GetSubmodules(repositoryPath); err != nil {
		msg.PrintErrorf("%+v\n", err)
		os.Exit(1)
	}

	composeFilePath := filepath.Join(repositoryPath, "docker-compose.yml")

	if _, err := os.Stat(composeFilePath); err != nil {
		msg.PrintErrorf("%+v\n", err)
		os.Exit(1)
	}

	msg.PrintTitle("docker-compose.yml was found")

	if err := rotateDeployments(etcd, application, config.MaxAppDeploy, config.DockerHost, config.RepositoryDir); err != nil {
		msg.PrintErrorf("%+v\n", err)
		os.Exit(1)
	}

	compose, err := model.NewCompose(config.DockerHost, composeFilePath, deployment.ProjectName, config.AWSRegion)

	if err != nil {
		msg.PrintErrorf("%+v\n", err)
		os.Exit(1)
	}

	if err := prepareComposeFile(application, deployment, compose); err != nil {
		msg.PrintErrorf("%+v\n", err)
		os.Exit(1)
	}

	webContainerID, err := deploy(application, compose, deployment, config.ClusterName)

	if err != nil {
		msg.PrintErrorf("%+v\n", err)
		os.Exit(1)
	}

	msg.PrintTitle("Application container is launched.")

	webContainer, err := model.ContainerFromID(config.DockerHost, webContainerID)

	if err != nil {
		msg.PrintErrorf("%+v\n", err)
		os.Exit(1)
	}

	path, interval, maxTry, err := application.HealthCheck()

	if err != nil {
		msg.PrintErrorf("%+v\n", err)
		os.Exit(1)
	}

	callback := func(path string, try int) {
		msg.Println(fmt.Sprintf("      Ping to %s (%d times) ...", path, try))
	}

	msg.PrintTitle("Start healthcheck ...")

	if !webContainer.ExecuteHealthCheck(path, interval, maxTry, callback) {
		msg.PrintError("=====> Web container is not active. Aborted.")
		compose.Stop()
		os.Exit(1)
	}

	msg.PrintTitle("Registering metadata ...")

	deployment.Timestamp = util.Timestamp()

	if err = deployment.Register(); err != nil {
		msg.PrintErrorf("%+v\n", err)
		os.Exit(1)
	}

	identifiers, err := vulcand.RegisterInformation(etcd, deployment, config.BaseDomain, webContainer)

	if err != nil {
		msg.PrintErrorf("%+v\n", err)
		os.Exit(1)
	}

	printDeployedURLs(application.Repository, config, identifiers)

	if err = util.RemoveUnpackedFiles(repositoryPath, deployment.ComposeFilePath); err != nil {
		msg.PrintErrorf("%+v\n", err)
		os.Exit(1)
	}
}
