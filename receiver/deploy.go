package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dtan4/paus-gitreceive/receiver/aws"
	"github.com/dtan4/paus-gitreceive/receiver/config"
	"github.com/dtan4/paus-gitreceive/receiver/model"
	"github.com/dtan4/paus-gitreceive/receiver/msg"
	"github.com/dtan4/paus-gitreceive/receiver/service"
	"github.com/dtan4/paus-gitreceive/receiver/store"
	"github.com/dtan4/paus-gitreceive/receiver/util"
	"github.com/dtan4/paus-gitreceive/receiver/vulcand"
)

func deploy(application *model.Application, deployment *model.Deployment, config *config.Config, etcd *store.Etcd) error {
	repositoryPath, err := util.UnpackReceivedFiles(config.RepositoryDir, application.Username, deployment.ProjectName, os.Stdin)
	if err != nil {
		return err
	}

	if err = os.Chdir(repositoryPath); err != nil {
		return err
	}

	msg.PrintTitle("Getting submodules...")

	if err = util.GetSubmodules(repositoryPath); err != nil {
		return err
	}

	composeFilePath := filepath.Join(repositoryPath, "docker-compose.yml")

	if _, err := os.Stat(composeFilePath); err != nil {
		return err
	}

	msg.PrintTitle("docker-compose.yml was found")

	// TODO: rotateDeployments

	compose, err := model.NewCompose(config.DockerHost, composeFilePath, deployment.ProjectName, config.AWSRegion)
	if err != nil {
		return err
	}

	if err := prepareComposeFile(application, deployment, compose); err != nil {
		return err
	}

	msg.PrintTitle("Building images...")

	images, err := compose.Build(deployment)
	if err != nil {
		return err
	}

	for _, image := range images {
		msg.Println("Build completed: " + image.String())
	}

	msg.PrintTitle("Pushing images...")

	if err = compose.Push(images); err != nil {
		return err
	}

	msg.PrintTitle("Rewrite compose yml to use built images...")

	compose.UpdateImages(images)

	msg.PrintTitle("Convert to TaskDefinition...")

	serviceName := application.ServiceName(util.Timestamp())

	taskDefinition, err := compose.TransformToTaskDefinition(application.TaskDefinitionName(), serviceName, config.AWSRegion)
	if err != nil {
		return err
	}

	msg.PrintTitle("Registering TaskDefinition...")

	td, err := service.RegisterTaskDefinition(taskDefinition)
	if err != nil {
		return err
	}

	msg.Println("TaskDefinition: " + *td.TaskDefinitionArn)

	msg.PrintTitle("Creating Log Group...")

	if err := service.CreateLogGroup(serviceName); err != nil {
		return err
	}

	msg.PrintTitle("Creating service ...")

	svc, err := service.CreateService(serviceName, config.ClusterName, *td.TaskDefinitionArn)
	if err != nil {
		return err
	}

	msg.Println("Service: " + *svc.ServiceArn)

	msg.PrintTitle("Wait for service becomes ACTIVE ...")

	if err := service.WaitUntilServicesStable(svc); err != nil {
		return err
	}

	webContainer, err := service.GetWebContainer(svc)
	if err != nil {
		return err
	}

	instanceID, err := service.GetRunningInstance(svc)
	if err != nil {
		return err
	}

	instance, err := aws.EC2().GetInstance(instanceID)
	if err != nil {
		return err
	}

	publicIP := *instance.PublicIpAddress
	port := *webContainer.NetworkBindings[0].HostPort

	serviceAddress := fmt.Sprintf("%s:%d", publicIP, port)

	msg.PrintTitle("Application container is launched!")

	msg.PrintTitle("Registering metadata...")

	identifiers, err := vulcand.RegisterInformation(etcd, deployment, config.BaseDomain, serviceAddress)
	if err != nil {
		return err
	}

	printDeployedURLs(application.Repository, config, identifiers)

	if err = util.RemoveUnpackedFiles(repositoryPath, deployment.ComposeFilePath); err != nil {
		return err
	}

	return nil
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

	msg.PrintTitle(repository + " was successfully deployed!")

	for _, identifier := range identifiers {
		url = strings.ToLower(config.URIScheme + "://" + identifier + "." + config.BaseDomain)
		msg.Println("  " + url)
	}
}
