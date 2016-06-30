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

	identifier := strings.ToLower(deployment.App.Username + "-" + deployment.App.AppName + "-" + deployment.Revision) // dtan4-app-19fb23cd

	if err := unsetFrontend(etcd, identifier); err != nil {
		return err
	}

	if err := unsetBackend(etcd, deployment.ProjectName); err != nil {
		return err
	}

	return nil
}

func RegisterInformation(etcd *store.Etcd, deployment *model.Deployment, baseDomain string, webContainer *model.Container) ([]string, error) {
	if err := setBackend(etcd, deployment.ProjectName); err != nil {
		return nil, err
	}

	identifiers := []string{
		strings.ToLower(deployment.App.Username + "-" + deployment.App.AppName + "-" + deployment.Branch),        // dtan4-app-master
		strings.ToLower(deployment.App.Username + "-" + deployment.App.AppName + "-" + deployment.Revision[0:8]), // dtan4-app-19fb23cd
	}

	if deployment.Branch == "master" {
		identifiers = append(identifiers, strings.ToLower(deployment.App.Username+"-"+deployment.App.AppName)) // dtan4-app
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
