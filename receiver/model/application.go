package model

import (
	"fmt"
	"strconv"
	"strings"
	"time"

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

func (app *Application) BuildArgs(etcd *store.Etcd) (map[string]string, error) {
	var args map[string]string

	userDirectoryKey := "/paus/users/" + app.Username

	if !etcd.HasKey(userDirectoryKey) {
		return nil, nil
	}

	appDirectoryKey := userDirectoryKey + "/apps/" + app.AppName

	if !etcd.HasKey(appDirectoryKey) {
		return nil, nil
	}

	buildArgsKey := appDirectoryKey + "/build-args/"

	if !etcd.HasKey(buildArgsKey) {
		return nil, nil
	}

	buildArgKeys, err := etcd.List(buildArgsKey, false)

	if err != nil {
		return nil, errors.Wrap(err, "Failed to get build arg keys.")
	}

	for _, key := range buildArgKeys {
		value, err := etcd.Get(key)

		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("Failed to get build arg value. key: %s", key))
		}

		args[strings.Replace(key, buildArgsKey, "", 1)] = value
	}

	return args, nil
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

func (app *Application) RegisterMetadata(etcd *store.Etcd) error {
	userDirectoryKey := "/paus/users/" + app.Username

	if !etcd.HasKey(userDirectoryKey) {
		_ = etcd.Mkdir(userDirectoryKey)
	}

	appDirectoryKey := userDirectoryKey + "/apps/" + app.AppName

	if !etcd.HasKey(appDirectoryKey) {
		_ = etcd.Mkdir(appDirectoryKey)
		_ = etcd.Mkdir(appDirectoryKey + "/envs")
		_ = etcd.Mkdir(appDirectoryKey + "/revisions")
	}

	if err := etcd.Set(appDirectoryKey+"/revisions/"+app.Revision, strconv.FormatInt(time.Now().Unix(), 10)); err != nil {
		return errors.Wrap(err, "Failed to set revisdion.")
	}

	return nil
}