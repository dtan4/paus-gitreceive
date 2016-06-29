DOCKER_REPOSITORY := quay.io
DOCKER_IMAGE_NAME := $(DOCKER_REPOSITORY)/dtan4/paus-gitreceive
DOCKER_IMAGE_TAG := latest
DOCKER_IMAGE := $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)

PAUS_USER ?= test
PAUS_APPNAME ?= test
SSH_PUBLIC_KEY ?= ~/.ssh/id_rsa.pub
DOCKER_HOST ?= localhost

.DEFAULT_GOAL := docker-release-build

.PHONY: ci-docker-release
ci-docker-release: docker-release-build
	@docker login -e="$(DOCKER_QUAY_EMAIL)" -u="$(DOCKER_QUAY_USERNAME)" -p="$(DOCKER_QUAY_PASSWORD)" $(DOCKER_REPOSITORY)
	docker push $(DOCKER_IMAGE)

.PHONY: compose-up
compose-up: compose-stop
	$(MAKE) -C receiver bin/receiver_linux-amd64
	docker-compose build gitreceive gitreceive-upload-key
	docker-compose up -d etcd gitreceive
	@docker-compose run --rm gitreceive-upload-key $(PAUS_USER) "$(shell cat $(SSH_PUBLIC_KEY))"
	etcdctl --endpoint "http://$(DOCKER_HOST):2379" mkdir /paus/users/$(PAUS_USER)/apps/$(PAUS_APPNAME) || true
	etcdctl --endpoint "http://$(DOCKER_HOST):2379" mkdir /paus/users/$(PAUS_USER)/apps/$(PAUS_APPNAME)/build-args || true
	etcdctl --endpoint "http://$(DOCKER_HOST):2379" mkdir /paus/users/$(PAUS_USER)/apps/$(PAUS_APPNAME)/deployments || true
	etcdctl --endpoint "http://$(DOCKER_HOST):2379" mkdir /paus/users/$(PAUS_USER)/apps/$(PAUS_APPNAME)/envs || true

.PHONY: compose-stop
compose-stop:
	docker-compose stop
	docker-compose rm -f

.PHONY: docker-push
docker-push:
	docker push $(DOCKER_IMAGE)

.PHONY: docker-release-build
docker-release-build:
	$(MAKE) -C receiver bin/receiver_linux-amd64
	docker build -f Dockerfile.release -t $(DOCKER_IMAGE) .
