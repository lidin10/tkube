# tkube

[![Go Report Card](https://goreportcard.com/badge/github.com/lidin10/tkube)](https://goreportcard.com/report/github.com/lidin10/tkube)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Release](https://img.shields.io/github/release/lidin10/tkube.svg)](https://github.com/lidin10/tkube/releases)
[![Coverage](https://img.shields.io/badge/Coverage-61.4%25-orange.svg)](https://github.com/lidin10/tkube)

🚀 Enhanced Teleport kubectl wrapper with auto-authentication

A Go binary for quickly logging into Kubernetes clusters via Teleport with intelligent configuration and cross-shell compatibility.

## Features

- 🚀 Quick login with `tkube <env> <cluster>`
- 🔐 Automatic Teleport authentication with session expiration tracking
- ⚙️ Simple JSON configuration with environment-specific settings
- 🎯 Smart tab completion with helpful messages for missing dependencies
- 🖥️ Cross-shell compatibility (works with any shell)
- 📦 Single binary - no shell-specific dependencies
- 🔧 **Multi-version tsh support** - use different tsh versions for different environments
- 📥 **Automatic tsh installation** - downloads the official tsh binaries from the Teleport CDN based on the server version.
- ⏰ **Session status display** - color-coded session health with remaining time

## Installation

### Using Homebrew (Recommended)
```bash
# Add the tap and install
brew tap lidin10/tap
brew install tkube

# Verify installation
tkube version
```

### From GitHub Releases
```bash
# Download the latest release for your platform
# For Apple Silicon Macs
curl -L https://github.com/lidin10/tkube/releases/latest/download/tkube_darwin_arm64.tar.gz | tar xz
sudo mv tkube /usr/local/bin/

# For Intel Macs
curl -L https://github.com/lidin10/tkube/releases/latest/download/tkube_darwin_amd64.tar.gz | tar xz
sudo mv tkube /usr/local/bin/

# For Linux
curl -L https://github.com/lidin10/tkube/releases/latest/download/tkube_linux_amd64.tar.gz | tar xz
sudo mv tkube /usr/local/bin/
```

### From Source
```bash
git clone https://github.com/lidin10/tkube
cd tkube
make build
make install
```

## Usage

```bash
# Connect to a cluster
tkube prod my-cluster

# Show available environments and auth status with session times
tkube status
# ✅ prod → teleport.prod.env:443 (10h59m left)
# ⚠️  test → teleport.test.env:443 (1h30m left)

# Show help
tkube help

# Show version
tkube version

# Log out from Teleport servers
tkube logout                 # Log out from all environments
tkube logout prod            # Log out from specific environment

# Generate shell completion
tkube completion bash   # for bash
tkube completion zsh    # for zsh
tkube completion fish   # for fish

# Manage tsh versions (now at root level)
tkube tsh-versions           # List installed tsh versions
tkube install-tsh 17.7.1     # Install specific tsh version

# Configuration management
tkube config show            # Show current configuration
tkube config path            # Show configuration file path
```

## Configuration

Configuration is stored in `~/.tkube/config.json`:

### Basic Configuration
```json
{
  "environments": {
    "prod": {
      "proxy": "teleport.prod.env:443",
      "tsh_version": "16.4.0",
      "user": "my-prod-username"
    },
    "test": {
      "proxy": "teleport.test.env:443",
      "tsh_version": "17.7.1"
    }
  },
  "auto_login": true,
  "default_user": "my-teleport-username"
}
```

### User Configuration
tkube supports three levels of user configuration for Teleport authentication:

1. **Environment-specific user**: Set `"user": "username"` in environment config
2. **Default user**: Set `"default_user": "username"` at the top level (applies to all environments without specific user)  
3. **System user**: Automatically uses your system username as fallback

**Priority**: Environment-specific user > Default user > System user

This is useful when:
- Your Teleport username differs from your system username
- You use different usernames for different environments (e.g., different identity providers)
- You want to explicitly control which user is used for authentication

The configuration file is automatically created with example values on first run.

## tsh Version Management

tkube supports using different versions of `tsh` for different environments, which is useful when:
- Different Teleport servers require different tsh versions
- You need to test compatibility with different versions
- Your system tsh version is incompatible with some servers

### Installing tsh Versions
```bash
# Install a specific tsh version (downloads real binaries from Teleport CDN)
tkube install-tsh 17.7.1

# Auto-detect and configure versions for all environments
tkube auto-detect-versions

# List installed versions and their usage
tkube tsh-versions
```

**Note:** The `install-tsh` command now downloads real tsh binaries from the official Teleport CDN (`https://cdn.teleport.dev/`) and supports multiple package formats (.tar.gz, .pkg, .app bundles) with automatic platform detection.

### Directory Structure
```
~/.tkube/
├── config.json
├── sessions/           # Isolated session directories per environment
│   ├── prod/           # Prod environment sessions  
│   └── test/           # Test environment sessions
└── tsh/                # Downloaded tsh binaries
    ├── 16.4.0/
    │   └── tsh
    └── 17.7.1/
        └── tsh
```

### Session Isolation
tkube keeps Teleport sessions completely isolated between environments:
- Each environment has its own session directory under `~/.tkube/sessions/<env>/`
- Login to one environment doesn't affect authentication status of others
- `tkube logout <env>` only affects the specified environment
- `tkube status` shows real authentication state per environment


## Requirements

- Teleport CLI (`tsh`) - will be downloaded automatically after your first attempt to connect to a cluster.
- kubectl - for automatic context switching.

## Shell Completion

Enable tab completion for your shell:

### Bash
```bash
# Load completion for current session
source <(tkube completion bash)

# Install permanently (Linux)
tkube completion bash > /etc/bash_completion.d/tkube

# Install permanently (macOS with Homebrew)
tkube completion bash > /usr/local/etc/bash_completion.d/tkube
```

### Zsh
```bash
# Load completion for current session
source <(tkube completion zsh)

# Install permanently
tkube completion zsh > "${fpath[1]}/_tkube"
```

### Fish
```bash
# Load completion for current session
tkube completion fish | source

# Install permanently
tkube completion fish > ~/.config/fish/completions/tkube.fish
```

## Examples

```bash
# Connect to production cluster (with smart tab completion!)
tkube prod <TAB>  # Shows available clusters or helpful messages
# If tsh version missing: "📦 tsh v16.4.0 not installed - run: tkube install-tsh 16.4.0"

# Connect to test environment
tkube test development

# Check authentication status with session times
tkube status
# ✅ prod → teleport.prod.env:443 (10h59m left)
# ⚠️  test → teleport.test.env:443 (1h30m left)
# ❌ dev → teleport.dev.env:443 (expired)

# Auto-detect required tsh versions
tkube auto-detect-versions
# Detected: prod requires tsh v16.4.0
# Detected: test requires tsh v17.7.1

# Install detected versions
tkube install-tsh 16.4.0
tkube install-tsh 17.7.1

# Configure different users for environments
tkube config show           # Check current config
# Edit ~/.tkube/config.json to add user settings:
# {
#   "environments": {
#     "prod": {
#       "proxy": "teleport.prod.env:443",
#       "tsh_version": "16.4.0", 
#       "user": "prod-username"
#     }
#   },
#   "default_user": "my-teleport-username"
# }

# Get help
tkube help
```

## Development

### Building
```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Run tests
make test

# Clean build artifacts
make clean
```

### Releases

This project uses [semantic-release](https://semantic-release.gitbook.io/) for automated versioning and releases. Releases are triggered automatically when commits are pushed to the `main` branch using [conventional commit messages](https://www.conventionalcommits.org/):

```bash
# Feature (minor version bump)
git commit -m "feat: add new cluster validation"

# Bug fix (patch version bump)
git commit -m "fix: resolve connection timeout issue"

# Breaking change (major version bump)
git commit -m "feat!: change configuration format"
# or
git commit -m "feat: change config format

BREAKING CHANGE: configuration file format has changed"
```

The automated release process:
1. Analyzes commit messages since the last release
2. Determines the version bump (patch/minor/major)
3. Generates changelog and release notes
4. Creates GitHub release with binaries
5. Updates Homebrew tap automatically

### Contributing

Please use conventional commit messages for your contributions to ensure proper automated releases.