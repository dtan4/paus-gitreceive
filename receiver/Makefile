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

GLIDE := glide
GLIDE_VERSION := 0.10.2

.DEFAULT_GOAL := $(BINARYDIR)/$(BINARY)

$(BINARYDIR)/$(GLIDE):
	if [ ! -d $(BINARYDIR) ]; then mkdir $(BINARYDIR); fi
ifeq ($(shell uname),Darwin)
	curl -fL https://github.com/Masterminds/glide/releases/download/$(GLIDE_VERSION)/glide-$(GLIDE_VERSION)-darwin-amd64.zip -o glide.zip
	unzip glide.zip
	mv ./darwin-amd64/glide $(BINARYDIR)/$(GLIDE)
	rm -fr ./darwin-amd64
	rm ./glide.zip
else
	curl -fL https://github.com/Masterminds/glide/releases/download/$(GLIDE_VERSION)/glide-$(GLIDE_VERSION)-linux-amd64.zip -o glide.zip
	unzip glide.zip
	mv ./linux-amd64/glide $(BINARYDIR)/$(GLIDE)
	rm -fr ./linux-amd64
	rm ./glide.zip
endif

$(BINARYDIR)/$(BINARY): $(SOURCES)
	go build $(LDFLAGS) -o $(BINARYDIR)/$(BINARY)

$(BINARYDIR)/$(BINARY)$(LINUX_AMD64_SUFFIX): $(SOURCES)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BINARYDIR)/$(BINARY)$(LINUX_AMD64_SUFFIX)

.PHONY: clean
clean:
	rm -fr $(BINARYDIR)

.PHONY: deps
deps: $(BINARYDIR)/$(GLIDE)
	$(BINARYDIR)/$(GLIDE) install

.PHONY: test
test:
	go test
