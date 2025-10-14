SHELL := /bin/bash
BIN_DIR := bin
BIN_NAME := flatbed

.PHONY: run dev test testdev build docker-build

run:
	@go run ./cmd/flatbed

dev:
	@find . -name '*.go' ! -name '*_test.go' | entr -r go run ./cmd/flatbed

test:
	go test ./...

testdev:
	@find . -type f -name '*.go' '!' -path './bin/*' \
	| entr -c sh -c 'go test -count=1 ./...; date "+%Y/%m/%d %H:%M:%S"'

build:
	go build -o $(BIN_DIR)/$(BIN_NAME) ./cmd/flatbed

IMAGE ?= flatbed
TAG	 ?= dev
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  ?= $(shell git rev-parse --short HEAD)
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

docker-build:
	DOCKER_BUILDKIT=1 docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		--build-arg DATE=$(DATE) \
		-t $(IMAGE):$(TAG) .
