.PHONY: build test lint clean install cross-compile help
.DEFAULT_GOAL := help

# Variables
APP_NAME := specify
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.1.0")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

# Go build flags
LDFLAGS := -s -w -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.commit=$(COMMIT)
BUILD_FLAGS := -ldflags "$(LDFLAGS)" -trimpath

build: ## Build the application for current platform
	@echo "Building $(APP_NAME) for current platform..."
	go build $(BUILD_FLAGS) -o bin/$(APP_NAME) ./cmd/specify

test: ## Run all tests
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...

test-contract: ## Run contract tests only
	@echo "Running contract tests..."
	go test -v -race ./tests/contract/...

test-integration: ## Run integration tests only
	@echo "Running integration tests..."
	go test -v -race ./tests/integration/...

test-unit: ## Run unit tests only
	@echo "Running unit tests..."
	go test -v -race ./tests/unit/...

lint: ## Run linter
	@echo "Running linter..."
	golangci-lint run

lint-fix: ## Run linter with auto-fix
	@echo "Running linter with auto-fix..."
	golangci-lint run --fix

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out

install: build ## Install the application locally
	@echo "Installing $(APP_NAME)..."
	cp bin/$(APP_NAME) $(GOPATH)/bin/$(APP_NAME)

cross-compile: ## Build for all platforms
	@echo "Cross-compiling for all platforms..."
	@mkdir -p bin
	@for platform in $(PLATFORMS); do \
		OS=$$(echo $$platform | cut -d'/' -f1); \
		ARCH=$$(echo $$platform | cut -d'/' -f2); \
		OUTPUT=bin/$(APP_NAME)-$$OS-$$ARCH; \
		if [ "$$OS" = "windows" ]; then OUTPUT=$$OUTPUT.exe; fi; \
		echo "Building for $$OS/$$ARCH..."; \
		GOOS=$$OS GOARCH=$$ARCH go build $(BUILD_FLAGS) -o $$OUTPUT ./cmd/specify; \
	done

deps: ## Download and tidy dependencies
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

fmt: ## Format Go code
	@echo "Formatting code..."
	go fmt ./...
	goimports -w .

vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

check: fmt vet lint test ## Run all checks (format, vet, lint, test)

bench: ## Run benchmarks
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

coverage: test ## Generate and view coverage report
	@echo "Generating coverage report..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

release: clean cross-compile ## Build release binaries
	@echo "Creating release artifacts..."
	@for binary in bin/*; do \
		if [ -f "$$binary" ]; then \
			echo "Compressing $$binary..."; \
			tar -czf "$$binary.tar.gz" -C bin "$$(basename $$binary)"; \
		fi; \
	done

dev: ## Start development mode with hot reload (requires air)
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "air not found. Install with: go install github.com/cosmtrek/air@latest"; \
		exit 1; \
	fi

help: ## Show this help message
	@echo "$(APP_NAME) - Makefile help"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "Version: $(VERSION)"
	@echo "Build time: $(BUILD_TIME)"
	@echo "Commit: $(COMMIT)"