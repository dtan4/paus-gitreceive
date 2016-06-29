package store

import (
	"github.com/coreos/etcd/client"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

type Etcd struct {
	keysAPI client.KeysAPI
}

func NewEtcd(etcdEndpoint string) (*Etcd, error) {
	config := client.Config{
		Endpoints: []string{etcdEndpoint},
		Transport: client.DefaultTransport,
	}

	c, err := client.New(config)

	if err != nil {
		return nil, errors.Wrap(err, "Failed to create etcd client.")
	}

	keysAPI := client.NewKeysAPI(c)

	return &Etcd{keysAPI}, nil
}

func (c *Etcd) Delete(key string) error {
	_, err := c.keysAPI.Delete(context.Background(), key, &client.DeleteOptions{})

	if err != nil {
		return errors.Wrapf(err, "Failed to delete etcd entry. key: %s", key)
	}

	return nil
}

func (c *Etcd) Get(key string) (string, error) {
	resp, err := c.keysAPI.Get(context.Background(), key, &client.GetOptions{})

	if err != nil {
		return "", errors.Wrapf(err, "Failed to get etcd value. key: %s", key)
	}

	return resp.Node.Value, nil
}

func (c *Etcd) HasKey(key string) bool {
	_, err := c.keysAPI.Get(context.Background(), key, &client.GetOptions{})

	return err == nil
}

func (c *Etcd) List(key string, recursive bool) ([]string, error) {
	result := []string{}

	resp, err := c.keysAPI.Get(context.Background(), key, &client.GetOptions{
		Recursive: recursive,
		Sort:      true,
	})

	if err != nil {
		return nil, errors.Wrapf(err, "Failed to list up etcd keys. key: %s, recursive: %v", key, recursive)
	}

	for _, node := range resp.Node.Nodes {
		result = append(result, node.Key)
	}

	return result, nil
}

func (c *Etcd) Mkdir(key string) error {
	_, err := c.keysAPI.Set(context.Background(), key, "", &client.SetOptions{Dir: true})

	if err != nil {
		return errors.Wrapf(err, "Failed to create etcd directory. key: %s", key)
	}

	return nil
}

func (c *Etcd) Set(key, value string) error {
	_, err := c.keysAPI.Set(context.Background(), key, value, &client.SetOptions{})

	if err != nil {
		return errors.Wrapf(err, "Failed to set etcd value. key: %s, value: %s", key, value)
	}

	return nil
}
