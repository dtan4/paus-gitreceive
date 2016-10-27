package model

import (
	"strconv"
	"strings"

	"github.com/dtan4/paus-gitreceive/receiver/modules/compose/ecs/utils"
	"github.com/dtan4/paus-gitreceive/receiver/service"
	"github.com/dtan4/paus-gitreceive/receiver/store"
	"github.com/pkg/errors"
)

const (
	buildArgsTable    = "paus-build-args"
	envsTable         = "paus-envs"
	healthchecksTable = "paus-healthchecks"
	userAppIndex      = "user-app-index"
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

// BuildArgs returns build args of given application
func (app *Application) BuildArgs() (map[string]string, error) {
	var args = make(map[string]string)

	items, err := service.Select(buildArgsTable, userAppIndex, map[string]string{
		"user": app.Username,
		"app":  app.AppName,
	})
	if err != nil {
		return nil, err
	}

	var key, value string

	for _, attrValue := range items {
		key = *attrValue["key"].S
		value = *attrValue["value"].S
		args[key] = value
	}

	return args, nil
}

// EnvironmentVariables returns environment variables of given application
func (app *Application) EnvironmentVariables() (map[string]string, error) {
	var envs = make(map[string]string)

	items, err := service.Select(envsTable, userAppIndex, map[string]string{
		"user": app.Username,
		"app":  app.AppName,
	})
	if err != nil {
		return nil, err
	}

	var key, value string

	for _, attrValue := range items {
		key = *attrValue["key"].S
		value = *attrValue["value"].S
		envs[key] = value
	}

	return envs, nil
}

// HealthCheck returns healthcheck parameters of given application
func (app *Application) HealthCheck() (string, int, int, error) {
	items, err := service.Select(healthchecksTable, userAppIndex, map[string]string{
		"user": app.Username,
		"app":  app.AppName,
	})
	if err != nil {
		return "", 0, 0, err
	}

	healthcheck := items[0]
	path := *healthcheck["path"].S

	interval, err := strconv.Atoi(*healthcheck["interval"].N)
	if err != nil {
		return "", 0, 0, err
	}

	maxTry, err := strconv.Atoi(*healthcheck["max-try"].N)
	if err != nil {
		return "", 0, 0, err
	}

	return path, interval, maxTry, nil
}

// TaskDefinitionName returns the name of TaskDefinition
func (app *Application) TaskDefinitionName() string {
	return utils.GetTaskDefinitionName(app.Username + "-" + app.AppName)
}

// ServiceName returns the name of Service with given suffix
func (app *Application) ServiceName(suffix string) string {
	return utils.GetServiceName(app.Username+"-"+app.AppName, suffix)
}
