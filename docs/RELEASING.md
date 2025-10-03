# Release Process

This document outlines how to create and publish releases for the Initiat CLI.

## Prerequisites

1. **GitHub Repository**: The repo must be hosted on GitHub
2. **GitHub Token**: For private repos, users will need a Personal Access Token (PAT) with `repo` scope to download releases
3. **Permissions**: You need write access to the repository to create releases

## Creating a Release

Releases are created manually through GitHub Actions to give you full control over when versions are published.

### Step 1: Update Changelog

First, update the changelog for your new version:

```bash
# This automatically adds a new version section with today's date
make changelog VERSION=v1.0.0

# Review the changes in CHANGELOG.md
git diff CHANGELOG.md

# Commit the changelog update
git add CHANGELOG.md
git commit -m "chore: update changelog for v1.0.0"
```

### Step 2: Decide on Version Number

Follow semantic versioning (e.g., `v1.0.0`, `v1.2.3-beta.1`):
- **Major version** (v2.0.0): Breaking changes
- **Minor version** (v1.1.0): New features, backwards compatible
- **Patch version** (v1.0.1): Bug fixes, backwards compatible
- **Pre-release** (v1.0.0-alpha.1, v1.0.0-beta.1): Testing versions

### Step 3: Trigger the Release Workflow

1. Go to your repository on GitHub
2. Navigate to **Actions** → **Release** workflow
3. Click **Run workflow**
4. Fill in the parameters:
   - **version**: The version tag (e.g., `v1.0.0`)
   - **draft**: Check this to create a draft release (you can publish it later)
   - **prerelease**: Check this for pre-release versions (alpha, beta, rc)
5. Click **Run workflow**

### Step 4: Monitor the Build

The workflow will:
1. Create a GitHub release with the specified version tag
2. Build binaries for all supported platforms:
   - macOS (Intel and Apple Silicon)
   - Linux (AMD64 and ARM64)
   - Windows (AMD64)
3. Upload all binaries as release assets
4. Generate and upload SHA256 checksums

The entire process takes about 5-10 minutes.

### Step 5: Verify the Release

Once complete:
1. Go to **Releases** in your GitHub repository
2. Find your new release
3. Verify all assets are present:
   - `initiat-darwin-amd64.tar.gz`
   - `initiat-darwin-arm64.tar.gz`
   - `initiat-linux-amd64.tar.gz`
   - `initiat-linux-arm64.tar.gz`
   - `initiat-windows-amd64.zip`
   - `checksums.txt`

### Step 6: Update Release Notes (Optional)

Edit the release description to add:
- Notable changes and new features
- Bug fixes
- Breaking changes
- Migration guides
- Known issues

## Local Release Testing

Before creating an actual release, you can test the build process locally:

```bash
# Test with a fake version
make release VERSION=v1.0.0-test

# Check the binaries in the dist/ folder
ls -la dist/

# Test a specific binary
./dist/initiat-darwin-arm64 version
```

## Installation Instructions

### For Public Repositories

Users can install directly using curl:

```bash
# macOS (Apple Silicon)
curl -L https://github.com/InitiatDev/initiat-cli/releases/download/v1.0.0/initiat-darwin-arm64.tar.gz | tar xz
sudo mv initiat /usr/local/bin/

# Linux (AMD64)
curl -L https://github.com/InitiatDev/initiat-cli/releases/download/v1.0.0/initiat-linux-amd64.tar.gz | tar xz
sudo mv initiat /usr/local/bin/
```

### For Private Repositories

Users need to authenticate with a GitHub token:

```bash
# Set your GitHub token
export GITHUB_TOKEN=ghp_your_token_here

# macOS (Apple Silicon)
curl -L -H "Authorization: token $GITHUB_TOKEN" \
  https://github.com/InitiatDev/initiat-cli/releases/download/v1.0.0/initiat-darwin-arm64.tar.gz | tar xz
sudo mv initiat /usr/local/bin/

# Or download manually
gh release download v1.0.0 --repo InitiatDev/initiat-cli --pattern "initiat-darwin-arm64.tar.gz"
tar xzf initiat-darwin-arm64.tar.gz
sudo mv initiat /usr/local/bin/
```

**Creating a GitHub Token:**
1. Go to GitHub Settings → Developer settings → Personal access tokens → Tokens (classic)
2. Click "Generate new token (classic)"
3. Give it a descriptive name (e.g., "Initiat CLI Downloads")
4. Select the `repo` scope (for private repos)
5. Click "Generate token"
6. Copy and save the token securely

## Verifying Downloads

Users should verify the integrity of downloaded binaries:

```bash
# Download the checksums file
curl -L -H "Authorization: token $GITHUB_TOKEN" \
  https://github.com/InitiatDev/initiat-cli/releases/download/v1.0.0/checksums.txt > checksums.txt

# Verify the binary (macOS/Linux)
sha256sum -c checksums.txt --ignore-missing

# Or manually
sha256sum initiat-darwin-arm64.tar.gz
# Compare with checksums.txt
```

## Troubleshooting

### Build Fails

Check the Actions logs for specific errors:
- Go compilation errors
- Module dependency issues
- Platform-specific build problems

### Assets Not Uploaded

Ensure the workflow has the correct permissions:
- Repository Settings → Actions → General → Workflow permissions
- Set to "Read and write permissions"

### Private Repo Access Issues

For private repositories:
- Users need a valid PAT with `repo` scope
- The token must not be expired
- The user must have at least read access to the repository

## Best Practices

1. **Test locally first**: Always run `make release VERSION=v1.0.0-test` locally
2. **Use draft releases**: Create drafts first, verify, then publish
3. **Write clear release notes**: Help users understand what changed
4. **Follow semantic versioning**: Makes it clear when breaking changes occur
5. **Don't delete releases**: Even if there's a problem, create a new version instead
6. **Tag thoughtfully**: Tags are permanent in Git history

## Version Bump Checklist

Before releasing:

- [ ] All tests pass (`make test`)
- [ ] Code is linted (`make lint`)
- [ ] Security checks pass (`make security`)
- [ ] Version number decided
- [ ] **Changelog updated** (`make changelog VERSION=v1.0.0`)
- [ ] Release notes drafted
- [ ] Local release build tested
- [ ] GitHub Actions workflow triggered
- [ ] Build completed successfully
- [ ] Assets verified
- [ ] Release notes published
- [ ] Team notified

## Automation Ideas (Future)

Consider adding:
- Automatic changelog generation from commit messages
- Automatic homebrew formula updates
- Automatic version bumping in code
- Integration with package managers (apt, yum, chocolatey)
- Notification to Slack/Discord on successful release

