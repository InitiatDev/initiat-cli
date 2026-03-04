# Setup Scripts Documentation

**Version 1** — Explicit, command-based development environment setup for Initiat CLI

`.initiat/setup.yml` allows you to define a complete, reproducible development environment setup that works across macOS, Linux, and Windows. This document provides a formal specification of the syntax and behavior.

## Table of Contents

1. [Overview](#overview)
2. [Generating setup from your project](#generating-setup-from-your-project)
3. [File Location](#file-location)
4. [Top-Level Structure](#top-level-structure)
5. [Execution Phases](#execution-phases)
6. [Step Configuration](#step-configuration)
7. [Actions](#actions)
8. [Conditions DSL](#conditions-dsl)
9. [Environment & Secrets](#environment--secrets)
10. [Defaults & Overrides](#defaults--overrides)
11. [Matrix Gating](#matrix-gating)
12. [Execution Flow](#execution-flow)
13. [Examples](#examples)
14. [Best Practices](#best-practices)

---

## Overview

Setup scripts define how to transform a bare system into a fully functional development environment. They are:

- **Explicit**: Commands are written directly, similar to GitHub Actions
- **Idempotent**: Safe to run multiple times (use conditions to check for existing installations)
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

## Generating setup from your project

Instead of writing `.initiat/setup.yml` by hand, you can generate it from the current repository:

```bash
initiat setup generate
```

**What it does:**

- **Detection:** Scans the current directory for stack markers and applies built-in templates:
  - **Go**: `go.mod`
  - **Node**: `package.json` or `assets/package.json` (e.g. Phoenix assets)
  - **Python**: `pyproject.toml` or `requirements.txt`
  - **Phoenix**: `mix.exs`
  - **Rails**: `Gemfile`
  - **Docker**: `docker-compose.yml` or `compose.yml`
- **Templates:** Each template adds phase steps (provision, setup, verify, post) with verify-only or minimal commands. Multiple stacks can be combined (e.g. Phoenix + Node).
- **Service inference:** If Docker Compose is present, steps for running `docker compose up -d` are added. Otherwise, dependency files (Gemfile, mix.exs, package.json, requirements.txt, pyproject.toml) are scanned for PostgreSQL, MySQL, SQLite, and Redis; corresponding verify steps (e.g. `psql --version`, `redis-cli --version`) are appended. When more than one database is detected, a post message suggests editing `.initiat/setup.yml` to remove unused steps.
- **Output:** Writes `.initiat/setup.yml` and `.initiat/config.yml` (project name is derived from the directory). Does not overwrite an existing `setup.yml` unless you pass `--force`.

After generating, run `initiat setup validate` and then `initiat setup run` (or `initiat project setup` if you use cloud secrets). See [Command Reference](COMMANDS.md#initiat-setup-generate---force) for details.

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
  secrets: [DATABASE_URL]

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
| `env`         | object | No       | Global environment variables; optional `env.secrets` for Initiat-hosted secrets |
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
  secrets: [DATABASE_URL]
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
| `secrets`  | array   | No       | Secret names to inject from Initiat (optional add-on; requires project context) |
| `optional_secrets` | bool    | No       | Don't fail if secret missing |
| `continue_on_error`  | bool    | No       | Continue even if step fails                      |
| `retries`           | object  | No       | Retry policy (see [Retries](#retries--timeouts)) |
| *one action*        | varies  | ✅ Yes   | Exactly one action field: `run` or `print`       |

---

## Actions

Each step must define exactly one action. Actions are mutually exclusive.

### `run` — Execute Shell Command

Run a shell command with environment variables from `env` and `secrets` (when using the cloud add-on) injected.

```yaml
run: "npm ci && npm run build"
```

For multi-line commands, use YAML block scalars:

```yaml
run: |
  asdf plugin add golang || true
  asdf install golang 1.25.7
  asdf global golang 1.25.7
```

**Behavior:**
- Uses the shell from `defaults.shell` (or `auto` → bash/sh/powershell)
- Inherits all `env` and (when project context exists) `secrets` values
- Secret values are never logged in plaintext
- Standard output/error are displayed
- Commands are explicit and visible in the YAML

**Best Practices:**
- Use `if` conditions with `cmd_ok()` to check if tools are already installed
- Use `|| exit 1` for assertions (e.g., `go version || exit 1`)
- Use `|| true` for idempotent operations (e.g., `asdf plugin add golang || true`)
- For OS-specific commands, use `if: os("macos")` conditions

### `print` — Print Message

Display a message to the user.

```yaml
print: "✅ Setup complete. Run: mix phx.server"
```

For multi-line messages:

```yaml
print: |
  ✅ Development environment setup complete!

  You can now build the project with:
    • make build
    • go build -o initiat .
```

**Behavior:**
- Always executes (unless `if` condition is false)
- No shell execution, just output
- Useful for completion messages and next steps

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
- `!` — NOT (must be quoted in YAML if at start: `'!os("windows")'`)
- `()` — Grouping

### Examples

```yaml
if: os("macos") && file_exists("mix.exs")
if: '!cmd_ok("command -v asdf")'  # Note: quote when starting with !
if: os("linux") || os("macos")
if: file_exists("package.json") && !file_exists("package-lock.json")
if: (os("macos") || os("linux")) && arch("arm64")
if: os("macos") && !cmd_ok("command -v brew")  # Install if not present
if: '!os("windows")'  # Unix-like systems
```

**Important:** When using `!` at the start of a condition, quote the entire condition string to avoid YAML tag syntax issues:
- ✅ `if: '!os("windows")'`
- ❌ `if: !os("windows")` (YAML interprets `!os` as a tag)

---

## Environment & Secrets

Use static **env** for variables. Optionally use Initiat-hosted secrets via **env.secrets** and **secrets** when you use the cloud add-on (project context + device registration).

### Global Environment

Define environment variables at the top level:

```yaml
env:
  NODE_ENV: development
  PROJECT_NAME: "myproject"
  secrets: [DATABASE_URL, OPENAI_API_KEY]
```

**Rules:**
- Static variables (e.g. `NODE_ENV`, `PROJECT_NAME`) are available to all steps.
- **secrets** declares which Initiat secrets to fetch when using the optional cloud add-on (requires project context and device registration).

### Step-Level Environment

Override or extend per step:

```yaml
setup:
  - run: mix ecto.setup
    env:
      MIX_ENV: dev
    secrets: [DATABASE_URL]
```

**Merging rules:** Step `env` overrides global `env`. `secrets` injects decrypted secrets from Initiat when the cloud add-on is used. Secrets are never logged or written to disk.

### Optional: Initiat-hosted secrets (env.secrets / secrets)

When using Initiat's optional cloud add-on, you can declare secret names in **env.secrets** and inject them per step with **secrets**. Requires project context (`.initiat/config.yml` with org/project or `initiat project init`) and device registration. See [Security](SECURITY.md) and [In-Repo YAML](IN_REPO_YAML.md).

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
- Useful for network operations or waiting for services

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
      - Load secrets from Initiat when `secrets` is used (requires project context)
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

bootstrap:
  - name: "Install asdf (macOS/Linux)"
    if: '!os("windows") && !cmd_ok("command -v asdf")'
    run: |
      git clone https://github.com/asdf-vm/asdf.git ~/.asdf --branch v0.14.0 || true
      echo '. "$HOME/.asdf/asdf.sh"' >> ~/.bashrc
      . "$HOME/.asdf/asdf.sh"

provision:
  - name: "Install Node.js via asdf"
    if: '!os("windows") && !cmd_ok("node --version")'
    run: |
      . "$HOME/.asdf/asdf.sh" || true
      asdf plugin add nodejs || true
      asdf install nodejs 20.10.0
      asdf global nodejs 20.10.0

  - name: "Install Node.js via Chocolatey (Windows)"
    if: os("windows") && !cmd_ok("node --version")
    run: choco install nodejs -y

setup:
  - name: "Install dependencies"
    run: npm ci

verify:
  - name: "Check Node version"
    run: node --version || exit 1
```

### Go Project

```yaml
version: 1
name: "Go Setup"

bootstrap:
  - name: "Install Git (macOS)"
    if: os("macos") && !cmd_ok("command -v git")
    run: brew install git

  - name: "Install Git (Linux)"
    if: os("linux") && !cmd_ok("command -v git")
    run: sudo apt-get install -y git

  - name: "Install asdf (macOS/Linux)"
    if: '!os("windows") && !cmd_ok("command -v asdf")'
    run: |
      git clone https://github.com/asdf-vm/asdf.git ~/.asdf --branch v0.14.0 || true
      echo '. "$HOME/.asdf/asdf.sh"' >> ~/.bashrc
      . "$HOME/.asdf/asdf.sh"

provision:
  - name: "Install Go via asdf"
    if: '!os("windows") && !cmd_ok("go --version")'
    run: |
      . "$HOME/.asdf/asdf.sh" || true
      asdf plugin add golang || true
      asdf install golang 1.25.7
      asdf global golang 1.25.7

  - name: "Install Go via Chocolatey (Windows)"
    if: os("windows") && !cmd_ok("go --version")
    run: choco install golang -y

setup:
  - name: "Download dependencies"
    run: go mod download

verify:
  - name: "Verify Go version"
    run: go version || exit 1

  - name: "Build the project"
    run: go build -o myapp .
```

### Phoenix/Elixir Project

```yaml
version: 1
name: "Phoenix Setup"

bootstrap:
  - name: "Install asdf (macOS/Linux)"
    if: '!os("windows") && !cmd_ok("command -v asdf")'
    run: |
      git clone https://github.com/asdf-vm/asdf.git ~/.asdf --branch v0.14.0 || true
      echo '. "$HOME/.asdf/asdf.sh"' >> ~/.bashrc
      . "$HOME/.asdf/asdf.sh"

provision:
  - name: "Install Erlang via asdf"
    if: '!os("windows") && !cmd_ok("erl -version")'
    run: |
      . "$HOME/.asdf/asdf.sh" || true
      asdf plugin add erlang || true
      asdf install erlang 26.1.2
      asdf global erlang 26.1.2

  - name: "Install Elixir via asdf"
    if: '!os("windows") && !cmd_ok("elixir --version")'
    run: |
      . "$HOME/.asdf/asdf.sh" || true
      asdf plugin add elixir || true
      asdf install elixir 1.16.1
      asdf global elixir 1.16.1

  - name: "Install PostgreSQL (macOS)"
    if: os("macos") && !cmd_ok("psql --version")
    run: |
      brew install postgresql@15
      brew services start postgresql@15

  - name: "Install PostgreSQL (Linux)"
    if: os("linux") && !cmd_ok("psql --version")
    run: |
      sudo apt-get install -y postgresql-15
      sudo systemctl start postgresql
      sudo systemctl enable postgresql

setup:
  - name: "Get dependencies"
    run: mix deps.get

  - name: "Setup database"
    secrets: [DATABASE_URL]
    run: mix ecto.setup

verify:
  - name: "Check Elixir version"
    run: elixir --version || exit 1

  - name: "Start server and check health"
    run: |
      mix phx.server &
      sleep 5
      curl -f http://localhost:4000/health || exit 1
```

---

## Best Practices

### ✅ Do

- **Use explicit commands** — Write commands directly in `run:` fields
- **Check for existing installations** — Use `if: !cmd_ok("command -v tool")` to make steps idempotent
- **Use conditions for OS-specific steps** — `if: os("macos")` for macOS-only commands
- **Quote conditions starting with `!`** — Use `'!os("windows")'` to avoid YAML tag syntax
- **Use `|| exit 1` for assertions** — `go version || exit 1` fails if command fails
- **Use `|| true` for idempotent operations** — `asdf plugin add golang || true` won't fail if already added
- **Use secrets** with Initiat when using the cloud add-on; declare names in `env.secrets`
- **Use descriptive step names** for better logging
- **Test on all target platforms**
- **Keep setup scripts version-controlled**

### ❌ Don't

- **Hard-code secret values** (use secrets with Initiat or supply via env from your own store)
- **Commit secret files** (use `.initiat/local/` and gitignore; never commit real secrets)
- **Skip condition checks** — Always check if tools are installed before installing
- **Assume OS-specific tools** without `if` conditions
- **Create non-idempotent steps** — Always check before installing/creating
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

- [Command Reference](COMMANDS.md) — CLI command documentation
- [Security Architecture](SECURITY.md) — Secret management details

---

**Last Updated**: Version 1.0
