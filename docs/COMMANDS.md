# Initiat CLI Commands

This document lists all Initiat CLI commands. **Offline-first commands** (init, setup, docs) are documented first; they require no account or server. **Cloud commands** (auth, device, project, env, secret) are optional and only needed when using hosted features.

## Table of Contents

### Offline-first (no account required)

- [Global Options](#global-options)
- [Init (scaffold)](#init-scaffold)
- [Setup Script Management](#setup-script-management)
- [Docs Generation](#docs-generation)
- [Configuration Management](#configuration-management)
- [Version Information](#version-information)

### Cloud commands (optional)

- [Authentication Commands](#authentication-commands)
- [Device Management](#device-management)
- [Project Management](#project-management)
- [Environment Management](#environment-management)
- [Secret Management](#secret-management)

## Global Options

All commands support these global flags:

| Flag | Short | Environment Variable | Default | Description |
|------|-------|---------------------|---------|-------------|
| `--config` | | | `~/.initiat/config.yaml` | Path to configuration file |
| `--api-url` | | `INITIAT_API_BASE_URL` | `https://www.initiat.dev` | API base URL |
| `--service-name` | | | `initiat-cli` | Keyring service name |
| `--project-path` | `-P` | | | Full project path (org/project) or alias |
| `--project` | `-p` | | | Project name (uses default org or --org) |
| `--org` | | | | Organization slug (used with --project) |

### Project Context Resolution

The CLI supports multiple ways to specify project context. If no project is specified, the CLI will intelligently prompt you to select from available projects.

#### Specifying Project Explicitly

```bash
# Method 1: Full project path
initiat secret list --project-path acme-corp/production

# Method 2: Separate org and project
initiat secret list --org acme-corp --project production

# Method 3: Project only (uses default org)
initiat secret list --project production

# Method 4: Short flags
initiat secret list -P acme-corp/production
initiat secret list --org acme-corp -p production
initiat secret list -p production
```

#### Interactive Project Selection

When no project is specified, the CLI will prompt you to select from available projects:

```bash
# No project specified - CLI will prompt
initiat secret list

# Output:
# ❓ Project context is required for this command.
# 💡 You can specify project using:
#    --project-path org/project
#    --org org --project project
#    Or configure defaults with 'initiat config set org <org>' and 'initiat config set project <project>'
#
# Available projects:
#   1. Production Environment (acme-corp/production)
#   2. Staging Environment (acme-corp/staging)
#   3. Development Environment (acme-corp/dev)
#   0. Enter custom project
#
# Select project (0 for custom): 
```

**Interactive Selection Options:**
- **Number Selection**: Choose from the numbered list of available projects
- **Custom Input**: Select option 0 to enter a custom project path manually
- **Fallback**: If project fetching fails, you'll be prompted to enter manually

**Benefits:**
- **Faster Workflow**: No need to remember exact project names
- **Discovery**: See all available projects at a glance
- **Flexible**: Can still enter custom projects when needed
- **User-Friendly**: Clear guidance and helpful error messages

**Note:** Project context is only required for cloud commands (secret, env sync, etc.). Offline-first commands (init, setup validate, setup run, docs generate) do not require project context.

## Init (scaffold)

### `initiat init`

Scaffold the `.initiat/` directory in the current repository with default config and optional setup/docs templates. Works offline; no account required.

**What it does:**
- Ensures current directory is a git repository
- Creates `.initiat/` if missing
- Creates or updates `.initiat/config.yml` with local-only metadata (e.g. repo name)
- Optionally creates `.initiat/setup.yml` and `.initiat/docs.yml` from templates
- Idempotent: safe to run multiple times

**Examples:**
```bash
# Scaffold .initiat/ in current repo
initiat init
```

**Output:**
```
✅ Initialized .initiat/ in this repository.
   Created: .initiat/config.yml
   Created: .initiat/setup.yml (from template)
   Next: Edit .initiat/setup.yml and run 'initiat setup validate'
```

---

## Cloud commands (optional)

The following commands require an Initiat account and device registration. They are used only when you opt in to hosted secrets, project management, or environment sync. Skip these if you use Initiat only for in-repo setup and docs.

## Authentication Commands

### `initiat auth login [--email EMAIL]`

Authenticate with your Initiat account credentials.

**Options:**
- `--email, -e`: Email address for login (optional, will prompt if not provided)

**Examples:**
```bash
# Login with email prompt
initiat auth login

# Login with email specified
initiat auth login --email user@example.com
initiat auth login -e user@example.com
```

**What it does:**
1. Prompts for your password (hidden input)
2. Authenticates with the Initiat API
3. Stores registration token securely in OS keychain
4. Displays next steps for device registration

**Output:**
```
✅ Login successful! Registration token expires in 15 minutes.
💡 Next: Register this device with 'initiat device register <name>'
```

## Device Management

### `initiat device register <device-name>`

Register this device with Initiat to enable secure secret access.

**Arguments:**
- `device-name`: Human-readable name for this device (required)

**Examples:**
```bash
initiat device register "my-laptop"
initiat device register "work-macbook"
```

**What it does:**
1. Generates Ed25519 signing keypair
2. Generates X25519 encryption keypair
3. Registers device with server using authentication token
4. Stores keys securely in system keychain
5. Clears temporary authentication token

**Output:**
```
🔑 Registering device: my-laptop
🔑 Generating Ed25519 signing keypair...
🔒 Generating X25519 encryption keypair...
📡 Registering device with server...
🔐 Storing keys securely in system keychain...
✅ Device registered successfully!

Device ID: dev_abc123
Device Name: my-laptop
Created: 2024-01-15T10:30:00Z

🔐 Keys stored securely in system keychain
💡 Next: Initialize project keys with 'initiat project list'
```

### `initiat device view`

View local device details including device ID, name, API environment, and key status.

**What it does:**
- Displays device information stored locally
- Shows device ID, name (if set), API base URL, and key status
- Helps verify you're working with the correct device

**Examples:**
```bash
# View local device details
initiat device view
```

**Output:**
```
Local Device Details

Device Name: my-laptop
Device ID: dev_abc123
API Base URL: https://www.initiat.dev

Keys Status:
  ✅ Ed25519 signing key: Present
  ✅ X25519 encryption key: Present
```

### `initiat device set-name <device-name>`

Set or update the device name stored locally.

**Arguments:**
- `device-name`: Human-readable name for this device (required)

**What it does:**
- Stores device name locally in the keychain
- Useful if you registered before device names were stored locally
- Allows backfilling device names for existing registrations

**Examples:**
```bash
# Set device name
initiat device set-name "my-laptop"

# Update existing device name
initiat device set-name "work-macbook"
```

**Output:**
```
✅ Device name set to: my-laptop
```

### `initiat device unregister`

Clear local device credentials from the system keychain.

**What it does:**
- Displays device details before unregistering
- Prompts for confirmation before clearing credentials
- Removes all device credentials stored locally
- Use when registering a fresh device or cleaning up after server deletion

**Examples:**
```bash
# Unregister device (with confirmation)
initiat device unregister
```

**Output:**
```
⚠️  You are about to unregister this device:

Device Details:

Device Name: my-laptop
Device ID: dev_abc123
API Base URL: https://www.initiat.dev

Keys Status:
  ✅ Ed25519 signing key: Present
  ✅ X25519 encryption key: Present

Are you sure you want to clear all local device credentials? (y/n): y

Clearing local device credentials...
✅ Device credentials cleared successfully!

💡 You can now register a new device with 'initiat device register <name>'
```

### `initiat device clear-token`

Clear stored authentication token.

**When to use:**
- Getting "Invalid or expired registration token" errors
- Need to re-authenticate

**Output:**
```
🔐 Clearing authentication token...
✅ Authentication token cleared successfully!
💡 You will need to authenticate again for device registration
```

### `initiat device approvals`

List all pending device approvals for projects where you are an admin.

**Output:**
```
📋 Pending Device Approvals (2)

ID  User           Device         Project      Requested
1   John Doe       work-laptop    acme/prod      Jan 15 10:30
2   Jane Smith     dev-machine    acme/staging   Jan 15 11:45

💡 Use 'initiat device approve --all' to approve all pending devices
💡 Use 'initiat device approve --id <id>' to approve a specific device
```

### `initiat device approve [--all] [--id ID]`

Approve device access to projects.

**Options:**
- `--all`: Approve all pending devices
- `--id`: Approve specific device by approval ID

**Examples:**
```bash
# Approve all pending devices
initiat device approve --all

# Approve specific device
initiat device approve --id 123
initiat device approve 123
```

**Output:**
```
🔐 Approving all pending devices...

Found 2 pending approvals:
  • work-laptop (acme-corp/production) - John Doe
  • dev-machine (acme-corp/staging) - Jane Smith

✅ Approved 2 devices successfully!
   All approved devices can now access their respective project secrets
```

### `initiat device reject [--all] [--id ID]`

Reject device access to projects.

**Options:**
- `--all`: Reject all pending devices
- `--id`: Reject specific device by approval ID

**Examples:**
```bash
# Reject all pending devices
initiat device reject --all

# Reject specific device
initiat device reject --id 123
initiat device reject 123
```

**Output:**
```
❌ Rejecting all pending devices...

Found 2 pending approvals to reject
❌ Rejected 2 devices
   Users will need to request approval again
```

### `initiat device approval --id ID`

Show detailed information about a specific device approval.

**Options:**
- `--id`: Device approval ID to show (required)

**Examples:**
```bash
initiat device approval --id 123
```

**Output:**
```
📋 Device Approval Details

User: John Doe (john.doe@example.com)
Device: work-laptop (ID: 456)
Project: Acme Corp / Production (acme-corp/production)
Requested: Jan 15 10:30:00Z
Status: pending

🔑 Device Public Keys:
  Ed25519: abc123def456... (for signing)
  X25519: def456ghi789... (for encryption)
```

## Project Management

### `initiat project list`

List all projects and their key initialization status.

**What it does:**
- Fetches all projects accessible to your account
- Shows key initialization status
- Displays your role in each project

**Output:**
```
🔍 Fetching projects...

Name           Composite Slug      Key Initialized  Role
Production     acme-corp/prod      ✅ Yes          admin
Staging        acme-corp/staging   ❌ No           member
Development    acme-corp/dev       ❌ No           member

💡 Initialize keys for projects marked "No" using:
   initiat project init <org-slug/project-slug>
```

### `initiat project init [project-path]`

Initialize a new project key for secure secret storage.

**Arguments:**
- `project-path`: Full project path (org/project) or use flags

**Options:**
- `--project-path, -P`: Full project path (org/project) or alias
- `--project, -p`: Project name (uses default org or --org)
- `--org`: Organization slug (used with --project)

**Examples:**
```bash
# Using positional argument
initiat project init acme-corp/production

# Using flags
initiat project init --org acme-corp --project production
initiat project init --org acme-corp -p production
initiat project init --project production  # Uses default org
initiat project init -p production
```

**What it does:**
1. Generates secure 256-bit project key
2. Encrypts project key with your device's X25519 key
3. Uploads encrypted key to server
4. Enables secret storage and retrieval for this project

**Output:**
```
🔐 Initializing project key for "acme-corp/production"...
⚡ Generating secure 256-bit project key...
🔒 Encrypting project key with your device's X25519 key...
📡 Uploading encrypted key to server...
✅ Project key initialized successfully!
🎯 You can now store and retrieve secrets in this project.

Next steps:
  • Add secrets: initiat secret set API_KEY --value your-secret
  • List secrets: initiat secret list
  • Invite devices: initiat project invite-device
```

### `initiat project setup`

Run the setup script from `.initiat/setup.yml` to configure the development environment.

**What it does:**
1. Reads and parses `.initiat/setup.yml`
2. Validates the setup configuration against the schema
3. Fetches required secrets from Initiat (if needed)
4. Executes the setup script (installs tools, runtimes, databases, etc.)
5. Runs all phases sequentially (bootstrap → provision → setup → verify → post)

**Prerequisites:**
- Must be inside a git repository
- Project context must be initialized (`.initiat/config.yml` must exist)
- Device must be registered and approved for the project

**Examples:**
```bash
# Run setup script (requires project context)
initiat project setup
```

**Output:**
```
📋 Loading setup script from .initiat/setup.yml...
🔍 Validating setup configuration...
✅ Setup configuration is valid
🚀 Executing setup script...

[bootstrap phase]
  ✓ Ensuring package manager...
  ✓ Ensuring git...

[provision phase]
  ✓ Installing Node.js runtime...
  ✓ Ensuring PostgreSQL database...

[setup phase]
  ✓ Installing dependencies...
  ✓ Running migrations...

[verify phase]
  ✓ Verifying installation...

[post phase]
  ✓ Setup complete!

✅ Setup completed successfully!
```

**Error Handling:**
- If setup file is not found, returns an error
- If validation fails, shows detailed validation errors
- If secrets are required but device is not registered, prompts for device registration
- If any step fails (and `continue_on_error` is false), execution stops

**Related Documentation:**
- See [Setup Scripts Documentation](SETUP_SCRIPTS.md) for complete syntax reference

## Setup Script Management

### `initiat setup validate [setup-file]`

Validate a setup script YAML file against the schema.

**Arguments:**
- `setup-file`: Path to setup file (optional, defaults to `.initiat/setup.yml`)

**Examples:**
```bash
# Validate default setup file
initiat setup validate

# Validate specific file
initiat setup validate .initiat/setup.yml

# Validate custom file
initiat setup validate custom-setup.yml
```

**What it does:**
1. Reads and parses the setup YAML file
2. Validates against JSON Schema
3. Performs additional Go-specific validation (secret references, path safety, etc.)
4. Shows detailed errors if validation fails

**Output (Success):**
```
Validating .initiat/setup.yml...
✅ Setup script is valid!
```

**Output (Failure):**
```
Validating .initiat/setup.yml...
❌ Validation failed:
  - version: must be 1
  - setup[0].ensure_runtime: install configuration is required
  - setup[0].secrets: secret 'FOO' not declared in env.secrets
```

**Use Cases:**
- Validate setup scripts before committing
- Check setup scripts in CI/CD pipelines
- Debug setup script issues

**Related Documentation:**
- See [Setup Scripts Documentation](SETUP_SCRIPTS.md) for syntax details

### `initiat setup schema [--output FILE]`

Output the JSON Schema for `.initiat/setup.yml` files.

**Options:**
- `--output, -o`: Save schema to file instead of stdout

**Examples:**
```bash
# Output schema to stdout
initiat setup schema

# Save schema to file
initiat setup schema --output schemas/setup-v1.json
initiat setup schema -o docs/schemas/setup-v1.json
```

**What it does:**
1. Generates JSON Schema v7 for setup scripts
2. Outputs to stdout (default) or saves to file
3. Schema includes all action types, validation rules, and constraints

**Output (Stdout):**
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "Setup Script Schema",
  "type": "object",
  "properties": {
    "version": {
      "type": "integer",
      "const": 1
    },
    ...
  }
}
```

**Output (File):**
```
✅ Schema saved to schemas/setup-v1.json
```

**Use Cases:**
- Generate schema for IDE autocomplete support
- Distribute schema separately (e.g., in separate repo)
- Validate setup files in other languages/tools
- Document the setup script format programmatically

**Related Documentation:**
- See [Setup Scripts Documentation](SETUP_SCRIPTS.md) for complete syntax reference

## Docs Generation

### `initiat docs generate [--output DIR]`

Generate onboarding docs (e.g. README fragments, runbook steps) from `.initiat/docs.yml` or `.initiat/onboard.yml`. Works offline; no account required.

**Options:**
- `--output, -o`: Output directory (default: current directory or a docs subdir)

**What it does:**
1. Reads `.initiat/docs.yml` (or `.initiat/onboard.yml`) if present
2. Renders templates or structured content into markdown (e.g. "Time to first commit", runbook steps)
3. Writes output files to the specified directory

**Examples:**
```bash
# Generate docs (default output)
initiat docs generate

# Generate to specific directory
initiat docs generate --output docs/onboarding
```

**Related:** Define runbook content and structure in `.initiat/docs.yml`. See the in-repo YAML contract documentation.

## Environment Management

The Initiat CLI provides local environment management to help developers work with different environments (development, staging, production) locally while keeping secrets secure and organized.

### `initiat env init`

Initialize environment management in the current project directory.

**What it does:**
- Creates `.initiat` directory structure for local environment management
- Sets up environment-specific secret storage
- Configures `.gitignore` to exclude environment files
- Generates `.envrc` file for automatic environment loading with direnv

**Examples:**
```bash
# Initialize environment management
initiat env init
```

**Output:**
```
🔧 Initializing environment management...
📁 Creating .initiat directory structure...
🔒 Setting up secure environment storage...
📝 Configuring .gitignore for environment files...
✅ Environment management initialized successfully!

💡 Next steps:
  • Switch to an environment: initiat env switch <env>
  • List available environments: initiat env list
  • Sync secrets: initiat env sync
```

### `initiat env list`

List all available environments with their sync status and last updated information.

**What it does:**
- Shows all environments available in the project
- Displays sync status (whether secrets have been synced)
- Shows last sync time for each environment
- Indicates which environment is currently active

**Examples:**
```bash
# List all environments
initiat env list
```

**Output:**
```
🌍 Available Environments

Name        Status    Last Synced    Active
production  synced    2h ago         ✅
staging     synced    1d ago         
development synced    3h ago         

💡 Switch environments with: initiat env switch <env>
💡 Sync secrets with: initiat env sync
```

### `initiat env switch <environment>`

Switch to a specific environment and update the active environment tracking.

**Arguments:**
- `environment`: Name of the environment to switch to (required)

**What it does:**
- Updates the active environment tracking
- Creates environment-specific directory if needed
- Updates symlinks (Unix) or files (Windows) for environment tracking
- Regenerates `.envrc` file for direnv integration
- Triggers direnv reload to load environment variables
- Shows confirmation of the switch

**Examples:**
```bash
# Switch to production environment
initiat env switch production

# Switch to staging environment
initiat env switch staging

# Switch to development environment
initiat env switch development
```

**Output:**
```
→ setting active -> environments/production
→ refreshing .envrc
→ direnv reload
Switched to "production"
```

**Note:** After switching, direnv will automatically load the environment variables from the local `secrets.env` file. Make sure to run `initiat env sync` first to download the latest secrets.

### `initiat env current`

Show the currently active environment.

**What it does:**
- Displays the name of the currently active environment
- Shows when the environment was last switched
- Provides helpful next steps

**Examples:**
```bash
# Show current environment
initiat env current
```

**Output:**
```
🌍 Current Environment: production
📅 Last switched: 2 hours ago

💡 Available commands:
  • Switch environment: initiat env switch <env>
  • List environments: initiat env list
  • Sync secrets: initiat env sync
```

### `initiat env sync [--env <slug>]`

Sync secrets from the remote project to the local environment(s).

**Options:**
- `--env <slug>`: Sync a specific environment (optional, syncs all environments if not specified)

**What it does:**
- Fetches all secrets from the remote project for the specified environment(s)
- Decrypts secrets using your local project key
- Stores them securely in the local environment directory as `secrets.env`
- Updates sync timestamps
- Shows summary of synced secrets

**Examples:**
```bash
# Sync secrets to current environment
initiat env sync

# Sync secrets to a specific environment
initiat env sync --env production

# Sync all environments
initiat env sync
```

**Output:**
```
🔄 Syncing secrets to environment 'production'...
📡 Fetching secrets from remote project...
🔒 Decrypting and storing secrets securely...
✅ Synced 5 secrets to environment 'production'

Synced secrets:
  • API_KEY
  • DB_PASSWORD
  • JWT_SECRET
  • REDIS_URL
  • SMTP_PASSWORD

💡 Environment is now up to date
💡 Switch to this environment with: initiat env switch production
```

**Important:** Secrets are stored in plaintext in the local `secrets.env` file, but they are:
- Encrypted when stored on Initiat servers
- Stored locally with restrictive file permissions (600 - owner read/write only)
- Automatically excluded from git via `.gitignore`
- Only accessible to processes that can read the file (your user account)

### `initiat env unset`

Clear the currently active environment and reload direnv.

**What it does:**
- Checks if environment management is initialized
- Verifies there's an active environment to unset
- Removes the active environment tracking (symlink or file)
- Runs direnv reload to clear environment variables
- Shows confirmation of the unset operation

**Examples:**
```bash
# Clear the active environment
initiat env unset
```

**Output:**
```
🔄 Unsetting active environment...
🧹 Clearing environment tracking...
🔄 Running direnv reload...
✅ Environment unset successfully!

💡 No environment is currently active
💡 Switch to an environment with: initiat env switch <env>
```

### Environment Directory Structure

The CLI creates the following directory structure for environment management:

```
.initiat/
├── environments/
│   ├── production/
│   │   └── secrets.env
│   ├── staging/
│   │   └── secrets.env
│   └── development/
│       └── secrets.env
└── active -> environments/production  (symlink on Unix, file on Windows)
```

The `active` symlink (or file on Windows) points to the currently active environment directory. This allows direnv to automatically load the correct `secrets.env` file.

### Direnv Integration

The CLI automatically generates `.envrc` files for seamless integration with direnv:

**Generated `.envrc` content (Unix):**
```bash
if [ -e ".initiat/active" ]; then
  dotenv ".initiat/active/secrets.env"
  export INITIAT_ENV=$(basename "$(readlink .initiat/active 2>/dev/null || cat .initiat/active)")
fi
```

**Generated `.envrc` content (Windows):**
```bash
if [ -e ".initiat/active" ]; then
  dotenv ".initiat/active/secrets.env"
  export INITIAT_ENV=$(cat .initiat/active)
fi
```

**How it works:**
1. When you run `initiat env switch <env>`, it creates/updates the `.initiat/active` symlink
2. The `.envrc` file checks if `.initiat/active` exists
3. If it exists, direnv loads the `secrets.env` file from the active environment directory
4. The `INITIAT_ENV` variable is set to the current environment name
5. All secrets are automatically loaded into your shell environment

**Benefits:**
- Automatic environment variable loading when entering the project directory
- Cross-platform compatibility (Unix and Windows)
- Secure local storage of environment-specific secrets (encrypted on disk)
- Integration with existing development workflows
- No need to manually export variables or use eval

### Security Features

- **Secure Storage**: All environment files use 600 permissions (owner read/write only)
- **Git Integration**: Automatic `.gitignore` management to prevent accidental commits (`.initiat/active` is gitignored)
- **Path Validation**: Protection against directory traversal vulnerabilities
- **Cross-Platform**: Symlink support on Unix, file-based tracking on Windows
- **Local Encryption**: While secrets are stored in plaintext locally for direnv compatibility, they remain encrypted on Initiat servers and require your device's private key to decrypt

**Security Considerations:**
- Secrets are stored in plaintext in `secrets.env` files locally to work with direnv
- Files are protected by restrictive permissions (600) and excluded from git
- If your device is compromised, an attacker with keychain access could decrypt secrets
- This is the same risk level as in-memory storage (both require device compromise)
- For maximum security, consider using `initiat secret get` to fetch secrets on-demand instead of syncing

## Secret Management

### `initiat secret set <secret-key> --value VALUE [options]`

Set a secret value in the specified project.

**Arguments:**
- `secret-key`: The key/name for the secret (required)

**Options:**
- `--value, -v`: Secret value (required)
- `--description, -d`: Optional description for the secret
- `--force, -f`: Overwrite existing secret without confirmation
- `--project-path, -P`: Full project path (org/project) or alias
- `--project, -p`: Project name (uses default org or --org)
- `--org`: Organization slug (used with --project)

**Examples:**
```bash
# Set secret with full project path
initiat secret set API_KEY --value "sk-1234567890abcdef" --project-path acme-corp/production

# Set secret with separate org/project
initiat secret set DB_PASSWORD --org acme-corp --project production \
  --value "super-secret-pass" --description "Production database password"

# Set secret with short flags
initiat secret set API_KEY -P acme-corp/production -v "sk-1234567890abcdef"

# Force overwrite existing secret
initiat secret set API_KEY -p production -v "new-value" --force
```

**What it does:**
1. Validates secret key and value
2. Retrieves project key from server
3. Encrypts secret value client-side
4. Uploads encrypted secret to server
5. Shows confirmation with metadata

**Output:**
```
🔐 Setting secret 'API_KEY' in project acme-corp/production...
🔒 Encrypting secret value...
📡 Uploading encrypted secret to server...
✅ Secret 'API_KEY' set successfully!
   Version: 1
   Updated: 2024-01-15T10:30:00Z
   Created by: my-laptop
```

### `initiat secret get <secret-key> [options]`

Get and decrypt a secret value from the specified project.

**Arguments:**
- `secret-key`: The key/name for the secret (required)

**Options:**
- `--copy, -c`: Copy value to clipboard instead of printing
- `--copy-kv`: Copy KEY=VALUE format to clipboard
- `--project-path, -P`: Full project path (org/project) or alias
- `--project, -p`: Project name (uses default org or --org)
- `--org`: Organization slug (used with --project)

**Examples:**
```bash
# Get secret with full project path
initiat secret get API_KEY --project-path acme-corp/production

# Get secret with short flags
initiat secret get API_KEY -P acme-corp/production

# Get secret and copy value to clipboard
initiat secret get API_KEY -p production --copy

# Get secret and copy KEY=VALUE format to clipboard
initiat secret get API_KEY -p production --copy-kv
```

**What it does:**
1. Retrieves encrypted secret from server
2. Gets project key and decrypts it
3. Decrypts secret value client-side
4. Outputs JSON with secret metadata (default)
5. Optionally copies value to clipboard (`--copy`)
6. Optionally copies KEY=VALUE format to clipboard (`--copy-kv`)

**Output:**
```
🔍 Getting secret 'API_KEY' from project acme-corp/production...
🔓 Decrypting secret value...
{
  "key": "API_KEY",
  "value": "sk-1234567890abcdef",
  "version": 1,
  "project_id": "ws_abc123",
  "updated_at": "2024-01-15T10:30:00Z",
  "created_by_device": "my-laptop"
}
```

### `initiat secret list [options]`

List all secrets in the specified project (metadata only, no values).

**Options:**
- `--project-path, -P`: Full project path (org/project) or alias
- `--project, -p`: Project name (uses default org or --org)
- `--org`: Organization slug (used with --project)

**Examples:**
```bash
# List secrets with full project path
initiat secret list --project-path acme-corp/production

# List secrets with short flags
initiat secret list -P acme-corp/production

# List secrets with project only
initiat secret list --project production
```

**What it does:**
1. Fetches all secrets for the project
2. Displays metadata in table format
3. Shows key, encrypted status, and version
4. Never exposes actual secret values

**Output:**
```
🔍 Listing secrets in project acme-corp/production...

Key        Value        Version
API_KEY    [encrypted]  1
DB_PASS    [encrypted]  1
JWT_SECRET [encrypted]  2
```

### `initiat secret delete <secret-key> [options]`

Delete a secret from the specified project.

**Arguments:**
- `secret-key`: The key/name for the secret (required)

**Options:**
- `--force, -f`: Skip confirmation prompt
- `--project-path, -P`: Full project path (org/project) or alias
- `--project, -p`: Project name (uses default org or --org)
- `--org`: Organization slug (used with --project)

**Examples:**
```bash
# Delete secret with confirmation
initiat secret delete API_KEY --project-path acme-corp/production

# Delete secret with short flags
initiat secret delete API_KEY -P acme-corp/production

# Force delete without confirmation
initiat secret delete OLD_API_KEY --project production --force
```

**What it does:**
1. Prompts for confirmation (unless --force is used)
2. Deletes secret from server
3. Shows confirmation message

**Output:**
```
⚠️  Are you sure you want to delete secret 'API_KEY' from project acme-corp/production? (y/N): y
🗑️  Deleting secret 'API_KEY' from project acme-corp/production...
✅ Secret 'API_KEY' deleted successfully!
```

### `initiat secret export <secret-key> --output FILE [options]`

Export a secret value to a file. Creates directories if needed and handles overwrite prompts.

**Arguments:**
- `secret-key`: The key/name for the secret (required)

**Options:**
- `--output, -o`: Output file path (required)
- `--force, -f`: Overwrite existing key without confirmation
- `--project-path, -P`: Full project path (org/project) or alias
- `--project, -p`: Project name (uses default org or --org)
- `--org`: Organization slug (used with --project)

**Examples:**
```bash
# Export secret to a file
initiat secret export API_KEY --output .env --project-path acme-corp/production

# Export to deep directory (creates folders)
initiat secret export API_KEY --output config/secrets.env -P acme-corp/production

# Export with force override
initiat secret export API_KEY --output secrets.txt --force
```

**What it does:**
1. Retrieves and decrypts secret from server
2. Creates output directory if it doesn't exist
3. Checks for existing key in file (prompts if found)
4. Writes secret in KEY=VALUE format
5. Detects git repository and suggests .gitignore

**Output:**
```
🔍 Getting secret 'API_KEY' from project acme-corp/production...
🔓 Decrypting secret value...
⚠️  File 'secrets.env' is not in .gitignore. Add it? (y/N): y
✅ Added 'secrets.env' to .gitignore
✅ Secret 'API_KEY' exported to secrets.env
```

## Configuration Management

The Initiat CLI stores configuration in `~/.initiat/config.yaml` and provides commands to manage settings, project defaults, and aliases.

### `initiat config set <key> <value>`

Set a configuration value using dot notation for nested keys.

**Arguments:**
- `key`: Configuration key to set (required)
- `value`: Value to set (required)

**Available Keys:**
- `api.url`: API base URL
- `api.timeout`: API timeout duration
- `org`: Default organization slug
- `project`: Default project slug
- `service`: Service name for keyring

**Examples:**
```bash
# Set API URL
initiat config set api.url "https://www.initiat.dev"

# Set API timeout
initiat config set api.timeout "60s"

# Set default organization
initiat config set org "my-company"

# Set default project
initiat config set project "production"

# Set service name
initiat config set service "my-custom-service"
```

**Output:**
```
✅ Set api.url = https://www.initiat.dev
```

### `initiat config get <key>`

Get a configuration value using dot notation for nested keys.

**Arguments:**
- `key`: Configuration key to get (required)

**Examples:**
```bash
# Get API URL
initiat config get api.url

# Get default organization
initiat config get org

# Get default project
initiat config get project
```

**Output:**
```
api.url: https://www.initiat.dev
org: my-company
project: production
```

### `initiat config show`

Show all current configuration values.

**Examples:**
```bash
# Show all configuration
initiat config show
```

**Output:**
```
Current configuration:
  api.url: https://www.initiat.dev
  api.timeout: 30s
  service: initiat-cli
  org: my-company
  project: production

Project aliases:
  prod: acme-corp/production
  staging: acme-corp/staging
```

### `initiat config clear <key>`

Clear a configuration value using dot notation for nested keys.

**Arguments:**
- `key`: Configuration key to clear (required)

**Options:**
- `--all`: Clear all configuration values (with confirmation)

**Examples:**
```bash
# Clear default organization
initiat config clear org

# Clear API timeout
initiat config clear api.timeout

# Clear all configuration
initiat config clear --all
```

**Output:**
```
✅ Cleared org
✅ Cleared api.timeout
```

### `initiat config reset`

Reset all configuration values to their default settings.

**What it does:**
- Resets all API settings to defaults
- Clears project defaults (org and project)
- Removes all project aliases
- Resets service name to default

**Examples:**
```bash
# Reset all configuration to defaults
initiat config reset
```

**Output:**
```
⚠️  Are you sure you want to reset all configuration to defaults? (y/N): y
✅ Configuration reset to defaults
```

**Safety Features:**
- Interactive confirmation prompt for safety
- Clear description of what will be reset
- Cancellation support if user doesn't confirm

### `initiat config alias set <alias> <project-path>`

Set a project alias to a full project path.

**Arguments:**
- `alias`: Alias name (required)
- `project-path`: Full project path in format 'org/project' (required)

**Examples:**
```bash
# Set production alias
initiat config alias set prod "acme-corp/production"

# Set staging alias
initiat config alias set staging "acme-corp/staging"

# Set development alias
initiat config alias set dev "acme-corp/development"
```

**Output:**
```
✅ Set alias 'prod' = acme-corp/production
```

### `initiat config alias get <alias>`

Get the project path for a specific alias.

**Arguments:**
- `alias`: Alias name to get (required)

**Examples:**
```bash
# Get production alias
initiat config alias get prod
```

**Output:**
```
prod: acme-corp/production
```

### `initiat config alias list`

List all configured project aliases.

**Examples:**
```bash
# List all aliases
initiat config alias list
```

**Output:**
```
Project aliases:
  prod: acme-corp/production
  staging: acme-corp/staging
  dev: acme-corp/development
```

### `initiat config alias remove <alias>`

Remove a project alias.

**Arguments:**
- `alias`: Alias name to remove (required)

**Examples:**
```bash
# Remove production alias
initiat config alias remove prod
```

**Output:**
```
✅ Removed alias 'prod'
```

### Configuration File Location

The CLI stores configuration in:
- **File**: `~/.initiat/config.yaml`
- **Format**: YAML
- **Permissions**: 600 (owner read/write only)

### Environment Variables

Configuration can also be set via environment variables:
- `INITIAT_API_BASE_URL`: API base URL
- `INITIAT_API_TIMEOUT`: API timeout
- `INITIAT_SERVICE_NAME`: Service name for keyring
- `INITIAT_PROJECT_DEFAULT_ORG`: Default organization
- `INITIAT_PROJECT_DEFAULT_PROJECT`: Default project

### Configuration Precedence

Settings are applied in this order (highest to lowest priority):
1. Command-line flags
2. Environment variables
3. Configuration file
4. Default values

## Version Information

### `initiat version`

Print the CLI version information.

**Output:**
```
initiat-cli v1.0.0
```

## Error Handling

The CLI provides clear error messages and suggestions for common issues:

### Authentication Errors
```
❌ Device not registered. Please run 'initiat device register <name>' first
❌ Failed to get project key: project key not initialized
```

### Network Errors
```
❌ Failed to set secret: network connection failed
❌ Failed to get secret: server returned 404
```

### Validation Errors
```
❌ Invalid secret key: must contain only alphanumeric characters and underscores
❌ Invalid secret value: cannot be empty
```

### Configuration Errors
```
❌ Failed to initialize config: permission denied
❌ Invalid project path: expected 'org-slug/project-slug'
```

## Best Practices

### Project Organization
- Use descriptive project names: `acme-corp/production`, `acme-corp/staging`
- Initialize project keys before storing secrets
- Use consistent naming conventions for secret keys

### Secret Management
- Use descriptive secret keys: `API_KEY`, `DB_PASSWORD`, `JWT_SECRET`
- Add descriptions for complex secrets
- Regularly rotate secrets and update versions

### Device Management
- Use descriptive device names: `john-macbook`, `ci-server-prod`
- Register devices before team members need access
- Approve device access promptly for team productivity

### Security Considerations
- Never share device credentials or project keys
- Use `--force` flag carefully with secret operations
- Regularly audit device access and remove unused devices
- Keep CLI updated to latest version for security patches
