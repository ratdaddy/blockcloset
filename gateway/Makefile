SHELL := /bin/bash
BIN_DIR := bin
BIN_NAME := gateway

.PHONY: run test build
run:
	go run ./cmd/gateway

test:
	go test ./cmd/gateway -v

build:
	go build -o $(BIN_DIR)/$(BIN_NAME) ./cmd/gateway
