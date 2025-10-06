# Initiat CLI

**The Developer Experience Platform that eliminates onboarding friction and accelerates time-to-productivity.**

[![Go Version](https://img.shields.io/badge/Go-1.22+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-AGPL--3.0-blue.svg)](LICENSE)
[![Security](https://img.shields.io/badge/Security-Zero--Knowledge-red.svg)](docs/SECURITY.md)

> **For Engineering Leaders**: Stop losing weeks to environment setup, secret management, and onboarding friction. Initiat transforms your developer experience from days of setup to minutes of productivity.

<img width="1024" height="1024" alt="ChatGPT Image Oct 6, 2025, 08_04_39 PM" src="https://github.com/user-attachments/assets/12018d13-f9ad-401f-9232-d6a6ba66d451" />

## The Problem: Developer Experience Debt

**Every engineering team faces the same productivity killers:**

- **Onboarding Hell**: New developers spend days (sometimes weeks) setting up environments
- **Secret Sprawl**: API keys scattered across Slack, emails, and sticky notes
- **Environment Drift**: "Works on my machine" becomes "works on my machine, sometimes"
- **Knowledge Silos**: Critical setup knowledge trapped in senior developers' heads
- **Security Gaps**: Secrets shared via insecure channels, no audit trails

**The Cost**: Lost productivity, frustrated developers, delayed releases, and security vulnerabilities.

## What's Available Now

**Zero-Knowledge Secret Management** 🔐 *Production ready - Invite only*

Transform how your team handles secrets with enterprise-grade security. Our zero-knowledge architecture ensures that even we can't decrypt your secrets - everything is encrypted client-side before transmission.

```bash
# Set a secret for your team (with workspace selection)
initiat secret set API_KEY --value "sk-1234567890abcdef"
# CLI will prompt: Select workspace (0 for custom): 1

# Get a secret (decrypted client-side)
initiat secret get API_KEY
# CLI will prompt: Select workspace (0 for custom): 1

# List all secrets in a workspace
initiat secret list
# CLI will prompt: Select workspace (0 for custom): 1

# Or specify workspace explicitly
initiat secret set API_KEY --value "sk-1234567890abcdef" --workspace-path acme-corp/production
```

- **Client-side encryption** with Ed25519/X25519 cryptography
- **Team-based access control** with device approval workflows
- **Audit trails** for security and compliance
- **Cross-platform CLI** for macOS, Linux, and Windows

**Team Management** 👥 *Production ready - Invite only*

Streamline device and workspace management with granular control over who can access what. Every device must be approved, and every workspace can be configured with specific permissions.

```bash
# Register a new device
initiat device register "my-laptop"

# Check pending device approvals
initiat device approvals

# Approve all pending devices
initiat device approve --all

# List available workspaces
initiat workspace list

# Initialize a new workspace
initiat workspace init acme-corp/production
```

- **Device registration** and approval workflows
- **Workspace-based organization** for teams and projects
- **Role-based access control** with granular permissions
- **Secure key storage** using OS keychain integration

**Interactive Workspace Selection**

Never remember workspace names again. When you don't specify a workspace, the CLI intelligently prompts you to select from available workspaces.

```bash
# No workspace specified - CLI will show interactive selection
initiat secret list

# Output:
# ❓ Workspace context is required for this command.
# 💡 You can specify workspace using:
#    --workspace-path org/workspace
#    --org org --workspace workspace
#    Or configure defaults with 'initiat config set org <org>' and 'initiat config set workspace <workspace>'
#
# Available workspaces:
#   1. Production Environment (acme-corp/production)
#   2. Staging Environment (acme-corp/staging)
#   3. Development Environment (acme-corp/dev)
#   0. Enter custom workspace
#
# Select workspace (0 for custom): 
```

**Benefits:**
- **Faster Workflow**: No need to remember exact workspace names
- **Discovery**: See all available workspaces at a glance
- **Flexible**: Can still enter custom workspaces when needed
- **User-Friendly**: Clear guidance and helpful error messages

## What's Coming Next

**Automated Environment Setup** 🚀 *Planned for 2026*

One command to rule them all. No more "works on my machine" - we'll automate the entire development environment setup process.

```bash
initiat setup my-project  # Sets up entire development environment
```

- Automated dependency installation
- Environment validation to prevent drift
- Docker and containerization support
- Custom templates per project and role

**Guided Onboarding Workflows** 📚 *Planned for 2026*

Interactive guides that actually work. Capture knowledge from your senior developers and turn it into step-by-step onboarding workflows.

- Step-by-step setup guides for new team members
- Progress tracking and completion verification
- Knowledge capture from experienced developers
- Custom templates per project and role

**Integration Ecosystem** 🔗 *Planned for 2026*

Connect everything together. We're building integrations with the tools you already use to create a seamless developer experience.

- GitHub/GitLab repository integration
- CI/CD pipeline configuration
- Kubernetes and cloud platform setup
- Popular development tools integration

**Team Visibility & Analytics** 📊 *Planned for 2026*

Data-driven developer experience. Understand where your team is spending time and identify bottlenecks in your development process.

- Onboarding dashboards showing progress and bottlenecks
- Time-to-productivity metrics
- Usage analytics and optimization insights
- Compliance reporting for security and audit

## Technical Foundation

**Cryptographic Security**

We use industry-standard cryptographic primitives to ensure your secrets are protected at the highest level. Every operation is designed with security-first principles.

- **Ed25519 signatures** for device authentication
- **X25519 key exchange** for workspace key wrapping
- **XSalsa20Poly1305** for secret value encryption
- **ChaCha20Poly1305** for workspace key encryption
- **HKDF-SHA256** for key derivation

**Zero-Knowledge Architecture**

Our zero-knowledge architecture means we can't see your secrets, even if we wanted to. Everything is encrypted client-side before it ever reaches our servers.

- Client-side encryption before transmission
- Server cannot decrypt secrets or workspace keys
- Private keys stored in OS keychain (macOS Keychain, Windows Credential Manager, Linux Secret Service)
- Forward secrecy - compromising one device doesn't affect others

## Quick Start

**Secret Management**

Manage your team's secrets with enterprise-grade security. Every secret is encrypted client-side before transmission.

```bash
# Set a secret
initiat secret set API_KEY --value "sk-1234567890abcdef" --workspace-path acme-corp/production

# Get a secret
initiat secret get API_KEY --workspace-path acme-corp/production

# List all secrets
initiat secret list --workspace-path acme-corp/production
```

**Device Management**

Control access with device approval workflows. Every device must be registered and approved before it can access secrets.

```bash
# Register your device
initiat device register "my-laptop"

# Check pending approvals
initiat device approvals

# Approve all pending devices
initiat device approve --all
```

**Workspace Management**

Organize your secrets by team and project. Each workspace can have its own access controls and permissions.

```bash
# List available workspaces
initiat workspace list

# Initialize a new workspace (with interactive selection)
initiat workspace init
# CLI will prompt: Select workspace (0 for custom): 1

# Or specify workspace explicitly
initiat workspace init acme-corp/production
```

## Getting Started

### **Installation**
1. **Create account** at [initiat.dev](https://initiat.dev) (coming soon)
2. **Download CLI** from [GitHub Releases](https://github.com/InitiatDev/initiat-cli/releases)
3. **Linux users**: Install X11 development libraries for clipboard support:
   ```bash
   # Ubuntu/Debian
   sudo apt-get install libx11-dev libxrandr-dev libxinerama-dev libxcursor-dev libxi-dev
   
   # CentOS/RHEL/Fedora
   sudo yum install libX11-devel libXrandr-devel libXinerama-devel libXcursor-devel libXi-devel
   # or for newer versions:
   sudo dnf install libX11-devel libXrandr-devel libXinerama-devel libXcursor-devel libXi-devel
   
   # Arch Linux
   sudo pacman -S libx11 libxrandr libxinerama libxcursor libxi
   ```
4. **Set up workspaces** for your teams and projects
5. **Configure device approval** workflows

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
- **Account Setup**: Create account at [initiat.dev](https://initiat.dev)
- **Support**: [GitHub Issues](https://github.com/InitiatDev/initiat-cli/issues)
- **Discussions**: [GitHub Discussions](https://github.com/InitiatDev/initiat-cli/discussions)

## Contributing

We welcome contributions! Here's how to get started:

### **Quick Start**
1. **Fork** the repository
2. **Create** a feature branch (`git checkout -b feature/amazing-feature`)
3. **Make** your changes with tests
4. **Run** the test suite (`make ci`)
5. **Commit** your changes (`git commit -m 'Add amazing feature'`)
6. **Push** to the branch (`git push origin feature/amazing-feature`)
7. **Open** a Pull Request

### **Development Setup**
```bash
# Clone your fork
git clone https://github.com/yourusername/initiat-cli.git
cd initiat-cli

# Install dependencies
go mod tidy

# Linux users: Install X11 dependencies for clipboard support
# Ubuntu/Debian
sudo apt-get install libx11-dev libxrandr-dev libxinerama-dev libxcursor-dev libxi-dev

# CentOS/RHEL/Fedora
sudo yum install libX11-devel libXrandr-devel libXinerama-devel libXcursor-devel libXi-devel
# or for newer versions:
sudo dnf install libX11-devel libXrandr-devel libXinerama-devel libXcursor-devel libXi-devel

# Arch Linux
sudo pacman -S libx11 libxrandr libxinerama libxcursor libxi

# Run tests to ensure everything works
make ci

# Build and test the CLI
go build -o initiat .
./initiat --help
```

### **What We're Looking For**
- **Bug fixes** and improvements
- **New features** that align with our roadmap
- **Documentation** improvements
- **Security** enhancements
- **Performance** optimizations

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
