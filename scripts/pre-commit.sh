#!/bin/bash

# Pre-commit hook for Initiat CLI
# This script runs formatting, linting, and tests before allowing a commit

set -e

echo "🔍 Running pre-commit checks..."

# Check if we're in a git repository
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    echo "❌ Not in a git repository"
    exit 1
fi

# Function to print colored output
print_status() {
    local status=$1
    local message=$2
    if [ "$status" = "success" ]; then
        echo "✅ $message"
    elif [ "$status" = "error" ]; then
        echo "❌ $message"
    elif [ "$status" = "info" ]; then
        echo "ℹ️  $message"
    fi
}

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check for required tools
print_status "info" "Checking for required tools..."

if ! command_exists go; then
    print_status "error" "Go is not installed"
    exit 1
fi

if ! command_exists golangci-lint; then
    print_status "info" "Installing golangci-lint..."
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
fi

if ! command_exists goimports; then
    print_status "info" "Installing goimports..."
    go install golang.org/x/tools/cmd/goimports@latest
fi

# 1. Format code
print_status "info" "Formatting code..."
if gofmt -s -w . && goimports -w .; then
    print_status "success" "Code formatted"
else
    print_status "error" "Failed to format code"
    exit 1
fi

# 2. Check for formatting issues
print_status "info" "Checking code formatting..."
if [ "$(gofmt -s -l . | wc -l)" -gt 0 ]; then
    print_status "error" "Code is not properly formatted:"
    gofmt -s -l .
    exit 1
else
    print_status "success" "Code is properly formatted"
fi

# 3. Tidy modules
print_status "info" "Tidying Go modules..."
if go mod tidy; then
    print_status "success" "Go modules tidied"
else
    print_status "error" "Failed to tidy Go modules"
    exit 1
fi

# 4. Run linter
print_status "info" "Running linter..."
if golangci-lint run; then
    print_status "success" "Linting passed"
else
    print_status "error" "Linting failed"
    exit 1
fi

# 5. Run tests
print_status "info" "Running tests..."
if go test -race ./...; then
    print_status "success" "All tests passed"
else
    print_status "error" "Tests failed"
    exit 1
fi

# 6. Check for security issues (optional, can be slow)
if command_exists gosec; then
    print_status "info" "Running security scan..."
    if gosec -quiet ./...; then
        print_status "success" "Security scan passed"
    else
        print_status "error" "Security issues found"
        exit 1
    fi
fi

# 7. Build check
print_status "info" "Testing build..."
if go build -o /tmp/initiat-test .; then
    rm -f /tmp/initiat-test
    print_status "success" "Build successful"
else
    print_status "error" "Build failed"
    exit 1
fi

# 8. Add any formatted files to git
git add -A

print_status "success" "All pre-commit checks passed! 🎉"
echo ""
echo "Ready to commit! 🚀"
