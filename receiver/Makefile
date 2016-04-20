default: build

build: clean
	go build -o bin/receiver

build-linux: clean
	GOOS=linux GOARCH=amd64 go build -o bin/receiver-linux_amd64

clean:
	rm -f bin/receiver
	rm -f bin/receiver-linux_amd64

deps:
	go get github.com/Masterminds/glide
	glide install

test:
	go test

.PHONY: build build-linux clean deps test
