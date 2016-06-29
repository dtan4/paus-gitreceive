package vulcand

import (
	"fmt"

	"github.com/dtan4/paus-gitreceive/receiver/store"
)

const (
	httpBackendJSON = "{\"Type\": \"http\"}"
)

type Backend struct {
	Type string `json:"Type"`
}

// {"Type": "http"}
func setBackend(etcd *store.Etcd, projectName string) error {
	key := fmt.Sprintf("%s/backends/%s/backend", vulcandKeyBase, projectName)

	if err := etcd.Set(key, httpBackendJSON); err != nil {
		return err
	}

	return nil
}

func unsetBackend(etcd *store.Etcd, projectName string) error {
	key := fmt.Sprintf("%s/backends/%s/backend", vulcandKeyBase, projectName)

	if err := etcd.Delete(key); err != nil {
		return err
	}

	return nil
}
