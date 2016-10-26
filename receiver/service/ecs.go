package service

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
)

var (
	svc = ecs.New(session.New(), &aws.Config{})
)

// CreateService creates new service and return ID of Deployment
func CreateService(serviceName, clusterName, taskDefinitionArn string) (*ecs.Service, error) {
	resp, err := svc.CreateService(&ecs.CreateServiceInput{
		DesiredCount:   aws.Int64(1),
		ServiceName:    aws.String(serviceName),
		TaskDefinition: aws.String(taskDefinitionArn),
		Cluster:        aws.String(clusterName),
	})

	if err != nil {
		return nil, err
	}

	return resp.Service, nil
}

// RegisterTaskDefinition creates new TaskDefinition and return ARN of it
func RegisterTaskDefinition(taskDefinition *ecs.TaskDefinition) (*ecs.TaskDefinition, error) {
	resp, err := svc.RegisterTaskDefinition(&ecs.RegisterTaskDefinitionInput{
		Family:               taskDefinition.Family,
		ContainerDefinitions: taskDefinition.ContainerDefinitions,
		Volumes:              taskDefinition.Volumes,
	})

	if err != nil {
		return nil, err
	}

	return resp.TaskDefinition, nil
}

// WaitUntilServicesStable waits services become active.
func WaitUntilServicesStable(service *ecs.Service) error {
	if err := svc.WaitUntilServicesStable(&ecs.DescribeServicesInput{
		Cluster:  service.ClusterArn,
		Services: aws.StringSlice([]string{*service.ServiceArn}),
	}); err != nil {
		return err
	}

	return nil
}
