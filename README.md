# Initiat CLI

A developer experience platform that accelerates team onboarding and reduces time-to-first-commit from days to minutes. The CLI provides secure secret sharing, environment setup, and guided onboarding workflows. Built with Go for cross-platform compatibility and security.

[![Go Version](https://img.shields.io/badge/Go-1.25+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

## 🚀 Quick Start

### Prerequisites
1. **Create an Initiat account** at [initiat.com](https://initiat.com)
2. **Set your password** during account creation or in your profile settings

### Installation & Login

#### Option 1: Download Pre-built Binary (Recommended - No Go Required!)

```bash
# Download for your platform from GitHub Releases
# macOS (Intel)
curl -L https://github.com/DylanBlakemore/initiat-cli/releases/latest/download/initiat-darwin-amd64.tar.gz | tar xz
sudo mv initiat-darwin-amd64 /usr/local/bin/initiat

# macOS (Apple Silicon)
curl -L https://github.com/DylanBlakemore/initiat-cli/releases/latest/download/initiat-darwin-arm64.tar.gz | tar xz
sudo mv initiat-darwin-arm64 /usr/local/bin/initiat

# Linux (x64)
curl -L https://github.com/DylanBlakemore/initiat-cli/releases/latest/download/initiat-linux-amd64.tar.gz | tar xz
sudo mv initiat-linux-amd64 /usr/local/bin/initiat

# Windows (download .zip from releases page)
```

#### Option 2: Install with Go

```bash
go install github.com/DylanBlakemore/initiat-cli@latest
```

#### Option 3: Build from Source

```bash
git clone https://github.com/DylanBlakemore/initiat-cli.git
cd initiat-cli
go build -o initiat .
```

#### Login

```bash
# Login with your Initiat account credentials
initiat auth login user@example.com
```

## 📋 Table of Contents

- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Linux Setup](#linux-setup)
- [Authentication](#authentication)
- [Configuration](#configuration)
- [Usage Examples](#usage-examples)
- [Developer Onboarding Features](#developer-onboarding-features)
- [Development](#development)
- [Security](#security)
- [Contributing](#contributing)

## 📋 Prerequisites

Before using the Initiat CLI, you need:

1. **Initiat Account**: Create an account at [initiat.com](https://initiat.com)
2. **Password Setup**: Set your password during registration or in your account settings
3. **System Requirements**:
   - Go 1.25 or later (for building from source)
   - OS keychain access (macOS Keychain, Windows Credential Manager, or Linux Secret Service)
   
   **Linux Users**: If you encounter keyring errors, see [Linux Setup](#linux-setup) below.

## 🛠 Installation

### Option 1: Pre-built Binaries (Recommended)

**No Go installation required!** Download the binary for your platform:

1. Go to [GitHub Releases](https://github.com/DylanBlakemore/initiat-cli/releases)
2. Download the archive for your platform:
   - `initiat-darwin-amd64.tar.gz` (macOS Intel)
   - `initiat-darwin-arm64.tar.gz` (macOS Apple Silicon)
   - `initiat-linux-amd64.tar.gz` (Linux x64)
   - `initiat-linux-arm64.tar.gz` (Linux ARM64)
   - `initiat-windows-amd64.zip` (Windows x64)
3. Extract and move to your PATH

### Option 2: Go Install

```bash
go install github.com/DylanBlakemore/initiat-cli@latest
```

### Option 3: Build from Source

```bash
git clone https://github.com/DylanBlakemore/initiat-cli.git
cd initiat-cli
go build -o initiat .
sudo mv initiat /usr/local/bin/  # Optional: add to PATH
```

## 🐧 Linux Setup

The Initiat CLI requires a secret service for secure credential storage. Most desktop Linux distributions include this by default, but some setups may need manual configuration.

### Check if Secret Service is Available

```bash
# Test if secret service is working
initiat auth login test@example.com
```

If you see an error like `The name org.freedesktop.secrets was not provided by any .service files`, follow the setup below.

### Install Secret Service (Ubuntu/Debian)

```bash
sudo apt-get update
sudo apt-get install gnome-keyring dbus-x11
```

### Install Secret Service (Fedora/RHEL)

```bash
sudo dnf install gnome-keyring
```

### Install Secret Service (Arch Linux)

```bash
sudo pacman -S gnome-keyring
```

### Start the Service

For desktop environments (GNOME, KDE, XFCE), the keyring usually starts automatically. For minimal setups:

```bash
# Start D-Bus session (if not running)
export $(dbus-launch)

# Start GNOME Keyring daemon
gnome-keyring-daemon --start --daemonize --components=secrets

# Add to your shell profile (~/.bashrc, ~/.zshrc) for persistence
echo 'export $(dbus-launch)' >> ~/.bashrc
echo 'gnome-keyring-daemon --start --daemonize --components=secrets >/dev/null 2>&1' >> ~/.bashrc
```

### Alternative: KWallet (KDE)

If you're using KDE and prefer KWallet:

```bash
# Install KWallet and bridge
sudo apt-get install kwalletmanager kwallet-pam  # Ubuntu/Debian
sudo dnf install kwalletmanager5                  # Fedora

# KWallet should provide the secret service automatically
```

### Headless/Server Environments

For servers or headless setups, you can use a minimal secret service:

```bash
# Install minimal keyring
sudo apt-get install gnome-keyring-daemon

# Create a simple startup script
cat > ~/.config/autostart/keyring.sh << 'EOF'
#!/bin/bash
export $(dbus-launch)
echo "defaultpassword" | gnome-keyring-daemon --unlock --daemonize
gnome-keyring-daemon --start --daemonize --components=secrets
EOF

chmod +x ~/.config/autostart/keyring.sh
```

### Troubleshooting

If you continue to have issues:

1. **Check if the service is running:**
   ```bash
   ps aux | grep keyring
   ```

2. **Verify D-Bus is available:**
   ```bash
   echo $DBUS_SESSION_BUS_ADDRESS
   ```

3. **Test secret service directly:**
   ```bash
   python3 -c "import secretstorage; print('Secret service working!')"
   ```

## 🔐 Authentication

### Login

Authenticate with your existing Initiat account credentials:

```bash
initiat auth login user@example.com
```

**Note**: You must have an account at [initiat.com](https://initiat.com) with a password set before using this command.

The CLI will:
1. Prompt for your password (hidden input)
2. Authenticate with the Initiat API
3. Store your registration token securely in the OS keychain
4. Display next steps for device registration

### Example Output

```
$ initiat auth login user@example.com
Password: ••••••••••
🔐 Authenticating...
✅ Login successful! Registration token expires in 15 minutes.
👋 Welcome, John Doe!
💡 Next: Register this device with 'initiat device register <name>'
```

## ⚙️ Configuration

The Initiat CLI supports multiple configuration methods with the following precedence (highest to lowest):

1. **Command-line flags**
2. **Environment variables**
3. **Configuration file**
4. **Default values**

### Configuration File

The CLI automatically creates and uses a configuration file at:
- **macOS/Linux**: `~/.initiat/config.yaml`
- **Windows**: `%USERPROFILE%\.initiat\config.yaml`

#### Example Configuration File

```yaml
# ~/.initiat/config.yaml
api_base_url: "https://www.initiat.dev"
```

### Environment Variables

All configuration options can be set via environment variables with the `INITIAT_` prefix:

```bash
# Set API base URL
export INITIAT_API_BASE_URL="http://localhost:4000"

# Use the CLI with environment config
initiat auth login user@example.com
```

### Command-Line Flags

Override any configuration option using global flags:

```bash
# Use localhost for development
initiat --api-url http://localhost:4000 auth login user@example.com

# Specify custom config file
initiat --config /path/to/custom-config.yaml auth login user@example.com
```

### Configuration Options

| Option | Flag | Environment Variable | Default | Description |
|--------|------|---------------------|---------|-------------|
| API Base URL | `--api-url` | `INITIAT_API_BASE_URL` | `https://www.initiat.dev` | Base URL for Initiat API |
| Config File | `--config` | N/A | `~/.initiat/config.yaml` | Path to configuration file |

### Development Configuration

For local development against a local Initiat server:

```bash
# Method 1: Environment variable
export INITIAT_API_BASE_URL="http://localhost:4000"
initiat auth login dev@example.com

# Method 2: Command-line flag
initiat --api-url http://localhost:4000 auth login dev@example.com

# Method 3: Configuration file
echo "api_base_url: http://localhost:4000" > ~/.initiat/config.yaml
initiat auth login dev@example.com
```

## 📚 Usage Examples

### Developer Onboarding Flow

```bash
# 1. Login to Initiat (requires existing account)
initiat auth login user@example.com

# 2. Register this device (coming soon)
initiat device register "My MacBook CLI"

# 3. List available workspaces (coming soon)
initiat workspace list

# 4. Initialize workspace key for secure secret access (coming soon)
initiat workspace init-key my-project

# 5. Set up development environment (coming soon)
initiat setup my-project

# 6. Fetch secrets and environment variables (coming soon)
initiat secrets fetch --workspace my-project --output .env
```

### Development Workflow

```bash
# Set up for local development
export INITIAT_API_BASE_URL="http://localhost:4000"

# Login with development server
initiat auth login dev@localhost.com

# All subsequent commands use localhost
initiat device register "Development Machine"
```

### Configuration Management

```bash
# View help for global options
initiat --help

# View current configuration (implied from defaults and environment)
initiat auth login --help  # Shows current API URL in global flags

# Test different API endpoints
initiat --api-url https://staging.initiat.com auth login user@example.com
```

## 🚀 Developer Onboarding Features

Initiat is designed to accelerate developer productivity and reduce onboarding friction. Secret management is just one component of a comprehensive developer experience platform:

### 🔐 **Secure Secret Sharing** (Current)
- Zero-knowledge architecture for maximum security
- OS keychain integration for secure local storage
- Team-based secret access controls
- Environment variable export (`.env` files)

### 🛠 **Environment Setup** (Coming Soon)
- One-command project setup (`initiat setup my-project`)
- Automated dependency installation
- Development environment configuration
- Docker and containerization support

### 📋 **Onboarding Workflows** (Coming Soon)
- Guided setup for new team members
- Interactive runbooks and documentation
- Progress tracking and completion verification
- Custom onboarding templates per project

### 🔗 **Integration Ecosystem** (Coming Soon)
- GitHub/GitLab repository integration
- CI/CD pipeline configuration
- Kubernetes and cloud platform setup
- Popular development tool integrations

### 📊 **Team Visibility** (Coming Soon)
- Onboarding progress dashboards
- Team access and permissions overview
- Usage analytics and time-to-productivity metrics
- Audit logs for compliance

### 🎯 **Key Benefits**
- **Reduce time-to-first-commit** from days to minutes
- **Eliminate onboarding friction** with automated setup
- **Improve security** with zero-knowledge secret management
- **Increase team velocity** with standardized workflows
- **Enhance compliance** with audit trails and access controls

## 🔧 Development

### Building from Source

```bash
# Clone the repository
git clone https://github.com/DylanBlakemore/initiat-cli.git
cd initiat-cli

# Install dependencies
go mod tidy

# Build the CLI
go build -o initiat .

# Run tests
go test ./...
```

### Project Structure

```
initiat-cli/
├── cmd/                    # CLI commands
│   ├── auth.go            # Authentication commands
│   ├── root.go            # Root command and global flags
│   └── version.go         # Version command
├── internal/
│   ├── client/            # HTTP client for Initiat API
│   ├── config/            # Configuration management
│   ├── routes/            # API route definitions
│   └── storage/           # Secure storage (OS keychain)
├── main.go                # Application entry point
├── go.mod                 # Go module definition
└── README.md              # This file
```

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage report
make test-coverage

# Run tests for specific package
go test ./internal/config

# Run all CI checks locally
make ci
```

### Code Quality

The project includes comprehensive code quality checks:

```bash
# Format code
make format

# Check formatting
make format-check

# Run linter
make lint

# Run security scan
make security

# Check for vulnerabilities
make vuln-check

# Install development tools
make install-tools
```

### Git Hooks

Set up pre-commit hooks for automatic code quality checks:

```bash
# Install git hooks
./scripts/setup-hooks.sh

# The pre-commit hook will automatically:
# - Format your code
# - Run linter
# - Run tests
# - Check for security issues
# - Verify build works
```

### CI/CD Workflows

The project includes automated GitHub Actions workflows:

#### **Continuous Integration (`.github/workflows/ci.yml`)**
Runs on every push and pull request:
- ✅ **Multi-version testing** (Go 1.25)
- ✅ **Code formatting** checks
- ✅ **Linting** with golangci-lint
- ✅ **Security scanning** with gosec
- ✅ **Vulnerability checking** with govulncheck
- ✅ **Cross-platform builds** (macOS, Linux, Windows)
- ✅ **Coverage reporting** to Codecov

#### **Release Automation (`.github/workflows/release.yml`)**
Runs on git tags (e.g., `v1.0.0`):
- 🚀 **Cross-platform binaries** for all supported platforms
- 📦 **Automated GitHub releases** with downloadable archives
- 📝 **Auto-generated release notes**

```bash
# Create a release
git tag v1.0.0
git push origin v1.0.0
# GitHub Actions automatically builds and releases!
```

### Adding New Commands

1. Create a new command file in `cmd/`
2. Add route constants to `internal/routes/`
3. Implement API client methods in `internal/client/`
4. Add comprehensive tests
5. Update this README

## 🔒 Security

### Token Storage

The CLI stores authentication tokens securely using the operating system's credential management system:

- **macOS**: Keychain Services
- **Windows**: Windows Credential Manager  
- **Linux**: Secret Service (GNOME Keyring, KDE Wallet, etc.)

### Zero-Knowledge Architecture

The Initiat CLI implements a zero-knowledge architecture:

- **Passwords**: Never stored, only used for authentication
- **Tokens**: Stored encrypted in OS keychain
- **Workspace Keys**: Generated client-side, encrypted before transmission
- **Secrets**: Encrypted client-side before storage

### Network Security

- All API requests use HTTPS in production
- Request signing with Ed25519 cryptographic signatures
- Timestamp validation to prevent replay attacks

## 🐛 Troubleshooting

### Common Issues

#### "Failed to store authentication token"

This usually indicates a problem with OS keychain access:

```bash
# macOS: Ensure Keychain Access is working
security find-generic-password -s "initiat-cli" 2>/dev/null || echo "No keychain entry found"

# Linux: Ensure secret service is running
systemctl --user status gnome-keyring-daemon
```

#### "Network connection failed"

Check your API URL configuration:

```bash
# Verify current configuration
initiat --api-url https://www.initiat.dev auth login --help

# Test with explicit URL
initiat --api-url https://www.initiat.dev auth login user@example.com
```

#### "Configuration file not found"

The CLI creates configuration files automatically, but you can create one manually:

```bash
# Create config directory
mkdir -p ~/.initiat

# Create basic config file
cat > ~/.initiat/config.yaml << EOF
api_base_url: "https://www.initiat.dev"
EOF
```

### Debug Mode

For troubleshooting, you can enable verbose output:

```bash
# Enable Go HTTP client debugging
export GODEBUG=http2debug=1
initiat auth login user@example.com
```

## 🤝 Contributing

We welcome contributions! Please see our contributing guidelines:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes with tests
4. Run the test suite (`go test ./...`)
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

### Development Setup

```bash
# Clone your fork
git clone https://github.com/yourusername/initiat-cli.git
cd initiat-cli

# Install dependencies
go mod tidy

# Run tests to ensure everything works
go test ./...

# Build and test the CLI
go build -o initiat .
./initiat --help
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🔗 Links

- [Initiat Website](https://initiat.com)
- [API Documentation](https://docs.initiat.com)
- [Issue Tracker](https://github.com/DylanBlakemore/initiat-cli/issues)
- [Discussions](https://github.com/DylanBlakemore/initiat-cli/discussions)

## 📞 Support

- **Documentation**: [docs.initiat.com](https://docs.initiat.com)
- **Community**: [GitHub Discussions](https://github.com/DylanBlakemore/initiat-cli/discussions)
- **Issues**: [GitHub Issues](https://github.com/DylanBlakemore/initiat-cli/issues)
- **Email**: support@initiat.com

---

**Initiat CLI** - Accelerating developer onboarding, one commit at a time. 🚀
