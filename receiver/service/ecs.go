package service

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
)

var (
	svc = ecs.New(session.New(), &aws.Config{})
)

// RegisterTaskDefinition creates new taskDefinition
func RegisterTaskDefinition(taskDefinition *ecs.TaskDefinition) error {
	_, err := svc.RegisterTaskDefinition(&ecs.RegisterTaskDefinitionInput{
		Family:               taskDefinition.Family,
		ContainerDefinitions: taskDefinition.ContainerDefinitions,
		Volumes:              taskDefinition.Volumes,
	})

	if err != nil {
		return err
	}

	return nil
}
