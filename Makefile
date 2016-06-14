DOCKER_REPOSITORY := quay.io
DOCKER_IMAGE_NAME := $(DOCKER_REPOSITORY)/dtan4/paus-gitreceive
DOCKER_IMAGE_TAG := latest
DOCKER_IMAGE = $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)

.PHONY: ci-docker-release
ci-docker-release: docker-release-build
	@docker login -e="$(DOCKER_QUAY_EMAIL)" -u="$(DOCKER_QUAY_USERNAME)" -p="$(DOCKER_QUAY_PASSWORD)" $(DOCKER_REPOSITORY)
	docker push $(DOCKER_IMAGE)

.PHONY: docker-push
docker-push:
	docker push $(DOCKER_IMAGE)

.PHONY: docker-release-build
docker-release-build:
	$(MAKE) -C receiver bin/receiver_linux-amd64
	docker build -f Dockerfile.release -t $(DOCKER_IMAGE) .
