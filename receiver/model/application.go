package model

import (
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
	etcd        *store.Etcd
}

func ApplicationFromArgs(args []string, etcd *store.Etcd) (*Application, error) {
	if len(args) < 3 {
		return nil, errors.Errorf("3 arguments (revision, username, appName) must be passed. got: %d", len(args))
	}

	repository := strings.Replace(args[0], "/", "-", -1)
	revision := args[1]
	username := args[2]
	appName := strings.Replace(repository, username+"-", "", 1)
	projectName := repository + "-" + revision[0:8]

	return &Application{
		Repository:  repository,
		Revision:    revision,
		Username:    username,
		AppName:     appName,
		ProjectName: projectName,
		etcd:        etcd,
	}, nil
}

func (app *Application) BuildArgs() (map[string]string, error) {
	var args = make(map[string]string)

	userDirectoryKey := "/paus/users/" + app.Username

	if !app.etcd.HasKey(userDirectoryKey) {
		return map[string]string{}, nil
	}

	appDirectoryKey := userDirectoryKey + "/apps/" + app.AppName

	if !app.etcd.HasKey(appDirectoryKey) {
		return map[string]string{}, nil
	}

	buildArgsKey := appDirectoryKey + "/build-args/"

	if !app.etcd.HasKey(buildArgsKey) {
		return map[string]string{}, nil
	}

	buildArgKeys, err := app.etcd.List(buildArgsKey, false)

	if err != nil {
		return nil, err
	}

	for _, key := range buildArgKeys {
		value, err := app.etcd.Get(key)

		if err != nil {
			return nil, err
		}

		args[strings.Replace(key, buildArgsKey, "", 1)] = value
	}

	return args, nil
}

func (app *Application) DirExists() bool {
	return app.etcd.HasKey("/paus/users/" + app.Username + "/apps/" + app.AppName)
}

func (app *Application) EnvironmentVariables() (map[string]string, error) {
	var envs = make(map[string]string)

	userDirectoryKey := "/paus/users/" + app.Username

	if !app.etcd.HasKey(userDirectoryKey) {
		return map[string]string{}, nil
	}

	appDirectoryKey := userDirectoryKey + "/apps/" + app.AppName

	if !app.etcd.HasKey(appDirectoryKey) {
		return map[string]string{}, nil
	}

	envDirectoryKey := appDirectoryKey + "/envs/"

	if !app.etcd.HasKey(envDirectoryKey) {
		return map[string]string{}, nil
	}

	envKeys, err := app.etcd.List(envDirectoryKey, false)

	if err != nil {
		return nil, err
	}

	for _, key := range envKeys {
		value, err := app.etcd.Get(key)

		if err != nil {
			return nil, err
		}

		envs[strings.Replace(key, envDirectoryKey, "", 1)] = value
	}

	return envs, nil
}

func (app *Application) RegisterMetadata(timestamp string) error {
	userDirectoryKey := "/paus/users/" + app.Username

	if !app.etcd.HasKey(userDirectoryKey) {
		_ = app.etcd.Mkdir(userDirectoryKey)
	}

	appDirectoryKey := userDirectoryKey + "/apps/" + app.AppName

	if !app.etcd.HasKey(appDirectoryKey) {
		_ = app.etcd.Mkdir(appDirectoryKey)
		_ = app.etcd.Mkdir(appDirectoryKey + "/envs")
		_ = app.etcd.Mkdir(appDirectoryKey + "/revisions")
	}

	if err := app.etcd.Set(appDirectoryKey+"/revisions/"+app.Revision, timestamp); err != nil {
		return err
	}

	return nil
}
