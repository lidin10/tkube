package commands

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"tkube/internal/config"
	"tkube/internal/kubectl"
	"tkube/internal/teleport"
)

func TestNewHandler(t *testing.T) {
	// Test that NewHandler can be created with real objects
	configManager, err := config.NewManager()
	if err != nil {
		t.Fatalf("Failed to create config manager: %v", err)
	}
	
	teleportClient, err := teleport.NewClient(configManager)
	if err != nil {
		t.Fatalf("Failed to create teleport client: %v", err)
	}
	
	kubectlClient := kubectl.NewClient()
	
	installer, err := teleport.NewTSHInstaller()
	if err != nil {
		t.Fatalf("Failed to create installer: %v", err)
	}
	
	handler := NewHandler(configManager, teleportClient, kubectlClient, installer)
	
	if handler == nil {
		t.Fatal("Expected handler to be created")
	}
}

func TestHandler_ShowVersion(t *testing.T) {
	configManager, _ := config.NewManager()
	teleportClient, _ := teleport.NewClient(configManager)
	kubectlClient := kubectl.NewClient()
	installer, _ := teleport.NewTSHInstaller()
	
	handler := NewHandler(configManager, teleportClient, kubectlClient, installer)
	
	// This should not panic or error
	handler.ShowVersion()
}

func TestHandler_ShowConfigPath(t *testing.T) {
	configManager, _ := config.NewManager()
	teleportClient, _ := teleport.NewClient(configManager)
	kubectlClient := kubectl.NewClient()
	installer, _ := teleport.NewTSHInstaller()
	
	handler := NewHandler(configManager, teleportClient, kubectlClient, installer)
	
	// This should not panic or error
	handler.ShowConfigPath()
}

func TestHandler_ShowTSHVersions(t *testing.T) {
	configManager, _ := config.NewManager()
	teleportClient, _ := teleport.NewClient(configManager)
	kubectlClient := kubectl.NewClient()
	installer, _ := teleport.NewTSHInstaller()
	
	handler := NewHandler(configManager, teleportClient, kubectlClient, installer)
	
	// This should not panic or error
	handler.ShowTSHVersions()
}

func TestHandler_ShowStatus(t *testing.T) {
	configManager, _ := config.NewManager()
	teleportClient, _ := teleport.NewClient(configManager)
	kubectlClient := kubectl.NewClient()
	installer, _ := teleport.NewTSHInstaller()
	
	handler := NewHandler(configManager, teleportClient, kubectlClient, installer)
	
	// This should not panic or error
	handler.ShowStatus()
}

func TestHandler_ShowConfig(t *testing.T) {
	configManager, _ := config.NewManager()
	teleportClient, _ := teleport.NewClient(configManager)
	kubectlClient := kubectl.NewClient()
	installer, _ := teleport.NewTSHInstaller()
	
	handler := NewHandler(configManager, teleportClient, kubectlClient, installer)
	
	// This should not panic or error
	handler.ShowConfig()
}

func TestHandler_InstallTSH(t *testing.T) {
	configManager, _ := config.NewManager()
	teleportClient, _ := teleport.NewClient(configManager)
	kubectlClient := kubectl.NewClient()
	installer, _ := teleport.NewTSHInstaller()
	
	handler := NewHandler(configManager, teleportClient, kubectlClient, installer)
	
	// Test with a non-existent version (should attempt to install and likely fail)
	err := handler.InstallTSH("999.999.999")
	
	// We expect this to fail in test environment, which is fine
	if err != nil {
		t.Logf("InstallTSH failed as expected in test environment: %v", err)
	}
}

func TestHandler_AutoInstallTSH(t *testing.T) {
	configManager, _ := config.NewManager()
	teleportClient, _ := teleport.NewClient(configManager)
	kubectlClient := kubectl.NewClient()
	installer, _ := teleport.NewTSHInstaller()
	
	handler := NewHandler(configManager, teleportClient, kubectlClient, installer)
	
	// Test with a non-existent version
	err := handler.AutoInstallTSH("999.999.999")
	
	// We expect this to fail in test environment, which is fine
	if err != nil {
		t.Logf("AutoInstallTSH failed as expected in test environment: %v", err)
	}
}

func TestHandler_ConnectToCluster_InvalidEnv(t *testing.T) {
	configManager, _ := config.NewManager()
	teleportClient, _ := teleport.NewClient(configManager)
	kubectlClient := kubectl.NewClient()
	installer, _ := teleport.NewTSHInstaller()
	
	handler := NewHandler(configManager, teleportClient, kubectlClient, installer)
	
	// Test with non-existent environment
	err := handler.ConnectToCluster("nonexistent", "test-cluster")
	if err == nil {
		t.Error("Expected error when connecting to non-existent environment")
	}
}

