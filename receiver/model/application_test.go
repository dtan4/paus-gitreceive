package model

import (
	"testing"

	"github.com/dtan4/paus-gitreceive/receiver/store"
)

func TestApplicationFromArgs(t *testing.T) {
	var args []string

	args = []string{}
	etcd, _ := store.NewEtcd("http://example.com:2379")

	_, err := ApplicationFromArgs(args, etcd)

	if err == nil {
		t.Fatalf("Error should be raised")
	}

	args = []string{
		"dtan4/rails-sample",
		"3e634e41d5a819a7586c621a6322ee4d5085232c",
		"dtan4",
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
