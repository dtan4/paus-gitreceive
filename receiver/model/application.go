package model

import (
	"strings"
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
