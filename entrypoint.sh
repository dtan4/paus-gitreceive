#!/bin/bash

if [ ! -d /home/git/.ssh ]; then
  mkdir /home/git/.ssh
  chown -R git /home/git/.ssh
fi

if [ ! -d /root/paus/config ]; then
  mkdir -p /root/paus/config
fi

if [ -n "$PAUS_BASE_DOMAIN" ]; then
  echo $PAUS_BASE_DOMAIN > /root/paus/config/BaseDomain
else
  echo "required key PAUS_BASE_DOMAIN missing value"
  exit 1
fi

if [ -n "$PAUS_DOCKER_HOST" ]; then
  echo $PAUS_DOCKER_HOST > /root/paus/config/DockerHost
fi

if [ -n "$PAUS_ETCD_ENDPOINT" ]; then
  echo $PAUS_ETCD_ENDPOINT > /root/paus/config/EtcdEndpoint
fi

if [ -n "$PAUS_REPOSITORY_DIR" ]; then
  echo $PAUS_REPOSITORY_DIR > /root/paus/config/RepositoryDir
fi

exec $@
