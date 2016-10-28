package model

import (
	"regexp"

	"github.com/dtan4/paus-gitreceive/receiver/aws"
	"github.com/pkg/errors"
)

const (
	deploymentsTable = "paus-deployments"
)

var (
	refnameRegexp = regexp.MustCompile(`^refs/heads/`)
)

type Deployment struct {
	App         *Application
	Branch      string
	ProjectName string
	Revision    string
	ServiceArn  string
}

// DeploymentFromArgs creates new Deployment from gitreceive arguments
// args:
//   user/app, 19fb23cd71a4cf2eab00ad1a393e40de4ed61531, user, 4c:1f:92:b9:43:2b:23:0b:c0:e8:ab:12:cd:34:ef:56, refs/heads/branch-name
func DeploymentFromArgs(app *Application, args []string) (*Deployment, error) {
	if len(args) < 5 {
		return nil, errors.Errorf("5 arguments (repository, revision, username, fingerprint, refname) must be passed. got: %d", len(args))
	}

	revision := args[1]
	branch := refnameRegexp.ReplaceAllString(args[4], "")

	return NewDeployment(app, branch, revision, ""), nil
}

// NewDeployment creates new Deployment object
func NewDeployment(app *Application, branch, revision, serviceArn string) *Deployment {
	projectName := app.Repository + "-" + revision[0:8]

	return &Deployment{
		App:         app,
		Branch:      branch,
		ProjectName: projectName,
		Revision:    revision,
		ServiceArn:  serviceArn,
	}
}

// Save saves Deployment to DynamoDB
func (d *Deployment) Save() error {
	fields := map[string]string{
		"app":         d.App.AppName,
		"user":        d.App.Username,
		"revision":    d.Revision,
		"service-arn": d.ServiceArn,
	}

	if err := aws.DynamoDB().Create(deploymentsTable, fields); err != nil {
		return err
	}

	return nil
}
