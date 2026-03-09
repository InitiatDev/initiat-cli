# Makefile for Initiat CLI

.PHONY: help build build-dev build-all install deps test test-coverage lint lint-fix format format-check security vuln-check clean tidy ci dev release-test release changelog install-tools docker-build docker-test cgo-check

# Tree-sitter grammars + go-tree-sitter require CGO.
CGO_ENABLED ?= 1
GO ?= env CGO_ENABLED=$(CGO_ENABLED) go

cgo-check: ## Fail fast if CGO/toolchain missing
	@if [ "$(CGO_ENABLED)" != "1" ]; then \
		echo "❌ CGO must be enabled for this project (Tree-sitter grammars require CGO)."; \
		echo "   Set CGO_ENABLED=1 (e.g. 'make ci CGO_ENABLED=1')."; \
		exit 1; \
	fi
	@if ! command -v cc >/dev/null 2>&1 && ! command -v gcc >/dev/null 2>&1 && ! command -v clang >/dev/null 2>&1; then \
		echo "❌ No C compiler found (cc/gcc/clang)."; \
		echo "   Install a compiler toolchain (e.g. build-essential / build-base) and rerun."; \
		exit 1; \
	fi

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Build targets
build: cgo-check ## Build the CLI binary
	@echo "🏗️  Building Initiat CLI..."
	$(GO) build -o initiat .

build-dev: cgo-check ## Build development version with localhost API URL
	@echo "🔧 Building Initiat CLI (dev mode)..."
	@echo "   API URL: http://localhost:4000"
	$(GO) build -tags dev -o initiat_dev .
	@echo "✅ Built: ./initiat_dev"

build-all: ## Build release binaries (native platform by default)
	@echo "🏗️  Building release binaries..."
	./scripts/build-release.sh

install: build ## Install the CLI to /usr/local/bin
	@echo "📦 Installing initiat to /usr/local/bin..."
	sudo mv initiat /usr/local/bin/

# Development targets
deps: ## Download and verify dependencies
	@echo "📦 Downloading dependencies..."
	$(GO) mod download
	$(GO) mod verify

test: cgo-check ## Run tests
	@echo "🧪 Running tests..."
	$(GO) test -race -coverprofile=coverage.out ./...

test-coverage: test ## Run tests and show coverage
	@echo "📊 Test coverage:"
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Code quality targets
lint: ## Run linter
	@echo "🔍 Running linter..."
	env CGO_ENABLED=$(CGO_ENABLED) golangci-lint run

lint-fix: ## Run linter with auto-fix
	@echo "🔧 Running linter with auto-fix..."
	env CGO_ENABLED=$(CGO_ENABLED) golangci-lint run --fix

format: ## Format code
	@echo "🎨 Formatting code..."
	@dirs="$$( $(GO) list -f '{{.Dir}}' ./... )"; \
	gofmt -s -w $$dirs; \
	goimports -w $$dirs

format-check: ## Check if code is formatted
	@echo "🎨 Checking code formatting..."
	@dirs="$$( $(GO) list -f '{{.Dir}}' ./... )"; \
	if [ "$$(gofmt -s -l $$dirs | wc -l)" -gt 0 ]; then \
		echo "❌ Code is not formatted. Run 'make format' to fix."; \
		gofmt -s -l $$dirs; \
		exit 1; \
	else \
		echo "✅ Code is properly formatted."; \
	fi

# Security targets
security: ## Run security scan
	@echo "🔒 Running security scan..."
	@if ! command -v gosec >/dev/null 2>&1; then \
		echo "Installing gosec..."; \
		$(GO) install github.com/securego/gosec/v2/cmd/gosec@latest; \
	fi
	@out=$$(mktemp); \
	if ! gosec -quiet \
		-exclude-dir=docs \
		-exclude-dir=internal/codeanalysis/testdata \
		./... >"$$out" 2>&1; then \
		cat "$$out"; \
		rm -f "$$out"; \
		exit 1; \
	fi; \
	rm -f "$$out"

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
	$(GO) mod tidy

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
	$(GO) install golang.org/x/tools/cmd/goimports@latest
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GO) install github.com/securego/gosec/v2/cmd/gosec@latest
	$(GO) install golang.org/x/vuln/cmd/govulncheck@latest
	@echo "✅ All development tools installed successfully!"

# Docker targets (if you want to add Docker support later)
docker-build: ## Build Docker image
	@echo "🐳 Building Docker image..."
	docker build -t initiat-cli .

docker-test: ## Test in Docker container
	@echo "🐳 Testing in Docker..."
	docker run --rm initiat-cli --help
