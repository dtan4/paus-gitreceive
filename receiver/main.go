package main

import (
	"fmt"
	"os"
	"os/exec"
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

	compose := NewCompose(dockerHost, composeFilePath, application.ProjectName)

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

	webContainerId, err := compose.GetContainerId("web")

	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("Failed to web container ID. appName: %s, composeFilePath: %s", application.AppName, composeFilePath))
	}

	return webContainerId, nil
}

func getSubmodules(repositoryPath string) error {
	dir := filepath.Join(repositoryPath, ".git")

	stat, err := os.Stat(dir)

	if err == nil && stat.IsDir() {
		if e := os.RemoveAll(dir); e != nil {
			return errors.Wrap(e, fmt.Sprintf("Failed to remove %s.", dir))
		}
	}

	cmd := exec.Command("/usr/local/bin/get-submodules")

	if err = RunCommand(cmd); err != nil {
		return errors.Wrap(err, "Failed to get submodules.")
	}

	return nil
}

func injectBuildArgs(application *model.Application, composeFile *ComposeFile, etcd *store.Etcd) error {
	userDirectoryKey := "/paus/users/" + application.Username

	if !etcd.HasKey(userDirectoryKey) {
		return nil
	}

	appDirectoryKey := userDirectoryKey + "/apps/" + application.AppName

	if !etcd.HasKey(appDirectoryKey) {
		return nil
	}

	buildArgsKey := appDirectoryKey + "/build-args/"

	if !etcd.HasKey(buildArgsKey) {
		return nil
	}

	buildArgKeys, err := etcd.List(buildArgsKey, false)

	if err != nil {
		return errors.Wrap(err, "Failed to get build arg keys.")
	}

	buildArgs := map[string]string{}

	for _, key := range buildArgKeys {
		value, err := etcd.Get(key)

		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Failed to get build arg value. key: %s", key))
		}

		buildArgs[strings.Replace(key, buildArgsKey, "", 1)] = value
	}

	composeFile.InjectBuildArgs(buildArgs)

	return nil
}

func injectEnvironmentVariables(application *model.Application, composeFile *ComposeFile, etcd *store.Etcd) error {
	userDirectoryKey := "/paus/users/" + application.Username

	if !etcd.HasKey(userDirectoryKey) {
		return nil
	}

	appDirectoryKey := userDirectoryKey + "/apps/" + application.AppName

	if !etcd.HasKey(appDirectoryKey) {
		return nil
	}

	envDirectoryKey := appDirectoryKey + "/envs/"

	if !etcd.HasKey(envDirectoryKey) {
		return nil
	}

	envKeys, err := etcd.List(envDirectoryKey, false)

	if err != nil {
		return errors.Wrap(err, "Failed to get environment variable keys.")
	}

	environmentVariables := map[string]string{}

	for _, key := range envKeys {
		value, err := etcd.Get(key)

		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Failed to get environment variable value. key: %s", key))
		}

		environmentVariables[strings.Replace(key, envDirectoryKey, "", 1)] = value
	}

	composeFile.InjectEnvironmentVariables(environmentVariables)

	return nil
}

func registerApplicationMetadata(application *model.Application, etcd *store.Etcd) error {
	userDirectoryKey := "/paus/users/" + application.Username

	if !etcd.HasKey(userDirectoryKey) {
		_ = etcd.Mkdir(userDirectoryKey)
	}

	appDirectoryKey := userDirectoryKey + "/apps/" + application.AppName

	if !etcd.HasKey(appDirectoryKey) {
		_ = etcd.Mkdir(appDirectoryKey)
		_ = etcd.Mkdir(appDirectoryKey + "/envs")
		_ = etcd.Mkdir(appDirectoryKey + "/revisions")
	}

	if err := etcd.Set(appDirectoryKey+"/revisions/"+application.Revision, strconv.FormatInt(time.Now().Unix(), 10)); err != nil {
		return errors.Wrap(err, "Failed to set revisdion.")
	}

	return nil
}

func main() {
	printVersion()

	config, err := config.LoadConfig()

	if err != nil {
		errors.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	etcd, err := store.NewEtcd(config.EtcdEndpoint)

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

	if err = getSubmodules(repositoryPath); err != nil {
		errors.Fprint(os.Stderr, errors.Wrap(err, fmt.Sprintf("Failed to get submodules. path: %s", repositoryPath)))
		os.Exit(1)
	}

	composeFilePath := filepath.Join(repositoryPath, "docker-compose.yml")

	if _, err := os.Stat(composeFilePath); err != nil {
		fmt.Fprintln(os.Stderr, "=====> docker-compose.yml was NOT found!")
		os.Exit(1)
	}

	fmt.Println("=====> docker-compose.yml was found")
	composeFile, err := NewComposeFile(composeFilePath)

	if err != nil {
		errors.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	if err = injectBuildArgs(application, composeFile, etcd); err != nil {
		errors.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	if err = injectEnvironmentVariables(application, composeFile, etcd); err != nil {
		errors.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	composeFile.RewritePortBindings()
	newComposeFilePath := filepath.Join(repositoryPath, "docker-compose-"+strconv.FormatInt(time.Now().Unix(), 10)+".yml")

	if err = composeFile.SaveAs(newComposeFilePath); err != nil {
		errors.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	webContainerId, err := deploy(config.DockerHost, application, newComposeFilePath)

	if err != nil {
		errors.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println("=====> Application container is launched.")

	webContainer, err := model.ContainerFromID(config.DockerHost, webContainerId)

	if err != nil {
		errors.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println("=====> Registering metadata ...")

	if err = registerApplicationMetadata(application, etcd); err != nil {
		errors.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	identifiers, err := vulcand.RegisterInformation(etcd, application, config.BaseDomain, webContainer)

	if err != nil {
		errors.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println("=====> " + application.Repository + " was successfully deployed at:")

	var url string

	for _, identifier := range identifiers {
		url = strings.ToLower(config.URIScheme + "://" + identifier + "." + config.BaseDomain)
		fmt.Println("         " + url)
	}

	if err = util.RemoveUnpackedFiles(repositoryPath, newComposeFilePath); err != nil {
		errors.Fprint(os.Stderr, err)
		os.Exit(1)
	}
}
