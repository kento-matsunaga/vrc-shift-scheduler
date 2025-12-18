# Makefile for VRC Shift Scheduler Backend
#
# Target name conventions:
#   public_target:   A public target, which should be used by human and/or other tools.
#                    A public target should have a comment starting with double pound sign (##)
#   _private_target: A private target, which should be used only by other target in this Makefile.
#                    A private target should have a comment starting with single pound sign (#)

SHELL := /bin/bash

# ============================================================
# Default
# ============================================================

## Runs default target ('help').
.PHONY: default
default: help

## Prints a help message
.PHONY: help
help:
	@echo "VRC Shift Scheduler - Backend Makefile"
	@echo ""
	@echo "Usage: make <target>"
	@echo ""
	@echo "Targets:"
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /'
	@echo ""

# ============================================================
# Development - Server
# ============================================================

## Start the development server
.PHONY: dev
dev:
	go run ./cmd/server/main.go

## Build the server binary
.PHONY: build
build:
	go build -o bin/server ./cmd/server/main.go

# ============================================================
# Database - Migration
# ============================================================

## Run database migrations
.PHONY: migrate
migrate:
	go run ./cmd/migrate/main.go

## Run database seed (insert initial data)
.PHONY: seed
seed:
	go run ./cmd/seed/main.go

## Reset database (drop all tables and re-migrate)
.PHONY: db-reset
db-reset:
	@echo "Resetting database..."
	go run ./cmd/migrate/main.go --reset
	@echo "Database reset complete."

# ============================================================
# Testing
# ============================================================

## Run all tests
.PHONY: test
test:
	go test ./...

## Run tests with verbose output
.PHONY: test-v
test-v:
	go test -v ./...

## Run tests with coverage
.PHONY: test-cov
test-cov:
	go test -cover ./...

## Run tests with coverage report (HTML)
.PHONY: test-cov-html
test-cov-html:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## Run integration tests (requires database)
.PHONY: test-integration
test-integration:
	go test -tags=integration ./internal/infra/db/...

# ============================================================
# Linting & Formatting
# ============================================================

## Run linter (golangci-lint)
.PHONY: lint
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint is not installed. Skipping lint."; \
		echo "Install: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

## Run linter with auto-fix
.PHONY: lint-fix
lint-fix:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --fix; \
	else \
		echo "golangci-lint is not installed."; \
		echo "Install: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

## Format code with go fmt
.PHONY: fmt
fmt:
	go fmt ./...

## Run go vet
.PHONY: vet
vet:
	go vet ./...

# ============================================================
# Dependencies
# ============================================================

## Download dependencies
.PHONY: deps
deps:
	go mod download

## Tidy dependencies
.PHONY: tidy
tidy:
	go mod tidy

# ============================================================
# Docker (for use inside container or with docker compose exec)
# ============================================================

## Run migrations in Docker container
.PHONY: docker-migrate
docker-migrate:
	docker compose exec backend go run ./cmd/migrate/main.go

## Run tests in Docker container
.PHONY: docker-test
docker-test:
	docker compose exec backend go test ./...

## Run seed in Docker container
.PHONY: docker-seed
docker-seed:
	docker compose exec backend go run ./cmd/seed/main.go

# ============================================================
# Utilities
# ============================================================

## Check Go version
.PHONY: check-go
check-go:
	@echo "Go version:"
	@go version
	@echo ""
	@echo "Required: Go 1.23+"

## Clean build artifacts
.PHONY: clean
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

## Run all checks (fmt, vet, lint, test)
.PHONY: check-all
check-all: fmt vet lint test
	@echo "All checks passed!"




