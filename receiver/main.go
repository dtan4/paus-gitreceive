package main

import (
	"os"

	"github.com/dtan4/paus-gitreceive/receiver/config"
	"github.com/dtan4/paus-gitreceive/receiver/model"
	"github.com/dtan4/paus-gitreceive/receiver/msg"
	"github.com/dtan4/paus-gitreceive/receiver/store"
)

func initialize() (*config.Config, *store.Etcd, error) {
	config, err := config.LoadConfig()
	if err != nil {
		return nil, nil, err
	}

	etcd, err := store.NewEtcd(config.EtcdEndpoint)
	if err != nil {
		return nil, nil, err
	}

	return config, etcd, nil
}

func main() {
	if len(os.Args) > 1 {
		if os.Args[1] == "-v" || os.Args[1] == "--version" {
			printVersion()
			os.Exit(0)
		}
	}

	config, etcd, err := initialize()
	if err != nil {
		printErrorAndExit(err)
	}

	application, err := model.ApplicationFromArgs(os.Args[1:])
	if err != nil {
		printErrorAndExit(err)
	}

	deployment, err := model.DeploymentFromArgs(application, os.Args[1:], "", config.RepositoryDir)
	if err != nil {
		printErrorAndExit(err)
	}

	if err := deploy(application, deployment, config, etcd); err != nil {
		printErrorAndExit(err)
	}
}

func printErrorAndExit(err error) {
	msg.PrintErrorf("%+v\n", err)
	os.Exit(1)
}