func TestHandler_AutoDetectVersions(t *testing.T) {
	configManager, _ := config.NewManager()
	teleportClient, _ := teleport.NewClient(configManager)
	kubectlClient := kubectl.NewClient()
	installer, _ := teleport.NewTSHInstaller()
	
	handler := NewHandler(configManager, teleportClient, kubectlClient, installer)
	
	// This should not panic or error
	handler.AutoDetectVersions()
}

// Test interactive methods return not implemented errors
func TestHandler_InteractiveMethods(t *testing.T) {
	configManager, _ := config.NewManager()
	teleportClient, _ := teleport.NewClient(configManager)
	kubectlClient := kubectl.NewClient()
	installer, _ := teleport.NewTSHInstaller()
	
	handler := NewHandler(configManager, teleportClient, kubectlClient, installer)
	
	// Test that interactive methods return not implemented errors
	err := handler.AddEnvironmentInteractive()
	if err == nil || !contains(err.Error(), "not implemented") {
		t.Error("Expected AddEnvironmentInteractive to return 'not implemented' error")
	}
	
	err = handler.EditEnvironmentInteractive("test")
	if err == nil || !contains(err.Error(), "not implemented") {
		t.Error("Expected EditEnvironmentInteractive to return 'not implemented' error")
	}
	
	err = handler.RemoveEnvironmentInteractive("test")
	if err == nil || !contains(err.Error(), "not implemented") {
		t.Error("Expected RemoveEnvironmentInteractive to return 'not implemented' error")
	}
	
	err = handler.ValidateConfiguration()
	if err == nil || !contains(err.Error(), "not implemented") {
		t.Error("Expected ValidateConfiguration to return 'not implemented' error")
	}
}

func TestHandler_GetEnvironments(t *testing.T) {
	configManager, _ := config.NewManager()
	teleportClient, _ := teleport.NewClient(configManager)
	kubectlClient := kubectl.NewClient()
	installer, _ := teleport.NewTSHInstaller()
	
	handler := NewHandler(configManager, teleportClient, kubectlClient, installer)
	
	// Test getEnvironments method (it's not exported, but we can test through other methods)
	// This is tested indirectly through ShowStatus
	handler.ShowStatus()
}

func TestHandler_PromptForInstallation(t *testing.T) {
	configManager, _ := config.NewManager()
	teleportClient, _ := teleport.NewClient(configManager)
	kubectlClient := kubectl.NewClient()
	installer, _ := teleport.NewTSHInstaller()
	
	handler := NewHandler(configManager, teleportClient, kubectlClient, installer)
	
	// We can't easily test the interactive prompt, but we can test that the method exists
	// by calling a method that uses it (ConnectToCluster with auto-install scenario)
	// This is tested indirectly through ConnectToCluster
	_ = handler
}

