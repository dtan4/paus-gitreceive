package model

import (
	"path/filepath"
	"regexp"

	"github.com/pkg/errors"
)

var (
	refnameRegexp = regexp.MustCompile(`^refs/heads/`)
)

type Deployment struct {
	Branch          string
	ComposeFilePath string
	ProjectName     string
	Revision        string
	Timestamp       string

	app *Application
}

// args:
//   user/app, 19fb23cd71a4cf2eab00ad1a393e40de4ed61531, user, 4c:1f:92:b9:43:2b:23:0b:c0:e8:ab:12:cd:34:ef:56, refs/heads/branch-name
func DeploymentFromArgs(app *Application, args []string, timestamp, repositoryDir string) (*Deployment, error) {
	if len(args) < 5 {
		return nil, errors.Errorf("5 arguments (repository, revision, username, fingerprint, refname) must be passed. got: %d", len(args))
	}

	revision := args[1]
	branch := refnameRegexp.ReplaceAllString(args[4], "")

	return NewDeployment(app, branch, revision, timestamp, repositoryDir), nil
}

func NewDeployment(app *Application, branch, revision, timestamp, repositoryDir string) *Deployment {
	projectName := app.Repository + "-" + revision[0:8]
	composeFilePath := filepath.Join(repositoryDir, app.Username, projectName, "docker-compose-"+timestamp+".yml")

	return &Deployment{
		app:             app,
		Branch:          branch,
		ComposeFilePath: composeFilePath,
		ProjectName:     projectName,
		Revision:        revision,
		Timestamp:       timestamp,
	}
}

func (d *Deployment) Register() error {
	return d.app.RegisterMetadata(d.Revision, d.Timestamp)
}
