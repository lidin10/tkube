package teleport

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"tkube/internal/config"
)

// Client handles Teleport operations
type Client struct {
	configManager *config.Manager
	installer     *TSHInstaller
}

// NewClient creates a new Teleport client
func NewClient(configManager *config.Manager) (*Client, error) {
	installer, err := NewTSHInstaller()
	if err != nil {
		return nil, fmt.Errorf("failed to create tsh installer: %w", err)
	}

	return &Client{
		configManager: configManager,
		installer:     installer,
	}, nil
}

// IsAuthenticated checks if the user is authenticated to a Teleport proxy
func (c *Client) IsAuthenticated(proxy string) bool {
	cmd := exec.Command("tsh", "status", "--proxy="+proxy)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}

	outputStr := string(output)
	return strings.Contains(outputStr, "logged in") || strings.Contains(outputStr, "Valid until")
}

// IsAuthenticatedWithEnv checks if the user is authenticated to a Teleport proxy using environment-specific tsh
func (c *Client) IsAuthenticatedWithEnv(env, proxy string) bool {
	// Ensure tsh version is installed
	if err := c.EnsureTSHVersion(env); err != nil {
		fmt.Printf("⚠️  Failed to ensure tsh version for environment %s: %v\n", env, err)
		return false
	}

	tshPath := c.getTSHPath(env)
	if tshPath == "" {
		fmt.Printf("⚠️  No tsh path available for environment %s\n", env)
		return false
	}

	cmd := exec.Command(tshPath, "status", "--proxy="+proxy)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}

	outputStr := string(output)
	return strings.Contains(outputStr, "logged in") || strings.Contains(outputStr, "Valid until")
}

// CheckAuthenticationStatus checks if the user is authenticated without auto-installing tsh
func (c *Client) CheckAuthenticationStatus(env, proxy string) bool {
	tshPath := c.getTSHPath(env)
	if tshPath == "" {
		return false
	}

	// Check if the tsh version is actually installed
	if !c.installer.IsVersionInstalled(c.getRequiredTSHVersion(env)) {
		return false
	}

	cmd := exec.Command(tshPath, "status", "--proxy="+proxy)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}

	outputStr := string(output)
	return strings.Contains(outputStr, "logged in") || strings.Contains(outputStr, "Valid until")
}

// SessionInfo represents session information for an environment
type SessionInfo struct {
	IsAuthenticated bool
	ValidUntil      string
	TimeRemaining   string
	IsExpired       bool
}

// GetSessionInfo returns detailed session information for an environment
func (c *Client) GetSessionInfo(env, proxy string) *SessionInfo {
	info := &SessionInfo{
		IsAuthenticated: false,
		ValidUntil:      "",
		TimeRemaining:   "",
		IsExpired:       false,
	}

	tshPath := c.getTSHPath(env)
	if tshPath == "" {
		return info
	}

	// Check if the tsh version is actually installed
	if !c.installer.IsVersionInstalled(c.getRequiredTSHVersion(env)) {
		return info
	}

	cmd := exec.Command(tshPath, "status", "--proxy="+proxy)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return info
	}

	outputStr := string(output)

	// Check if authenticated
	if !strings.Contains(outputStr, "logged in") && !strings.Contains(outputStr, "Valid until") {
		return info
	}

	info.IsAuthenticated = true

	// Parse "Valid until" line
	lines := strings.Split(outputStr, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "Valid until:") {
			// Extract the time part
			parts := strings.Split(line, "Valid until:")
			if len(parts) > 1 {
				timeInfo := strings.TrimSpace(parts[1])

				// Check if expired
				if strings.Contains(timeInfo, "EXPIRED") {
					info.IsExpired = true
					info.TimeRemaining = "EXPIRED"
					info.ValidUntil = timeInfo
				} else {
					// Extract time remaining from brackets like [valid for 11h29m0s]
					if strings.Contains(timeInfo, "[valid for ") && strings.Contains(timeInfo, "]") {
						start := strings.Index(timeInfo, "[valid for ") + len("[valid for ")
						end := strings.Index(timeInfo, "]")
						if start < end {
							info.TimeRemaining = timeInfo[start:end]
						}
					}

					// Extract the actual expiry time (before the bracket)
					if bracketIndex := strings.Index(timeInfo, " ["); bracketIndex > 0 {
						info.ValidUntil = strings.TrimSpace(timeInfo[:bracketIndex])
					} else {
						info.ValidUntil = timeInfo
					}
				}
			}
			break
		}
	}

	return info
}

