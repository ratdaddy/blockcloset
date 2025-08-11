SHELL := /bin/bash
BIN_DIR := bin
BIN_NAME := gateway

.PHONY: run dev test testdev build
run:
	go run ./cmd/gateway

dev:
	@find . -name '*.go' ! -name '*_test.go' | entr -r go run ./cmd/gateway

test:
	go test ./...

testdev:
	@find . -type f -name '*.go' '!' -path './bin/*' \
	| entr -c sh -c 'go test -count=1 ./...; date "+%Y/%m/%d %H:%M:%S"'

build:
	go build -o $(BIN_DIR)/$(BIN_NAME) ./cmd/gateway
