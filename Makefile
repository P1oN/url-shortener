SHELL := /bin/sh

.PHONY: help run migrate-up migrate-down migrate-version build test test-integration migrate-up-integration openapi-generate openapi-check

help:
	@printf "Targets:\n"
	@printf "  run                Run API server\n"
	@printf "  migrate-up         Apply all migrations\n"
	@printf "  migrate-down       Roll back all migrations\n"
	@printf "  migrate-version    Migrate to specific version (VERSION=2)\n"
	@printf "  build              Build server and migrate binaries\n"
	@printf "  test               Run unit tests\n"
	@printf "  test-integration   Run integration tests\n"
	@printf "  migrate-up-integration   Run migration integration test\n"
	@printf "  openapi-generate   Generate OpenAPI server contract code\n"
	@printf "  openapi-check      Verify OpenAPI generated files are up to date\n"

run:
	go run ./cmd/server

migrate-up:
	go run ./cmd/migrate --up

migrate-down:
	go run ./cmd/migrate --down

migrate-version:
	@if [ -z "$$VERSION" ]; then printf "VERSION is required\n"; exit 1; fi
	go run ./cmd/migrate --version $$VERSION

build:
	go build ./cmd/server
	go build ./cmd/migrate

test:
	go test ./...

test-integration:
	INTEGRATION_TESTS=1 go test ./...

migrate-up-integration:
	INTEGRATION_TESTS=1 go test ./internal/repo/postgres -run TestMigrations_Up

openapi-generate:
	go run ./cmd/openapi-gen

openapi-check:
	$(MAKE) openapi-generate
	git diff --exit-code
