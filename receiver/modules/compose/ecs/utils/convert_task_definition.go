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

const (
	readOnlyVolumeAccessMode  = "ro"
	readWriteVolumeAccessMode = "rw"
	volumeFromContainerKey    = "container"
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
	volumesFrom, err := convertToVolumesFrom(svc.VolumesFrom)
	if err != nil {
		return nil, err
	}

	// mount points

	// extra hosts
	extraHosts, err := convertToExtraHosts(svc.ExtraHosts)
	if err != nil {
		return nil, err
	}

	// logs

	// ulimits

	// popular container definition

	return &ecs.ContainerDefinition{
		Name:         aws.String(name),
		PortMappings: portMappings,
		VolumesFrom:  volumesFrom,
		ExtraHosts:   extraHosts,
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

func convertToVolumesFrom(cfgVolumesFrom []string) ([]*ecs.VolumeFrom, error) {
	volumesFrom := []*ecs.VolumeFrom{}

	for _, cfgVolumeFrom := range cfgVolumesFrom {
		parts := strings.Split(cfgVolumeFrom, ":")

		var containerName, accessModeStr string

		parseErr := fmt.Errorf(
			"expected format [container:]SERVICE|CONTAINER[:ro|rw]. could not parse cfgVolumeFrom: %s", cfgVolumeFrom)

		switch len(parts) {
		// for the following volumes_from formats (supported by compose file formats v1 and v2),
		// name: refers to either service_name or container_name
		// container: is a keyword thats introduced in v2 to differentiate between service_name and container:container_name
		// ro|rw: read-only or read-write access
		case 1: // Format: name
			containerName = parts[0]
		case 2: // Format: name:ro|rw (OR) container:name
			if parts[0] == volumeFromContainerKey {
				containerName = parts[1]
			} else {
				containerName = parts[0]
				accessModeStr = parts[1]
			}
		case 3: // Format: container:name:ro|rw
			if parts[0] != volumeFromContainerKey {
				return nil, parseErr
			}
			containerName = parts[1]
			accessModeStr = parts[2]
		default:
			return nil, parseErr
		}

		// parse accessModeStr
		var readOnly bool
		if accessModeStr != "" {
			if accessModeStr == readOnlyVolumeAccessMode {
				readOnly = true
			} else if accessModeStr == readWriteVolumeAccessMode {
				readOnly = false
			} else {
				return nil, fmt.Errorf("Could not parse access mode %s", accessModeStr)
			}
		}
		volumesFrom = append(volumesFrom, &ecs.VolumeFrom{
			SourceContainer: aws.String(containerName),
			ReadOnly:        aws.Bool(readOnly),
		})
	}
	return volumesFrom, nil
}

func convertToExtraHosts(cfgExtraHosts []string) ([]*ecs.HostEntry, error) {
	extraHosts := []*ecs.HostEntry{}
	for _, cfgExtraHost := range cfgExtraHosts {
		parts := strings.Split(cfgExtraHost, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf(
				"expected format HOSTNAME:IPADDRESS. could not parse ExtraHost: %s", cfgExtraHost)
		}
		extraHost := &ecs.HostEntry{
			Hostname:  aws.String(parts[0]),
			IpAddress: aws.String(parts[1]),
		}
		extraHosts = append(extraHosts, extraHost)
	}

	return extraHosts, nil
}
