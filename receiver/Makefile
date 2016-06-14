VERSION := 0.1.0
REVISION := $(shell git rev-parse --short HEAD)
BUILDTIME := $(shell date '+%Y/%m/%d %H:%M:%S %Z')
GOVERSION := $(subst go version ,,$(shell go version))

BINARYDIR := bin
BINARY := receiver
LINUX_AMD64_SUFFIX := _linux-amd64

SOURCEDIR := .
SOURCES := $(shell find $(SOURCEDIR) -name '*.go' -type f)

LDFLAGS := -ldflags="-w -X \"main.Version=$(VERSION)\" -X \"main.Revision=$(REVISION)\" -X \"main.BuildTime=$(BUILDTIME)\" -X \"main.GoVersion=$(GOVERSION)\""

.DEFAULT_GOAL := $(BINARYDIR)/$(BINARY)

$(BINARYDIR)/$(BINARY): $(SOURCES)
	go build $(LDFLAGS) -o $(BINARYDIR)/$(BINARY)

$(BINARYDIR)/$(BINARY)$(LINUX_AMD64_SUFFIX): $(SOURCES)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BINARYDIR)/$(BINARY)$(LINUX_AMD64_SUFFIX)

.PHONY: clean
clean:
	rm -fr $(BINARYDIR)

.PHONY: deps
deps:
	go get github.com/Masterminds/glide
	glide install

.PHONY: test
test:
	go test
