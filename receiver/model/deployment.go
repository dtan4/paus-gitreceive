package model

import (
	"path/filepath"
)

type Deployment struct {
	Revision  string
	Timestamp string

	app *Application
}

func NewDeployment(app *Application, revision, timestamp string) *Deployment {
	return &Deployment{
		app:       app,
		Revision:  revision,
		Timestamp: timestamp,
	}
}

func (d *Deployment) ComposeFilePath(repositoryDir string) string {
	return filepath.Join(repositoryDir, d.app.Username, d.ProjectName(), "docker-compose-"+d.Timestamp+".yml")
}

func (d *Deployment) ProjectName() string {
	return d.app.Repository + "-" + d.Revision[0:8]
}
