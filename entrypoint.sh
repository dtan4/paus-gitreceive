#!/bin/bash

if [ ! -d /home/git/.ssh ]; then
  mkdir /home/git/.ssh
  chown -R git /home/git/.ssh
fi

if [ ! -d /paus ]; then
  mkdir -p /paus
fi

touch /etc/profile.d/envs.sh

if [ -n "$AWS_ACCESS_KEY_ID" ]; then
  echo "export AWS_ACCESS_KEY_ID=$AWS_ACCESS_KEY_ID" >> /etc/profile.d/envs.sh
fi

if [ -n "$AWS_SECRET_ACCESS_KEY" ]; then
  echo "export AWS_SECRET_ACCESS_KEY=$AWS_SECRET_ACCESS_KEY" >> /etc/profile.d/envs.sh
fi

if [ -n "$AWS_REGION" ]; then
  echo "export AWS_REGION=$AWS_REGION" >> /etc/profile.d/envs.sh
fi

if [ -n "$PAUS_BASE_DOMAIN" ]; then
  echo "export PAUS_BASE_DOMAIN=$PAUS_BASE_DOMAIN" >> /etc/profile.d/envs.sh
else
  echo "required key PAUS_BASE_DOMAIN missing value"
  exit 1
fi

if [ -n "$PAUS_DOCKER_HOST" ]; then
  echo "export PAUS_DOCKER_HOST=$PAUS_DOCKER_HOST" >> /etc/profile.d/envs.sh
fi

if [ -n "$PAUS_ETCD_ENDPOINT" ]; then
  echo "export PAUS_ETCD_ENDPOINT=$PAUS_ETCD_ENDPOINT" >> /etc/profile.d/envs.sh
fi

if [ -n "$PAUS_MAX_APP_DEPLOY" ]; then
  echo "export PAUS_MAX_APP_DEPLOY=$PAUS_MAX_APP_DEPLOY" >> /etc/profile.d/envs.sh
fi

if [ -n "$PAUS_REPOSITORY_DIR" ]; then
  echo "export PAUS_REPOSITORY_DIR=$PAUS_REPOSITORY_DIR" >> /etc/profile.d/envs.sh
  chown -R git:git $PAUS_REPOSITORY_DIR
fi

if [ -n "$PAUS_URI_SCHEME" ]; then
  echo "export PAUS_URI_SCHEME=$PAUS_URI_SCHEME" >> /etc/profile.d/envs.sh
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
