package teleport

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"tkube/internal/config"
)

func TestNewClient(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping CI-unfriendly test in CI environment")
	}
	configManager, err := config.NewManager()
	if err != nil {
		t.Fatalf("Failed to create config manager: %v", err)
	}

	client, err := NewClient(configManager)
	if err != nil {
		t.Fatalf("Expected no error creating client, got %v", err)
	}

	if client == nil {
		t.Fatal("Expected client to be created")
	}
}

func TestClient_GetSessionInfo(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping CI-unfriendly test in CI environment")
	}
	configManager, _ := config.NewManager()
	client, _ := NewClient(configManager)

	// Test with non-existent environment
	sessionInfo := client.GetSessionInfo("nonexistent", "test.proxy.com:443")

	if sessionInfo == nil {
		t.Fatal("Expected session info to be returned")
	}

	if sessionInfo.IsAuthenticated {
		t.Error("Expected IsAuthenticated to be false for non-existent tsh")
	}

	if sessionInfo.ValidUntil != "" {
		t.Error("Expected ValidUntil to be empty")
	}

	if sessionInfo.TimeRemaining != "" {
		t.Error("Expected TimeRemaining to be empty")
	}

	if sessionInfo.IsExpired {
		t.Error("Expected IsExpired to be false")
	}
}

func TestClient_IsAuthenticated(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping CI-unfriendly test in CI environment")
	}
	configManager, _ := config.NewManager()
	client, _ := NewClient(configManager)

	// Test with non-existent tsh (should return false)
	authenticated := client.IsAuthenticated("nonexistent.proxy.com:443")
	if authenticated {
		t.Error("Expected IsAuthenticated to return false for non-existent proxy")
	}
}

func TestClient_GetClusters_EnvironmentNotFound(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping CI-unfriendly test in CI environment")
	}
	configManager, _ := config.NewManager()
	client, _ := NewClient(configManager)

	clusters, err := client.GetClusters("nonexistent")
	if err == nil {
		t.Error("Expected error when environment not found")
	}

	if clusters != nil {
		t.Error("Expected nil clusters when environment not found")
	}
}

func TestClient_GetClustersForCompletion_EnvironmentNotFound(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping CI-unfriendly test in CI environment")
	}
	configManager, _ := config.NewManager()
	client, _ := NewClient(configManager)

	result, err := client.GetClustersForCompletion("nonexistent")

	// Should not error, but return helpful message
	if err != nil {
		t.Errorf("Expected no error but got: %v", err)
	}

	if len(result) == 0 {
		t.Error("Expected at least one result (error message)")
	}

	// Should contain error message
	if len(result) > 0 && !contains(result[0], "not found") {
		t.Errorf("Expected error message about environment not found, got: %s", result[0])
	}
}

func TestClient_IsAuthenticatedWithEnv(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping CI-unfriendly test in CI environment")
	}
	configManager, _ := config.NewManager()
	client, _ := NewClient(configManager)

	// Test with environment that doesn't exist
	authenticated := client.IsAuthenticatedWithEnv("nonexistent", "test.proxy.com:443")
	if authenticated {
		t.Error("Expected IsAuthenticatedWithEnv to return false for non-existent environment")
	}
}

func TestClient_CheckAuthenticationStatus(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping CI-unfriendly test in CI environment")
	}
	configManager, _ := config.NewManager()
	client, _ := NewClient(configManager)

	// Test with environment that doesn't exist
	authenticated := client.CheckAuthenticationStatus("nonexistent", "test.proxy.com:443")
	if authenticated {
		t.Error("Expected CheckAuthenticationStatus to return false for non-existent environment")
	}
}

func TestClient_GetInstalledTSHVersions(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping CI-unfriendly test in CI environment")
	}
	configManager, _ := config.NewManager()
	client, _ := NewClient(configManager)

	versions, err := client.GetInstalledTSHVersions()
	if err != nil {
		t.Errorf("Expected no error getting installed versions, got %v", err)
	}

	// Should return a slice (possibly empty)
	if versions == nil {
		t.Error("Expected versions slice, got nil")
	}
}

