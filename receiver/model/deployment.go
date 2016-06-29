package model

import (
	"path/filepath"
)

type Deployment struct {
	ComposeFilePath string
	ProjectName     string
	Revision        string
	Timestamp       string

	app *Application
}

func NewDeployment(app *Application, revision, timestamp, repositoryDir string) *Deployment {
	projectName := app.Repository + "-" + revision[0:8]
	composeFilePath := filepath.Join(repositoryDir, app.Username, projectName, "docker-compose-"+timestamp+".yml")

	return &Deployment{
		app:             app,
		ComposeFilePath: composeFilePath,
		ProjectName:     projectName,
		Revision:        revision,
		Timestamp:       timestamp,
	}
}

func (d *Deployment) Register() error {
	return d.app.RegisterMetadata(d.Revision, d.Timestamp)
}
