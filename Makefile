GO ?= go
CONFIG ?= config.example.toml

.PHONY: fmt test run build docker-build

fmt:
	$(GO) fmt ./...

test:
	$(GO) test ./...

run:
	$(GO) run . --config $(CONFIG)

build:
	mkdir -p bin
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build -o bin/multipanel .

docker-build:
	docker build -t multipanel:debug .
