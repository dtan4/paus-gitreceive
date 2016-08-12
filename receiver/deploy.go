package main

import (
	"fmt"
	"strings"

	"github.com/dtan4/paus-gitreceive/receiver/config"
	"github.com/dtan4/paus-gitreceive/receiver/model"
	"github.com/dtan4/paus-gitreceive/receiver/store"
	"github.com/dtan4/paus-gitreceive/receiver/util"
	"github.com/dtan4/paus-gitreceive/receiver/vulcand"
)

func deploy(application *model.Application, compose *model.Compose, dockerHost, registryDomain string, deployment *model.Deployment) (string, error) {
	var err error

	fmt.Println("=====> Building ...")

	if err = compose.Build(dockerHost, registryDomain, deployment); err != nil {
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

func injectBuildArgs(application *model.Application, compose *model.Compose) error {
	args, err := application.BuildArgs()

	if err != nil {
		return err
	}

	compose.InjectBuildArgs(args)

	return nil
}

func injectEnvironmentVariables(application *model.Application, compose *model.Compose) error {
	envs, err := application.EnvironmentVariables()

	if err != nil {
		return err
	}

	compose.InjectEnvironmentVariables(envs)

	return nil
}

func prepareComposeFile(application *model.Application, deployment *model.Deployment, compose *model.Compose) error {
	if err := injectBuildArgs(application, compose); err != nil {
		return err
	}

	if err := injectEnvironmentVariables(application, compose); err != nil {
		return err
	}

	compose.RewritePortBindings()

	if err := compose.SaveAs(deployment.ComposeFilePath); err != nil {
		return err
	}

	return nil
}

func printDeployedURLs(repository string, config *config.Config, identifiers []string) {
	var url string

	fmt.Println("=====> " + repository + " was successfully deployed at:")

	for _, identifier := range identifiers {
		url = strings.ToLower(config.URIScheme + "://" + identifier + "." + config.BaseDomain)
		fmt.Println("         " + url)
	}
}

func rotateDeployments(etcd *store.Etcd, application *model.Application, maxAppDeploy int64, dockerHost string, repositoryDir string) error {
	deployments, err := application.Deployments()

	if err != nil {
		return err
	}

	if len(deployments) == 0 || int64(len(deployments)) < maxAppDeploy {
		return nil
	}

	fmt.Println("=====> Max deploy limit reached.")

	oldestTimestamp := util.SortKeys(deployments)[0]
	oldestDeployment := model.NewDeployment(application, "", deployments[oldestTimestamp], oldestTimestamp, repositoryDir)

	fmt.Println("=====> Stop " + oldestDeployment.Revision + " ...")

	compose, err := model.NewCompose(dockerHost, oldestDeployment.ComposeFilePath, oldestDeployment.ProjectName)

	if err != nil {
		return err
	}

	if err := compose.Stop(); err != nil {
		return err
	}

	if err := application.DeleteDeployment(oldestTimestamp); err != nil {
		return err
	}

	if err := vulcand.DeregisterInformation(etcd, oldestDeployment); err != nil {
		return err
	}

	return nil
}
