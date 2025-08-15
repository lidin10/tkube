package teleport

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

// VersionDetector handles automatic tsh version detection
type VersionDetector struct{}

// NewVersionDetector creates a new version detector
func NewVersionDetector() *VersionDetector {
	return &VersionDetector{}
}

// ServerInfo represents Teleport server information
type ServerInfo struct {
	ServerVersion string `json:"server_version"`
	Version       string `json:"version"`
	Build         string `json:"build"`
}

// DetectTSHVersion detects the required tsh version for a Teleport proxy
func (vd *VersionDetector) DetectTSHVersion(proxy string) (string, error) {
	// Try to get version from server info endpoint
	version, err := vd.getVersionFromServer(proxy)
	if err != nil {
		// Fallback: try to extract version from proxy hostname or other methods
		version, err = vd.extractVersionFromProxy(proxy)
		if err != nil {
			return "", fmt.Errorf("failed to detect tsh version: %w", err)
		}
	}

	return version, nil
}

// getVersionFromServer attempts to get version from Teleport server
func (vd *VersionDetector) getVersionFromServer(proxy string) (string, error) {
	// Primary endpoint for Teleport server version
	endpoint := fmt.Sprintf("https://%s/webapi/ping", proxy)

	version, err := vd.queryEndpoint(endpoint)
	if err == nil && version != "" {
		return version, nil
	}

	return "", fmt.Errorf("could not determine version from server endpoint: %s", endpoint)
}

// queryEndpoint queries a specific endpoint for version information
func (vd *VersionDetector) queryEndpoint(endpoint string) (string, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return "", err
	}

	// Add common headers
	req.Header.Set("User-Agent", "tkube/1.1.0")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Check if response contains version information
	if resp.StatusCode == 200 {
		// Try to parse JSON response
		var serverInfo ServerInfo
		if err := json.NewDecoder(resp.Body).Decode(&serverInfo); err == nil {
			// Priority: server_version (from /webapi/ping) > version > build
			if serverInfo.ServerVersion != "" {
				return vd.normalizeVersion(serverInfo.ServerVersion), nil
			}
			if serverInfo.Version != "" {
				return vd.normalizeVersion(serverInfo.Version), nil
			}
			if serverInfo.Build != "" {
				return vd.normalizeVersion(serverInfo.Build), nil
			}
		}
	}

	return "", fmt.Errorf("no version information found")
}

// extractVersionFromProxy extracts version from proxy hostname or other sources
func (vd *VersionDetector) extractVersionFromProxy(proxy string) (string, error) {
	// Try to extract version from hostname patterns like:
	// teleport-v14.prod.company.com
	// teleport-14-0-0.prod.company.com
	// teleport14.prod.company.com

	patterns := []*regexp.Regexp{
		regexp.MustCompile(`teleport-v?(\d+)(?:\.(\d+))?(?:\.(\d+))?`),
		regexp.MustCompile(`teleport(\d+)(?:\.(\d+))?(?:\.(\d+))?`),
		regexp.MustCompile(`tsh-v?(\d+)(?:\.(\d+))?(?:\.(\d+))?`),
	}

	for _, pattern := range patterns {
		matches := pattern.FindStringSubmatch(proxy)
		if len(matches) >= 2 {
			// Construct version string
			version := matches[1]
			if len(matches) >= 3 && matches[2] != "" {
				version += "." + matches[2]
			}
			if len(matches) >= 4 && matches[3] != "" {
				version += "." + matches[3]
			}
			return version, nil
		}
	}

	// If no pattern matches, try to get from environment variables
	if version := vd.getVersionFromEnvironment(proxy); version != "" {
		return version, nil
	}

	return "", fmt.Errorf("could not extract version from proxy: %s", proxy)
}

// getVersionFromEnvironment tries to get version from environment variables
func (vd *VersionDetector) getVersionFromEnvironment(proxy string) string {
	// Check for common environment variable patterns
	envVars := []string{
		"TELEPORT_VERSION",
		"TSH_VERSION",
		"TELEPORT_TSH_VERSION",
	}

	for _, envVar := range envVars {
		if version := os.Getenv(envVar); version != "" {
			return vd.normalizeVersion(version)
		}
	}

	// Check for proxy-specific environment variables
	proxyKey := strings.ReplaceAll(proxy, ":", "_")
	proxyKey = strings.ReplaceAll(proxyKey, ".", "_")
	proxyKey = strings.ToUpper(proxyKey)

	envVars = []string{
		fmt.Sprintf("%s_TELEPORT_VERSION", proxyKey),
		fmt.Sprintf("%s_TSH_VERSION", proxyKey),
	}

	for _, envVar := range envVars {
		if version := os.Getenv(envVar); version != "" {
			return vd.normalizeVersion(version)
		}
	}

	return ""
}

// normalizeVersion normalizes version string but preserves full version
func (vd *VersionDetector) normalizeVersion(version string) string {
	// Remove common prefixes
	version = strings.TrimPrefix(version, "v")
	version = strings.TrimPrefix(version, "teleport-")
	version = strings.TrimPrefix(version, "tsh-")

	// Return the full normalized version (e.g., "17.7.1" instead of just "17.7")
	return version
}

// SuggestInstallation suggests installing the required tsh version
func (vd *VersionDetector) SuggestInstallation(requiredVersion string) string {
	return fmt.Sprintf(`💡 Required tsh version %s is not installed.

To install it, run:
  tkube config install-tsh %s

Or for automatic installation:
  tkube config auto-install-tsh %s

After installation, the version will be automatically configured for this environment.`,
		requiredVersion, requiredVersion, requiredVersion)
}