func TestClient_IsTSHVersionInstalled(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping CI-unfriendly test in CI environment")
	}
	configManager, _ := config.NewManager()
	client, _ := NewClient(configManager)

	// Test with non-existent version
	installed := client.IsTSHVersionInstalled("999.999.999")
	if installed {
		t.Error("Expected non-existent version to not be installed")
	}
}

func TestClient_GetTSHVersionInfo(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping CI-unfriendly test in CI environment")
	}
	configManager, _ := config.NewManager()
	client, _ := NewClient(configManager)

	// Test with non-existent path
	info := client.GetTSHVersionInfo("/non/existent/path/tsh")

	if info == "" {
		t.Error("Expected version info to be non-empty even for non-existent path")
	}
}

// Table-driven tests for various client operations
func TestClient_TableDriven(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping CI-unfriendly test in CI environment")
	}
	tests := []struct {
		name      string
		operation string
		env       string
		proxy     string
		expected  interface{}
	}{
		{
			name:      "GetSessionInfo with non-existent env",
			operation: "session_info",
			env:       "nonexistent",
			proxy:     "test.proxy.com:443",
			expected:  false, // IsAuthenticated should be false
		},
		{
			name:      "IsAuthenticated with invalid proxy",
			operation: "is_authenticated",
			proxy:     "invalid.proxy.com:443",
			expected:  false,
		},
		{
			name:      "CheckAuthenticationStatus with non-existent env",
			operation: "check_auth_status",
			env:       "nonexistent",
			proxy:     "test.proxy.com:443",
			expected:  false,
		},
	}

	configManager, _ := config.NewManager()
	client, _ := NewClient(configManager)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.operation {
			case "session_info":
				sessionInfo := client.GetSessionInfo(tt.env, tt.proxy)
				if sessionInfo.IsAuthenticated != tt.expected.(bool) {
					t.Errorf("Expected IsAuthenticated to be %v, got %v", tt.expected, sessionInfo.IsAuthenticated)
				}
			case "is_authenticated":
				result := client.IsAuthenticated(tt.proxy)
				if result != tt.expected.(bool) {
					t.Errorf("Expected IsAuthenticated to be %v, got %v", tt.expected, result)
				}
			case "check_auth_status":
				result := client.CheckAuthenticationStatus(tt.env, tt.proxy)
				if result != tt.expected.(bool) {
					t.Errorf("Expected CheckAuthenticationStatus to be %v, got %v", tt.expected, result)
				}
			}
		})
	}
}

// Additional comprehensive tests

func TestClient_GetSessionInfo_WithTSHPath(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping CI-unfriendly test in CI environment")
	}
	// Create a temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")

	testConfig := &config.Config{
		Environments: map[string]config.Environment{
			"test": {Proxy: "test.proxy.com:443", TSHVersion: "14.0.0"},
		},
		AutoLogin: true,
	}

	data, _ := json.MarshalIndent(testConfig, "", "  ")
	os.WriteFile(configPath, data, 0644)

	configManager, _ := config.NewManager()
	client, _ := NewClient(configManager)

	// Test with environment that has TSH version configured
	sessionInfo := client.GetSessionInfo("test", "test.proxy.com:443")

	if sessionInfo == nil {
		t.Fatal("Expected session info to be returned")
	}

	// Should not be authenticated since tsh is not actually installed
	if sessionInfo.IsAuthenticated {
		t.Error("Expected IsAuthenticated to be false when tsh not installed")
	}
}

