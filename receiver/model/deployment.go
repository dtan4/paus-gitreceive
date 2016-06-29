package model

import (
	"path/filepath"
)

type Deployment struct {
	App       *Application
	Revision  string
	Timestamp string
}

func (d *Deployment) ComposeFilePath(repositoryDir string) string {
	return filepath.Join(repositoryDir, d.App.Username, d.ProjectName(), "docker-compose-"+d.Timestamp+".yml")
}

func (d *Deployment) ProjectName() string {
	return d.App.Repository + "-" + d.Revision[0:8]
}
