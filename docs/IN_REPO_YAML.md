# In-Repo YAML Contract

This document defines the `.initiat/` directory layout and the YAML files used for offline-first workflows. All of this lives in the repository and runs without an Initiat server.

## Directory Layout

```
.initiat/
├── config.yml      # Local metadata (repo name, owner, links); optional remote context only if using cloud
├── setup.yml       # Development environment setup (phases, steps, commands)
└── local/          # Optional: local-only files (gitignored)
```

- **config.yml**: Local project metadata. May include optional `org`/`project` only when using cloud features.
- **setup.yml**: Canonical setup script. See [Setup Scripts](SETUP_SCRIPTS.md).
- **local/**: Directory for local-only, gitignored files (e.g. `.initiat/local/dev.env`). Never committed.

## config.yml

Used for local context and optional cloud context.

**Local-only (default, offline):**

```yaml
# Optional: display name and links (no server required)
name: "My Project"
repo: "my-org/my-repo"
links:
  - label: "Contributing"
    url: "CONTRIBUTING.md"
  - label: "API"
    url: "https://api.example.com/docs"
```

**Optional remote context (only when using cloud):**

```yaml
org: my-org
project: my-project
```

When `org` and `project` are present, cloud commands (e.g. secret, env sync) may use them. They are not required for `initiat setup generate`, `initiat setup validate`, or `initiat setup run`.

## setup.yml

Fully specified in [Setup Scripts](SETUP_SCRIPTS.md). Summary:

- **version**: Must be `1`.
- **matrix**: Optional OS/arch restriction.
- **defaults**: timeout, shell, continue_on_error, cwd.
- **Phases**: bootstrap, provision, setup, verify, post (each a list of steps).
- **Steps**: Each step has optional `name`, `if`, `timeout`, `cwd`, `env`, **secrets** (optional, when using cloud), and exactly one action: `run` or `print`.

### Optional: Initiat-hosted secrets (env.secrets / secrets)

When using Initiat's optional cloud add-on, you can declare secret names in **env.secrets** (top-level in setup.yml) and inject them per step with **secrets**. Requires project context (e.g. `org`/`project` in config.yml or `initiat project init`) and device registration. See [Setup Scripts](SETUP_SCRIPTS.md) and [Security](SECURITY.md).

**Example:**

```yaml
env:
  NODE_ENV: development
  secrets: [DATABASE_URL, OPENAI_API_KEY]

setup:
  - name: "Run migrations"
    secrets: [DATABASE_URL]
    run: "mix ecto.migrate"
```

## Migration

- **config.yml**: Prefer local-only fields (`name`, `repo`, `links`). Add `org`/`project` only if you use cloud.
- **setup.yml**: Optional `env.secrets` and `secrets` require project context when used; see [Setup Scripts](SETUP_SCRIPTS.md).
- **New repos**: Use `initiat setup generate` to scaffold `.initiat/setup.yml` and `.initiat/config.yml` from detected language/framework templates.

## Related

- [Setup Scripts](SETUP_SCRIPTS.md) — Full setup.yml reference and conditions DSL
- [Command Reference](COMMANDS.md) — Offline-first and cloud commands
