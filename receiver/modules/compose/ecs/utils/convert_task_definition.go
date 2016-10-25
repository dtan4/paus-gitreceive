package utils

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/docker/libcompose/config"
	"github.com/docker/libcompose/project"
)

// ConvertToTaskDefinition converts docker-compose.yml to ECS TaskDefinition
func ConvertToTaskDefinition(prj *project.Project) (*ecs.TaskDefinition, error) {
	containerDefinitions := []*ecs.ContainerDefinition{}

	for _, name := range prj.ServiceConfigs.Keys() {
		svc, _ := prj.ServiceConfigs.Get(name)

		containerDef, err := convertToContainerDef(name, svc)
		if err != nil {
			return nil, err
		}

		containerDefinitions = append(containerDefinitions, containerDef)
	}

	taskDefinition := &ecs.TaskDefinition{
		ContainerDefinitions: containerDefinitions,
	}

	return taskDefinition, nil
}

func convertToContainerDef(name string, svc *config.ServiceConfig) (*ecs.ContainerDefinition, error) {
	containerDef := &ecs.ContainerDefinition{
		Name: aws.String(name),
	}

	// memory

	// environment variables

	// volumes from

	// mount points

	// extra hosts

	// logs

	// ulimits

	// popular container definition

	return containerDef, nil
}
