BINARY := receiver
BINARY_DIR := bin

default: build

build: clean
	go build -o $(BINARY_DIR)/$(BINARY)

build-linux: clean
	GOOS=linux GOARCH=amd64 go build -o $(BINARY_DIR)/$(BINARY)_linux-amd64

clean:
	rm -fr $(BINARY_DIR)

deps:
	go get github.com/Masterminds/glide
	glide install

test:
	go test

.PHONY: build build-linux clean deps test
