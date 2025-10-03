# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.0]

### Changed
- **BREAKING**: Updated CLI command structure for better consistency
  - `initiat device register <name>` - Now uses positional argument for device name
  - `initiat device approve <id>` - Now uses positional argument for approval ID
  - `initiat device reject <id>` - Now uses positional argument for approval ID
  - `initiat secret get <key>` - Now uses positional argument for secret key
  - `initiat secret delete <key>` - Now uses positional argument for secret key
  - `initiat secret set <key>` - Now uses positional argument for secret key
  - `initiat workspace init <workspace-path>` - Now uses positional argument for workspace path
- Updated all documentation to reflect new command structure
- Improved command examples and help text

## [0.1.0] - 2025-10-03

### Added
- Development build system with `make build-dev` for localhost API URL
- GitHub Actions workflow for manual releases
- Multi-platform binary builds (macOS Intel/ARM, Linux AMD64/ARM64, Windows)
- SHA256 checksums for release verification
- Comprehensive release documentation

- Initial CLI implementation
- Authentication system with email/password login
- Device registration and management
- Workspace initialization and key management
- Secret management (set, get, list, delete)
- Configuration management with YAML config files
- Cross-platform support (macOS, Linux, Windows)
- Comprehensive test suite
- Documentation and examples

### Security
- Ed25519 signing keys for device authentication
- X25519 encryption keys for secret encryption
- Secure key storage using OS keychain/credential store
- Workspace key wrapping for secret access control

[Unreleased]: https://github.com/DylanBlakemore/initiat-cli/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/DylanBlakemore/initiat-cli/releases/tag/v0.1.0
