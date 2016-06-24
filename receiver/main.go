package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/dtan4/paus-gitreceive/receiver/config"
	"github.com/dtan4/paus-gitreceive/receiver/model"
	"github.com/dtan4/paus-gitreceive/receiver/store"
	"github.com/dtan4/paus-gitreceive/receiver/util"
	"github.com/dtan4/paus-gitreceive/receiver/vulcand"
)

func appDirExists(application *model.Application, etcd *store.Etcd) bool {
	return etcd.HasKey("/paus/users/" + application.Username + "/apps/" + application.AppName)
}

func deploy(application *model.Application, compose *model.Compose) (string, error) {
	var err error

	fmt.Println("=====> Building ...")

	if err = compose.Build(); err != nil {
		return "", err
	}

	fmt.Println("=====> Pulling ...")

	if err = compose.Pull(); err != nil {
		return "", err
	}

	fmt.Println("=====> Deploying ...")

	if err = compose.Up(); err != nil {
		return "", err
	}

	webContainerID, err := compose.GetContainerID("web")

	if err != nil {
		return "", err
	}

	return webContainerID, nil
}

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

func injectBuildArgs(application *model.Application, compose *model.Compose, etcd *store.Etcd) error {
	args, err := application.BuildArgs(etcd)

	if err != nil {
		return err
	}

	compose.InjectBuildArgs(args)

	return nil
}

func injectEnvironmentVariables(application *model.Application, compose *model.Compose, etcd *store.Etcd) error {
	envs, err := application.EnvironmentVariables(etcd)

	if err != nil {
		return err
	}

	compose.InjectEnvironmentVariables(envs)

	return nil
}

func prepareComposeFile(application *model.Application, compose *model.Compose, etcd *store.Etcd) (string, error) {
	if err := injectBuildArgs(application, compose, etcd); err != nil {
		return "", err
	}

	if err := injectEnvironmentVariables(application, compose, etcd); err != nil {
		return "", err
	}

	compose.RewritePortBindings()
	newComposeFilePath := filepath.Join(filepath.Dir(compose.ComposeFilePath), "docker-compose-"+strconv.FormatInt(time.Now().Unix(), 10)+".yml")

	if err := compose.SaveAs(newComposeFilePath); err != nil {
		return "", err
	}

	return newComposeFilePath, nil
}

func printDeployedURLs(repository string, config *config.Config, identifiers []string) {
	var url string

	fmt.Println("=====> " + repository + " was successfully deployed at:")

	for _, identifier := range identifiers {
		url = strings.ToLower(config.URIScheme + "://" + identifier + "." + config.BaseDomain)
		fmt.Println("         " + url)
	}
}

func main() {
	// printVersion()

	config, etcd, err := initialize()

	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}

	application, err := model.ApplicationFromArgs(os.Args[1:])

	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}

	if !appDirExists(application, etcd) {
		fmt.Fprintln(os.Stderr, "=====> Application not found: "+application.AppName)
		os.Exit(1)
	}

	repositoryPath, err := util.UnpackReceivedFiles(config.RepositoryDir, application.Username, application.ProjectName, os.Stdin)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}

	if err = os.Chdir(repositoryPath); err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}

	fmt.Println("=====> Getting submodules ...")

	if err = util.GetSubmodules(repositoryPath); err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}

	composeFilePath := filepath.Join(repositoryPath, "docker-compose.yml")

	if _, err := os.Stat(composeFilePath); err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}

	fmt.Println("=====> docker-compose.yml was found")

	compose, err := model.NewCompose(config.DockerHost, composeFilePath, application.ProjectName)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}

	newComposeFilePath, err := prepareComposeFile(application, compose, etcd)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}

	webContainerID, err := deploy(application, compose)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}

	fmt.Println("=====> Application container is launched.")

	webContainer, err := model.ContainerFromID(config.DockerHost, webContainerID)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}

	fmt.Println("=====> Registering metadata ...")

	if err = application.RegisterMetadata(etcd); err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}

	identifiers, err := vulcand.RegisterInformation(etcd, application, config.BaseDomain, webContainer)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}

	printDeployedURLs(application.Repository, config, identifiers)

	if err = util.RemoveUnpackedFiles(repositoryPath, newComposeFilePath); err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
}
