package model

import (
	"testing"
)

func TestComposeFilePath(t *testing.T) {
	application := &Application{
		Repository: "user-repository",
		Username:   "user",
		AppName:    "app",
	}

	deployment := &Deployment{
		App:       application,
		Revision:  "19fb23cd71a4cf2eab00ad1a393e40de4ed61531",
		Timestamp: "1467181319",
	}

	repositoryDir := "/repos"

	expected := "/repos/user/user-repository-19fb23cd/docker-compose-1467181319.yml"
	actual := deployment.ComposeFilePath(repositoryDir)

	if actual != expected {
		t.Fatalf("ComposeFilePath does not match. expected: %s actual: %s", expected, actual)
	}
}

func TestProjectName(t *testing.T) {
	application := &Application{
		Repository: "user-repository",
		Username:   "user",
		AppName:    "app",
	}

	deployment := &Deployment{
		App:       application,
		Revision:  "19fb23cd71a4cf2eab00ad1a393e40de4ed61531",
		Timestamp: "1467181319",
	}

	expected := "user-repository-19fb23cd"
	actual := deployment.ProjectName()

	if actual != expected {
		t.Fatalf("ProjectName does not match. expected: %s actual: %s", expected, actual)
	}
}
