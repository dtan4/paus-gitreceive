#!/bin/bash

if [ ! -d /home/git/.ssh ]; then
  mkdir /home/git/.ssh
  chown -R git /home/git/.ssh
fi

if [ ! -d /root/paus ]; then
  mkdir -p /root/paus
fi

touch /root/paus/config

if [ -n "$PAUS_BASE_DOMAIN" ]; then
  echo "BaseDomain=$PAUS_BASE_DOMAIN" >> /root/paus/config
else
  echo "required key PAUS_BASE_DOMAIN missing value"
  exit 1
fi

if [ -n "$PAUS_DOCKER_HOST" ]; then
  echo "DockerHost=$PAUS_DOCKER_HOST" >> /root/paus/config
fi

if [ -n "$PAUS_ETCD_ENDPOINT" ]; then
  echo "EtcdEndpoint=$PAUS_ETCD_ENDPOINT" >> /root/paus/config
fi

if [ -n "$PAUS_REPOSITORY_DIR" ]; then
  echo "RepositoryDir=$PAUS_REPOSITORY_DIR" >> /root/paus/config
  chown -R git:git $PAUS_REPOSITORY_DIR
fi

exec $@
