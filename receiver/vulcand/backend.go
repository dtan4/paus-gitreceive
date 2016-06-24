package vulcand

import (
	"fmt"

	"github.com/dtan4/paus-gitreceive/receiver/model"
	"github.com/dtan4/paus-gitreceive/receiver/store"
)

const (
	httpBackendJSON = "{\"Type\": \"http\"}"
)

type Backend struct {
	Type string `json:"Type"`
}

// {"Type": "http"}
func setBackend(etcd *store.Etcd, application *model.Application, baseDomain string) error {
	key := fmt.Sprintf("%s/backends/%s/backend", vulcandKeyBase, application.ProjectName)

	if err := etcd.Set(key, httpBackendJSON); err != nil {
		return err
	}

	return nil
}
