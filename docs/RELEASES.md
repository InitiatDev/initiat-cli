# Release Process

This document explains how to create new releases of the Initiat CLI. The process is automated via GitHub Actions - you just need to update the changelog and create a release tag.

## Table of Contents

- [Release Overview](#release-overview)
- [Versioning Strategy](#versioning-strategy)
- [Pre-Release Checklist](#pre-release-checklist)
- [Creating a Release](#creating-a-release)
- [GitHub Actions Automation](#github-actions-automation)
- [Post-Release Tasks](#post-release-tasks)
- [Rollback Procedures](#rollback-procedures)

## Release Overview

The Initiat CLI uses **automated releases via GitHub Actions**. When you create a release tag, GitHub Actions automatically:

- **Builds cross-platform binaries** for macOS, Linux, and Windows (AMD64 and ARM64)
- **Runs all tests** and security scans
- **Creates GitHub release** with downloadable archives
- **Generates release notes** from changelog

**Your job**: Update the changelog and create a release tag. GitHub Actions does the rest!

## Versioning Strategy

### Semantic Versioning

We follow [Semantic Versioning 2.0.0](https://semver.org/):

- **MAJOR** (X.0.0): Breaking changes, incompatible API changes
- **MINOR** (X.Y.0): New features, backward-compatible functionality
- **PATCH** (X.Y.Z): Bug fixes, backward-compatible

### Version Examples

```
v1.0.0    # Initial stable release
v1.1.0    # New features, backward compatible
v1.1.1    # Bug fixes only
v2.0.0    # Breaking changes
v1.2.0-beta.1  # Pre-release version
```

## Pre-Release Checklist

Before creating a release, ensure all items are completed:

### Code Quality
- [ ] All tests pass (`make test`)
- [ ] Code is formatted (`make format-check`)
- [ ] Linting passes (`make lint`)
- [ ] Security scan passes (`make security`)
- [ ] All CI checks pass on main branch

### Documentation
- [ ] README.md is updated
- [ ] CHANGELOG.md is updated with new version
- [ ] All new features are documented

### Release Preparation
- [ ] Version number determined
- [ ] Release notes prepared
- [ ] Breaking changes documented (if any)

## Creating a Release

### Step 1: Update Changelog

Update `CHANGELOG.md` with your new version:

```markdown
## [Unreleased]

## [v1.1.0] - 2024-01-15

### Added
- New feature 1
- New feature 2

### Changed
- Improved performance
- Updated dependencies

### Fixed
- Fixed bug 1
- Fixed bug 2
```

### Step 2: Create Release Tag

```bash
# Create and push the release tag
git tag -a v1.1.0 -m "Release v1.1.0"
git push origin v1.1.0
```

### Step 3: Monitor GitHub Actions

- Go to [Actions tab](https://github.com/InitiatDev/initiat-cli/actions)
- Watch the release workflow run
- Wait for all checks to pass

### Step 4: Verify Release

- Go to [Releases page](https://github.com/InitiatDev/initiat-cli/releases)
- Verify the release was created with all binaries
- Check that release notes were generated correctly

**That's it!** GitHub Actions handles all the building, testing, and distribution.

## GitHub Actions Automation

### What Happens Automatically

When you push a release tag, GitHub Actions automatically:

1. **Runs all tests** and security scans
2. **Builds binaries** for all platforms:
   - `initiat-darwin-amd64.tar.gz` (macOS Intel)
   - `initiat-darwin-arm64.tar.gz` (macOS Apple Silicon)
   - `initiat-linux-amd64.tar.gz` (Linux x64)
   - `initiat-linux-arm64.tar.gz` (Linux ARM64)
   - `initiat-windows-amd64.zip` (Windows x64)
3. **Creates GitHub release** with all binaries
4. **Generates release notes** from your changelog

### Monitoring the Process

- **Actions Tab**: [github.com/InitiatDev/initiat-cli/actions](https://github.com/InitiatDev/initiat-cli/actions)
- **Releases Page**: [github.com/InitiatDev/initiat-cli/releases](https://github.com/InitiatDev/initiat-cli/releases)

### If Something Goes Wrong

- Check the Actions tab for error details
- Fix any issues and create a new tag
- GitHub Actions will retry the release process

## Post-Release Tasks

### Immediate Tasks

- [ ] Verify release appears on GitHub
- [ ] Test download and installation
- [ ] Announce release to team

### Communication

- [ ] Send release announcement
- [ ] Update project website (if applicable)
- [ ] Notify stakeholders

## Rollback Procedures

### If Release Has Issues

```bash
# Delete the release tag
git tag -d v1.1.0
git push origin :refs/tags/v1.1.0

# Delete GitHub release
gh release delete v1.1.0

# Create hotfix and new release
git checkout -b hotfix/v1.1.1
# ... make fixes ...
git tag -a v1.1.1 -m "Release v1.1.1 (hotfix)"
git push origin v1.1.1
```

## Release Notes

### Format

Release notes should follow this format:

```markdown
# Release v1.1.0

## What's New
- New feature 1
- New feature 2
- Performance improvements

## Bug Fixes
- Fixed issue 1
- Fixed issue 2
- Security improvements

## Breaking Changes
- None (or list breaking changes)

## Installation
Download the latest release from [GitHub Releases](https://github.com/InitiatDev/initiat-cli/releases)
```

### Content Guidelines

- **Clear and Concise**: Use simple, clear language
- **User-Focused**: Focus on user benefits, not technical details
- **Categorized**: Group changes by type (features, fixes, etc.)
- **Actionable**: Include installation and upgrade instructions
- **Complete**: Include all significant changes

## Best Practices

### Release Frequency

- **Major Releases**: Every 6-12 months
- **Minor Releases**: Every 1-3 months
- **Patch Releases**: As needed for critical fixes
- **Pre-releases**: For testing new features

### Quality Gates

- **All Tests Pass**: No failing tests
- **Security Scan**: No high-severity vulnerabilities
- **Code Quality**: Passes all linting and formatting checks
- **Documentation**: All changes documented
- **Manual Testing**: Tested on target platforms

### Communication

- **Release Announcements**: Notify users of new releases
- **Breaking Changes**: Clearly communicate breaking changes
- **Migration Guides**: Provide upgrade instructions
- **Support**: Be available for user questions

## Conclusion

The Initiat CLI release process is designed to ensure:

- **Quality**: All releases are thoroughly tested
- **Security**: Security scans and vulnerability checks
- **Reliability**: Automated testing and validation
- **Usability**: Clear documentation and migration guides
- **Support**: Responsive issue resolution

By following these procedures, we ensure that every release maintains the high quality and security standards expected by our users.
