package vulcand

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dtan4/paus-gitreceive/receiver/model"
	"github.com/dtan4/paus-gitreceive/receiver/store"
	"github.com/pkg/errors"
)

const (
	VulcandKeyBase = "/vulcand"
)

var (
	httpBackendJSON string
)

func RegisterInformation(etcd *store.Etcd, application *model.Application, baseDomain string, webContainer *model.Container) ([]string, error) {
	if err := setBackend(etcd, application, baseDomain); err != nil {
		return nil, errors.Wrap(err, "Failed to set vulcand backend.")
	}

	identifiers := []string{
		strings.ToLower(application.ProjectName),
		strings.ToLower(application.Username + "-" + application.AppName),
	}

	for _, identifier := range identifiers {
		if err := setFrontend(etcd, application, identifier, baseDomain); err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("Failed to set vulcand frontend. identifier: %s", identifier))
		}
	}

	if err := setServer(etcd, application, webContainer, baseDomain); err != nil {
		return nil, errors.Wrap(err, "Failed to set vulcand backend.")
	}

	return identifiers, nil
}

// {"Type": "http"}
func setBackend(etcd *store.Etcd, application *model.Application, baseDomain string) error {
	key := fmt.Sprintf("%s/backends/%s/backend", VulcandKeyBase, application.ProjectName)

	if err := etcd.Set(key, httpBackendJSON); err != nil {
		return errors.Wrap(err, "Failed to set vulcand backend in etcd.")
	}

	return nil
}

// {"Type": "http", "BackendId": "$identifier", "Route": "Host(`$identifier.$base_domain`) && PathRegexp(`/`)", "Settings": {"TrustForwardHeader": true}}
func setFrontend(etcd *store.Etcd, application *model.Application, identifier, baseDomain string) error {
	key := fmt.Sprintf("%s/frontends/%s/frontend", VulcandKeyBase, identifier)
	frontend := Frontend{
		Type:      "http",
		BackendId: application.ProjectName,
		Route:     fmt.Sprintf("Host(`%s.%s`) && PathRegexp(`/`)", strings.ToLower(identifier), strings.ToLower(baseDomain)),
		Settings: FrontendSettings{
			TrustForwardHeader: true,
		},
	}

	b, err := json.Marshal(frontend)

	if err != nil {
		return errors.Wrap(err, "Failed to generate vulcand frontend JSON.")
	}

	// json.Marshal generates HTML-escaped JSON string
	b = bytes.Replace(b, []byte("\\u0026"), []byte("&"), -1)
	json := string(b)

	if err := etcd.Set(key, json); err != nil {
		return errors.Wrap(err, "Failed to set vulcand frontend in etcd.")
	}

	return nil
}

// {"URL": "http://$web_container_host_ip:$web_container_port"}
func setServer(etcd *store.Etcd, application *model.Application, container *model.Container, baseDomain string) error {
	key := fmt.Sprintf("%s/backends/%s/servers/%s", VulcandKeyBase, application.ProjectName, container.ContainerId)
	server := Server{
		URL: fmt.Sprintf("http://%s:%s", container.HostIP(), container.HostPort()),
	}

	b, err := json.Marshal(server)

	if err != nil {
		return errors.Wrap(err, "Failed to generate vulcand server JSON.")
	}

	json := string(b)

	if err := etcd.Set(key, json); err != nil {
		return errors.Wrap(err, "Failed to set vulcand server in etcd.")
	}

	return nil
}
