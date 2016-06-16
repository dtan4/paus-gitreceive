# paus-gitreceive
[![Build Status](https://travis-ci.org/dtan4/paus-gitreceive.svg?branch=master)](https://travis-ci.org/dtan4/paus-gitreceive)
[![Docker Repository on Quay](https://quay.io/repository/dtan4/paus-gitreceive/status "Docker Repository on Quay")](https://quay.io/repository/dtan4/paus-gitreceive)

Git server of [Paus](https://github.com/dtan4/paus)

## What is this?

paus-gitreceive does:

- Receive `git push` and extract commit metadata
- Build Docker image from `Dockerfile` in the repository
- Deploy application using [Docker Compose](https://docs.docker.com/compose/)
- Register application metadata and [Vulcand](https://github.com/vulcand/vulcand) routing information in etcd

## Environment variables

| Key                  | Required | Description                                    | Default                 | Example                 |
|----------------------|----------|------------------------------------------------|-------------------------|-------------------------|
| `PAUS_BASE_DOMAIN`   | Required | Base domain for application URL                |                         | `pausapp.com`           |
| `PAUS_DOCKER_HOST` |          | Endpoint of Docker daemon                       | `tcp://127.0.0.1:2375` | `tcp://127.0.0.1:2377` (Docker Swarm) |
| `PAUS_ETCD_ENDPOINT` |          | Endpoint of etcd cluster                       | `http://127.0.0.1:2379` | `http://127.0.0.1:2379` |
| `PAUS_REPOSITORY_DIR`    |          | Directory to store repository files | `/repos`                   | `/repos`                  |
| `PAUS_URI_SCHEME`        |          | URI scheme of application URL (`http`&#124;`https`) | `http`     | `http`                    |

## Development

### Build receiver

```bash
$ cd receiver
$ make deps
$ make
```

### Run on local using Docker Compose

```bash
# Set these as you like
$ export PAUS_USER=dtan4
$ export PAUS_APPNAME=docker-service-rails
$ export SSH_PUBLIC_KEY=~/.ssh/id_rsa.github.pub

$ make compose-up

$ vim ~/.ssh/config
$ cat ~/.ssh/config
Host pausapp.com
     Hostname 127.0.0.1
     User git
     Port 2222
     IdentityFile ~/.ssh/id_rsa.github

# at your application repository
$ git remote add paus git@pausapp.com:dtan4/docker-service-rails
$ git push paus master
```
