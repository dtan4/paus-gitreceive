package vulcand

import (
	"strings"

	"github.com/dtan4/paus-gitreceive/receiver/model"
	"github.com/dtan4/paus-gitreceive/receiver/store"
	"github.com/pkg/errors"
)

const (
	vulcandKeyBase = "/vulcand"
)

func RegisterInformation(etcd *store.Etcd, application *model.Application, baseDomain string, webContainer *model.Container) ([]string, error) {
	if err := setBackend(etcd, application, baseDomain); err != nil {
		return nil, err
	}

	identifiers := []string{
		strings.ToLower(application.ProjectName),
		strings.ToLower(application.Username + "-" + application.AppName),
	}

	for _, identifier := range identifiers {
		if err := setFrontend(etcd, application, identifier, baseDomain); err != nil {
			return nil, err
		}
	}

	if err := setServer(etcd, application, webContainer, baseDomain); err != nil {
		return nil, errors.Wrap(err, "Failed to set vulcand backend.")
	}

	return identifiers, nil
}
