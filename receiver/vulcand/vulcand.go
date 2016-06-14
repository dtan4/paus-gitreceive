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

type Vulcand struct {
	etcd *store.Etcd
}

func NewVulcand(etcd *store.Etcd) *Vulcand {
	backend := Backend{
		Type: "http",
	}

	b, _ := json.Marshal(backend)
	httpBackendJSON = string(b)

	return &Vulcand{
		etcd: etcd,
	}
}

// {"Type": "http"}
func (v *Vulcand) SetBackend(application *model.Application, baseDomain string) error {
	key := fmt.Sprintf("%s/backends/%s/backend", VulcandKeyBase, application.ProjectName)

	if err := v.etcd.Set(key, httpBackendJSON); err != nil {
		return errors.Wrap(err, "Failed to set vulcand backend in etcd.")
	}

	return nil
}

// {"Type": "http", "BackendId": "$identifier", "Route": "Host(`$identifier.$base_domain`) && PathRegexp(`/`)", "Settings": {"TrustForwardHeader": true}}
func (v *Vulcand) SetFrontend(application *model.Application, identifier, baseDomain string) error {
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

	if err := v.etcd.Set(key, json); err != nil {
		return errors.Wrap(err, "Failed to set vulcand frontend in etcd.")
	}

	return nil
}

// {"URL": "http://$web_container_host_ip:$web_container_port"}
func (v *Vulcand) SetServer(application *model.Application, container *model.Container, baseDomain string) error {
	key := fmt.Sprintf("%s/backends/%s/servers/%s", VulcandKeyBase, application.ProjectName, container.ContainerId)
	server := Server{
		URL: fmt.Sprintf("http://%s:%s", container.HostIP(), container.HostPort()),
	}

	b, err := json.Marshal(server)

	if err != nil {
		return errors.Wrap(err, "Failed to generate vulcand server JSON.")
	}

	json := string(b)

	if err := v.etcd.Set(key, json); err != nil {
		return errors.Wrap(err, "Failed to set vulcand server in etcd.")
	}

	return nil
}
