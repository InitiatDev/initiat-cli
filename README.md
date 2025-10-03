# Initiat CLI

**The Developer Experience Platform that eliminates onboarding friction and accelerates time-to-productivity.**

[![Go Version](https://img.shields.io/badge/Go-1.22+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-AGPL--3.0-blue.svg)](LICENSE)
[![Security](https://img.shields.io/badge/Security-Zero--Knowledge-red.svg)](docs/SECURITY.md)

> **For Engineering Leaders**: Stop losing weeks to environment setup, secret management, and onboarding friction. Initiat transforms your developer experience from days of setup to minutes of productivity.

## The Problem: Developer Experience Debt

**Every engineering team faces the same productivity killers:**

- **Onboarding Hell**: New developers spend days (sometimes weeks) setting up environments
- **Secret Sprawl**: API keys scattered across Slack, emails, and sticky notes
- **Environment Drift**: "Works on my machine" becomes "works on my machine, sometimes"
- **Knowledge Silos**: Critical setup knowledge trapped in senior developers' heads
- **Security Gaps**: Secrets shared via insecure channels, no audit trails

**The Cost**: Lost productivity, frustrated developers, delayed releases, and security vulnerabilities.

## The Solution: Unified Developer Experience

**Initiat is building a unified developer experience platform:**

### **Zero-Knowledge Secret Management** (Available Now)
- Client-side encryption with Ed25519/X25519 cryptography
- Team-based access control with device approval workflows
- Audit trails for security and compliance
- Zero-knowledge architecture - server cannot decrypt secrets

### **Automated Environment Setup** (Planned)
- One-command project setup (`initiat setup my-project`)
- Automated dependency installation
- Environment validation to prevent drift
- Docker and containerization support

### **Guided Onboarding Workflows** (Planned)
- Interactive setup guides for new team members
- Progress tracking and completion verification
- Custom templates per project and role
- Knowledge capture from experienced developers

### **Integration Ecosystem** (Planned)
- GitHub/GitLab repository integration
- CI/CD pipeline configuration
- Kubernetes and cloud platform setup
- Popular development tools integration

### **Team Visibility & Analytics** (Planned)
- Onboarding dashboards showing progress and bottlenecks
- Time-to-productivity metrics
- Usage analytics and optimization insights
- Compliance reporting for security and audit

## Current Capabilities

### **Secret Management** (Production Ready - Invite only)
- Zero-knowledge secret sharing with client-side encryption
- Team-based access control with device approval workflows
- Audit trails for security and compliance
- Cross-platform CLI with macOS, Linux, and Windows support

### **Team Management** (Production Ready - Invite only)
- Device registration and approval workflows
- Workspace-based organization
- Role-based access control
- Secure key storage using OS keychain integration

## Technical Implementation

### **Cryptographic Security**
- Ed25519 signatures for device authentication
- X25519 key exchange for workspace key wrapping
- XSalsa20Poly1305 for secret value encryption
- ChaCha20Poly1305 for workspace key encryption
- HKDF-SHA256 for key derivation

### **Zero-Knowledge Architecture**
- Client-side encryption before transmission
- Server cannot decrypt secrets or workspace keys
- Private keys stored in OS keychain (macOS Keychain, Windows Credential Manager, Linux Secret Service)
- Forward secrecy - compromising one device doesn't affect others

### **Command Examples**
```bash
# Secret management
initiat secret set API_KEY --value "sk-1234567890abcdef" --workspace-path acme-corp/production
initiat secret get API_KEY --workspace-path acme-corp/production
initiat secret list --workspace-path acme-corp/production

# Device management
initiat device register "my-laptop"
initiat device approvals
initiat device approve --all

# Workspace management
initiat workspace list
initiat workspace init acme-corp/production
```

## Roadmap: Developer Experience Platform

**Initiat is building toward a complete developer experience platform:**

### **Phase 1: Secret Management** (Current)
- Zero-knowledge secret sharing
- Team-based access control
- Enterprise security

### **Phase 2: Environment Automation** (Planned)
- One-command project setup
- Automated dependency management
- Environment validation

### **Phase 3: Onboarding Intelligence** (Planned)
- Interactive setup guides
- Progress tracking
- Knowledge capture

### **Phase 4: Developer Analytics** (Planned)
- Productivity insights
- Bottleneck identification
- Optimization recommendations

## Getting Started

### **Installation**
1. **Create account** at [initiat.dev](https://initiat.dev) (coming soon)
2. **Download CLI** from [GitHub Releases](https://github.com/InitiatDev/initiat-cli/releases)
3. **Set up workspaces** for your teams and projects
4. **Configure device approval** workflows

### **For Teams**
1. **Evaluate** current secret management process
2. **Plan migration** from insecure channels (Slack, email, etc.)
3. **Train teams** on new workflows
4. **Monitor usage** and security improvements

## Documentation

### **Complete Guides**
- **[Command Reference](docs/COMMANDS.md)**: Complete CLI command documentation
- **[Security Architecture](docs/SECURITY.md)**: Detailed security and cryptographic implementation
- **[Release Process](docs/RELEASES.md)**: How to create and manage releases

### **Quick Links**
- **Account Setup**: Create account at [initiat.com](https://initiat.com)
- **Support**: [GitHub Issues](https://github.com/InitiatDev/initiat-cli/issues)
- **Discussions**: [GitHub Discussions](https://github.com/InitiatDev/initiat-cli/discussions)

## Contributing

We welcome contributions! Please see our contributing guidelines:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes with tests
4. Run the test suite (`make ci`)
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
make ci

# Build and test the CLI
go build -o initiat .
./initiat --help
```

## License

This project is licensed under the GNU Affero General Public License v3.0 (AGPL-3.0) - see the [LICENSE](LICENSE) file for details.

**Important**: This license allows you to use, modify, and distribute the software, but requires that any derivative works or network services using this software must also be open source under the same license. This protects the open source nature of the project while allowing commercial use of the web application.

## Support

- **Documentation**: [GitHub Repository](https://github.com/InitiatDev/initiat-cli)
- **Issues**: [GitHub Issues](https://github.com/InitiatDev/initiat-cli/issues)
- **Discussions**: [GitHub Discussions](https://github.com/InitiatDev/initiat-cli/discussions)
- **Website**: [initiat.dev](https://initiat.dev)

---

**Initiat CLI** - Transforming developer experience, one team at a time. 🚀
