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
	"github.com/pkg/errors"
)

func appDirExists(application *model.Application, etcd *store.Etcd) bool {
	return etcd.HasKey("/paus/users/" + application.Username + "/apps/" + application.AppName)
}

func deploy(dockerHost string, application *model.Application, composeFilePath string) (string, error) {
	var err error

	compose, err := model.NewCompose(dockerHost, composeFilePath, application.ProjectName)

	if err != nil {
		return "", errors.Wrap(err, "Failed to initialize Compose object")
	}

	fmt.Println("=====> Building ...")

	if err = compose.Build(); err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("Failed to build application image. appName: %s, composeFilePath: %s", application.AppName, composeFilePath))
	}

	fmt.Println("=====> Pulling ...")

	if err = compose.Pull(); err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("Failed to pull application image. appName: %s, composeFilePath: %s", application.AppName, composeFilePath))
	}

	fmt.Println("=====> Deploying ...")

	if err = compose.Up(); err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("Failed to start application. appName: %s, composeFilePath: %s", application.AppName, composeFilePath))
	}

	webContainerID, err := compose.GetContainerID("web")

	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("Failed to web container ID. appName: %s, composeFilePath: %s", application.AppName, composeFilePath))
	}

	return webContainerID, nil
}

func initialize() (*config.Config, *store.Etcd, error) {
	config, err := config.LoadConfig()

	if err != nil {
		return nil, nil, errors.Wrap(err, "Failed to load config.")
	}

	etcd, err := store.NewEtcd(config.EtcdEndpoint)

	if err != nil {
		return nil, nil, errors.Wrap(err, "Failed to initialize etcd store.")
	}

	return config, etcd, nil
}

func injectBuildArgs(application *model.Application, composeFile *model.ComposeFile, etcd *store.Etcd) error {
	args, err := application.BuildArgs(etcd)

	if err != nil {
		return errors.Wrap(err, "Failed to get environment build args.")
	}

	composeFile.InjectBuildArgs(args)

	return nil
}

func injectEnvironmentVariables(application *model.Application, composeFile *model.ComposeFile, etcd *store.Etcd) error {
	envs, err := application.EnvironmentVariables(etcd)

	if err != nil {
		return errors.Wrap(err, "Failed to get environment variables.")
	}

	composeFile.InjectEnvironmentVariables(envs)

	return nil
}

func prepareComposeFile(repositoryPath string, application *model.Application, etcd *store.Etcd) (string, error) {
	composeFilePath := filepath.Join(repositoryPath, "docker-compose.yml")

	if _, err := os.Stat(composeFilePath); err != nil {
		return "", errors.Wrap(err, "docker-compose.yml was not found! path: "+composeFilePath)
	}

	fmt.Println("=====> docker-compose.yml was found")
	composeFile, err := model.NewComposeFile(composeFilePath)

	if err != nil {
		return "", errors.Wrap(err, "Failed to load docker-compose.yml")
	}

	if err = injectBuildArgs(application, composeFile, etcd); err != nil {
		return "", errors.Wrap(err, "Failed to inject build args.")
	}

	if err = injectEnvironmentVariables(application, composeFile, etcd); err != nil {
		return "", errors.Wrap(err, "Failed to inject environment variables.")
	}

	composeFile.RewritePortBindings()
	newComposeFilePath := filepath.Join(repositoryPath, "docker-compose-"+strconv.FormatInt(time.Now().Unix(), 10)+".yml")

	if err = composeFile.SaveAs(newComposeFilePath); err != nil {
		return "", errors.Wrap(err, "Failed to save docker-compose.yml. path: "+newComposeFilePath)
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
	printVersion()

	config, etcd, err := initialize()

	if err != nil {
		errors.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	application := model.ApplicationFromArgs(os.Args[1:])

	if !appDirExists(application, etcd) {
		fmt.Fprintln(os.Stderr, "=====> Application not found: "+application.AppName)
		os.Exit(1)
	}

	repositoryPath, err := util.UnpackReceivedFiles(config.RepositoryDir, application.Username, application.ProjectName, os.Stdin)

	if err != nil {
		errors.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	if err = os.Chdir(repositoryPath); err != nil {
		errors.Fprint(os.Stderr, errors.Wrap(err, fmt.Sprintf("Failed to chdir to %s.", repositoryPath)))
		os.Exit(1)
	}

	fmt.Println("=====> Getting submodules ...")

	if err = util.GetSubmodules(repositoryPath); err != nil {
		errors.Fprint(os.Stderr, errors.Wrap(err, fmt.Sprintf("Failed to get submodules. path: %s", repositoryPath)))
		os.Exit(1)
	}

	newComposeFilePath, err := prepareComposeFile(repositoryPath, application, etcd)

	if err != nil {
		errors.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	webContainerID, err := deploy(config.DockerHost, application, newComposeFilePath)

	if err != nil {
		errors.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println("=====> Application container is launched.")

	webContainer, err := model.ContainerFromID(config.DockerHost, webContainerID)

	if err != nil {
		errors.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println("=====> Registering metadata ...")

	if err = application.RegisterMetadata(etcd); err != nil {
		errors.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	identifiers, err := vulcand.RegisterInformation(etcd, application, config.BaseDomain, webContainer)

	if err != nil {
		errors.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	printDeployedURLs(application.Repository, config, identifiers)

	if err = util.RemoveUnpackedFiles(repositoryPath, newComposeFilePath); err != nil {
		errors.Fprint(os.Stderr, err)
		os.Exit(1)
	}
}
