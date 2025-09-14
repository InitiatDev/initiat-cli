# Makefile for init.Flow CLI

.PHONY: help build test lint format clean install deps security vuln-check

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Build targets
build: ## Build the CLI binary
	@echo "🏗️  Building init.Flow CLI..."
	go build -o initflow .

build-all: ## Build for all platforms
	@echo "🏗️  Building for all platforms..."
	./scripts/build-release.sh

install: build ## Install the CLI to /usr/local/bin
	@echo "📦 Installing initflow to /usr/local/bin..."
	sudo mv initflow /usr/local/bin/

# Development targets
deps: ## Download and verify dependencies
	@echo "📦 Downloading dependencies..."
	go mod download
	go mod verify

test: ## Run tests
	@echo "🧪 Running tests..."
	go test -v -race -coverprofile=coverage.out ./...

test-coverage: test ## Run tests and show coverage
	@echo "📊 Test coverage:"
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Code quality targets
lint: ## Run linter
	@echo "🔍 Running linter..."
	golangci-lint run

lint-fix: ## Run linter with auto-fix
	@echo "🔧 Running linter with auto-fix..."
	golangci-lint run --fix

format: ## Format code
	@echo "🎨 Formatting code..."
	gofmt -s -w .
	goimports -w .

format-check: ## Check if code is formatted
	@echo "🎨 Checking code formatting..."
	@if [ "$$(gofmt -s -l . | wc -l)" -gt 0 ]; then \
		echo "❌ Code is not formatted. Run 'make format' to fix."; \
		gofmt -s -l .; \
		exit 1; \
	else \
		echo "✅ Code is properly formatted."; \
	fi

# Security targets
security: ## Run security scan
	@echo "🔒 Running security scan..."
	@if ! command -v gosec >/dev/null 2>&1; then \
		echo "Installing gosec..."; \
		go install github.com/securego/gosec/v2/cmd/gosec@latest; \
	fi
	gosec ./...

vuln-check: ## Check for vulnerabilities
	@echo "🛡️  Checking for vulnerabilities..."
	govulncheck ./...

# Utility targets
clean: ## Clean build artifacts
	@echo "🧹 Cleaning build artifacts..."
	rm -f initflow
	rm -rf dist/
	rm -f coverage.out coverage.html

tidy: ## Tidy go modules
	@echo "🧹 Tidying go modules..."
	go mod tidy

# CI targets (run all checks)
ci: deps format-check lint test security vuln-check build ## Run all CI checks locally

# Development workflow
dev: deps format lint test build ## Quick development workflow

# Release targets
release-test: ## Test release build process
	@echo "🚀 Testing release build..."
	./scripts/build-release.sh test

# Tool installation targets
install-tools: ## Install development tools
	@echo "🔧 Installing development tools..."
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest
	@echo "✅ All development tools installed successfully!"

# Docker targets (if you want to add Docker support later)
docker-build: ## Build Docker image
	@echo "🐳 Building Docker image..."
	docker build -t initflow-cli .

docker-test: ## Test in Docker container
	@echo "🐳 Testing in Docker..."
	docker run --rm initflow-cli --help
