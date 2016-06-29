package vulcand

import (
	"encoding/json"
	"fmt"

	"github.com/dtan4/paus-gitreceive/receiver/model"
	"github.com/dtan4/paus-gitreceive/receiver/store"
	"github.com/pkg/errors"
)

type Server struct {
	URL string `json:"URL"`
}

// {"URL": "http://$web_container_host_ip:$web_container_port"}
func setServer(etcd *store.Etcd, projectName string, container *model.Container, baseDomain string) error {
	key := fmt.Sprintf("%s/backends/%s/servers/%s", vulcandKeyBase, projectName, container.ContainerId)
	server := Server{
		URL: fmt.Sprintf("http://%s:%s", container.HostIP(), container.HostPort()),
	}

	b, err := json.Marshal(server)

	if err != nil {
		return errors.Wrap(err, "Failed to generate vulcand server JSON.")
	}

	json := string(b)

	if err := etcd.Set(key, json); err != nil {
		return err
	}

	return nil
}

func unsetServer(etcd *store.Etcd, projectName string) error {
	key := fmt.Sprintf("%s/backends/%s/servers", vulcandKeyBase, projectName)

	if err := etcd.DeleteDir(key, true); err != nil {
		return err
	}

	return nil
}
