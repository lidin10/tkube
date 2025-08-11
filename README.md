# tkube

🚀 Enhanced Teleport kubectl wrapper with auto-authentication

A Go binary for quickly logging into Kubernetes clusters via Teleport with intelligent configuration and cross-shell compatibility.

## Features

- 🚀 Quick login with `tkube <env> <cluster>`
- 🔐 Automatic Teleport authentication
- ⚙️ Simple JSON configuration
- 🎯 Clear error messages and usage hints
- 🖥️ Cross-shell compatibility (works with any shell)
- 📦 Single binary - no shell-specific dependencies

## Installation

### Using Homebrew (Recommended)
```bash
# Add the tap
brew tap lidin10/tap

# Install tkube
brew install tkube

# Enable shell completion (automatically installed with Homebrew)
# Completions are automatically available after installation
```

### From GitHub Releases
```bash
# Download the latest release for your platform
curl -L https://github.com/lidin10/tkube/releases/latest/download/tkube_v1.0.0_darwin_amd64.tar.gz | tar xz
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

# Show available environments and auth status
tkube status

# Show help
tkube help

# Show version
tkube version

# Generate shell completion
tkube completion bash   # for bash
tkube completion zsh    # for zsh
tkube completion fish   # for fish
```

## Configuration

Configuration is stored in `~/.tkube/config.json`:

```json
{
  "environments": {
    "prod": "teleport.prod.company.com:443",
    "test": "teleport.test.company.com:443",
    "dev": "teleport.dev.company.com:443"
  },
  "auto_login": true
}
```

The configuration file is automatically created with example values on first run.

## Requirements

- Teleport CLI (`tsh`)
- kubectl (optional, for Kubernetes operations)

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
# Connect to production cluster (with tab completion!)
tkube prod <TAB>  # Shows available clusters

# Connect to test environment
tkube test development

# Check authentication status
tkube status

# Get help
tkube help
```

## Building

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
