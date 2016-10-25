package utils

import (
	"fmt"
	"strconv"
	"strings"

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
	// memory

	// environment variables

	// port mappings
	portMappings, err := convertToPortMappings(svc.Ports)
	if err != nil {
		return nil, err
	}

	// volumes from

	// mount points

	// extra hosts

	// logs

	// ulimits

	// popular container definition

	return &ecs.ContainerDefinition{
		Name:         aws.String(name),
		PortMappings: portMappings,
	}, nil
}

func convertToPortMappings(ports []string) ([]*ecs.PortMapping, error) {
	portMappings := []*ecs.PortMapping{}

	for _, portMapping := range ports {
		protocol := ecs.TransportProtocolTcp
		tcp := strings.HasSuffix(portMapping, "/"+ecs.TransportProtocolTcp)
		udp := strings.HasSuffix(portMapping, "/"+ecs.TransportProtocolUdp)
		if tcp || udp {
			protocol = portMapping[len(portMapping)-3:]
			portMapping = portMapping[0 : len(portMapping)-4]
		}

		// either has 1 part (just the containerPort) or has 2 parts (hostPort:containerPort)
		parts := strings.Split(portMapping, ":")
		var containerPort, hostPort int
		var portErr error
		switch len(parts) {
		case 1: // Format "containerPort" Example "8000"
			containerPort, portErr = strconv.Atoi(parts[0])
		case 2: // Format "hostPort:containerPort" Example "8000:8000"
			hostPort, portErr = strconv.Atoi(parts[0])
			containerPort, portErr = strconv.Atoi(parts[1])
		case 3: // Format "ipAddr:hostPort:containerPort" Example "127.0.0.0.1:8000:8000"
			hostPort, portErr = strconv.Atoi(parts[1])
			containerPort, portErr = strconv.Atoi(parts[2])
		default:
			return nil, fmt.Errorf(
				"expected format [hostPort]:containerPort. Could not parse portmappings: %s", portMapping)
		}
		if portErr != nil {
			return nil, fmt.Errorf("Could not convert port into integer in portmappings: %v", portErr)
		}

		portMappings = append(portMappings, &ecs.PortMapping{
			Protocol:      aws.String(protocol),
			ContainerPort: aws.Int64(int64(containerPort)),
			HostPort:      aws.Int64(int64(hostPort)),
		})
	}

	return portMappings, nil
}
