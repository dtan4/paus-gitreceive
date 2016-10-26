NAME := receiver
VERSION := 0.1.0
REVISION := $(shell git rev-parse --short HEAD)
GOVERSION := $(subst go version ,,$(shell go version))

LINUX_AMD64_SUFFIX := _linux-amd64

LDFLAGS := -ldflags="-s -w -X \"main.Version=$(VERSION)\" -X \"main.Revision=$(REVISION)\" -X \"main.GoVersion=$(GOVERSION)\""

.DEFAULT_GOAL := bin/$(NAME)

bin/$(NAME): deps
	go build $(LDFLAGS) -o bin/$(NAME)

bin/$(NAME)$(LINUX_AMD64_SUFFIX): deps
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(NAME)$(LINUX_AMD64_SUFFIX)

.PHONY: clean
clean:
	rm -fr bin/*
	rm -fr vendor/*

.PHONY: deps
deps: glide
	glide install

.PHONY: glide
glide:
ifeq ($(shell command -v glide 2> /dev/null),)
	curl https://glide.sh/get | sh
endif

.PHONY: test
test:
	go test -cover -v `glide novendor`

.PHONY: update-deps
update-deps: glide
	glide update
