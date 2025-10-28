.PHONY: build test clean install help version

# Base semantic version; build metadata is a zero-padded counter
BASE_VERSION := v0.2.0
# Compute build number from commits that touched the current folder (works from repo root or this dir)
BUILD_NUM := $(shell (git rev-list --count HEAD -- . 2>/dev/null || echo 0) | awk '{printf "%04d", $$1}')
SPACECTL_VERSION := $(BASE_VERSION)-$(BUILD_NUM)
LDFLAGS := -X 'spacectl/internal/version.Version=$(SPACECTL_VERSION)'

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the spacectl binary
	@echo "Building spacectl $(SPACECTL_VERSION)"
	go build -ldflags "$(LDFLAGS)" -o bin/spacectl main.go

install: ## Install spacectl to $GOPATH/bin
	go install -ldflags "$(LDFLAGS)"

test: ## Run Go unit tests
	@echo "Running unit tests"
	go test -v ./...

clean: ## Clean build artifacts
	rm -rf bin/

deps: ## Install dependencies
	go mod download
	go mod tidy

lint: ## Run linter
	golangci-lint run

run: ## Run spacectl with arguments (use: make run ARGS="--help")
	go run -ldflags "$(LDFLAGS)" main.go $(ARGS)

# Development targets
dev-setup: ## Setup development environment
	go mod download
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Cross-compilation targets
build-linux: ## Build for Linux
	@echo "Building spacectl-linux $(SPACECTL_VERSION)"
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/spacectl-linux main.go

build-darwin: ## Build for macOS
	@echo "Building spacectl-darwin $(SPACECTL_VERSION)"
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/spacectl-darwin main.go

build-windows: ## Build for Windows
	@echo "Building spacectl-windows $(SPACECTL_VERSION)"
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/spacectl-windows.exe main.go

build-all: ## Build for all platforms
	make build-linux
	make build-darwin
	make build-windows

version: ## Print computed version
	@echo $(SPACECTL_VERSION)

setup-completion: ## Setup shell autocompletion
	@./scripts/setup-completion.sh
