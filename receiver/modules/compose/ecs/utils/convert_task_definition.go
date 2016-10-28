// This file convert_task_definition.go is derived from amazon-ecs-cli project, Copyright 2015-2016 Amazon.com, Inc.
// The original code may be found:
// https://github.com/aws/amazon-ecs-cli/blob/204c9687175f651a981f88ab188fbb68f87240b4/ecs-cli/modules/compose/ecs/utils/convert_task_definition.go
// https://github.com/aws/amazon-ecs-cli/blob/204c9687175f651a981f88ab188fbb68f87240b4/ecs-cli/modules/compose/ecs/utils/name.go

package utils

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/docker/libcompose/config"
	"github.com/docker/libcompose/project"
	"github.com/docker/libcompose/yaml"
)

const (
	defaultMemLimit = 512
	kiB             = 1024

	readOnlyVolumeAccessMode  = "ro"
	readWriteVolumeAccessMode = "rw"
	volumeFromContainerKey    = "container"

	ecsVolumeNamePrefix = "volume"
)

// ConvertToTaskDefinition converts docker-compose.yml to ECS TaskDefinition
func ConvertToTaskDefinition(taskDefinitionName string, context *project.Context, prj *project.Project, serviceName, region string) (*ecs.TaskDefinition, error) {
	containerDefinitions := []*ecs.ContainerDefinition{}
	volumes := make(map[string]string)

	for _, name := range prj.ServiceConfigs.Keys() {
		svc, _ := prj.ServiceConfigs.Get(name)

		containerDef, err := convertToContainerDef(name, context, svc, volumes, serviceName, region)
		if err != nil {
			return nil, err
		}

		containerDefinitions = append(containerDefinitions, containerDef)
	}

	taskDefinition := &ecs.TaskDefinition{
		Family:               aws.String(taskDefinitionName),
		ContainerDefinitions: containerDefinitions,
		Volumes:              convertToECSVolumes(volumes),
	}

	return taskDefinition, nil
}

func convertToContainerDef(name string, context *project.Context, svc *config.ServiceConfig, volumes map[string]string, serviceName, region string) (*ecs.ContainerDefinition, error) {
	// memory
	var mem int64
	if svc.MemLimit != 0 {
		mem = int64(svc.MemLimit) / kiB / kiB // convert bytes to MiB
	}
	if mem == 0 {
		mem = defaultMemLimit
	}

	// environment variables
	environment := convertToKeyValuePairs(context, svc.Environment, name)

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
	mountPoints, err := convertToMountPoints(svc.Volumes, volumes)
	if err != nil {
		return nil, err
	}

	// extra hosts
	extraHosts, err := convertToExtraHosts(svc.ExtraHosts)
	if err != nil {
		return nil, err
	}

	// logs
	logConfig := &ecs.LogConfiguration{
		LogDriver: aws.String("awslogs"),
		Options: aws.StringMap(map[string]string{
			"awslogs-group":  serviceName,
			"awslogs-region": region,
		}),
	}

	// ulimits
	ulimits, err := convertToULimits(svc.Ulimits)
	if err != nil {
		return nil, err
	}

	containerDefinition := &ecs.ContainerDefinition{
		Cpu:                    aws.Int64(int64(svc.CPUShares)),
		Command:                aws.StringSlice(svc.Command),
		DnsSearchDomains:       aws.StringSlice(svc.DNSSearch),
		DnsServers:             aws.StringSlice(svc.DNS),
		DockerLabels:           aws.StringMap(svc.Labels),
		DockerSecurityOptions:  aws.StringSlice(svc.SecurityOpt),
		EntryPoint:             aws.StringSlice(svc.Entrypoint),
		Environment:            environment,
		ExtraHosts:             extraHosts,
		Image:                  aws.String(svc.Image),
		Links:                  aws.StringSlice(svc.Links),
		LogConfiguration:       logConfig,
		Memory:                 aws.Int64(mem),
		MountPoints:            mountPoints,
		Name:                   aws.String(name),
		Privileged:             aws.Bool(svc.Privileged),
		PortMappings:           portMappings,
		ReadonlyRootFilesystem: aws.Bool(svc.ReadOnly),
		Ulimits:                ulimits,
		VolumesFrom:            volumesFrom,
	}

	if svc.Hostname != "" {
		containerDefinition.Hostname = aws.String(svc.Hostname)
	}

	if svc.User != "" {
		containerDefinition.User = aws.String(svc.User)
	}

	if svc.WorkingDir != "" {
		containerDefinition.WorkingDirectory = aws.String(svc.WorkingDir)
	}

	return containerDefinition, nil
}

