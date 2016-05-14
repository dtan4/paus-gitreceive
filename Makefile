DOCKER_IMAGE_NAME := quay.io/dtan4/paus-gitreceive
DOCKER_IMAGE_TAG := latest
DOCKER_IMAGE = $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)

ci-docker-release: docker-release-build
	@docker login -e="$(DOCKER_QUAY_EMAIL)" -u="$(DOCKER_QUAY_USERNAME)" -p="$(DOCKER_QUAY_PASSWORD)" quay.io
	docker push $(DOCKER_IMAGE)

docker-push:
	docker push $(DOCKER_IMAGE)

docker-release-build:
	cd receiver; make build-linux
	docker build -f Dockerfile.release -t $(DOCKER_IMAGE) .

.PHONY: ci-docker-release docker-release-build docker-push
