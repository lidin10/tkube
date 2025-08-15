package teleport

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// TSHInstaller handles downloading and installing tsh clients
type TSHInstaller struct {
	baseDir string
}

// NewTSHInstaller creates a new tsh installer
func NewTSHInstaller() (*TSHInstaller, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	baseDir := filepath.Join(homeDir, ".tkube", "tsh")
	return &TSHInstaller{baseDir: baseDir}, nil
}

// PackageInfo represents information about a Teleport package
type PackageInfo struct {
	Version    string
	URL        string
	PackageExt string // "pkg" for macOS, "tar.gz" for Linux
}

// getPackageInfo determines the correct package URL and type for a given version
func (installer *TSHInstaller) getPackageInfo(version string) (*PackageInfo, error) {
	// Normalize version (remove 'v' prefix if present)
	version = strings.TrimPrefix(version, "v")

	var packageExt string
	var packageName string
	baseURL := "https://cdn.teleport.dev"

	// Parse version to determine major version for naming convention
	versionParts := strings.Split(version, ".")
	if len(versionParts) == 0 {
		return nil, fmt.Errorf("invalid version format: %s", version)
	}

	switch runtime.GOOS {
	case "darwin":
		// Try tar.gz first for macOS as it's easier to extract
		packageExt = "tar.gz"
		arch := runtime.GOARCH
		if arch == "amd64" {
			arch = "x86_64"
		}
		packageName = fmt.Sprintf("teleport-v%s-darwin-%s-bin.tar.gz", version, arch)

		// Fallback to .pkg if tar.gz doesn't exist
		// We'll implement this as a secondary attempt in InstallTSH
	case "linux":
		packageExt = "tar.gz"
		// For Linux tar.gz packages
		arch := runtime.GOARCH
		if arch == "amd64" {
			arch = "x86_64"
		}
		packageName = fmt.Sprintf("teleport-v%s-linux-%s-bin.tar.gz", version, arch)
	default:
		return nil, fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	url := fmt.Sprintf("%s/%s", baseURL, packageName)

	return &PackageInfo{
		Version:    version,
		URL:        url,
		PackageExt: packageExt,
	}, nil
}

// getPkgPackageInfo returns package info for macOS .pkg files as fallback
func (installer *TSHInstaller) getPkgPackageInfo(version string) (*PackageInfo, error) {
	version = strings.TrimPrefix(version, "v")
	versionParts := strings.Split(version, ".")
	if len(versionParts) == 0 {
		return nil, fmt.Errorf("invalid version format: %s", version)
	}

	majorVersionStr := versionParts[0]
	baseURL := "https://cdn.teleport.dev"

	var packageName string
	if majorVersionStr >= "17" {
		// v17+ uses teleport-X.Y.Z.pkg
		packageName = fmt.Sprintf("teleport-%s.pkg", version)
	} else {
		// v16 and below use tsh-X.Y.Z.pkg
		packageName = fmt.Sprintf("tsh-%s.pkg", version)
	}

	url := fmt.Sprintf("%s/%s", baseURL, packageName)

	return &PackageInfo{
		Version:    version,
		URL:        url,
		PackageExt: "pkg",
	}, nil
}

// InstallTSH downloads and installs a specific version of tsh
func (installer *TSHInstaller) InstallTSH(version string) error {

	// Create version directory
	versionDir := filepath.Join(installer.baseDir, version)
	if err := os.MkdirAll(versionDir, 0755); err != nil {
		return fmt.Errorf("failed to create version directory: %w", err)
	}

	// Try different package formats
	var lastErr error

	// First try: tar.gz (works for both Linux and macOS)
	packageInfo, err := installer.getPackageInfo(version)
	if err == nil {
		lastErr = installer.tryInstallPackage(packageInfo, versionDir)
		if lastErr == nil {
			return nil
		}
	}

	// Second try for macOS: .pkg files
	if runtime.GOOS == "darwin" {
		pkgInfo, err := installer.getPkgPackageInfo(version)
		if err == nil {
			lastErr = installer.tryInstallPackage(pkgInfo, versionDir)
			if lastErr == nil {
				return nil
			}
		}
	}

	return fmt.Errorf("failed to install tsh version %s: %w", version, lastErr)
}

// tryInstallPackage attempts to install from a specific package
func (installer *TSHInstaller) tryInstallPackage(packageInfo *PackageInfo, versionDir string) error {
	// Download package
	packagePath := filepath.Join(versionDir, fmt.Sprintf("teleport-%s.%s", packageInfo.Version, packageInfo.PackageExt))
	if err := installer.downloadPackage(packageInfo.URL, packagePath); err != nil {
		return fmt.Errorf("failed to download package: %w", err)
	}

	// Extract and install
	var err error
	switch packageInfo.PackageExt {
	case "pkg":
		err = installer.extractFromPkg(packagePath, versionDir)
	case "tar.gz":
		err = installer.extractFromTarGz(packagePath, versionDir)
	default:
		err = fmt.Errorf("unsupported package type: %s", packageInfo.PackageExt)
	}

	if err != nil {
		os.Remove(packagePath) // Clean up on failure
		return fmt.Errorf("failed to extract package: %w", err)
	}

	// Verify installation
	tshPath := filepath.Join(versionDir, "tsh")
	if _, err := os.Stat(tshPath); err != nil {
		os.Remove(packagePath) // Clean up on failure
		return fmt.Errorf("tsh binary not found after installation: %w", err)
	}

	// Make executable
	if err := os.Chmod(tshPath, 0755); err != nil {
		return fmt.Errorf("failed to make tsh executable: %w", err)
	}

	// Clean up package file
	os.Remove(packagePath)

	return nil
}

// downloadPackage downloads a package from the given URL
func (installer *TSHInstaller) downloadPackage(url, destPath string) error {

	client := &http.Client{
		Timeout: 10 * time.Minute, // Large timeout for big packages
	}

	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	file, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	return nil
}

// extractFromPkg extracts tsh from a macOS .pkg file
func (installer *TSHInstaller) extractFromPkg(pkgPath, destDir string) error {
	// Create a temporary directory for extraction
	tempDir := filepath.Join(destDir, "temp_extract")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Use pkgutil to expand the .pkg file
	fmt.Println("🔧 Extracting .pkg file...")
	if err := installer.runShellCommand("pkgutil", "--expand-full", pkgPath, filepath.Join(tempDir, "extracted")); err != nil {
		return fmt.Errorf("failed to extract pkg: %w", err)
	}

	// Find tsh binary in extracted contents
	tshPath, err := installer.findTSHInDirectory(filepath.Join(tempDir, "extracted"))
	if err != nil {
		return fmt.Errorf("tsh not found in package: %w", err)
	}

	// Copy tsh to destination (handle both .app bundles and direct binaries)
	if err := installer.copyTSHToDestination(tshPath, destDir); err != nil {
		return fmt.Errorf("failed to copy tsh: %w", err)
	}

	return nil
}

// extractVersionFromPath extracts version from package path
func (installer *TSHInstaller) extractVersionFromPath(pkgPath string) string {
	// Extract version from filename like "teleport-17.7.1.pkg" or "tsh-16.5.12.pkg"
	filename := filepath.Base(pkgPath)
	filename = strings.TrimSuffix(filename, ".pkg")
	filename = strings.TrimSuffix(filename, ".tar.gz")

	// Remove prefixes
	filename = strings.TrimPrefix(filename, "teleport-")
	filename = strings.TrimPrefix(filename, "tsh-")

	return filename
}

// extractFromTarGz extracts tsh from a .tar.gz file
func (installer *TSHInstaller) extractFromTarGz(tarPath, destDir string) error {
	// Create a temporary directory for extraction
	tempDir := filepath.Join(destDir, "temp_extract")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Extract the entire archive first
	if err := installer.extractTarGz(tarPath, tempDir); err != nil {
		return fmt.Errorf("failed to extract tar.gz: %w", err)
	}

	// Find tsh binary in extracted contents
	tshPath, err := installer.findTSHInDirectory(tempDir)
	if err != nil {
		return fmt.Errorf("tsh not found in package: %w", err)
	}

	// Copy tsh to destination (handle both .app bundles and direct binaries)
	return installer.copyTSHToDestination(tshPath, destDir)
}

// extractTarGz extracts a tar.gz archive to a directory
func (installer *TSHInstaller) extractTarGz(tarPath, destDir string) error {
	file, err := os.Open(tarPath)
	if err != nil {
		return fmt.Errorf("failed to open tar file: %w", err)
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar: %w", err)
		}

		// Create the full path
		path := filepath.Join(destDir, header.Name)

		// Ensure the directory exists
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			// Create directory
			if err := os.MkdirAll(path, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
		case tar.TypeReg:
			// Create file
			file, err := os.Create(path)
			if err != nil {
				return fmt.Errorf("failed to create file: %w", err)
			}

			_, err = io.Copy(file, tr)
			file.Close()

			if err != nil {
				return fmt.Errorf("failed to extract file: %w", err)
			}

			// Set file permissions
			if err := os.Chmod(path, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("failed to set file permissions: %w", err)
			}
		}
	}

	return nil
}

