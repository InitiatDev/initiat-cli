# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.7.3] - 2025-11-15

### Added
- **Device Management Enhancements**: New commands for viewing and managing local device information
  - `initiat device view` - View local device details including device ID, name, API environment, and key status
  - `initiat device set-name <name>` - Set or update the device name stored locally
  - Device name is now stored locally during registration (no API call needed to view it)
- **Enhanced Device Unregistration**: Improved `initiat device unregister` command
  - Now displays device details before unregistering
  - Requires user confirmation before clearing credentials
  - Helps prevent accidental device unregistration

### Changed
- **Device Name Storage**: Device name is now stored locally in the keychain
  - Eliminates need for API calls when viewing device information
  - Works offline and faster than fetching from server
  - Allows backfilling device names for devices registered before this feature

### Fixed
- **Test Reliability**: Improved storage tests to properly handle keyring operations
  - Tests now fail on actual errors instead of skipping
  - Better error detection and reporting

## [0.7.2] - 2025-11-10

### Changed
- **BREAKING**: Simplified setup scripts to explicit command-based approach
  - Removed declarative action types: `ensure_package_manager`, `ensure_tool`, `ensure_runtime`, `ensure_database`, `assert_command`, `assert_http`
  - Setup scripts now use explicit `run:` commands similar to GitHub Actions
  - Commands are now explicit shell commands rather than high-level declarations
  - Significantly reduced codebase complexity and maintenance burden
- **Setup Scripts Condition DSL**: Updated condition syntax to use function calls
  - Changed from variable comparison (`os == "macos"`) to function calls (`os("macos")`)
  - Added `cmd_ok()` function for checking command availability
  - Improved condition evaluation with proper function registration
  - Fixed YAML parsing for conditions starting with `!` by requiring quotes

## [0.7.1] - 2025-11-09

### Changed
- **Project Initialization**: Updated `initiat project init` behavior
  - Project initialization now creates the project if it does not exist remotely.
- **Setup scripts**: Added setup script
  - Added script to set up InitiatCLI locally

### Fixed
- **Setup scripts**: Fixed package managers
  - Separated package managers into system managers and runtime managers
  - System managers install base level system tools, like git
  - Runtime managers install versioned language runtimes
  - Runtime managers can fall back to using asdf
  - Package managers now know how to install themselves

## [0.7.0] - 2025-11-02

### Added
- **Setup Scripts System**: Complete declarative development environment setup feature
  - New `initiat project setup` command to execute setup scripts from `.initiat/setup.yml`
  - New `initiat setup validate` command to validate setup script syntax and schema
  - New `initiat setup schema` command to generate JSON Schema for IDE support
  - Support for 5 execution phases: bootstrap, provision, setup, verify, and post
  - 8 action types: `run`, `print`, `ensure_package_manager`, `ensure_tool`, `ensure_runtime`, `ensure_database`, `assert_command`, `assert_http`
  - Conditions DSL with OS/arch detection, file existence checks, and command validation
  - Matrix gating to restrict execution to specific OS/architecture combinations
  - Secret injection from Initiat directly into setup scripts as environment variables
  - Cross-platform support (macOS, Linux, Windows) with automatic package manager detection
  - Idempotent execution - safe to run multiple times
  - Timeout and retry support for all steps
  - Comprehensive validation with detailed error messages

### Changed
- **Project Initialization**: Updated `initiat project init` behavior
  - Now requires being inside a git repository to create `.initiat` directory
  - Creates `.initiat/config.yml` instead of `.initiat` file
  - Idempotent operation - skips creation if already initialized
  - Remote project key initialization is also idempotent

## [0.6.1] - 2025-10-26

### Added
- **Environment Unset Command**: New `initiat env unset` command to clear active environment
  - Clears currently active environment and reloads direnv
  - Cross-platform support (removes symlinks on Unix, clears file content on Windows)
  - Automatic direnv reload after unsetting environment
  - Graceful handling when no environment is currently set

### Enhanced
- **Environment Initialization**: Improved `initiat env init` command with enhanced setup
  - Automatic direnv hook setup with shell type detection (zsh/bash)
  - Interactive shell configuration with user prompts
  - Automatic `.envrc` file creation and `direnv allow` execution
  - Enhanced init completion verification (checks both directory and gitignore setup)
- **Environment Sync Validation**: Added init completion check to `initiat env sync`
  - Prevents sync operations before proper environment initialization
  - Clear error messages guiding users to run `initiat env init` first
- **Gitignore Management**: Updated to ignore entire environments folder
  - Changed from `.initiat/environments/*/secrets.env` to `.initiat/environments/`
  - More comprehensive environment file protection

### Changed
- **Error Message Formatting**: Standardized error message capitalization throughout CLI
  - All error messages now use lowercase formatting for consistency
  - Improved readability and adherence to Go error message conventions

## [0.6.0] - 2025-10-26

### Added
- **Environment Management System**: New `initiat env` command suite for local environment management
  - `initiat env init` - Initialize environment management in current project
  - `initiat env list` - List available environments with sync status
  - `initiat env switch <env>` - Switch active environment
  - `initiat env current` - Show currently active environment
  - `initiat env sync` - Sync secrets from remote to local environment
- **Local Environment Storage**: Secure local environment management
  - Environment-specific secret storage in `.initiat/environments/`
  - Active environment tracking with symlinks (Unix) or files (Windows)
  - Automatic `.gitignore` management for environment files
