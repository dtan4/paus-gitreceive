package main

import (
	"testing"
)

func TestNewCommitMetadataFromArgs(t *testing.T) {
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

	commitMetadata := CommitMetadataFromArgs(args)

	if expectedRepository != commitMetadata.Repository {
		t.Fatalf("Repository is not matched. Expected: %s, Actual: %s", expectedRepository, commitMetadata.Repository)
	}

	if expectedRevision != commitMetadata.Revision {
		t.Fatalf("Revision is not matched. Expected: %s, Actual: %s", expectedRevision, commitMetadata.Revision)
	}

	if expectedUsername != commitMetadata.Username {
		t.Fatalf("Username is not matched. Expected: %s, Actual: %s", expectedUsername, commitMetadata.Username)
	}

	if expectedAppName != commitMetadata.AppName {
		t.Fatalf("AppName is not matched. Expected: %s, Actual: %s", expectedAppName, commitMetadata.AppName)
	}

	if expectedProjectName != commitMetadata.ProjectName {
		t.Fatalf("ProjectName is not matched. Expected: %s, Actual: %s", expectedProjectName, commitMetadata.ProjectName)
	}
}
