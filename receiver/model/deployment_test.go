package model

import (
	"testing"
)

func TestNewDeployment(t *testing.T) {
	var (
		actual   string
		expected string
	)

	app := &Application{
		Repository: "user-repository",
		Username:   "user",
		AppName:    "app",
	}

	revision := "19fb23cd71a4cf2eab00ad1a393e40de4ed61531"
	timestamp := "1467181319"
	repositoryDir := "/repos"

	deployment := NewDeployment(app, revision, timestamp, repositoryDir)

	expected = "/repos/user/user-repository-19fb23cd/docker-compose-1467181319.yml"
	actual = deployment.ComposeFilePath

	if actual != expected {
		t.Fatalf("ComposeFilePath does not match. expected: %s actual: %s", expected, actual)
	}

	expected = "user-repository-19fb23cd"
	actual = deployment.ProjectName

	if actual != expected {
		t.Fatalf("ProjectName does not match. expected: %s actual: %s", expected, actual)
	}
}
