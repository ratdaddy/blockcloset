SHELL := /bin/bash
BIN_DIR := bin
BIN_NAME := gantry
DATA_DIR := data
DEV_DB_FILE := $(DATA_DIR)/dev.db
TEST_DB_FILE := $(DATA_DIR)/test.db
MIGRATE_DB_FILES := $(DEV_DB_FILE) $(TEST_DB_FILE)
.PHONY: run dev test testdev build migrate rollback migrate-status

run:
	@go run ./cmd/gantry

dev:
	@find . -name '*.go' ! -name '*_test.go' | ENABLE_REFLECTION=true entr -r go run ./cmd/gantry

test:
	go test ./...

testdev:
	@find . -type f -name '*.go' '!' -path './bin/*' \
	| entr -c sh -c 'go test -count=1 ./...; date "+%Y/%m/%d %H:%M:%S"'

build:
	go build -o $(BIN_DIR)/$(BIN_NAME) ./cmd/gantry

migrate: | $(DATA_DIR)
	@for db in $(MIGRATE_DB_FILES); do \
		echo "migrating $$db"; \
		migrate -verbose -path migrations -database sqlite3://$(PWD)/$$db up; \
	done

rollback: | $(DATA_DIR)
	@for db in $(MIGRATE_DB_FILES); do \
		echo "rollback $$db"; \
		migrate -verbose -path migrations -database sqlite3://$(PWD)/$$db down 1; \
	done

migrate-status: | $(DATA_DIR)
	@for db in $(MIGRATE_DB_FILES); do \
		echo "status $$db"; \
		migrate -verbose -path migrations -database sqlite3://$(PWD)/$$db version; \
	done

$(DATA_DIR):
	@mkdir -p $(DATA_DIR)
