default: build

docker-build-release:
	cd receiver; make build-linux
	docker build -f Dockerfile.release -t quay.io/dtan4/paus-gitreceive:latest .