func convertToKeyValuePairs(context *project.Context, envVars yaml.MaporEqualSlice, svcName string) []*ecs.KeyValuePair {
	environment := []*ecs.KeyValuePair{}

	for _, env := range envVars {
		parts := strings.SplitN(env, "=", 2)
		key := parts[0]

		if len(parts) > 1 && parts[1] != "" {
			environment = append(environment, createKeyValuePair(key, parts[1]))
			continue
		}

		if context.EnvironmentLookup != nil {
			resolvedEnvVars := context.EnvironmentLookup.Lookup(key, svcName, nil)

			if len(resolvedEnvVars) == 0 {
				environment = append(environment, createKeyValuePair(key, ""))
				continue
			}

			value := resolvedEnvVars[0]
			lookupParts := strings.SplitN(value, "=", 2)
			environment = append(environment, createKeyValuePair(key, lookupParts[1]))
		}
	}

	return environment
}

func createKeyValuePair(key, value string) *ecs.KeyValuePair {
	return &ecs.KeyValuePair{
		Name:  aws.String(key),
		Value: aws.String(value),
	}
}

func convertToECSVolumes(hostPaths map[string]string) []*ecs.Volume {
	volumes := []*ecs.Volume{}

	for hostPath, volName := range hostPaths {
		ecsVolume := &ecs.Volume{
			Name: aws.String(volName),
			Host: &ecs.HostVolumeProperties{
				SourcePath: aws.String(hostPath),
			},
		}

		volumes = append(volumes, ecsVolume)
	}

	return volumes
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

func convertToMountPoints(cfgVolumes *yaml.Volumes, volumes map[string]string) ([]*ecs.MountPoint, error) {
	mountPoints := []*ecs.MountPoint{}
	if cfgVolumes == nil {
		return mountPoints, nil
	}

	for _, v := range cfgVolumes.Volumes {
		hostPath := v.Source
		containerPath := v.Destination

		accessMode := v.AccessMode
		var readOnly bool

		if accessMode != "" {
			if accessMode == readOnlyVolumeAccessMode {
				readOnly = true
			} else if accessMode == readWriteVolumeAccessMode {
				readOnly = false
			} else {
				return nil, fmt.Errorf(
					"expected format [HOST:]CONTAINER[:ro|rw]. could not parse volume: %s", v)
			}
		}

		var volumeName string

		if len(volumes) > 0 {
			volumeName = volumes[hostPath]
		}

		if volumeName == "" {
			volumeName = getVolumeName(len(volumes))
			volumes[hostPath] = volumeName
		}

		mountPoints = append(mountPoints, &ecs.MountPoint{
			ContainerPath: aws.String(containerPath),
			SourceVolume:  aws.String(volumeName),
			ReadOnly:      aws.Bool(readOnly),
		})
	}

	return mountPoints, nil
}

func getVolumeName(suffixNum int) string {
	return ecsVolumeNamePrefix + "-" + strconv.Itoa(suffixNum)
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

func convertToULimits(cfgUlimits yaml.Ulimits) ([]*ecs.Ulimit, error) {
	ulimits := []*ecs.Ulimit{}
	for _, cfgUlimit := range cfgUlimits.Elements {
		ulimit := &ecs.Ulimit{
			Name:      aws.String(cfgUlimit.Name),
			SoftLimit: aws.Int64(cfgUlimit.Soft),
			HardLimit: aws.Int64(cfgUlimit.Hard),
		}
		ulimits = append(ulimits, ulimit)
	}

	return ulimits, nil
}
