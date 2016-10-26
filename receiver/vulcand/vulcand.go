package vulcand

import (
	"regexp"
	"strings"

	"github.com/dtan4/paus-gitreceive/receiver/model"
	"github.com/dtan4/paus-gitreceive/receiver/store"
)

const (
	vulcandKeyBase = "/vulcand"
)

var (
	branchRegexp = regexp.MustCompile(`[^a-zA-Z0-9.-]`)
)

// DeregisterInformation removes Vulcand routing information to service container
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

// RegisterInformation registers Vulcand routing information to service container
func RegisterInformation(etcd *store.Etcd, deployment *model.Deployment, baseDomain, serviceAddress string) ([]string, error) {
	if err := setBackend(etcd, deployment.ProjectName); err != nil {
		return nil, err
	}

	branchIdentifier := strings.ToLower(deployment.App.Username + "-" + deployment.App.AppName + "-" + branchRegexp.ReplaceAllString(deployment.Branch, "-"))

	if len(branchIdentifier) > 63 {
		branchIdentifier = branchIdentifier[0:63]
	}

	lastChar := string(branchIdentifier[len(branchIdentifier)-1])

	if lastChar == "." || lastChar == "-" {
		branchIdentifier = branchIdentifier[0:(len(branchIdentifier) - 1)]
	}

	identifiers := []string{
		branchIdentifier, // dtan4-app-master
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

	if err := setServer(etcd, deployment.ProjectName, baseDomain, serviceAddress); err != nil {
		return nil, err
	}

	return identifiers, nil
}
