# Changelog

## [1.2.0] - 2025-08-15

### Added
- **Session expiration display**: `tkube status` now shows remaining session time for each environment
- **Color-coded session status**: Visual indicators for session health (green: >2h, yellow: <2h, expired)
- **Automatic tsh installation**: Real tsh binaries are now downloaded and installed from official Teleport CDN
- **Smart completion**: Tab completion shows helpful messages when tsh versions are missing or authentication is required
- **Version auto-detection**: Automatically detect required tsh versions from Teleport servers

### Changed
- **Command structure reorganization**: 
  - Moved `install-tsh`, `tsh-versions`, `auto-detect-versions` to root level
  - `config` now only handles configuration file operations (`show`, `path`)
- **Optimized user messages**: Reduced verbose output, cleaner and more actionable feedback
- **Enhanced tsh installation**: 
  - Downloads real tsh binaries from `https://cdn.teleport.dev/`
  - Supports both .tar.gz (Linux/macOS) and .pkg (macOS fallback) formats
  - Handles different package naming conventions across Teleport versions
  - Extracts .app bundles correctly on macOS (v17+)
- **Improved completion behavior**: No automatic installation during tab completion, shows informative messages instead

### Technical Improvements
- **TSHInstaller**: Complete rewrite with real package download and extraction
- **SessionInfo**: New struct for tracking session expiration and status
- **Platform detection**: Automatic architecture detection (arm64, x86_64)
- **Package format handling**: Support for different Teleport package naming conventions:
  - v18+: `teleport-18.1.3.pkg`
  - v17: `teleport-17.7.1.pkg` 
  - v16-: `tsh-16.5.13.pkg`
- **macOS .app bundle support**: Proper extraction and symlinking of tsh.app bundles

### Fixed
- **macOS compatibility**: Fixed tsh v17+ installation by properly handling .app bundle structure
- **Completion performance**: Removed blocking operations from tab completion
- **Message consistency**: Unified message format and reduced duplication

### Command Changes
```bash
# Old commands (no longer available)
tkube config install-tsh <version>
tkube config tsh-versions
tkube config auto-detect-versions

# New commands (root level)
tkube install-tsh <version>
tkube tsh-versions  
tkube auto-detect-versions

# Config now only handles configuration
tkube config show
tkube config path
```

### Migration from 1.1.0
- **Command updates**: Update any scripts using old `tkube config install-tsh` to `tkube install-tsh`
- **No configuration changes**: Existing config.json files work without modification
- **Automatic detection**: Run `tkube auto-detect-versions` to automatically configure tsh versions for all environments
- **Session display**: `tkube status` now shows session expiration times automatically

### New Workflow Examples
```bash
# Check status with session times
tkube status
# ✅ prod → teleport.prod.env:443 (10h59m left)
# ✅ test → teleport.test.env:443 (10h59m left)

# Install specific tsh version
tkube install-tsh 17.7.1

# List all installed versions
tkube tsh-versions

# Auto-detect versions for all environments
tkube auto-detect-versions

# Tab completion shows helpful messages
tkube prod <TAB>
# 📦 tsh v16.4.0 not installed - run: tkube install-tsh 16.4.0
```

## [1.1.0] - 2024-12-19

### Added
- **Multi-version tsh support**: Use different tsh versions for different environments
- New command `tkube config install-tsh <version>`: Creates installation scripts for specific tsh versions
- New command `tkube config auto-install-tsh <version>`: Auto-downloads and installs tsh versions
- New command `tkube config tsh-versions`: Lists installed tsh versions and environment usage
- Enhanced configuration format with support for environment-specific tsh versions
- Automatic fallback to system tsh when specific versions are not configured or installed

### Changed
- **Breaking Change**: Configuration format updated to support tsh versions
  - Old format: `"prod": "teleport.prod.company.com:443"`
  - New format: `"prod": {"proxy": "teleport.prod.company.com:443", "tsh_version": "14.0.0"}`
- Enhanced status display to show tsh version information for each environment
- Improved error messages with guidance on installing required tsh versions

### Technical Improvements
- New `Environment` struct with `Proxy` and `TSHVersion` fields
- Functions now use environment-specific tsh versions when available
- Automatic platform detection for tsh downloads (Linux/macOS, AMD64/ARM64)
- Installation scripts generated with proper error handling and cleanup

## [1.0.0] - Initial Release

### Features
- Basic Teleport kubectl wrapper functionality
- Environment-based configuration
- Auto-authentication support
- Cross-shell compatibility
- Tab completion support
