package utils

import (
	"testing"
)

var (
	projectName = "dtan4-rails"
)

func TestGetServiceName(t *testing.T) {
	suffix := "suffix"

	expected := "paus-dtan4-rails-suffix"
	actual := GetServiceName(projectName, suffix)

	if expected != actual {
		t.Fatalf("Service name mismatched. expected: %s, actual: %s", expected, actual)
	}
}

func TestGetTaskDefinitionName(t *testing.T) {
	expected := "paus-dtan4-rails"
	actual := GetTaskDefinitionName(projectName)

	if expected != actual {
		t.Fatalf("TaskDefinition name mismatched. expected: %s, actual: %s", expected, actual)
	}
}
