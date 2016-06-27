package model

import (
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

func ApplicationFromArgs(args []string) (*Application, error) {
	if len(args) != 3 {
		return nil, errors.Errorf("3 arguments (revision, username, appName) must be passed. got: %d", len(args))
	}

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
	}, nil
}

func (app *Application) BuildArgs(etcd *store.Etcd) (map[string]string, error) {
	var args = make(map[string]string)

	userDirectoryKey := "/paus/users/" + app.Username

	if !etcd.HasKey(userDirectoryKey) {
		return map[string]string{}, nil
	}

	appDirectoryKey := userDirectoryKey + "/apps/" + app.AppName

	if !etcd.HasKey(appDirectoryKey) {
		return map[string]string{}, nil
	}

	buildArgsKey := appDirectoryKey + "/build-args/"

	if !etcd.HasKey(buildArgsKey) {
		return map[string]string{}, nil
	}

	buildArgKeys, err := etcd.List(buildArgsKey, false)

	if err != nil {
		return nil, err
	}

	for _, key := range buildArgKeys {
		value, err := etcd.Get(key)

		if err != nil {
			return nil, err
		}

		args[strings.Replace(key, buildArgsKey, "", 1)] = value
	}

	return args, nil
}

func (app *Application) DirExists(etcd *store.Etcd) bool {
	return etcd.HasKey("/paus/users/" + app.Username + "/apps/" + app.AppName)
}

func (app *Application) EnvironmentVariables(etcd *store.Etcd) (map[string]string, error) {
	var envs = make(map[string]string)

	userDirectoryKey := "/paus/users/" + app.Username

	if !etcd.HasKey(userDirectoryKey) {
		return map[string]string{}, nil
	}

	appDirectoryKey := userDirectoryKey + "/apps/" + app.AppName

	if !etcd.HasKey(appDirectoryKey) {
		return map[string]string{}, nil
	}

	envDirectoryKey := appDirectoryKey + "/envs/"

	if !etcd.HasKey(envDirectoryKey) {
		return map[string]string{}, nil
	}

	envKeys, err := etcd.List(envDirectoryKey, false)

	if err != nil {
		return nil, err
	}

	for _, key := range envKeys {
		value, err := etcd.Get(key)

		if err != nil {
			return nil, err
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
		return err
	}

	return nil
}
