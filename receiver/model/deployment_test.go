package model

import (
	"testing"
)

func TestDeploymentFromArgs(t *testing.T) {
	var (
		args []string

		actual   string
		expected string
	)

	app := &Application{
		Repository: "user-repository",
		Username:   "user",
		AppName:    "app",
	}

	timestamp := "1467181319"
	repositoryDir := "/repos"

	args = []string{}

	_, err := DeploymentFromArgs(app, args, timestamp, repositoryDir)

	if err == nil {
		t.Fatalf("Error should be raised when empty args is passed")
	}

	args = []string{
		"dtan4/rails-sample",
		"3e634e41d5a819a7586c621a6322ee4d5085232c",
		"dtan4",
		"4c:1f:92:b9:43:2b:23:0b:c0:e8:ab:12:cd:34:ef:56",
		"refs/heads/branch",
	}

	deployment, err := DeploymentFromArgs(app, args, timestamp, repositoryDir)

	if err != nil {
		t.Fatalf("Error should not be raised.")
	}

	expected = "branch"
	actual = deployment.Branch

	if actual != expected {
		t.Fatalf("Branch does not match. expected: %s actual: %s", expected, actual)
	}
}

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

	branch := "branch"
	revision := "19fb23cd71a4cf2eab00ad1a393e40de4ed61531"
	timestamp := "1467181319"
	repositoryDir := "/repos"

	deployment := NewDeployment(app, branch, revision, timestamp, repositoryDir)

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
