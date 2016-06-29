package vulcand

import (
	"strings"

	"github.com/dtan4/paus-gitreceive/receiver/model"
	"github.com/dtan4/paus-gitreceive/receiver/store"
)

const (
	vulcandKeyBase = "/vulcand"
)

func DeregisterInformation(etcd *store.Etcd, deployment *model.Deployment) error {
	if err := unsetServer(etcd, deployment.ProjectName); err != nil {
		return err
	}

	identifier := strings.ToLower(deployment.ProjectName)

	if err := unsetFrontend(etcd, identifier); err != nil {
		return err
	}

	if err := unsetBackend(etcd, deployment.ProjectName); err != nil {
		return err
	}

	return nil
}

func RegisterInformation(etcd *store.Etcd, application *model.Application, deployment *model.Deployment, baseDomain string, webContainer *model.Container) ([]string, error) {
	if err := setBackend(etcd, deployment.ProjectName); err != nil {
		return nil, err
	}

	identifiers := []string{
		strings.ToLower(deployment.ProjectName),
		strings.ToLower(application.Username + "-" + application.AppName),
	}

	for _, identifier := range identifiers {
		if err := setFrontend(etcd, deployment.ProjectName, identifier, baseDomain); err != nil {
			return nil, err
		}
	}

	if err := setServer(etcd, deployment.ProjectName, webContainer, baseDomain); err != nil {
		return nil, err
	}

	return identifiers, nil
}
