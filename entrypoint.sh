#!/bin/bash

if [ ! -d /home/git/.ssh ]; then
  mkdir /home/git/.ssh
  chown -R git /home/git/.ssh
fi

if [ ! -d /paus ]; then
  mkdir -p /paus
fi

touch /paus/config

if [ -n "$PAUS_BASE_DOMAIN" ]; then
  echo "BaseDomain=$PAUS_BASE_DOMAIN" >> /paus/config
else
  echo "required key PAUS_BASE_DOMAIN missing value"
  exit 1
fi

if [ -n "$PAUS_DOCKER_HOST" ]; then
  echo "DockerHost=$PAUS_DOCKER_HOST" >> /paus/config
fi

if [ -n "$PAUS_ETCD_ENDPOINT" ]; then
  echo "EtcdEndpoint=$PAUS_ETCD_ENDPOINT" >> /paus/config
fi

if [ -n "$PAUS_MAX_APP_DEPLOY" ]; then
  echo "MaxAppDeploy=$PAUS_MAX_APP_DEPLOY" >> /paus/config
fi

if [ -n "$PAUS_REPOSITORY_DIR" ]; then
  echo "RepositoryDir=$PAUS_REPOSITORY_DIR" >> /paus/config
  chown -R git:git $PAUS_REPOSITORY_DIR
fi

if [ -n "$PAUS_URI_SCHEME" ]; then
  echo "URIScheme=$PAUS_URI_SCHEME" >> /paus/config
fi

if [ -n "$PAUS_DOCKER_CONFIG_BASE64" ]; then
  if [ ! -d /home/git/.docker ]; then
    mkdir /home/git/.docker
  fi

  echo $PAUS_DOCKER_CONFIG_BASE64 | base64 -d > /home/git/.docker/config.json
  chown -R git /home/git/.docker
fi

exec $@
