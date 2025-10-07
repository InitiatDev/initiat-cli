# Initiat CLI Commands

This document provides comprehensive documentation for all Initiat CLI commands, their options, and usage examples.

## Table of Contents

- [Global Options](#global-options)
- [Authentication Commands](#authentication-commands)
- [Device Management](#device-management)
- [Project Management](#project-management)
- [Secret Management](#secret-management)
- [Configuration Management](#configuration-management)
- [Version Information](#version-information)

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

### `initiat device unregister`

Clear local device credentials from the system keychain.

**What it does:**
- Removes all device credentials stored locally
- Use when registering a fresh device or cleaning up after server deletion

**Output:**
```
🔐 Clearing local device credentials...
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
