VERSION := 0.1.0
REVISION := $(shell git rev-parse --short HEAD)
BUILDTIME := $(shell date '+%Y/%m/%d %H:%M:%S %Z')
GOVERSION := $(subst go version ,,$(shell go version))

BINARY := receiver
BINARY_DIR := bin

LDFLAGS := -ldflags="-w -X \"main.Version=$(VERSION)\" -X \"main.Revision=$(REVISION)\" -X \"main.BuildTime=$(BUILDTIME)\" -X \"main.GoVersion=$(GOVERSION)\""

default: build

build: clean
	go build $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY)

build-linux: clean
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY)_linux-amd64

clean:
	rm -fr $(BINARY_DIR)

deps:
	go get github.com/Masterminds/glide
	glide install

test:
	go test

.PHONY: build build-linux clean deps test
