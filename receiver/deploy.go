package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/dtan4/paus-gitreceive/receiver/config"
	"github.com/dtan4/paus-gitreceive/receiver/model"
)

func deploy(application *model.Application, compose *model.Compose, maxAppDeploy int64) (string, error) {
	var err error

	revisions, err := application.Revisions()

	if err != nil {
		return "", err
	}

	if int64(len(revisions)) >= maxAppDeploy {
		fmt.Println("Rate Limit Exceeded!!!!!!!!!!!!")
	}

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

func prepareComposeFile(application *model.Application, compose *model.Compose, timestamp string) (string, error) {
	if err := injectBuildArgs(application, compose); err != nil {
		return "", err
	}

	if err := injectEnvironmentVariables(application, compose); err != nil {
		return "", err
	}

	compose.RewritePortBindings()
	newComposeFilePath := filepath.Join(filepath.Dir(compose.ComposeFilePath), "docker-compose-"+timestamp+".yml")

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
