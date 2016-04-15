package main

import (
	"github.com/coreos/etcd/client"
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
		return nil, err
	}

	keysAPI := client.NewKeysAPI(c)

	return &Etcd{keysAPI}, nil
}

func (c *Etcd) Get(key string) (string, error) {
	resp, err := c.keysAPI.Get(context.Background(), key, &client.GetOptions{})

	if err != nil {
		return "", err
	}

	return resp.Node.Value, nil
}

func (c *Etcd) HasKey(key string) bool {
	_, err := c.keysAPI.Get(context.Background(), key, &client.GetOptions{})

	return err == nil
}

func (c *Etcd) Mkdir(key string) error {
	_, err := c.keysAPI.Set(context.Background(), key, "", &client.SetOptions{Dir: true})

	return err
}

func (c *Etcd) Set(key, value string) error {
	_, err := c.keysAPI.Set(context.Background(), key, value, &client.SetOptions{})

	return err
}
