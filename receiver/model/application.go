package model

import (
	"strings"

	"github.com/dtan4/paus-gitreceive/receiver/store"
	"github.com/pkg/errors"
)

type Application struct {
	Repository string
	Username   string
	AppName    string

	etcd *store.Etcd
}

// args:
//  user/app, 19fb23cd71a4cf2eab00ad1a393e40de4ed61531, user, 4c:1f:92:b9:43:2b:23:0b:c0:e8:ab:12:cd:34:ef:56, refs/heads/branch-name
func ApplicationFromArgs(args []string, etcd *store.Etcd) (*Application, error) {
	if len(args) < 5 {
		return nil, errors.Errorf("5 arguments (repository, revision, username, fingerprint, refname) must be passed. got: %d", len(args))
	}

	repository := strings.Replace(args[0], "/", "-", -1)
	username := args[2]
	appName := strings.Replace(repository, username+"-", "", 1)

	return &Application{
		Repository: repository,
		Username:   username,
		AppName:    appName,
		etcd:       etcd,
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

func (app *Application) DeleteDeployment(deployment string) error {
	key := "/paus/users/" + app.Username + "/apps/" + app.AppName + "/deployments/" + deployment

	if err := app.etcd.Delete(key); err != nil {
		return err
	}

	return nil
}

func (app *Application) Deployments() (map[string]string, error) {
	var deployments = make(map[string]string)

	deploymentsKey := "/paus/users/" + app.Username + "/apps/" + app.AppName + "/deployments/"
	keys, err := app.etcd.List(deploymentsKey, false)

	if err != nil {
		return nil, err
	}

	for _, key := range keys {
		value, err := app.etcd.Get(key)

		if err != nil {
			return nil, err
		}

		deployments[strings.Replace(key, deploymentsKey, "", 1)] = value
	}

	return deployments, nil
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

func (app *Application) RegisterMetadata(revision, timestamp string) error {
	userDirectoryKey := "/paus/users/" + app.Username

	if !app.etcd.HasKey(userDirectoryKey) {
		_ = app.etcd.Mkdir(userDirectoryKey)
	}

	appDirectoryKey := userDirectoryKey + "/apps/" + app.AppName

	if !app.etcd.HasKey(appDirectoryKey) {
		_ = app.etcd.Mkdir(appDirectoryKey)
		_ = app.etcd.Mkdir(appDirectoryKey + "/deployments")
		_ = app.etcd.Mkdir(appDirectoryKey + "/envs")
	}

	if err := app.etcd.Set(appDirectoryKey+"/deployments/"+timestamp, revision); err != nil {
		return err
	}

	return nil
}
