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

type Frontend struct {
	Type      string           `json:"Type"`
	BackendId string           `json:"BackendId"`
	Route     string           `json:"Route"`
	Settings  FrontendSettings `json:"Settings"`
}

type FrontendSettings struct {
	TrustForwardHeader bool `json:"TrustForwardHeader"`
}

// {"Type": "http", "BackendId": "$identifier", "Route": "Host(`$identifier.$base_domain`) && PathRegexp(`/`)", "Settings": {"TrustForwardHeader": true}}
func setFrontend(etcd *store.Etcd, application *model.Application, identifier, baseDomain string) error {
	key := fmt.Sprintf("%s/frontends/%s/frontend", vulcandKeyBase, identifier)
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
		return err
	}

	return nil
}
