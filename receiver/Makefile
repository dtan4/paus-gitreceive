VERSION := 0.1.0
REVISION := $(shell git rev-parse --short HEAD)
BUILDTIME := $(shell date '+%Y/%m/%d %H:%M:%S %Z')
GOVERSION := $(subst go version ,,$(shell go version))

BINARY := receiver
LINUX_AMD64_SUFFIX := _linux-amd64

SOURCES := $(shell find . -name '*.go' -type f)

LDFLAGS := -ldflags="-w -X \"main.Version=$(VERSION)\" -X \"main.Revision=$(REVISION)\" -X \"main.BuildTime=$(BUILDTIME)\" -X \"main.GoVersion=$(GOVERSION)\""

GLIDE_VERSION := 0.11.1

.DEFAULT_GOAL := bin/$(BINARY)

glide:
ifeq ($(shell uname),Darwin)
	curl -fL https://github.com/Masterminds/glide/releases/download/v$(GLIDE_VERSION)/glide-v$(GLIDE_VERSION)-darwin-amd64.zip -o glide.zip
	unzip glide.zip
	mv ./darwin-amd64/glide ./glide
	rm -fr ./darwin-amd64
	rm ./glide.zip
else
	curl -fL https://github.com/Masterminds/glide/releases/download/v$(GLIDE_VERSION)/glide-v$(GLIDE_VERSION)-linux-amd64.zip -o glide.zip
	unzip glide.zip
	mv ./linux-amd64/glide ./glide
	rm -fr ./linux-amd64
	rm ./glide.zip
endif

bin/$(BINARY): deps $(SOURCES)
	go build $(LDFLAGS) -o bin/$(BINARY)

bin/$(BINARY)$(LINUX_AMD64_SUFFIX): deps $(SOURCES)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY)$(LINUX_AMD64_SUFFIX)

.PHONY: clean
clean:
	rm -fr bin/*
	rm -fr vendor/*

.PHONY: deps
deps: glide
	./glide install

.PHONY: test
test:
	go test -v . ./model
