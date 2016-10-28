package ecs

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
)

type ECSClient struct {
	client *ecs.ECS
}

// NewECSClient creates new ECSClient object
func NewECSClient() *ECSClient {
	return &ECSClient{
		client: ecs.New(session.New(), &aws.Config{}),
	}
}

// CreateService creates new service and return ID of Deployment
func (c *ECSClient) CreateService(serviceName, clusterName, taskDefinitionArn string) (*ecs.Service, error) {
	resp, err := c.client.CreateService(&ecs.CreateServiceInput{
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

// GetRunningInstance returns instance ID where the service task runs
func (c *ECSClient) GetRunningInstance(service *ecs.Service) (string, error) {
	taskArns, err := c.listTasks(service)
	if err != nil {
		return "", err
	}
	taskArn := taskArns[0]

	tasks, err := c.describeTasks(service, *taskArn)
	if err != nil {
		return "", err
	}
	containerInstanceArn := tasks[0].ContainerInstanceArn

	containerInstances, err := c.describeContainerInstances(service, *containerInstanceArn)
	if err != nil {
		return "", err
	}

	return *containerInstances[0].Ec2InstanceId, nil
}

// GetWebContainer returns web container
func (c *ECSClient) GetWebContainer(service *ecs.Service) (*ecs.Container, error) {
	taskArns, err := c.listTasks(service)
	if err != nil {
		return nil, err
	}
	taskArn := taskArns[0]

	tasks, err := c.describeTasks(service, *taskArn)
	if err != nil {
		return nil, err
	}

	for _, container := range tasks[0].Containers {
		if *container.Name == "web" {
			return container, nil
		}
	}

	return nil, fmt.Errorf("Web container not found!")
}

func (c *ECSClient) describeContainerInstances(service *ecs.Service, containerInstanceArn string) ([]*ecs.ContainerInstance, error) {
	resp, err := c.client.DescribeContainerInstances(&ecs.DescribeContainerInstancesInput{
		Cluster:            service.ClusterArn,
		ContainerInstances: aws.StringSlice([]string{containerInstanceArn}),
	})
	if err != nil {
		return []*ecs.ContainerInstance{}, err
	}

	return resp.ContainerInstances, nil
}

func (c *ECSClient) describeTasks(service *ecs.Service, taskArn string) ([]*ecs.Task, error) {
	resp, err := c.client.DescribeTasks(&ecs.DescribeTasksInput{
		Cluster: service.ClusterArn,
		Tasks:   aws.StringSlice([]string{taskArn}),
	})
	if err != nil {
		return []*ecs.Task{}, err
	}

	return resp.Tasks, nil
}

func (c *ECSClient) listTasks(service *ecs.Service) ([]*string, error) {
	resp, err := c.client.ListTasks(&ecs.ListTasksInput{
		Cluster:     service.ClusterArn,
		ServiceName: service.ServiceArn,
	})
	if err != nil {
		return []*string{}, err
	}

	return resp.TaskArns, nil
}

// RegisterTaskDefinition creates new TaskDefinition and return ARN of it
func (c *ECSClient) RegisterTaskDefinition(taskDefinition *ecs.TaskDefinition) (*ecs.TaskDefinition, error) {
	resp, err := c.client.RegisterTaskDefinition(&ecs.RegisterTaskDefinitionInput{
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
func (c *ECSClient) WaitUntilServicesStable(service *ecs.Service) error {
	if err := c.client.WaitUntilServicesStable(&ecs.DescribeServicesInput{
		Cluster:  service.ClusterArn,
		Services: aws.StringSlice([]string{*service.ServiceArn}),
	}); err != nil {
		return err
	}

	return nil
}
