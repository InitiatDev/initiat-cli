# Setup Scripts Documentation

**Version 1** — Declarative development environment setup for Initiat CLI

`.initiat/setup.yml` allows you to define a complete, reproducible development environment setup that works across macOS, Linux, and Windows. This document provides a formal specification of the syntax and behavior.

## Table of Contents

1. [Overview](#overview)
2. [File Location](#file-location)
3. [Top-Level Structure](#top-level-structure)
4. [Execution Phases](#execution-phases)
5. [Step Configuration](#step-configuration)
6. [Actions](#actions)
7. [Conditions DSL](#conditions-dsl)
8. [Environment & Secrets](#environment--secrets)
9. [Defaults & Overrides](#defaults--overrides)
10. [Matrix Gating](#matrix-gating)
11. [Execution Flow](#execution-flow)
12. [Examples](#examples)
13. [Best Practices](#best-practices)

---

## Overview

Setup scripts define how to transform a bare system into a fully functional development environment. They are:

- **Declarative**: Describe what you want, not how to get it
- **Idempotent**: Safe to run multiple times
- **Cross-platform**: Works on macOS, Linux, and Windows
- **Secure**: Integrates with Initiat's zero-knowledge secret management

### How It Works

1. **Validation**: The script is parsed and validated against a JSON Schema
2. **Matrix Check**: Verifies the host OS/arch matches the script's requirements
3. **Phase Execution**: Phases run sequentially (bootstrap → provision → setup → verify → post)
4. **Step Processing**: Each step's condition is evaluated, and if true, the action is executed
5. **Secret Injection**: Secrets are fetched from Initiat, decrypted locally, and injected as environment variables
6. **Command Execution**: Commands run with timeouts, retries, and error handling

---

## File Location

| Property     | Value                    |
| ------------ | ------------------------ |
| **Path**     | `.initiat/setup.yml`     |
| **Format**   | YAML 1.2                 |
| **Encoding** | UTF-8                    |
| **Required** | No (only if using setup) |

---

## Top-Level Structure

```yaml
version: 1
name: "My Project Setup"
description: "Zero-to-ready dev environment"

matrix:
  os: [macos, linux, windows]
  arch: [x86_64, arm64]

defaults:
  timeout: "15m"
  shell: auto
  continue_on_error: false
  cwd: "."

env:
  PROJECT_NAME: "myproject"
  secrets: [DATABASE_URL, OPENAI_API_KEY]

bootstrap:
  - ...steps...
provision:
  - ...steps...
setup:
  - ...steps...
verify:
  - ...steps...
post:
  - ...steps...
```

### Top-Level Fields

| Field         | Type   | Required | Description                                  |
| ------------- | ------ | -------- | -------------------------------------------- |
| `version`     | int    | ✅ Yes   | Must be `1`                                  |
| `name`        | string | No       | Human-readable name for the setup            |
| `description` | string | No       | Description of what this setup does          |
| `matrix`      | object | No       | OS/architecture requirements (see [Matrix](#matrix-gating)) |
| `defaults`    | object | No       | Default values for steps (see [Defaults](#defaults--overrides)) |
| `env`         | object | No       | Global environment variables and secrets      |
| `bootstrap`   | array  | No       | Base system prerequisites                    |
| `provision`    | array  | No       | Runtimes, databases, system services         |
| `setup`       | array  | No       | Project dependencies, migrations, seeding     |
| `verify`      | array  | No       | Health checks and assertions                 |
| `post`        | array  | No       | Completion messages and next steps            |

---

## Execution Phases

Phases execute in a fixed order, with each phase running to completion before the next begins.

| Phase       | Purpose                                                                  | Typical Use Cases                                    |
| ----------- | ------------------------------------------------------------------------ | ---------------------------------------------------- |
| **bootstrap** | Base system prerequisites                                               | Package managers, Git, build tools                  |
| **provision** | Language runtimes, databases, system services                           | Node.js, PostgreSQL, Redis                          |
| **setup**     | Project-specific dependencies and configuration                          | `npm install`, `mix deps.get`, database migrations  |
| **verify**    | Assertions and health checks                                            | Version checks, HTTP endpoints, command execution   |
| **post**      | User-facing messages and instructions                                   | Success messages, next steps                         |

### Phase Execution Rules

- Empty or missing phases are skipped
- If any step fails and `continue_on_error` is `false`, the phase stops
- Later phases are not executed if earlier phases fail (unless `continue_on_error` is true)
- Steps within a phase execute sequentially

---

## Step Configuration

Each step defines metadata and exactly one action.

```yaml
- name: "Install dependencies"
  if: file_exists("package.json")
  timeout: "10m"
  cwd: "frontend"
  env:
    NODE_ENV: development
  env_from_secrets: [DATABASE_URL]
  optional_secrets: false
  continue_on_error: false
  retries:
    attempts: 3
    backoff: "2s"
  run: "npm install"
```

### Step Fields

| Field               | Type    | Required | Description                                      |
| ------------------- | ------- | -------- | ------------------------------------------------ |
| `name`              | string  | No       | Display name in logs                             |
| `if`                | string  | No       | Condition DSL expression (see [Conditions](#conditions-dsl)) |
| `timeout`           | string  | No       | Step timeout (overrides default, e.g., `"10m"`)  |
| `cwd`               | string  | No       | Working directory (relative to repo root)        |
| `env`               | object  | No       | Static environment variables                     |
| `env_from_secrets`  | array   | No       | Secret names to inject from Initiat              |
| `optional_secrets` | bool    | No       | Don't fail if secret missing                      |
| `continue_on_error`  | bool    | No       | Continue even if step fails                      |
| `retries`           | object  | No       | Retry policy (see [Retries](#retries--timeouts)) |
| *one action*        | varies  | ✅ Yes   | Exactly one action field (see [Actions](#actions)) |

---

## Actions

Each step must define exactly one action. Actions are mutually exclusive.

### `run` — Execute Shell Command

Run a shell command with environment variables and secrets injected.

```yaml
run: "npm ci && npm run build"
```

- Uses the shell from `defaults.shell` (or `auto` → bash/sh/powershell)
- Inherits all `env` variables and secrets
- Secret values are redacted in logs
- Standard output/error are displayed

### `print` — Print Message

Display a message to the user.

```yaml
print: "✅ Setup complete. Run: mix phx.server"
```

- Always executes (unless `if` condition is false)
- No shell execution, just output

### `ensure_package_manager` — Ensure Package Manager

Install a system package manager if not present.

```yaml
ensure_package_manager:
  type: auto  # auto|brew|apt|choco
```

**Supported Types:**
- `auto` — Automatically detect and install based on OS (brew for macOS, apt for Linux, choco for Windows)
- `brew` — Homebrew (macOS)
- `apt` — APT package manager (Linux)
- `choco` — Chocolatey (Windows)

**Behavior:**
- Checks if the package manager is available
- If not present, installs it automatically
- Idempotent: safe to run multiple times

### `ensure_tool` — Ensure CLI Tool

Install a command-line tool with optional version checking.

```yaml
ensure_tool:
  name: git
  version: ">=2.34"
  install:
    brew: { formula: git }
    apt: { packages: [git], update: true }
    choco: { packages: [git] }
    fallback_url: "https://example.com/git.zip"
    checksum: "sha256:abc123..."
```

**Fields:**

| Field            | Type   | Required | Description                              |
| ---------------- | ------ | -------- | -------------------------------------- |
| `name`           | string | ✅ Yes   | Tool name (for version checking)       |
| `version`        | string | No       | Version constraint (e.g., `">=2.34"`)  |
| `install.brew`   | object | No       | Homebrew installation config           |
| `install.apt`    | object | No       | APT installation config                |
| `install.choco`  | object | No       | Chocolatey installation config          |
| `install.fallback_url` | string | No | URL if package managers fail (requires `checksum`) |
| `install.checksum` | string | No | Checksum for fallback URL (`sha256:...`) |

**Install Config:**

```yaml
brew:
  formula: git  # Homebrew formula name

apt:
  packages: [git, git-core]  # Package names
  update: true                # Run apt update first

choco:
  packages: [git]  # Chocolatey package names
```

**Behavior:**
- Checks if tool exists and version matches
- If missing or version mismatch, installs using the appropriate package manager for the OS
- Falls back to `fallback_url` if package manager installation fails (requires checksum verification)

### `ensure_runtime` — Ensure Language Runtime

Install a language runtime (Node.js, Python, Go, etc.).

```yaml
ensure_runtime:
  name: elixir
  version: "1.16.x"
  manager:
    asdf: true
  fallback_installers:
    - brew: { formula: elixir }
    - apt: { packages: [elixir] }
    - choco: { packages: [elixir] }
```

**Supported Runtimes:**
- `node` — Node.js
- `python` — Python
- `go` — Go
- `elixir` — Elixir
- `erlang` — Erlang
- `java` — Java (OpenJDK)
- `rust` — Rust
- `dotnet` — .NET

**Fields:**

| Field                | Type    | Required | Description                                  |
| -------------------- | ------- | -------- | -------------------------------------------- |
| `name`               | string  | ✅ Yes   | Runtime name (see list above)               |
| `version`            | string  | No       | Version constraint (e.g., `"1.16.x"`, `">=20"`) |
| `manager.asdf`       | bool    | No       | Prefer asdf for version management            |
| `fallback_installers` | array  | No       | Alternative installers if asdf unavailable      |

**Behavior:**
- Prefers `asdf` if available and `manager.asdf: true`
- Falls back to OS package managers in order if asdf unavailable
- Versions are pinned in `.tool-versions` (asdf) or via package manager
- Install steps are executed sequentially in `fallback_installers` order

**Note:** For dependent runtimes (e.g., Elixir requires Erlang), ensure dependencies are installed in separate steps earlier in the phase.

### `ensure_database` — Ensure Database Service

Install and configure a database.

```yaml
ensure_database:
  engine: postgres
  version: "15"
  service_name: "postgres"
  ensure_running: true
  create_db: ["app_dev", "app_test"]
```

**Supported Engines:**
- `postgres` — PostgreSQL
- `mysql` — MySQL/MariaDB
- `sqlite` — SQLite (no installation needed)

**Fields:**

| Field           | Type     | Required | Description                                  |
| --------------- | -------- | -------- | -------------------------------------------- |
| `engine`        | string   | ✅ Yes   | Database engine (see list above)            |
| `version`       | string   | No       | Version constraint                            |
| `service_name`  | string   | No       | System service name                           |
| `ensure_running` | bool    | No       | Ensure the service is running                 |
| `create_db`     | array    | No       | Database names to create                     |

**Behavior:**
- Installs the database if not present
- Starts the service if `ensure_running: true`
- Creates databases listed in `create_db`
- Service management uses OS-appropriate tools (systemd, brew services, Windows services)

### `assert_command` — Assert Command Success

Verify that a command exits successfully.

```yaml
assert_command: "mix --version"
```

**Behavior:**
- Runs the command
- Fails the step if exit code is non-zero
- Useful for health checks and verification

### `assert_http` — Assert HTTP Endpoint

Poll an HTTP endpoint until it returns the expected status code.

```yaml
assert_http:
  url: "http://localhost:4000/health"
  expect_status: 200
  retries:
    attempts: 20
    backoff: "1s"
```

**Fields:**

| Field           | Type   | Required | Description                                  |
| --------------- | ------ | -------- | -------------------------------------------- |
| `url`           | string | ✅ Yes   | HTTP/HTTPS URL to check                      |
| `expect_status` | int    | No       | Expected status code (default: `200`)        |
| `retries`       | object | No       | Retry policy (see [Retries](#retries--timeouts)) |

**Behavior:**
- Uses `curl` on macOS/Linux, `PowerShell` on Windows
- Retries with exponential backoff until success or max attempts
- Useful for waiting for services to start

---

## Conditions DSL

The `if` field accepts a boolean expression that determines whether a step executes.

### Functions

| Function              | Syntax                          | Description                                  |
| --------------------- | ------------------------------- | -------------------------------------------- |
| `os()`                | `os("macos")` or `os("linux")` or `os("windows")` | Match current operating system                |
| `arch()`               | `arch("arm64")` or `arch("x86_64")` | Match processor architecture                  |
| `file_exists()`        | `file_exists("path/to/file")`   | Check if file or directory exists            |
| `cmd_ok()`             | `cmd_ok("command")`              | True if command exits with code 0            |

### Logical Operators

- `&&` — AND
- `||` — OR
- `!` — NOT
- `()` — Grouping

### Examples

```yaml
if: os("macos") && file_exists("mix.exs")
if: !cmd_ok("asdf --version")
if: os("linux") || os("macos")
if: file_exists("package.json") && !file_exists("package-lock.json")
if: (os("macos") || os("linux")) && arch("arm64")
```

---

## Environment & Secrets

### Global Environment

Define environment variables at the top level:

```yaml
env:
  NODE_ENV: development
  PROJECT_NAME: "myproject"
  secrets: [DATABASE_URL, OPENAI_API_KEY]
```

**Rules:**
- Static variables are available to all steps
- `secrets` declares which Initiat secrets to fetch
- Secrets are encrypted and decrypted locally (zero-knowledge)

### Step-Level Environment

Override or extend environment per step:

```yaml
setup:
  - run: mix ecto.setup
    env:
      MIX_ENV: dev
    env_from_secrets: [DATABASE_URL]
```

**Merging Rules:**
- Step `env` overrides global `env`
- `env_from_secrets` injects decrypted secrets
- Secrets are never logged or written to disk
- Secret values are redacted in command output

### Secret Workflow

1. Declare secret names in `env.secrets`
2. Reference them in steps via `env_from_secrets`
3. CLI fetches secrets from Initiat API
4. Secrets are decrypted client-side using project keys
5. Values are injected as environment variables
6. Values are redacted in all logs

---

## Defaults & Overrides

Set default values for all steps:

```yaml
defaults:
  timeout: "15m"
  shell: auto
  continue_on_error: false
  cwd: "."
```

**Defaults:**

| Field               | Type    | Default      | Description                                      |
| ------------------- | ------- | ------------ | ------------------------------------------------ |
| `timeout`           | string  | `"5m"`       | Default step timeout (e.g., `"10m"`, `"30s"`)    |
| `shell`             | string  | `"auto"`     | Shell to use (`auto` → bash/sh/powershell)      |
| `continue_on_error` | bool    | `false`      | Continue execution if step fails                |
| `cwd`               | string  | `"."`        | Default working directory (relative to repo)    |

**Overrides:**

Any default can be overridden per step:

```yaml
- name: "Long-running task"
  timeout: "30m"  # Override default
  continue_on_error: true
  run: "npm run build"
```

---

## Matrix Gating

Restrict execution to specific OS/architecture combinations:

```yaml
matrix:
  os: [macos, linux]
  arch: [arm64, x86_64]
```

**Supported Values:**

| Field | Values                              |
| ----- | ----------------------------------- |
| `os`  | `macos`, `linux`, `windows`        |
| `arch` | `x86_64`, `arm64`                  |

**Behavior:**
- If `matrix` is specified, execution halts early if the host doesn't match
- Empty `matrix` means "run on all platforms"
- Useful for OS-specific setup scripts

**Example:**

```yaml
# Only run on macOS
matrix:
  os: [macos]

# Only run on Linux ARM
matrix:
  os: [linux]
  arch: [arm64]
```

---

## Retries & Timeouts

### Retries

Retry failed steps automatically:

```yaml
retries:
  attempts: 3
  backoff: "2s"
```

**Fields:**
- `attempts` — Number of retry attempts (minimum: 1)
- `backoff` — Delay between retries (format: `"1s"`, `"2m"`, `"30s"`)

**Behavior:**
- Exponential backoff: delay doubles after each retry
- Only transient failures should be retried
- HTTP assertions (`assert_http`) use retries by default

### Timeouts

Limit how long a step can run:

```yaml
timeout: "10m"  # 10 minutes
```

**Format:** `"<number><unit>"` where unit is `s` (seconds), `m` (minutes), or `h` (hours)

**Behavior:**
- Applies to the entire command process tree
- Step is killed if timeout exceeded
- Default timeout is `5m` if not specified

---

## Execution Flow

1. **Parse & Validate**: YAML is parsed and validated against JSON Schema
2. **Matrix Check**: If `matrix` is specified, verify host OS/arch matches
3. **Phase Loop**: For each phase (bootstrap, provision, setup, verify, post):
   a. **Step Loop**: For each step in the phase:
      - Evaluate `if` condition (skip if false)
      - Fetch secrets if `env_from_secrets` is specified
      - Execute action with timeout and retries
      - Handle errors per `continue_on_error`
   b. **Phase Complete**: Move to next phase
4. **Success**: All phases completed

**Error Handling:**
- If `continue_on_error: false` and a step fails, phase stops and execution aborts
- If `continue_on_error: true`, failed steps are logged but execution continues
- Retries are attempted before considering a step failed

---

## Examples

### Minimal Setup

```yaml
version: 1
name: "Minimal Setup"
```

### Node.js Project

```yaml
version: 1
name: "Node.js Setup"

provision:
  - ensure_runtime:
      name: node
      version: ">=20"
      manager: { asdf: true }
      fallback_installers:
        - brew: { formula: node }
        - apt: { packages: [nodejs] }

setup:
  - run: npm ci

verify:
  - assert_command: "node --version"
```

### Phoenix/Elixir Project

See `test/fixtures/setup_examples/phoenix_basic.yml` for a complete example.

---

## Best Practices

### ✅ Do

- **Use conditions** to make steps OS-specific when needed
- **Declare all secrets** in `env.secrets` upfront
- **Make actions idempotent** (safe to run multiple times)
- **Use descriptive step names** for better logging
- **Test on all target platforms**
- **Use `assert_http` with retries** for waiting on services
- **Keep setup scripts version-controlled**

### ❌ Don't

- **Hard-code secret values** (use Initiat secrets)
- **Write secrets to `.env` files** or commit them
- **Use `sudo` in `run` commands** (package managers handle elevation)
- **Assume OS-specific tools** without `if` conditions
- **Create non-idempotent steps** (e.g., creating files that already exist)
- **Skip validation** (`initiat setup validate` before committing)

---

## Validation

Validate your setup script before committing:

```bash
initiat setup validate
```

This checks:
- YAML syntax
- JSON Schema compliance
- Secret references
- Path safety
- Required fields

Generate the JSON Schema:

```bash
initiat setup schema --output schemas/setup-v1.json
```

---

## Related Documentation

- [Setup Implementation Plan](SETUP_IMPLEMENTATION_PLAN.md) — Internal implementation details
- [Command Reference](COMMANDS.md) — CLI command documentation
- [Security Architecture](SECURITY.md) — Secret management details

---

**Last Updated**: Version 1.0

