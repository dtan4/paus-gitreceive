package model

import (
	"fmt"
	"strings"

	"github.com/dtan4/paus-gitreceive/receiver/store"
	"github.com/pkg/errors"
)

type Application struct {
	Repository  string
	Revision    string
	Username    string
	AppName     string
	ProjectName string
}

func ApplicationFromArgs(args []string) *Application {
	repository := strings.Replace(args[0], "/", "-", -1)
	revision := args[1]
	username := args[2]
	appName := strings.Replace(repository, username+"-", "", 1)
	projectName := repository + "-" + revision[0:8]

	return &Application{
		repository,
		revision,
		username,
		appName,
		projectName,
	}
}

func (app *Application) EnvironmentVariables(etcd *store.Etcd) (map[string]string, error) {
	var envs map[string]string

	userDirectoryKey := "/paus/users/" + app.Username

	if !etcd.HasKey(userDirectoryKey) {
		return nil, nil
	}

	appDirectoryKey := userDirectoryKey + "/apps/" + app.AppName

	if !etcd.HasKey(appDirectoryKey) {
		return nil, nil
	}

	envDirectoryKey := appDirectoryKey + "/envs/"

	if !etcd.HasKey(envDirectoryKey) {
		return nil, nil
	}

	envKeys, err := etcd.List(envDirectoryKey, false)

	if err != nil {
		return nil, errors.Wrap(err, "Failed to get environment variable keys.")
	}

	for _, key := range envKeys {
		value, err := etcd.Get(key)

		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("Failed to get environment variable value. key: %s", key))
		}

		envs[strings.Replace(key, envDirectoryKey, "", 1)] = value
	}

	return envs, nil
}