// findTSHInDirectory recursively searches for tsh binary in a directory
func (installer *TSHInstaller) findTSHInDirectory(dir string) (string, error) {
	var tshPaths []string
	var appBundlePaths []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			// Look for .app bundles
			if strings.HasSuffix(info.Name(), "tsh.app") {
				appBundlePaths = append(appBundlePaths, path)
			}
		} else {
			filename := info.Name()
			// Look for tsh binary (exact match)
			if filename == "tsh" {
				// Check if it's likely an executable
				if info.Mode()&0111 != 0 || runtime.GOOS == "darwin" {
					tshPaths = append(tshPaths, path)
				}
			}
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	// Priority order for macOS:
	// 1. Return .app bundle path (v17+) - we'll copy the whole bundle
	// 2. Return direct tsh binary path (v16 and below)
	if runtime.GOOS == "darwin" {
		// First, look for .app bundle (v17+)
		for _, appPath := range appBundlePaths {
			// Verify the executable exists inside
			execPath := filepath.Join(appPath, "Contents", "MacOS", "tsh")
			if _, err := os.Stat(execPath); err == nil {
				return appPath, nil // Return the .app bundle path
			}
		}

		// Then, look for direct tsh in teleport directory (v16-)
		for _, path := range tshPaths {
			if strings.Contains(path, "teleport/tsh") && !strings.Contains(path, ".app") {
				return path, nil
			}
		}
	}

	if len(tshPaths) == 0 && len(appBundlePaths) == 0 {
		return "", fmt.Errorf("tsh binary not found in directory %s", dir)
	}

	// Fallback: return the first found binary
	if len(tshPaths) > 0 {
		return tshPaths[0], nil
	}

	return "", fmt.Errorf("no suitable tsh found")
}

