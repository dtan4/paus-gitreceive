package main

import (
	"strings"

	"github.com/dtan4/paus-gitreceive/receiver/config"
	"github.com/dtan4/paus-gitreceive/receiver/model"
	"github.com/dtan4/paus-gitreceive/receiver/msg"
	"github.com/dtan4/paus-gitreceive/receiver/store"
	"github.com/dtan4/paus-gitreceive/receiver/util"
	"github.com/dtan4/paus-gitreceive/receiver/vulcand"
)

func deploy(application *model.Application, compose *model.Compose, deployment *model.Deployment) (string, error) {
	var err error

	msg.PrintTitle("Building ...")

	images, err := compose.Build(deployment)

	if err != nil {
		return "", err
	}

	msg.PrintTitle("Pushing ...")

	if err = compose.Push(images); err != nil {
		return "", err
	}

	msg.PrintTitle("Pulling ...")

	if err = compose.Pull(); err != nil {
		return "", err
	}

	msg.PrintTitle("Deploying ...")

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

	msg.PrintTitle(repository + " was successfully deployed at:")

	for _, identifier := range identifiers {
		url = strings.ToLower(config.URIScheme + "://" + identifier + "." + config.BaseDomain)
		msg.Println("         " + url)
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

	msg.PrintTitle("Max deploy limit reached.")

	oldestTimestamp := util.SortKeys(deployments)[0]
	oldestDeployment := model.NewDeployment(application, "", deployments[oldestTimestamp], oldestTimestamp, repositoryDir)

	msg.PrintTitle("Stop " + oldestDeployment.Revision + " ...")

	// TODO: set registryDomain
	compose, err := model.NewCompose(dockerHost, oldestDeployment.ComposeFilePath, oldestDeployment.ProjectName, "")

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
