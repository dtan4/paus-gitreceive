#!/bin/bash

set -eu

if [ ! -f id_rsa.pub ]; then
  echo "id_rsa.pub does not exist!"
  exit 1
fi

docker-compose -f docker-compose-coreos.yml stop || true
docker-compose -f docker-compose-coreos.yml rm -f || true

docker-compose -f docker-compose-coreos.yml build gitreceive gitreceive-upload-key
docker-compose -f docker-compose-coreos.yml up -d gitreceive
docker-compose run --rm gitreceive-upload-key $PAUS_USER "$(cat id_rsa.pub)"

etcdctl ls --recursive /paus/users/$PAUS_USER/apps/$PAUS_APPNAME/build-args/ | xargs -r etcdctl rm || true
etcdctl rmdir /paus/users/$PAUS_USER/apps/$PAUS_APPNAME/build-args || true

etcdctl ls --recursive /paus/users/$PAUS_USER/apps/$PAUS_APPNAME/deployments/ | xargs -r etcdctl rm || true
etcdctl rmdir /paus/users/$PAUS_USER/apps/$PAUS_APPNAME/deployments || true

etcdctl ls --recursive /paus/users/$PAUS_USER/apps/$PAUS_APPNAME/envs/ | xargs -r etcdctl rm || true
etcdctl rmdir /paus/users/$PAUS_USER/apps/$PAUS_APPNAME/envs || true

etcdctl rmdir /paus/users/$PAUS_USER/apps/$PAUS_APPNAME || true

etcdctl mkdir /paus/users/$PAUS_USER/apps/$PAUS_APPNAME || true
etcdctl mkdir /paus/users/$PAUS_USER/apps/$PAUS_APPNAME/build-args || true
etcdctl mkdir /paus/users/$PAUS_USER/apps/$PAUS_APPNAME/deployments || true
etcdctl mkdir /paus/users/$PAUS_USER/apps/$PAUS_APPNAME/envs || true

docker ps | grep -v gitreceive | grep -v IMAGE | awk '{ print $1 }' | xargs -r docker stop | xargs -r docker rm || true
