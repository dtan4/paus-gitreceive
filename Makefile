ci-docker-release: docker-build-release
	@docker login -e="$(DOCKER_QUAY_EMAIL)" -u="$(DOCKER_QUAY_USERNAME)" -p="$(DOCKER_QUAY_PASSWORD)" quay.io
	docker push quay.io/dtan4/paus-gitreceive:latest

docker-build-release:
	cd receiver; make build-linux
	docker build -f Dockerfile.release -t quay.io/dtan4/paus-gitreceive:latest .

.PHONY: ci-docker-release docker-build-release