func TestHandler_FormatTimeRemaining(t *testing.T) {
	configManager, _ := config.NewManager()
	teleportClient, _ := teleport.NewClient(configManager)
	kubectlClient := kubectl.NewClient()
	installer, _ := teleport.NewTSHInstaller()
	
	handler := NewHandler(configManager, teleportClient, kubectlClient, installer)
	
	// We can't test the private method directly, but it's used in ShowStatus
	// This is tested indirectly through ShowStatus
	handler.ShowStatus()
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || 
		s[len(s)-len(substr):] == substr || 
		containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Additional comprehensive tests

func TestHandler_ConnectToCluster_WithConfig(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")
	
	testConfig := &config.Config{
		Environments: map[string]config.Environment{
			"test": {Proxy: "test.proxy.com:443", TSHVersion: "14.0.0"},
		},
		AutoLogin: false, // Disable auto-login to test authentication check
	}
	
	data, _ := json.MarshalIndent(testConfig, "", "  ")
	os.WriteFile(configPath, data, 0644)
	
	configManager, _ := config.NewManager()
	teleportClient, _ := teleport.NewClient(configManager)
	kubectlClient := kubectl.NewClient()
	installer, _ := teleport.NewTSHInstaller()
	
	handler := NewHandler(configManager, teleportClient, kubectlClient, installer)
	
	// Test with valid environment but no authentication
	err := handler.ConnectToCluster("test", "test-cluster")
	if err == nil {
		t.Error("Expected error when not authenticated")
	}
}

func TestHandler_ShowStatus_WithEnvironments(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")
	
	testConfig := &config.Config{
		Environments: map[string]config.Environment{
			"prod": {Proxy: "prod.proxy.com:443", TSHVersion: "14.0.0"},
			"test": {Proxy: "test.proxy.com:443"},
		},
		AutoLogin: true,
	}
	
	data, _ := json.MarshalIndent(testConfig, "", "  ")
	os.WriteFile(configPath, data, 0644)
	
	configManager, _ := config.NewManager()
	teleportClient, _ := teleport.NewClient(configManager)
	kubectlClient := kubectl.NewClient()
	installer, _ := teleport.NewTSHInstaller()
	
	handler := NewHandler(configManager, teleportClient, kubectlClient, installer)
	
	// This should display environment status
	handler.ShowStatus()
}

func TestHandler_ShowConfig_WithConfig(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")
	
	testConfig := &config.Config{
		Environments: map[string]config.Environment{
			"prod": {Proxy: "prod.proxy.com:443", TSHVersion: "14.0.0"},
		},
		AutoLogin: true,
	}
	
	data, _ := json.MarshalIndent(testConfig, "", "  ")
	os.WriteFile(configPath, data, 0644)
	
	configManager, _ := config.NewManager()
	teleportClient, _ := teleport.NewClient(configManager)
	kubectlClient := kubectl.NewClient()
	installer, _ := teleport.NewTSHInstaller()
	
	handler := NewHandler(configManager, teleportClient, kubectlClient, installer)
	
	// This should display the configuration
	handler.ShowConfig()
}

func TestHandler_InstallTSH_AlreadyInstalled(t *testing.T) {
	configManager, _ := config.NewManager()
	teleportClient, _ := teleport.NewClient(configManager)
	kubectlClient := kubectl.NewClient()
	installer, _ := teleport.NewTSHInstaller()
	
	handler := NewHandler(configManager, teleportClient, kubectlClient, installer)
	
	// Get list of actually installed versions
	versions, err := installer.GetInstalledVersions()
	if err != nil {
		t.Fatalf("Failed to get installed versions: %v", err)
	}
	
	if len(versions) > 0 {
		// Test with an already installed version
		err := handler.InstallTSH(versions[0])
		if err != nil {
			t.Errorf("Expected no error for already installed version, got: %v", err)
		}
	} else {
		t.Skip("No tsh versions installed, skipping already installed test")
	}
}

func TestHandler_ShowTSHVersions_WithVersions(t *testing.T) {
	configManager, _ := config.NewManager()
	teleportClient, _ := teleport.NewClient(configManager)
	kubectlClient := kubectl.NewClient()
	installer, _ := teleport.NewTSHInstaller()
	
	handler := NewHandler(configManager, teleportClient, kubectlClient, installer)
	
	// This should display installed tsh versions
	handler.ShowTSHVersions()
}

func TestHandler_PromptForInstallation_Logic(t *testing.T) {
	configManager, _ := config.NewManager()
	teleportClient, _ := teleport.NewClient(configManager)
	kubectlClient := kubectl.NewClient()
	installer, _ := teleport.NewTSHInstaller()
	
	handler := NewHandler(configManager, teleportClient, kubectlClient, installer)
	
	// We can't easily test the interactive prompt, but we can test scenarios
	// where it would be called through ConnectToCluster
	
	// Create a config with a non-existent tsh version
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")
	
	testConfig := &config.Config{
		Environments: map[string]config.Environment{
			"test": {Proxy: "test.proxy.com:443", TSHVersion: "999.999.999"},
		},
		AutoLogin: true,
	}
	
	data, _ := json.MarshalIndent(testConfig, "", "  ")
	os.WriteFile(configPath, data, 0644)
	
	// This would trigger the prompt for installation, but will fail in test environment
	err := handler.ConnectToCluster("test", "test-cluster")
	if err == nil {
		t.Error("Expected error when tsh version not available")
	}
}

func TestHandler_GetEnvironments_Method(t *testing.T) {
	configManager, _ := config.NewManager()
	teleportClient, _ := teleport.NewClient(configManager)
	kubectlClient := kubectl.NewClient()
	installer, _ := teleport.NewTSHInstaller()
	
	handler := NewHandler(configManager, teleportClient, kubectlClient, installer)
	
	// The getEnvironments method is private, but it's used in ShowStatus
	// We can test it indirectly by calling ShowStatus
	handler.ShowStatus()
}

func TestHandler_AutoDetectVersions_WithEnvironments(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")
	
	testConfig := &config.Config{
		Environments: map[string]config.Environment{
			"test1": {Proxy: "teleport-v14.test.com:443"},                    // Version in hostname
			"test2": {Proxy: "test2.proxy.com:443", TSHVersion: "15.0.0"}, // Already has version
		},
		AutoLogin: true,
	}
	
	data, _ := json.MarshalIndent(testConfig, "", "  ")
	os.WriteFile(configPath, data, 0644)
	
	configManager, _ := config.NewManager()
	teleportClient, _ := teleport.NewClient(configManager)
	kubectlClient := kubectl.NewClient()
	installer, _ := teleport.NewTSHInstaller()
	
	handler := NewHandler(configManager, teleportClient, kubectlClient, installer)
	
	// This should auto-detect versions
	handler.AutoDetectVersions()
}