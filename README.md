# Initiat CLI

**In-repo developer onboarding: setup and docs from YAML. No server required.**

[![Go Version](https://img.shields.io/badge/Go-1.22+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-AGPL--3.0-blue.svg)](LICENSE)
[![Security](https://img.shields.io/badge/Security-Offline--First-green.svg)](docs/SECURITY.md)

Define your project's setup and onboarding in `.initiat/` YAML committed to the repo. Run setup, validate configs, and generate docs locally. No account, no login, no sending code or secrets to any server.

## The Problem

**Every engineering team faces the same productivity killers:**

- **Onboarding Hell**: New developers spend days (sometimes weeks) setting up environments
- **Environment Drift**: "Works on my machine" becomes "works on my machine, sometimes"
- **Knowledge Silos**: Critical setup knowledge trapped in senior developers' heads
- **Scattered Instructions**: READMEs, wikis, and Slack threads that go stale

**The Cost**: Lost productivity, frustrated developers, and delayed time-to-first-commit.

## What Initiat Does (Offline First)

Initiat's core value is **in-repo workflows that run offline**:

### Reproducible Setup from Repo YAML

Define your development environment in `.initiat/setup.yml`. Same file works on macOS, Linux, and Windows. Explicit, GitHub Actions-style commands—no magic.

**Key features:**
- Explicit command-based setup (run exactly what you define)
- Cross-platform support with OS/arch conditions
- Idempotent operations (safe to run multiple times)
- Validate and generate JSON Schema for IDE support
- Dry-run (plan) to see what would run without executing

**Learn more:** [Setup Scripts](docs/SETUP_SCRIPTS.md)

### Validation and Planning

- `initiat setup validate` — validate `.initiat/setup.yml` against the schema
- `initiat setup schema` — output JSON Schema for autocomplete and tooling

### Onboarding Docs from Repo

Generate onboarding runbooks and README fragments from `.initiat/docs.yml` (or `.initiat/onboard.yml`). Keep "time-to-first-commit" steps in version control next to the setup that makes them possible.

### No Server Required

Core workflows need no account, no device registration, and no network. Clone the repo, run the CLI, and you're done.

## Quick Start (Offline)

### Installation

Install the Initiat CLI (e.g. from [releases](https://github.com/InitiatDev/initiat-cli/releases) or your package manager). No signup required.

### Basic Usage (No Login)

```bash
# Scaffold .initiat/ in the current repo
initiat init

# Validate setup config
initiat setup validate

# Run the setup script from .initiat/setup.yml
initiat setup run

# Generate onboarding docs from .initiat/docs.yml
initiat docs generate
```

All of the above work with no account and no connection to any Initiat server.

## Optional: Cloud Add-ons

If your team wants hosted secret storage, device approval, or project management, those features are available as **opt-in** add-ons. They are not required for setup, validation, or docs.

- **Hosted secrets (optional)**: Zero-knowledge secret storage with device approval. See [Security & Cloud](docs/SECURITY.md#optional-cloud-add-on).
- **Team & project management (optional)**: Organize access by project; approve devices. See [Command Reference](docs/COMMANDS.md) under "Cloud commands (optional)".

You can use Initiat entirely without ever creating an account or sending secrets to our servers.

## Documentation

### Core (Offline-First)

- **[In-Repo YAML](docs/IN_REPO_YAML.md)**: Directory layout, `config.yml`, `setup.yml`, `docs.yml`, and provider-agnostic `env_from`
- **[Setup Scripts](docs/SETUP_SCRIPTS.md)**: Complete guide to `.initiat/setup.yml`
- **[Command Reference](docs/COMMANDS.md)**: All CLI commands (offline-first commands first)
- **[Security](docs/SECURITY.md)**: Offline-first security model and optional cloud add-on

### Additional

- **[Release Process](docs/RELEASES.md)**: How to create and manage releases
- **[Support](https://github.com/InitiatDev/initiat-cli/issues)**: GitHub Issues

## Contributing

1. **Fork** the repository
2. **Create** a feature branch (`git checkout -b feature/amazing-feature`)
3. **Make** your changes with tests
4. **Run** the test suite (`make ci`)
5. **Commit** your changes (`git commit -m 'Add amazing feature'`)
6. **Push** the branch (`git push origin feature/amazing-feature`)
7. **Open** a Pull Request

### Development Setup

```bash
git clone https://github.com/yourusername/initiat-cli.git
cd initiat-cli

go mod tidy
initiat setup run
make ci
go build -o initiat .
```

See [Setup Scripts](docs/SETUP_SCRIPTS.md) for the full development environment definition in `.initiat/setup.yml`.

## License

This project is licensed under the GNU Affero General Public License v3.0 (AGPL-3.0) - see the [LICENSE](LICENSE) file for details.

**Important**: This license allows you to use, modify, and distribute the software, but requires that any derivative works or network services using this software must also be open source under the same license.

## Support

- **Documentation**: [GitHub Repository](https://github.com/InitiatDev/initiat-cli)
- **Issues**: [GitHub Issues](https://github.com/InitiatDev/initiat-cli/issues)
- **Discussions**: [GitHub Discussions](https://github.com/InitiatDev/initiat-cli/discussions)

---

**Initiat CLI** — In-repo onboarding, offline by default.
