package model

import (
	"testing"
)

func TestApplicationFromArgs(t *testing.T) {
	args := []string{
		"dtan4/rails-sample",
		"3e634e41d5a819a7586c621a6322ee4d5085232c",
		"dtan4",
	}

	expectedRepository := "dtan4-rails-sample"
	expectedRevision := "3e634e41d5a819a7586c621a6322ee4d5085232c"
	expectedUsername := "dtan4"
	expectedAppName := "rails-sample"
	expectedProjectName := "dtan4-rails-sample-3e634e41"

	application := ApplicationFromArgs(args)

	if expectedRepository != application.Repository {
		t.Fatalf("Repository is not matched. Expected: %s, Actual: %s", expectedRepository, application.Repository)
	}

	if expectedRevision != application.Revision {
		t.Fatalf("Revision is not matched. Expected: %s, Actual: %s", expectedRevision, application.Revision)
	}

	if expectedUsername != application.Username {
		t.Fatalf("Username is not matched. Expected: %s, Actual: %s", expectedUsername, application.Username)
	}

	if expectedAppName != application.AppName {
		t.Fatalf("AppName is not matched. Expected: %s, Actual: %s", expectedAppName, application.AppName)
	}

	if expectedProjectName != application.ProjectName {
		t.Fatalf("ProjectName is not matched. Expected: %s, Actual: %s", expectedProjectName, application.ProjectName)
	}
}
