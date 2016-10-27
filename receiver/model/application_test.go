package model

import (
	"testing"
)

func TestApplicationFromArgs(t *testing.T) {
	var args []string

	args = []string{}

	_, err := ApplicationFromArgs(args)

	if err == nil {
		t.Fatalf("Error should be raised")
	}

	args = []string{
		"dtan4/rails-sample",
		"3e634e41d5a819a7586c621a6322ee4d5085232c",
		"dtan4",
		"4c:1f:92:b9:43:2b:23:0b:c0:e8:ab:12:cd:34:ef:56",
		"refs/heads/branch",
	}

	expectedRepository := "dtan4-rails-sample"
	expectedUsername := "dtan4"
	expectedAppName := "rails-sample"

	application, err := ApplicationFromArgs(args, etcd)

	if err != nil {
		t.Fatalf("Unexpected error has been raised. error: %s", err)
	}

	if expectedRepository != application.Repository {
		t.Fatalf("Repository is not matched. Expected: %s, Actual: %s", expectedRepository, application.Repository)
	}

	if expectedUsername != application.Username {
		t.Fatalf("Username is not matched. Expected: %s, Actual: %s", expectedUsername, application.Username)
	}

	if expectedAppName != application.AppName {
		t.Fatalf("AppName is not matched. Expected: %s, Actual: %s", expectedAppName, application.AppName)
	}
}