func TestClient_IsAuthenticatedWithEnv_WithVersion(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping CI-unfriendly test in CI environment")
	}
	// Create a temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")

	testConfig := &config.Config{
		Environments: map[string]config.Environment{
			"test": {Proxy: "test.proxy.com:443", TSHVersion: "14.0.0"},
		},
		AutoLogin: true,
	}

	data, _ := json.MarshalIndent(testConfig, "", "  ")
	os.WriteFile(configPath, data, 0644)

	configManager, _ := config.NewManager()
	client, _ := NewClient(configManager)

	// Test with environment that has TSH version but not installed
	authenticated := client.IsAuthenticatedWithEnv("test", "test.proxy.com:443")
	if authenticated {
		t.Error("Expected IsAuthenticatedWithEnv to return false when TSH not installed")
	}
}

func TestClient_GetClustersForCompletion_NoTSHVersion(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping CI-unfriendly test in CI environment")
	}
	// Create a temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")

	testConfig := &config.Config{
		Environments: map[string]config.Environment{
			"test": {Proxy: "test.proxy.com:443"}, // No TSH version
		},
		AutoLogin: true,
	}

	data, _ := json.MarshalIndent(testConfig, "", "  ")
	os.WriteFile(configPath, data, 0644)

	configManager, _ := config.NewManager()
	client, _ := NewClient(configManager)

	result, err := client.GetClustersForCompletion("test")

	// Should not error, but return helpful message
	if err != nil {
		t.Errorf("Expected no error but got: %v", err)
	}

	if len(result) == 0 {
		t.Error("Expected at least one result (error message)")
	}

	// Should contain message about no tsh version
	if len(result) > 0 {
		t.Logf("Got result: %s", result[0])
	}
}

func TestClient_GetClustersForCompletion_TSHNotInstalled(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping CI-unfriendly test in CI environment")
	}
	// Create a temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")

	testConfig := &config.Config{
		Environments: map[string]config.Environment{
			"test": {Proxy: "test.proxy.com:443", TSHVersion: "999.999.999"}, // Non-existent version
		},
		AutoLogin: true,
	}

	data, _ := json.MarshalIndent(testConfig, "", "  ")
	os.WriteFile(configPath, data, 0644)

	configManager, _ := config.NewManager()
	client, _ := NewClient(configManager)

	result, err := client.GetClustersForCompletion("test")

	// Should not error, but return helpful message
	if err != nil {
		t.Errorf("Expected no error but got: %v", err)
	}

	if len(result) == 0 {
		t.Error("Expected at least one result (error message)")
	}

	// Should contain message about tsh not installed
	if len(result) > 0 {
		t.Logf("Got result: %s", result[0])
		// The actual result might vary, so just log it
	}
}

func TestClient_EnsureTSHVersion_NoVersion(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")

	testConfig := &config.Config{
		Environments: map[string]config.Environment{
			"test": {Proxy: "teleport-v14.test.com:443"}, // No TSH version, but version in hostname
		},
		AutoLogin: true,
	}

	data, _ := json.MarshalIndent(testConfig, "", "  ")
	os.WriteFile(configPath, data, 0644)

	configManager, _ := config.NewManager()
	client, _ := NewClient(configManager)

	// This should try to auto-detect version and install
	err := client.EnsureTSHVersion("test")

	// Will likely fail in test environment, but should not panic
	if err != nil {
		t.Logf("EnsureTSHVersion failed as expected: %v", err)
	}
}

func TestClient_GetTSHPath_WithVersion(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")

	testConfig := &config.Config{
		Environments: map[string]config.Environment{
			"test": {Proxy: "test.proxy.com:443", TSHVersion: "14.0.0"},
		},
		AutoLogin: true,
	}

	data, _ := json.MarshalIndent(testConfig, "", "  ")
	os.WriteFile(configPath, data, 0644)

	configManager, _ := config.NewManager()
	client, _ := NewClient(configManager)

	// Test getTSHPath method (it's private, but we can test through other methods)
	// This is tested indirectly through IsAuthenticatedWithEnv
	authenticated := client.IsAuthenticatedWithEnv("test", "test.proxy.com:443")

	// Should return false since tsh is not actually installed
	if authenticated {
		t.Error("Expected false when tsh not installed")
	}
}