- **Direnv Integration**: Automatic environment loading with direnv
  - `initiat env init` generates `.envrc` files for automatic environment loading
  - Cross-platform direnv detection and installation guidance
  - Automatic environment variable loading when entering project directories

## [0.5.2] - 2025-10-26

### Fixed
- **Build Tag Separation**: Fixed dev and production builds to use separate storage keys
  - Dev builds (`make build-dev`) now use isolated storage from production builds
  - Prevents accidental device unregistration when using dev builds
  - Dev builds always point to localhost:4000 regardless of configuration
  - Production builds continue to use configured API endpoints

### Changed
- **API URL Resolution**: Updated all components to use build-tag-aware API URL resolution
  - Client, storage, and config commands now respect build tags
  - Ensures consistent API endpoint usage across all components

## [0.5.1] - 2025-10-25

### Changed
- **Local Configuration Structure**: Updated local config from `.initiat` file to `.initiat/config.yml` directory structure
  - Improved scalability for future local configuration options
  - Maintains same functionality with better organization
  - Project setup now creates `.initiat/config.yml` instead of `.initiat`

## [0.5.0] - 2025-10-08

### Added
- **Local Configuration Support**: New `.initiat` file for project-specific overrides
  - Create `.initiat` files in project directories to override global config
  - Supports organization and project context overrides
  - Automatic detection and priority handling (command flags > local config > global config)
  - Enhanced project context resolution with local config support
- **Guided Project Setup**: New `initiat project setup` command
  - Interactive project setup with smart defaults
  - Prompts for organization (uses default if configured)
  - Auto-detects project name from current directory
  - Creates `.initiat` file with project context
  - Handles remote project existence and key initialization
  - Comprehensive error handling and user guidance
- **Enhanced Project Management**: Improved project initialization workflow
  - Better separation of concerns between command logic and business logic
  - Comprehensive test coverage for project setup functionality
  - Refactored code structure for better maintainability

### Changed
- **Configuration Priority**: Updated configuration resolution order
  - Command-line flags take highest priority
  - Local `.initiat` file takes second priority
  - Global config defaults take third priority
  - Environment variables take lowest priority
- **Project Context Resolution**: Enhanced to support local configuration
  - Improved error messages mentioning `.initiat` file option
  - Better fallback handling when project context is missing

## [0.4.1] - 2025-10-07

### Changed
- **Flag Aliases**: Updated short flags for better project terminology alignment
  - `-w` → `-p` (for `--project`)
  - `-W` → `-P` (for `--project-path`)
  - Updated all documentation and examples to use new flag aliases

## [0.4.0] - 2025-10-07

### Changed
- **BREAKING**: Renamed "workspace" terminology to "project" throughout the CLI
  - All commands now use "project" instead of "workspace" for consistency
  - `initiat workspace` commands are now `initiat project` commands
  - Updated all documentation, help text, and examples
  - Maintained backwards compatibility for existing encrypted data
- **Enhanced Error Handling**: Improved error messages for API communication issues
  - Better debugging information when API endpoints return non-JSON responses
  - More descriptive error messages for authentication and device registration issues

### Added
- **Configuration Management**: New `initiat config reset` command
  - Reset configuration to default values
  - Clear all stored settings and return to initial state
  - Useful for troubleshooting configuration issues

## [0.3.0] - 2025-10-05

### Added
- **Interactive Project Selection**: When no project is specified, the CLI now intelligently prompts users to select from available projects
  - Shows numbered list of available projects for easy selection
  - Option to enter custom project path manually
  - Graceful fallback when project fetching fails
  - Improved user experience with clear guidance and helpful error messages
- **Secret Export Functionality**: New `initiat secret export` command to export secrets to files
  - Export secrets to `.env` files with proper formatting
  - Automatic directory creation for nested paths
  - Git integration with automatic `.gitignore` management
  - Overwrite protection with user confirmation prompts
- **Clipboard Integration**: Enhanced secret retrieval with clipboard support
  - `--copy` flag to copy secret values directly to clipboard
  - `--copy-kv` flag to copy secrets in KEY=VALUE format
  - Cross-platform clipboard support using `golang.design/x/clipboard`

## [0.2.1] - 2025-10-04
- Trigger homebrew publish on release

## [0.2.0] - 2025-10-04

### Changed
- **BREAKING**: Updated CLI command structure for better consistency
  - `initiat device register <name>` - Now uses positional argument for device name
  - `initiat device approve <id>` - Now uses positional argument for approval ID
  - `initiat device reject <id>` - Now uses positional argument for approval ID
  - `initiat secret get <key>` - Now uses positional argument for secret key
  - `initiat secret delete <key>` - Now uses positional argument for secret key
  - `initiat secret set <key>` - Now uses positional argument for secret key
  - `initiat project init <project-path>` - Now uses positional argument for project path
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
- Project initialization and key management
- Secret management (set, get, list, delete)
- Configuration management with YAML config files
- Cross-platform support (macOS, Linux, Windows)
- Comprehensive test suite
- Documentation and examples

### Security
- Ed25519 signing keys for device authentication
- X25519 encryption keys for secret encryption
- Secure key storage using OS keychain/credential store
- Project key wrapping for secret access control

[Unreleased]: https://github.com/InitiatDev/initiat-cli/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/InitiatDev/initiat-cli/releases/tag/v0.1.0
