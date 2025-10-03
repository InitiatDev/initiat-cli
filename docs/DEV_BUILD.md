# Development Build

This document explains how to build and use a development version of the Initiat CLI that defaults to a local API server.

## Quick Start

```bash
# Build the dev binary
make build-dev

# Use it immediately
./initiat_dev auth login dev@example.com
./initiat_dev workspace list
```

## What's Different?

The development binary (`initiat_dev`) is identical to the production binary except:

- **Default API URL**: `http://localhost:4000` (instead of `https://www.initiat.dev`)
- **Binary Name**: `initiat_dev` (to avoid confusion with production binary)

Everything else works the same way:
- Same commands and flags
- Same configuration file support
- Same environment variable overrides
- Can still override the API URL with `--api-url` flag

## Use Cases

### Local Backend Development

When developing the Initiat backend API:

```bash
# Build once
make build-dev

# Start your local backend on port 4000
cd ../initiat-backend
make dev

# Use the dev CLI naturally
./initiat_dev auth login dev@example.com
./initiat_dev device register "Dev Machine"
./initiat_dev workspace list
```

### Testing Against Different Environments

```bash
# Dev binary defaults to localhost
./initiat_dev auth login dev@example.com

# Override to test staging
./initiat_dev --api-url https://staging.initiat.dev auth login dev@example.com

# Production binary defaults to production
./initiat auth login user@example.com
```

## How It Works

The dev build uses Go's `-ldflags` to inject a different default API URL at compile time:

```bash
go build \
  -ldflags "-X github.com/DylanBlakemore/initiat-cli/internal/config.defaultAPIBaseURL=http://localhost:4000" \
  -o initiat_dev .
```

The `defaultAPIBaseURL` variable in `internal/config/config.go` is set at build time, so the binary permanently defaults to localhost without any configuration changes needed.

## Configuration Priority

Even with the dev binary, the configuration priority remains:

1. **Command-line flag** `--api-url` (highest priority)
2. **Environment variable** `INITIAT_API_BASE_URL`
3. **Config file** `~/.initiat/config.yaml`
4. **Build-time default** (localhost for dev build, production for regular build)

## Comparison

| Feature | `initiat` (production) | `initiat_dev` |
|---------|----------------------|---------------|
| Default API URL | `https://www.initiat.dev` | `http://localhost:4000` |
| Binary Name | `initiat` | `initiat_dev` |
| Override with `--api-url` | ✅ Yes | ✅ Yes |
| Override with env var | ✅ Yes | ✅ Yes |
| Override with config | ✅ Yes | ✅ Yes |
| All commands | ✅ Same | ✅ Same |
| All flags | ✅ Same | ✅ Same |

## Building Both Versions

```bash
# Build production binary
make build

# Build development binary
make build-dev

# Now you have both:
./initiat version       # Production
./initiat_dev version   # Development

# Clean both
make clean
```

## Tips

1. **Keep both binaries**: Having both `initiat` and `initiat_dev` lets you easily switch between production and local environments

2. **Add to PATH**: You can add `initiat_dev` to your PATH if you primarily work with local development:
   ```bash
   sudo cp initiat_dev /usr/local/bin/
   ```

3. **Aliases**: Create shell aliases for convenience:
   ```bash
   alias initiat-dev='./initiat_dev'
   alias initiat-prod='./initiat'
   ```

4. **Git ignore**: The `.gitignore` is already configured to ignore both `initiat` and `initiat_dev` binaries

## Troubleshooting

### Connection Refused

If you get "connection refused" errors:
```bash
# Check if your backend is running
curl http://localhost:4000/health

# Start your backend
cd ../initiat-backend
make dev
```

### Wrong API URL

Verify which API URL is being used:
```bash
# The help shows the flag, but the actual default is set at build time
./initiat_dev --help

# To truly verify, check the config at runtime:
# Look at the actual requests being made or enable debug logging
```

### Still Using Production URL

If the dev binary seems to use production:
```bash
# Rebuild to ensure the ldflags are applied
make clean
make build-dev

# Verify it was built correctly
ls -la initiat_dev
```