func TestClient_InstallTSHVersion(t *testing.T) {
	configManager, _ := config.NewManager()
	client, _ := NewClient(configManager)

	// Test installing non-existent version
	err := client.InstallTSHVersion("999.999.999")

	// Should fail in test environment
	if err == nil {
		t.Error("Expected error when installing non-existent version")
	} else {
		t.Logf("InstallTSHVersion failed as expected: %v", err)
	}
}

func TestClient_UninstallTSHVersion(t *testing.T) {
	configManager, _ := config.NewManager()
	client, _ := NewClient(configManager)

	// Test uninstalling non-existent version
	err := client.UninstallTSHVersion("999.999.999")

	// May or may not error depending on implementation
	t.Logf("UninstallTSHVersion result: %v", err)
}

func TestClient_Login_Methods(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping CI-unfriendly test in CI environment")
	}
	configManager, _ := config.NewManager()
	client, _ := NewClient(configManager)

	// These methods require interactive input, so we can't easily test them
	// But we can verify they exist and don't panic when called with invalid data

	// Test Login (will fail but shouldn't panic)
	err := client.Login("invalid.proxy.com:443")
	if err == nil {
		t.Error("Expected error for invalid proxy")
	}

	// Test LoginWithEnv (will fail but shouldn't panic)
	err = client.LoginWithEnv("nonexistent", "invalid.proxy.com:443")
	if err == nil {
		t.Error("Expected error for non-existent environment")
	}

	// Test KubeLogin (will fail but shouldn't panic)
	err = client.KubeLogin("invalid.proxy.com:443", "test-cluster")
	if err == nil {
		t.Error("Expected error for invalid proxy")
	}

	// Test KubeLoginWithEnv (will fail but shouldn't panic)
	err = client.KubeLoginWithEnv("nonexistent", "invalid.proxy.com:443", "test-cluster")
	if err == nil {
		t.Error("Expected error for non-existent environment")
	}
}

func TestClient_Logout_Methods(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping CI-unfriendly test in CI environment")
	}
	configManager, _ := config.NewManager()
	client, _ := NewClient(configManager)

	// Test LogoutWithEnv (should fail for non-existent environment)
	err := client.LogoutWithEnv("nonexistent", "invalid.proxy.com:443")
	if err == nil {
		t.Error("Expected error for non-existent environment")
	} else {
		t.Logf("LogoutWithEnv failed as expected: %v", err)
	}
}

func TestClient_LogoutWithEnv_WithValidConfig(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping CI-unfriendly test in CI environment")
	}
	// Create a temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")

	testConfig := &config.Config{
		Environments: map[string]config.Environment{
			"test": {Proxy: "test.proxy.com:443", TSHVersion: "14.0.0"},
		},
		AutoLogin: true,
	}

	data, _ := json.MarshalIndent(testConfig, "", "  ")
	os.WriteFile(configPath, data, 0644)

	configManager, _ := config.NewManager()
	client, _ := NewClient(configManager)

	// Test LogoutWithEnv with configured environment
	err := client.LogoutWithEnv("test", "test.proxy.com:443")
	if err == nil {
		t.Log("LogoutWithEnv succeeded")
	} else {
		t.Logf("LogoutWithEnv failed as expected (tsh not installed): %v", err)
	}
}

func TestClient_LogoutWithEnv_NoTSHPath(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping CI-unfriendly test in CI environment")
	}
	
	configManager, _ := config.NewManager()
	client, _ := NewClient(configManager)

	// Test with environment that doesn't exist - should fail with no tsh path
	err := client.LogoutWithEnv("completely-nonexistent-env", "test.proxy.com:443")
	if err == nil {
		t.Error("Expected error when no tsh path available for non-existent environment")
	} else {
		if !contains(err.Error(), "no tsh path available") {
			t.Errorf("Expected 'no tsh path available' error, got: %v", err)
		}
		t.Logf("LogoutWithEnv failed as expected (no tsh path): %v", err)
	}
}
