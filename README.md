# paus-gitreceive

[![Docker Repository on Quay](https://quay.io/repository/dtan4/paus-gitreceive/status "Docker Repository on Quay")](https://quay.io/repository/dtan4/paus-gitreceive)

Git server of [Paus](https://github.com/dtan4/paus)

## What is this?

paus-gitreceive does:

- Receive `git push` and extract commit metadata
- Build Docker image from `Dockerfile` in the repository
- Deploy application using [Docker Compose](https://docs.docker.com/compose/)
- Register application metadata and [Vulcand](https://github.com/vulcand/vulcand) routing information in etcd
