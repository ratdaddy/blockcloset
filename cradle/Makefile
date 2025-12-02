SHELL := /bin/bash
BIN_DIR := bin
BIN_NAME := cradle
OBJECTS_ROOT := data/cradle
.PHONY: run dev test testdev build

run:
	@go run ./cmd/cradle

dev:
	@find . -name '*.go' ! -name '*_test.go' | ENABLE_REFLECTION=true entr -r go run ./cmd/cradle

test:
	go test ./...

testdev:
	@find . -type f -name '*.go' '!' -path './bin/*' \
	| entr -c sh -c 'go test -count=1 ./...; date "+%Y/%m/%d %H:%M:%S"'

build:
	go build -o $(BIN_DIR)/$(BIN_NAME) ./cmd/cradle

$(OBJECTS_ROOT):
	@mkdir -p $(OBJECTS_ROOT)
