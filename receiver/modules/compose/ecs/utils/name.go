package utils

import (
	"fmt"
)

const (
	serviceNamePrefix        = "paus"
	taskDefinitionNamePrefix = "paus"
)

// GetServiceName returns the name of Service with predefined prefix and given suffix
func GetServiceName(projectName, suffix string) string {
	return fmt.Sprintf("%s-%s-%s", serviceNamePrefix, projectName, suffix)
}

// GetTaskDefinitionName returns the name of TaskDefinition with predefined prefix
func GetTaskDefinitionName(projectName string) string {
	return fmt.Sprintf("%s-%s", taskDefinitionNamePrefix, projectName)
}
