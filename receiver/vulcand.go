package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

const (
	VulcandKeyBase = "/vulcand"
)

var (
	httpBackendJSON string
)

type Vulcand struct {
	etcd *Etcd
}

type VulcandBackend struct {
	Type string `json:"Type"`
}

type VulcandServer struct {
	URL string `json:"URL"`
}

type VulcandFrontend struct {
	Type      string                  `json:"Type"`
	BackendId string                  `json:"BackendId"`
	Route     string                  `json:"Route"`
	Settings  VulcandFrontendSettings `json:"Settings"`
}

type VulcandFrontendSettings struct {
	TrustForwardHeader bool `json:"TrustForwardHeader"`
}

func NewVulcand(etcd *Etcd) *Vulcand {
	backend := VulcandBackend{
		Type: "http",
	}

	b, _ := json.Marshal(backend)
	httpBackendJSON = string(b)

	return &Vulcand{
		etcd: etcd,
	}
}

// {"Type": "http"}
func (v *Vulcand) SetBackend(application *Application, baseDomain string) error {
	key := fmt.Sprintf("%s/backends/%s/backend", VulcandKeyBase, application.ProjectName)

	if err := v.etcd.Set(key, httpBackendJSON); err != nil {
		return errors.Wrap(err, "Failed to set vulcand backend in etcd.")
	}

	return nil
}

// {"Type": "http", "BackendId": "$identifier", "Route": "Host(`$identifier.$base_domain`) && PathRegexp(`/`)", "Settings": {"TrustForwardHeader": true}}
func (v *Vulcand) SetFrontend(application *Application, identifier, baseDomain string) error {
	key := fmt.Sprintf("%s/frontends/%s/frontend", VulcandKeyBase, identifier)
	frontend := VulcandFrontend{
		Type:      "http",
		BackendId: application.ProjectName,
		Route:     fmt.Sprintf("Host(`%s.%s`) && PathRegexp(`/`)", strings.ToLower(identifier), strings.ToLower(baseDomain)),
		Settings: VulcandFrontendSettings{
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
func (v *Vulcand) SetServer(application *Application, container *Container, baseDomain string) error {
	key := fmt.Sprintf("%s/backends/%s/servers/%s", VulcandKeyBase, application.ProjectName, container.ContainerId)
	server := VulcandServer{
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
