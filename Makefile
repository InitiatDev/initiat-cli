# Makefile for Initiat CLI

.PHONY: help build test lint format clean install deps security vuln-check

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Build targets
build: ## Build the CLI binary
	@echo "🏗️  Building Initiat CLI..."
	go build -o initiat .

build-dev: ## Build development version with localhost API URL
	@echo "🔧 Building Initiat CLI (dev mode)..."
	@echo "   API URL: http://localhost:4000"
	go build \
		-ldflags "-X github.com/InitiatDev/initiat-cli/internal/config.defaultAPIBaseURL=http://localhost:4000" \
		-o initiat_dev .
	@echo "✅ Built: ./initiat_dev"

build-all: ## Build for all platforms
	@echo "🏗️  Building for all platforms..."
	./scripts/build-release.sh

install: build ## Install the CLI to /usr/local/bin
	@echo "📦 Installing initiat to /usr/local/bin..."
	sudo mv initiat /usr/local/bin/

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
	rm -f initiat initiat_dev
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

release: ## Build release binaries (usage: make release VERSION=v1.0.0)
	@if [ -z "$(VERSION)" ]; then \
		echo "❌ VERSION is required. Usage: make release VERSION=v1.0.0"; \
		exit 1; \
	fi
	@echo "🚀 Building release $(VERSION)..."
	./scripts/build-release.sh $(VERSION)

changelog: ## Update changelog for new version (usage: make changelog VERSION=v1.0.0)
	@if [ -z "$(VERSION)" ]; then \
		echo "❌ VERSION is required. Usage: make changelog VERSION=v1.0.0"; \
		exit 1; \
	fi
	@echo "📝 Updating changelog for $(VERSION)..."
	@sed -i.bak "s/## \[Unreleased\]/## [Unreleased]\n\n## [$(VERSION)] - $(shell date +%Y-%m-%d)/" CHANGELOG.md
	@rm CHANGELOG.md.bak
	@echo "✅ Changelog updated. Please review and commit changes."

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
	docker build -t initiat-cli .

docker-test: ## Test in Docker container
	@echo "🐳 Testing in Docker..."
	docker run --rm initiat-cli --help
