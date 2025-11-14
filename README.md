# Initiat CLI

**The Developer Experience Platform that eliminates onboarding friction and accelerates time-to-productivity.**

[![Go Version](https://img.shields.io/badge/Go-1.22+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-AGPL--3.0-blue.svg)](LICENSE)
[![Security](https://img.shields.io/badge/Security-Zero--Knowledge-red.svg)](docs/SECURITY.md)

> **For Engineering Leaders**: Stop losing weeks to environment setup, secret management, and onboarding friction. Initiat transforms your developer experience from days of setup to minutes of productivity.

## The Problem

**Every engineering team faces the same productivity killers:**

- **Onboarding Hell**: New developers spend days (sometimes weeks) setting up environments
- **Secret Sprawl**: API keys scattered across Slack, emails, and sticky notes
- **Environment Drift**: "Works on my machine" becomes "works on my machine, sometimes"
- **Knowledge Silos**: Critical setup knowledge trapped in senior developers' heads
- **Security Gaps**: Secrets shared via insecure channels, no audit trails

**The Cost**: Lost productivity, frustrated developers, delayed releases, and security vulnerabilities.

## What Initiat Does

Initiat solves these problems with three core capabilities:

### 🔐 Zero-Knowledge Secret Management

Manage team secrets with enterprise-grade security. Secrets are encrypted on your device before transmission, and we can never decrypt them - even with full server access.

**Key Features:**
- Client-side encryption with Ed25519/X25519 cryptography
- Team-based access control with device approval workflows
- Audit trails for security and compliance
- Cross-platform CLI for macOS, Linux, and Windows

**Learn more:** [Security Documentation](docs/SECURITY.md)

### 🚀 Automated Environment Setup

Define your development environment in `.initiat/setup.yml` and let Initiat handle the rest. Works across macOS, Linux, and Windows with explicit, GitHub Actions-style commands.

**Key Features:**
- Explicit command-based setup (no magic, just clear commands)
- Cross-platform support with OS-specific conditions
- Idempotent operations (safe to run multiple times)
- Integrates with secret management for secure configuration

**Learn more:** [Setup Scripts Documentation](docs/SETUP_SCRIPTS.md)

### 👥 Team & Project Management

Organize secrets and environments by team and project. Control who can access what with device approval workflows and granular permissions.

**Key Features:**
- Device registration and approval workflows
- Project-based organization for teams and projects
- Role-based access control
- Secure key storage using OS keychain integration

**Learn more:** [Command Reference](docs/COMMANDS.md)

## Quick Start

### Installation

1. **Create account** at [initiat.dev](https://initiat.dev)
2. **Follow the setup instructions** at [initiat.dev](https://initiat.dev/docs/setup)
3. **Register your device**: `initiat device register "my-laptop"`
4. **Get approved** by your team administrator
5. **Start using secrets**: `initiat secret set API_KEY --value "sk-..."`

### Basic Usage

```bash
# Set a secret
initiat secret set API_KEY --value "sk-1234567890abcdef"

# Get a secret
initiat secret get API_KEY

# List all secrets
initiat secret list

# Initialize a project
initiat project init acme-corp/production

# Run setup script
initiat project setup
```

## Documentation

### Core Documentation

- **[Security & Secret Management](docs/SECURITY.md)**: How secrets are encrypted and protected
- **[Setup Scripts](docs/SETUP_SCRIPTS.md)**: Complete guide to `.initiat/setup.yml`
- **[Command Reference](docs/COMMANDS.md)**: All CLI commands and options

### Additional Resources

- **[Release Process](docs/RELEASES.md)**: How to create and manage releases
- **[Account Setup](https://initiat.dev)**: Create account and get started
- **[Support](https://github.com/InitiatDev/initiat-cli/issues)**: GitHub Issues

## Technical Foundation

**Cryptographic Security**

- **Ed25519 signatures** for device authentication
- **X25519 key exchange** for project key wrapping
- **XSalsa20Poly1305** for secret value encryption
- **ChaCha20Poly1305** for project key encryption
- **HKDF-SHA256** for key derivation

**Zero-Knowledge Architecture**

- Client-side encryption before transmission
- Server cannot decrypt secrets or project keys
- Private keys stored in OS keychain
- Forward secrecy - compromising one device doesn't affect others

**Learn more:** [Security Documentation](docs/SECURITY.md)

## Contributing

We welcome contributions! Here's how to get started:

1. **Fork** the repository
2. **Create** a feature branch (`git checkout -b feature/amazing-feature`)
3. **Make** your changes with tests
4. **Run** the test suite (`make ci`)
5. **Commit** your changes (`git commit -m 'Add amazing feature'`)
6. **Push** to the branch (`git push origin feature/amazing-feature`)
7. **Open** a Pull Request

### Development Setup

```bash
# Clone your fork
git clone https://github.com/yourusername/initiat-cli.git
cd initiat-cli

# Install dependencies
go mod tidy

# Run the setup script (installs dependencies, tools, etc.)
initiat setup run

# Run tests
make ci

# Build the CLI
go build -o initiat .
```

**Learn more:** See [Setup Scripts Documentation](docs/SETUP_SCRIPTS.md) for the complete development environment setup.

## License

This project is licensed under the GNU Affero General Public License v3.0 (AGPL-3.0) - see the [LICENSE](LICENSE) file for details.

**Important**: This license allows you to use, modify, and distribute the software, but requires that any derivative works or network services using this software must also be open source under the same license.

## Support

- **Documentation**: [GitHub Repository](https://github.com/InitiatDev/initiat-cli)
- **Issues**: [GitHub Issues](https://github.com/InitiatDev/initiat-cli/issues)
- **Discussions**: [GitHub Discussions](https://github.com/InitiatDev/initiat-cli/discussions)
- **Website**: [initiat.dev](https://initiat.dev)

---

**Initiat CLI** - Transforming developer experience, one team at a time. 🚀