// copyFile copies a file from src to dst
func (installer *TSHInstaller) copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// copyTSHToDestination copies tsh (file or .app bundle) to destination
func (installer *TSHInstaller) copyTSHToDestination(srcPath, destDir string) error {
	srcInfo, err := os.Stat(srcPath)
	if err != nil {
		return fmt.Errorf("failed to stat source: %w", err)
	}

	if srcInfo.IsDir() {
		// This is a .app bundle, copy the entire directory
		appName := filepath.Base(srcPath)
		destAppPath := filepath.Join(destDir, appName)

		if err := installer.copyDirectory(srcPath, destAppPath); err != nil {
			return fmt.Errorf("failed to copy app bundle: %w", err)
		}

		// Create a symlink or wrapper script for easy access
		tshExecutable := filepath.Join(destAppPath, "Contents", "MacOS", "tsh")
		symlinkPath := filepath.Join(destDir, "tsh")

		// Remove existing symlink if it exists
		os.Remove(symlinkPath)

		// Create symlink to the actual executable
		if err := os.Symlink(tshExecutable, symlinkPath); err != nil {
			return fmt.Errorf("failed to create symlink: %w", err)
		}

		return nil
	} else {
		// This is a direct binary file
		destTSHPath := filepath.Join(destDir, "tsh")
		if err := installer.copyFile(srcPath, destTSHPath); err != nil {
			return fmt.Errorf("failed to copy tsh binary: %w", err)
		}

		// Make sure it's executable
		if err := os.Chmod(destTSHPath, 0755); err != nil {
			return fmt.Errorf("failed to make tsh executable: %w", err)
		}

		return nil
	}
}

