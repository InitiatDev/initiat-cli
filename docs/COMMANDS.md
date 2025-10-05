# Initiat CLI Commands

This document provides comprehensive documentation for all Initiat CLI commands, their options, and usage examples.

## Table of Contents

- [Global Options](#global-options)
- [Authentication Commands](#authentication-commands)
- [Device Management](#device-management)
- [Workspace Management](#workspace-management)
- [Secret Management](#secret-management)
- [Version Information](#version-information)

## Global Options

All commands support these global flags:

| Flag | Short | Environment Variable | Default | Description |
|------|-------|---------------------|---------|-------------|
| `--config` | | | `~/.initiat/config.yaml` | Path to configuration file |
| `--api-url` | | `INITIAT_API_BASE_URL` | `https://www.initiat.dev` | API base URL |
| `--service-name` | | | `initiat-cli` | Keyring service name |
| `--workspace-path` | `-W` | | | Full workspace path (org/workspace) or alias |
| `--workspace` | `-w` | | | Workspace name (uses default org or --org) |
| `--org` | | | | Organization slug (used with --workspace) |

### Workspace Context Resolution

The CLI supports multiple ways to specify workspace context. If no workspace is specified, the CLI will intelligently prompt you to select from available workspaces.

#### Specifying Workspace Explicitly

```bash
# Method 1: Full workspace path
initiat secret list --workspace-path acme-corp/production

# Method 2: Separate org and workspace
initiat secret list --org acme-corp --workspace production

# Method 3: Workspace only (uses default org)
initiat secret list --workspace production

# Method 4: Short flags
initiat secret list -W acme-corp/production
initiat secret list --org acme-corp -w production
initiat secret list -w production
```

#### Interactive Workspace Selection

When no workspace is specified, the CLI will prompt you to select from available workspaces:

```bash
# No workspace specified - CLI will prompt
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

**Interactive Selection Options:**
- **Number Selection**: Choose from the numbered list of available workspaces
- **Custom Input**: Select option 0 to enter a custom workspace path manually
- **Fallback**: If workspace fetching fails, you'll be prompted to enter manually

**Benefits:**
- **Faster Workflow**: No need to remember exact workspace names
- **Discovery**: See all available workspaces at a glance
- **Flexible**: Can still enter custom workspaces when needed
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
💡 Next: Initialize workspace keys with 'initiat workspace list'
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

List all pending device approvals for workspaces where you are an admin.

**Output:**
```
📋 Pending Device Approvals (2)

ID  User           Device         Workspace      Requested
1   John Doe       work-laptop    acme/prod      Jan 15 10:30
2   Jane Smith     dev-machine    acme/staging   Jan 15 11:45

💡 Use 'initiat device approve --all' to approve all pending devices
💡 Use 'initiat device approve --id <id>' to approve a specific device
```

### `initiat device approve [--all] [--id ID]`

Approve device access to workspaces.

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
   All approved devices can now access their respective workspace secrets
```

### `initiat device reject [--all] [--id ID]`

Reject device access to workspaces.

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
Workspace: Acme Corp / Production (acme-corp/production)
Requested: Jan 15 10:30:00Z
Status: pending

🔑 Device Public Keys:
  Ed25519: abc123def456... (for signing)
  X25519: def456ghi789... (for encryption)
```

## Workspace Management

### `initiat workspace list`

List all workspaces and their key initialization status.

**What it does:**
- Fetches all workspaces accessible to your account
- Shows key initialization status
- Displays your role in each workspace

**Output:**
```
🔍 Fetching workspaces...

Name           Composite Slug      Key Initialized  Role
Production     acme-corp/prod      ✅ Yes          admin
Staging        acme-corp/staging   ❌ No           member
Development    acme-corp/dev       ❌ No           member

💡 Initialize keys for workspaces marked "No" using:
   initiat workspace init <org-slug/workspace-slug>
```

### `initiat workspace init [workspace-path]`

Initialize a new workspace key for secure secret storage.

**Arguments:**
- `workspace-path`: Full workspace path (org/workspace) or use flags

**Options:**
- `--workspace-path, -W`: Full workspace path (org/workspace) or alias
- `--workspace, -w`: Workspace name (uses default org or --org)
- `--org`: Organization slug (used with --workspace)

**Examples:**
```bash
# Using positional argument
initiat workspace init acme-corp/production

# Using flags
initiat workspace init --org acme-corp --workspace production
initiat workspace init --org acme-corp -w production
initiat workspace init --workspace production  # Uses default org
initiat workspace init -w production
```

**What it does:**
1. Generates secure 256-bit workspace key
2. Encrypts workspace key with your device's X25519 key
3. Uploads encrypted key to server
4. Enables secret storage and retrieval for this workspace

**Output:**
```
🔐 Initializing workspace key for "acme-corp/production"...
⚡ Generating secure 256-bit workspace key...
🔒 Encrypting workspace key with your device's X25519 key...
📡 Uploading encrypted key to server...
✅ Workspace key initialized successfully!
🎯 You can now store and retrieve secrets in this workspace.

Next steps:
  • Add secrets: initiat secret set API_KEY --value your-secret
  • List secrets: initiat secret list
  • Invite devices: initiat workspace invite-device
```

## Secret Management

### `initiat secret set <secret-key> --value VALUE [options]`

Set a secret value in the specified workspace.

**Arguments:**
- `secret-key`: The key/name for the secret (required)

**Options:**
- `--value, -v`: Secret value (required)
- `--description, -d`: Optional description for the secret
- `--force, -f`: Overwrite existing secret without confirmation
- `--workspace-path, -W`: Full workspace path (org/workspace) or alias
- `--workspace, -w`: Workspace name (uses default org or --org)
- `--org`: Organization slug (used with --workspace)

**Examples:**
```bash
# Set secret with full workspace path
initiat secret set API_KEY --value "sk-1234567890abcdef" --workspace-path acme-corp/production

# Set secret with separate org/workspace
initiat secret set DB_PASSWORD --org acme-corp --workspace production \
  --value "super-secret-pass" --description "Production database password"

# Set secret with short flags
initiat secret set API_KEY -W acme-corp/production -v "sk-1234567890abcdef"

# Force overwrite existing secret
initiat secret set API_KEY -w production -v "new-value" --force
```

**What it does:**
1. Validates secret key and value
2. Retrieves workspace key from server
3. Encrypts secret value client-side
4. Uploads encrypted secret to server
5. Shows confirmation with metadata

**Output:**
```
🔐 Setting secret 'API_KEY' in workspace acme-corp/production...
🔒 Encrypting secret value...
📡 Uploading encrypted secret to server...
✅ Secret 'API_KEY' set successfully!
   Version: 1
   Updated: 2024-01-15T10:30:00Z
   Created by: my-laptop
```

### `initiat secret get <secret-key> [options]`

Get and decrypt a secret value from the specified workspace.

**Arguments:**
- `secret-key`: The key/name for the secret (required)

**Options:**
- `--copy, -c`: Copy value to clipboard instead of printing
- `--copy-kv`: Copy KEY=VALUE format to clipboard
- `--workspace-path, -W`: Full workspace path (org/workspace) or alias
- `--workspace, -w`: Workspace name (uses default org or --org)
- `--org`: Organization slug (used with --workspace)

**Examples:**
```bash
# Get secret with full workspace path
initiat secret get API_KEY --workspace-path acme-corp/production

# Get secret with short flags
initiat secret get API_KEY -W acme-corp/production

# Get secret and copy value to clipboard
initiat secret get API_KEY -w production --copy

# Get secret and copy KEY=VALUE format to clipboard
initiat secret get API_KEY -w production --copy-kv
```

**What it does:**
1. Retrieves encrypted secret from server
2. Gets workspace key and decrypts it
3. Decrypts secret value client-side
4. Outputs JSON with secret metadata (default)
5. Optionally copies value to clipboard (`--copy`)
6. Optionally copies KEY=VALUE format to clipboard (`--copy-kv`)

**Output:**
```
🔍 Getting secret 'API_KEY' from workspace acme-corp/production...
🔓 Decrypting secret value...
{
  "key": "API_KEY",
  "value": "sk-1234567890abcdef",
  "version": 1,
  "workspace_id": "ws_abc123",
  "updated_at": "2024-01-15T10:30:00Z",
  "created_by_device": "my-laptop"
}
```

### `initiat secret list [options]`

List all secrets in the specified workspace (metadata only, no values).

**Options:**
- `--workspace-path, -W`: Full workspace path (org/workspace) or alias
- `--workspace, -w`: Workspace name (uses default org or --org)
- `--org`: Organization slug (used with --workspace)

**Examples:**
```bash
# List secrets with full workspace path
initiat secret list --workspace-path acme-corp/production

# List secrets with short flags
initiat secret list -W acme-corp/production

# List secrets with workspace only
initiat secret list --workspace production
```

**What it does:**
1. Fetches all secrets for the workspace
2. Displays metadata in table format
3. Shows key, encrypted status, and version
4. Never exposes actual secret values

**Output:**
```
🔍 Listing secrets in workspace acme-corp/production...

Key        Value        Version
API_KEY    [encrypted]  1
DB_PASS    [encrypted]  1
JWT_SECRET [encrypted]  2
```

### `initiat secret delete <secret-key> [options]`

Delete a secret from the specified workspace.

**Arguments:**
- `secret-key`: The key/name for the secret (required)

**Options:**
- `--force, -f`: Skip confirmation prompt
- `--workspace-path, -W`: Full workspace path (org/workspace) or alias
- `--workspace, -w`: Workspace name (uses default org or --org)
- `--org`: Organization slug (used with --workspace)

**Examples:**
```bash
# Delete secret with confirmation
initiat secret delete API_KEY --workspace-path acme-corp/production

# Delete secret with short flags
initiat secret delete API_KEY -W acme-corp/production

# Force delete without confirmation
initiat secret delete OLD_API_KEY --workspace production --force
```

**What it does:**
1. Prompts for confirmation (unless --force is used)
2. Deletes secret from server
3. Shows confirmation message

**Output:**
```
⚠️  Are you sure you want to delete secret 'API_KEY' from workspace acme-corp/production? (y/N): y
🗑️  Deleting secret 'API_KEY' from workspace acme-corp/production...
✅ Secret 'API_KEY' deleted successfully!
```

### `initiat secret export <secret-key> --output FILE [options]`

Export a secret value to a file. Creates directories if needed and handles overwrite prompts.

**Arguments:**
- `secret-key`: The key/name for the secret (required)

**Options:**
- `--output, -o`: Output file path (required)
- `--force, -f`: Overwrite existing key without confirmation
- `--workspace-path, -W`: Full workspace path (org/workspace) or alias
- `--workspace, -w`: Workspace name (uses default org or --org)
- `--org`: Organization slug (used with --workspace)

**Examples:**
```bash
# Export secret to a file
initiat secret export API_KEY --output .env --workspace-path acme-corp/production

# Export to deep directory (creates folders)
initiat secret export API_KEY --output config/secrets.env -W acme-corp/production

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
🔍 Getting secret 'API_KEY' from workspace acme-corp/production...
🔓 Decrypting secret value...
⚠️  File 'secrets.env' is not in .gitignore. Add it? (y/N): y
✅ Added 'secrets.env' to .gitignore
✅ Secret 'API_KEY' exported to secrets.env
```

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
❌ Failed to get workspace key: workspace key not initialized
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
❌ Invalid workspace path: expected 'org-slug/workspace-slug'
```

## Best Practices

### Workspace Organization
- Use descriptive workspace names: `acme-corp/production`, `acme-corp/staging`
- Initialize workspace keys before storing secrets
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
- Never share device credentials or workspace keys
- Use `--force` flag carefully with secret operations
- Regularly audit device access and remove unused devices
- Keep CLI updated to latest version for security patches
