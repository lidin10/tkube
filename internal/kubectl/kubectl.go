package kubectl

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Client handles kubectl operations
type Client struct{}

// NewClient creates a new kubectl client
func NewClient() *Client {
	return &Client{}
}

// CheckVersion checks if kubectl is available and returns its version
func (c *Client) CheckVersion() (string, error) {
	cmd := exec.Command("kubectl", "version", "--client", "--short")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("kubectl not found: %w", err)
	}

	versionStr := strings.TrimSpace(string(output))
	if strings.Contains(versionStr, "Client Version:") {
		parts := strings.Fields(versionStr)
		if len(parts) >= 3 {
			return parts[2], nil
		}
	}

	return "installed", nil
}

// IsAvailable checks if kubectl is available
func (c *Client) IsAvailable() bool {
	cmd := exec.Command("kubectl", "version", "--client", "--short")
	return cmd.Run() == nil
}

// GetContext returns the current kubectl context
func (c *Client) GetContext() (string, error) {
	cmd := exec.Command("kubectl", "config", "current-context")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current context: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// GetContexts returns a list of available kubectl contexts
func (c *Client) GetContexts() ([]string, error) {
	cmd := exec.Command("kubectl", "config", "get-contexts", "-o", "name")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get contexts: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var contexts []string
	for _, line := range lines {
		if line = strings.TrimSpace(line); line != "" {
			contexts = append(contexts, line)
		}
	}

	return contexts, nil
}

// SetContext sets the current kubectl context
func (c *Client) SetContext(context string) error {
	cmd := exec.Command("kubectl", "config", "use-context", context)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// TestConnection tests the connection to the current cluster
func (c *Client) TestConnection() error {
	cmd := exec.Command("kubectl", "cluster-info")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// GetClusterInfo returns information about the current cluster
func (c *Client) GetClusterInfo() (string, error) {
	cmd := exec.Command("kubectl", "cluster-info")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get cluster info: %w", err)
	}

	return string(output), nil
}