// copyDirectory recursively copies a directory
func (installer *TSHInstaller) copyDirectory(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate the destination path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		destPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			// Create directory
			return os.MkdirAll(destPath, info.Mode())
		} else {
			// Copy file
			if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
				return err
			}

			if err := installer.copyFile(path, destPath); err != nil {
				return err
			}

			// Set file permissions
			return os.Chmod(destPath, info.Mode())
		}
	})
}

// runShellCommand executes a shell command with arguments
func (installer *TSHInstaller) runShellCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command '%s %v' failed: %w\nOutput: %s", name, args, err, string(output))
	}
	return nil
}

// IsVersionInstalled checks if a specific tsh version is already installed
func (installer *TSHInstaller) IsVersionInstalled(version string) bool {
	versionDir := filepath.Join(installer.baseDir, version)
	tshPath := filepath.Join(versionDir, "tsh")

	if _, err := os.Stat(tshPath); err != nil {
		return false
	}

	// Verify it's executable and working
	return installer.verifyTSHBinary(tshPath)
}

// verifyTSHBinary verifies that the tsh binary is working
func (installer *TSHInstaller) verifyTSHBinary(tshPath string) bool {
	// Check if file exists
	if _, err := os.Stat(tshPath); err != nil {
		return false
	}

	// Check if file is executable
	info, err := os.Stat(tshPath)
	if err != nil {
		return false
	}

	mode := info.Mode()
	if mode&0111 == 0 {
		return false // Not executable
	}

	// Try to run tsh version to verify it actually works
	cmd := exec.Command(tshPath, "version", "--client")
	err = cmd.Run()
	return err == nil
}

// GetInstalledVersions returns a list of installed tsh versions
func (installer *TSHInstaller) GetInstalledVersions() ([]string, error) {
	if _, err := os.Stat(installer.baseDir); os.IsNotExist(err) {
		return []string{}, nil
	}

	entries, err := os.ReadDir(installer.baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read tsh directory: %w", err)
	}

	var versions []string
	for _, entry := range entries {
		if entry.IsDir() {
			version := entry.Name()
			if installer.IsVersionInstalled(version) {
				versions = append(versions, version)
			}
		}
	}

	return versions, nil
}

// UninstallVersion removes a specific tsh version
func (installer *TSHInstaller) UninstallVersion(version string) error {
	versionDir := filepath.Join(installer.baseDir, version)

	if _, err := os.Stat(versionDir); os.IsNotExist(err) {
		return fmt.Errorf("version %s is not installed", version)
	}

	if err := os.RemoveAll(versionDir); err != nil {
		return fmt.Errorf("failed to remove version directory: %w", err)
	}

	return nil
}

// GetTSHPath returns the path to a specific tsh version
func (installer *TSHInstaller) GetTSHPath(version string) string {
	return filepath.Join(installer.baseDir, version, "tsh")
}

// AutoInstallForEnvironment automatically installs tsh for an environment if needed
func (installer *TSHInstaller) AutoInstallForEnvironment(envName, requiredVersion string) error {
	if installer.IsVersionInstalled(requiredVersion) {
		return nil
	}

	return installer.InstallTSH(requiredVersion)
}

// GetTSHVersionInfo returns version information for the given tsh path
func (installer *TSHInstaller) GetTSHVersionInfo(tshPath string) string {
	// Try to get actual version from the binary
	cmd := exec.Command(tshPath, "version", "--client")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Sprintf("installed at %s (version check failed)", tshPath)
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) > 0 && strings.TrimSpace(lines[0]) != "" {
		return strings.TrimSpace(lines[0])
	}

	return fmt.Sprintf("installed at %s", tshPath)
}
