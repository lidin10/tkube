BINARY_NAME=tkube
VERSION=1.0.0

.PHONY: build clean install test

build:
	go mod tidy
	go build -ldflags="-s -w" -o $(BINARY_NAME) .

build-all:
	go mod tidy
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o $(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o $(BINARY_NAME)-darwin-arm64 .
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $(BINARY_NAME)-linux-amd64 .

install: build
	cp $(BINARY_NAME) /usr/local/bin/

clean:
	rm -f $(BINARY_NAME) $(BINARY_NAME)-*

test:
	go test -v ./...

.DEFAULT_GOAL := build