// getRequiredTSHVersion returns the required tsh version for an environment
func (c *Client) getRequiredTSHVersion(env string) string {
	config, err := c.configManager.Load()
	if err != nil {
		return ""
	}

	envConfig, exists := config.Environments[env]
	if !exists {
		return ""
	}

	return envConfig.TSHVersion
}

// Login authenticates to a Teleport proxy
func (c *Client) Login(proxy string) error {
	cmd := exec.Command("tsh", "login", "--proxy="+proxy)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// LoginWithEnv authenticates to a Teleport proxy using environment-specific tsh
func (c *Client) LoginWithEnv(env, proxy string) error {
	// Ensure tsh version is installed
	if err := c.EnsureTSHVersion(env); err != nil {
		return fmt.Errorf("failed to ensure tsh version for environment %s: %w", env, err)
	}

	tshPath := c.getTSHPath(env)
	if tshPath == "" {
		return fmt.Errorf("no tsh path available for environment %s", env)
	}

	cmd := exec.Command(tshPath, "login", "--proxy="+proxy)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// KubeLogin authenticates to a Kubernetes cluster via Teleport
func (c *Client) KubeLogin(proxy, cluster string) error {
	cmd := exec.Command("tsh", "--proxy="+proxy, "kube", "login", cluster)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// KubeLoginWithEnv authenticates to a Kubernetes cluster via Teleport using environment-specific tsh
func (c *Client) KubeLoginWithEnv(env, proxy, cluster string) error {
	// Ensure tsh version is installed
	if err := c.EnsureTSHVersion(env); err != nil {
		return fmt.Errorf("failed to ensure tsh version for environment %s: %w", env, err)
	}

	tshPath := c.getTSHPath(env)
	if tshPath == "" {
		return fmt.Errorf("no tsh path available for environment %s", env)
	}

	cmd := exec.Command(tshPath, "--proxy="+proxy, "kube", "login", cluster)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// GetClusters returns a list of available Kubernetes clusters for an environment
func (c *Client) GetClusters(env string) ([]string, error) {
	envConfig, err := c.configManager.GetEnvironment(env)
	if err != nil {
		return nil, err
	}

	// Ensure tsh version is installed
	if err := c.EnsureTSHVersion(env); err != nil {
		return nil, fmt.Errorf("failed to ensure tsh version for environment %s: %w", env, err)
	}

	tshPath := c.getTSHPath(env)
	if tshPath == "" {
		return nil, fmt.Errorf("no tsh path available for environment %s", env)
	}

	cmd := exec.Command(tshPath, "--proxy="+envConfig.Proxy, "kube", "ls", "--format=json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get clusters: %w", err)
	}

	// Parse JSON output to extract cluster names
	var clusters []map[string]interface{}
	if err := json.Unmarshal(output, &clusters); err != nil {
		return nil, fmt.Errorf("failed to parse cluster output: %w", err)
	}

	var clusterNames []string
	for _, cluster := range clusters {
		if name, ok := cluster["kube_cluster_name"].(string); ok {
			clusterNames = append(clusterNames, name)
		}
	}

	return clusterNames, nil
}

// GetClustersForCompletion returns a list of clusters for shell completion
func (c *Client) GetClustersForCompletion(env string) ([]string, error) {
	envConfig, err := c.configManager.GetEnvironment(env)
	if err != nil {
		return []string{"❌ Environment '" + env + "' not found"}, nil
	}

	// Check if tsh version is configured, auto-detect if not
	requiredVersion := c.getRequiredTSHVersion(env)
	if requiredVersion == "" {
		// Try to auto-detect version
		detector := NewVersionDetector()
		version, err := detector.DetectTSHVersion(envConfig.Proxy)
		if err != nil {
			return []string{"⚠️  No tsh version configured and auto-detection failed"}, nil
		}

		// Update config with detected version
		if err := c.configManager.UpdateEnvironmentTSHVersion(env, version); err != nil {
			return []string{"⚠️  Auto-detected version but failed to save config"}, nil
		}

		requiredVersion = version
	}

	// Check if tsh is installed - don't auto-install for completion
	if !c.installer.IsVersionInstalled(requiredVersion) {
		return []string{"📦 tsh v" + requiredVersion + " not installed - run: tkube install-tsh " + requiredVersion}, nil
	}

	tshPath := c.getTSHPath(env)
	if tshPath == "" {
		return []string{"⚠️  No tsh path available for environment '" + env + "'"}, nil
	}

	// Check if user is authenticated
	if !c.CheckAuthenticationStatus(env, envConfig.Proxy) {
		return []string{"🔐 Not authenticated - run: tkube " + env + " <cluster> to authenticate"}, nil
	}

	// Get clusters using the specific tsh version for this environment
	cmd := exec.Command(tshPath, "--proxy="+envConfig.Proxy, "kube", "ls", "--format=json")
	output, err := cmd.Output()
	if err != nil {
		return []string{"⚠️  Failed to get clusters - check connection to " + envConfig.Proxy}, nil
	}

	// Parse JSON output to extract cluster names
	var clusters []map[string]interface{}
	if err := json.Unmarshal(output, &clusters); err != nil {
		return []string{"⚠️  Failed to parse cluster list for environment '" + env + "'"}, nil
	}

	var clusterNames []string
	for _, cluster := range clusters {
		if name, ok := cluster["kube_cluster_name"].(string); ok {
			clusterNames = append(clusterNames, name)
		}
	}

	// If no clusters found, return helpful message
	if len(clusterNames) == 0 {
		return []string{"ℹ️  No clusters available in environment '" + env + "'"}, nil
	}

	return clusterNames, nil
}

// getTSHPath returns the path to the appropriate tsh version for the given environment
func (c *Client) getTSHPath(env string) string {
	config, err := c.configManager.Load()
	if err != nil {
		return ""
	}

	envConfig, exists := config.Environments[env]
	if !exists || envConfig.TSHVersion == "" {
		return ""
	}

	// Return path to specific tsh version from our managed installations
	tshPath := c.installer.GetTSHPath(envConfig.TSHVersion)

	// Check if the specific version exists and is installed
	if c.installer.IsVersionInstalled(envConfig.TSHVersion) {
		return tshPath
	}

	// Return empty string if version is not installed - this will trigger installation
	return ""
}

// IsTSHVersionInstalled checks if a specific tsh version is installed
func (c *Client) IsTSHVersionInstalled(version string) bool {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	tshPath := filepath.Join(homeDir, ".tkube", "tsh", version, "tsh")
	if _, err := os.Stat(tshPath); err != nil {
		return false
	}

	// Check if it's executable and not just a placeholder
	cmd := exec.Command(tshPath, "version", "--client")
	return cmd.Run() == nil
}

// GetTSHVersionInfo returns version information for the given tsh path
func (c *Client) GetTSHVersionInfo(tshPath string) string {
	cmd := exec.Command(tshPath, "version", "--client")
	output, err := cmd.Output()
	if err != nil {
		return "unknown version"
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0])
	}
	return "unknown version"
}

// GetInstalledTSHVersions returns a list of installed tsh versions
func (c *Client) GetInstalledTSHVersions() ([]string, error) {
	return c.installer.GetInstalledVersions()
}

// InstallTSHVersion installs a specific tsh version
func (c *Client) InstallTSHVersion(version string) error {
	return c.installer.InstallTSH(version)
}

// UninstallTSHVersion removes a specific tsh version
func (c *Client) UninstallTSHVersion(version string) error {
	return c.installer.UninstallVersion(version)
}

// EnsureTSHVersion ensures that the required tsh version is installed for an environment
func (c *Client) EnsureTSHVersion(env string) error {
	config, err := c.configManager.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	envConfig, exists := config.Environments[env]
	if !exists {
		return fmt.Errorf("environment '%s' not found", env)
	}

	if envConfig.TSHVersion == "" {
		// Try to auto-detect version
		detector := NewVersionDetector()
		version, err := detector.DetectTSHVersion(envConfig.Proxy)
		if err != nil {
			return fmt.Errorf("no tsh version configured for environment '%s' and auto-detection failed: %w", env, err)
		}

		// Update config with detected version
		if err := c.configManager.UpdateEnvironmentTSHVersion(env, version); err != nil {
			return fmt.Errorf("failed to update config with detected version: %w", err)
		}

		envConfig.TSHVersion = version

	}

	// Check if version is installed
	if !c.installer.IsVersionInstalled(envConfig.TSHVersion) {
		if err := c.installer.InstallTSH(envConfig.TSHVersion); err != nil {
			return fmt.Errorf("failed to install tsh version %s: %w", envConfig.TSHVersion, err)
		}
	}

	return nil
}
