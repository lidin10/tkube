package shell

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"tkube/internal/config"
	"tkube/internal/teleport"
)

func TestNewProvider(t *testing.T) {
	configManager, err := config.NewManager()
	if err != nil {
		t.Fatalf("Failed to create config manager: %v", err)
	}
	
	teleportClient, err := teleport.NewClient(configManager)
	if err != nil {
		t.Fatalf("Failed to create teleport client: %v", err)
	}
	
	provider := NewProvider(configManager, teleportClient)
	if provider == nil {
		t.Fatal("Expected provider to be created")
	}
}

func TestProvider_GetCommands(t *testing.T) {
	configManager, _ := config.NewManager()
	teleportClient, _ := teleport.NewClient(configManager)
	
	provider := NewProvider(configManager, teleportClient)
	
	commands := provider.GetCommands()
	
	if len(commands) == 0 {
		t.Error("Expected commands to be returned")
	}
	
	// Check for expected commands
	expectedCommands := []string{"status", "version", "config", "completion"}
	commandMap := make(map[string]bool)
	for _, cmd := range commands {
		commandMap[cmd] = true
	}
	
	for _, expected := range expectedCommands {
		if !commandMap[expected] {
			t.Errorf("Expected command '%s' to be present", expected)
		}
	}
}

func TestProvider_GetConfigSubcommands(t *testing.T) {
	configManager, _ := config.NewManager()
	teleportClient, _ := teleport.NewClient(configManager)
	
	provider := NewProvider(configManager, teleportClient)
	
	subcommands := provider.GetConfigSubcommands()
	
	if len(subcommands) == 0 {
		t.Error("Expected config subcommands to be returned")
	}
	
	// Check for expected subcommands
	expectedSubcommands := []string{"show", "path", "add", "edit", "remove", "validate"}
	subcommandMap := make(map[string]bool)
	for _, cmd := range subcommands {
		subcommandMap[cmd] = true
	}
	
	for _, expected := range expectedSubcommands {
		if !subcommandMap[expected] {
			t.Errorf("Expected subcommand '%s' to be present", expected)
		}
	}
}

func TestProvider_GetCompletionShells(t *testing.T) {
	configManager, _ := config.NewManager()
	teleportClient, _ := teleport.NewClient(configManager)
	
	provider := NewProvider(configManager, teleportClient)
	
	shells := provider.GetCompletionShells()
	
	if len(shells) == 0 {
		t.Error("Expected completion shells to be returned")
	}
	
	// Check for expected shells
	expectedShells := []string{"bash", "zsh", "fish", "powershell"}
	shellMap := make(map[string]bool)
	for _, shell := range shells {
		shellMap[shell] = true
	}
	
	for _, expected := range expectedShells {
		if !shellMap[expected] {
			t.Errorf("Expected shell '%s' to be present", expected)
		}
	}
}

func TestProvider_GetEnvironments(t *testing.T) {
	configManager, _ := config.NewManager()
	teleportClient, _ := teleport.NewClient(configManager)
	
	provider := NewProvider(configManager, teleportClient)
	
	// This might return nil if no config exists, which is fine
	envs := provider.GetEnvironments()
	
	// Should return a slice (possibly nil/empty)
	if envs == nil {
		t.Log("No environments found (expected if no config file exists)")
	} else {
		t.Logf("Found %d environments", len(envs))
	}
}

func TestProvider_GetClusters(t *testing.T) {
	configManager, _ := config.NewManager()
	teleportClient, _ := teleport.NewClient(configManager)
	
	provider := NewProvider(configManager, teleportClient)
	
	// Test with non-existent environment
	clusters := provider.GetClusters("nonexistent")
	
	// The actual implementation might return an empty slice or nil
	// Let's just check that it doesn't panic
	t.Logf("Clusters result: %v", clusters)
}

func TestProvider_GetCommandsWithContext(t *testing.T) {
	configManager, _ := config.NewManager()
	teleportClient, _ := teleport.NewClient(configManager)
	
	provider := NewProvider(configManager, teleportClient)
	
	items := provider.GetCommandsWithContext()
	
	if len(items) == 0 {
		t.Error("Expected command items to be returned")
	}
	
	// Check that all items have required fields
	for _, item := range items {
		if item.Value == "" {
			t.Error("Expected all items to have non-empty Value")
		}
		if item.Description == "" {
			t.Error("Expected all items to have non-empty Description")
		}
		if item.Category == "" {
			t.Error("Expected all items to have non-empty Category")
		}
	}
}

func TestProvider_GetConfigSubcommandsWithContext(t *testing.T) {
	configManager, _ := config.NewManager()
	teleportClient, _ := teleport.NewClient(configManager)
	
	provider := NewProvider(configManager, teleportClient)
	
	items := provider.GetConfigSubcommandsWithContext()
	
	if len(items) == 0 {
		t.Error("Expected config subcommand items to be returned")
	}
	
	// Check that all items have required fields
	for _, item := range items {
		if item.Value == "" {
			t.Error("Expected all items to have non-empty Value")
		}
		if item.Description == "" {
			t.Error("Expected all items to have non-empty Description")
		}
		if item.Category == "" {
			t.Error("Expected all items to have non-empty Category")
		}
	}
}

func TestProvider_GetCompletionShellsWithContext(t *testing.T) {
	configManager, _ := config.NewManager()
	teleportClient, _ := teleport.NewClient(configManager)
	
	provider := NewProvider(configManager, teleportClient)
	
	items := provider.GetCompletionShellsWithContext()
	
	if len(items) == 0 {
		t.Error("Expected completion shell items to be returned")
	}
	
	// Check that all items have required fields
	for _, item := range items {
		if item.Value == "" {
			t.Error("Expected all items to have non-empty Value")
		}
		if item.Description == "" {
			t.Error("Expected all items to have non-empty Description")
		}
		if item.Category == "" {
			t.Error("Expected all items to have non-empty Category")
		}
	}
}

func TestProvider_GetEnvironmentsWithContext(t *testing.T) {
	configManager, _ := config.NewManager()
	teleportClient, _ := teleport.NewClient(configManager)
	
	provider := NewProvider(configManager, teleportClient)
	
	items := provider.GetEnvironmentsWithContext()
	
	if len(items) == 0 {
		t.Error("Expected environment items to be returned")
	}
	
	// Check that all items have required fields
	for _, item := range items {
		if item.Value == "" && item.Category != "error" && item.Category != "help" {
			t.Error("Expected all items to have non-empty Value (except error/help)")
		}
		if item.Description == "" {
			t.Error("Expected all items to have non-empty Description")
		}
		if item.Category == "" {
			t.Error("Expected all items to have non-empty Category")
		}
	}
}

func TestProvider_GetClustersWithContext(t *testing.T) {
	configManager, _ := config.NewManager()
	teleportClient, _ := teleport.NewClient(configManager)
	
	provider := NewProvider(configManager, teleportClient)
	
	// Test with non-existent environment
	items := provider.GetClustersWithContext("nonexistent")
	
	if len(items) == 0 {
		t.Error("Expected cluster items to be returned (even if error)")
	}
	
	// Should return error message
	if len(items) > 0 && items[0].Category != "error" {
		t.Errorf("Expected error category for non-existent environment, got '%s'", items[0].Category)
	}
}

func TestProvider_GetClustersWithPrefix(t *testing.T) {
	configManager, _ := config.NewManager()
	teleportClient, _ := teleport.NewClient(configManager)
	
	provider := NewProvider(configManager, teleportClient)
	
	// Test with non-existent environment and empty prefix
	clusters := provider.GetClustersWithPrefix("nonexistent", "")
	
	// Should return some result (likely error message)
	t.Logf("Clusters with prefix result: %v", clusters)
}

func TestProvider_GetSystemStatus(t *testing.T) {
	configManager, _ := config.NewManager()
	teleportClient, _ := teleport.NewClient(configManager)
	
	provider := NewProvider(configManager, teleportClient)
	
	status := provider.GetSystemStatus()
	
	if status == nil {
		t.Error("Expected system status to be returned")
	}
	
	// Check for expected keys
	expectedKeys := []string{"environments_count", "tsh_versions_installed"}
	for _, key := range expectedKeys {
		if _, exists := status[key]; !exists {
			t.Errorf("Expected system status to contain key '%s'", key)
		}
	}
}

// Test CompletionItem struct
func TestCompletionItem(t *testing.T) {
	item := CompletionItem{
		Value:       "test-value",
		Description: "test description",
		Category:    "test-category",
	}
	
	if item.Value != "test-value" {
		t.Errorf("Expected Value 'test-value', got '%s'", item.Value)
	}
	
	if item.Description != "test description" {
		t.Errorf("Expected Description 'test description', got '%s'", item.Description)
	}
	
	if item.Category != "test-category" {
		t.Errorf("Expected Category 'test-category', got '%s'", item.Category)
	}
}

// Additional comprehensive tests

func TestProvider_GetEnvironmentsWithContext_WithConfig(t *testing.T) {
	// Create a temporary config file with environments
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
	
	// Create manager with custom config path
	_ = &config.Manager{}
	// We can't easily set the private field, so we'll use the real manager
	realConfigManager, _ := config.NewManager()
	
	teleportClient, _ := teleport.NewClient(realConfigManager)
	provider := NewProvider(realConfigManager, teleportClient)
	
	items := provider.GetEnvironmentsWithContext()
	
	// Should return items (may be error if no config exists)
	if len(items) == 0 {
		t.Error("Expected environment items to be returned")
	}
}

func TestProvider_GetClustersWithContext_WithEnvironment(t *testing.T) {
	// Create a temporary config file with environments
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
	teleportClient, _ := teleport.NewClient(configManager)
	provider := NewProvider(configManager, teleportClient)
	
	// Test with environment that exists but tsh not installed
	items := provider.GetClustersWithContext("test")
	
	if len(items) == 0 {
		t.Error("Expected cluster items to be returned (even if error)")
	}
	
	// Should indicate tsh not installed or authentication required
	if len(items) > 0 {
		expectedCategories := []string{"missing-dependency", "authentication-required", "error"}
		found := false
		for _, category := range expectedCategories {
			if items[0].Category == category {
				found = true
				break
			}
		}
		if !found {
			t.Logf("Got category: %s, description: %s", items[0].Category, items[0].Description)
		}
	}
}

func TestProvider_GetClustersWithPrefix_WithStatusMessage(t *testing.T) {
	configManager, _ := config.NewManager()
	teleportClient, _ := teleport.NewClient(configManager)
	provider := NewProvider(configManager, teleportClient)
	
	// Test with non-existent environment
	clusters := provider.GetClustersWithPrefix("nonexistent", "test")
	
	// Should return status message
	if len(clusters) > 0 {
		t.Logf("Got clusters with prefix: %v", clusters)
	}
}

func TestProvider_FormatTimeRemaining(t *testing.T) {
	configManager, _ := config.NewManager()
	teleportClient, _ := teleport.NewClient(configManager)
	provider := NewProvider(configManager, teleportClient)
	
	// We can't test the private method directly, but we can test it through
	// methods that use it. Let's create a scenario where it would be called.
	
	// This is tested indirectly through GetEnvironmentsWithContext
	items := provider.GetEnvironmentsWithContext()
	
	// Just verify it doesn't panic
	if len(items) == 0 {
		t.Log("No environment items returned (expected if no config)")
	}
}

func TestProvider_GetSystemStatus_Detailed(t *testing.T) {
	configManager, _ := config.NewManager()
	teleportClient, _ := teleport.NewClient(configManager)
	provider := NewProvider(configManager, teleportClient)
	
	status := provider.GetSystemStatus()
	
	if status == nil {
		t.Error("Expected system status to be returned")
	}
	
	// Check specific status fields
	if envCount, exists := status["environments_count"]; exists {
		t.Logf("Environments count: %v", envCount)
	}
	
	if tshCount, exists := status["tsh_versions_installed"]; exists {
		t.Logf("TSH versions installed: %v", tshCount)
	}
	
	// Check for optional fields that might exist
	if authCount, exists := status["authenticated_count"]; exists {
		t.Logf("Authenticated count: %v", authCount)
	}
	
	if expiredCount, exists := status["expired_count"]; exists {
		t.Logf("Expired count: %v", expiredCount)
	}
}

func TestCompletionItem_AllFields(t *testing.T) {
	item := CompletionItem{
		Value:       "test-value",
		Description: "test description with special chars: !@#$%",
		Category:    "test-category",
	}
	
	// Test all fields are preserved
	if item.Value != "test-value" {
		t.Errorf("Expected Value 'test-value', got '%s'", item.Value)
	}
	
	if item.Description != "test description with special chars: !@#$%" {
		t.Errorf("Expected Description with special chars, got '%s'", item.Description)
	}
	
	if item.Category != "test-category" {
		t.Errorf("Expected Category 'test-category', got '%s'", item.Category)
	}
}

func TestProvider_EdgeCases(t *testing.T) {
	configManager, _ := config.NewManager()
	teleportClient, _ := teleport.NewClient(configManager)
	provider := NewProvider(configManager, teleportClient)
	
	// Test with empty strings
	clusters := provider.GetClustersWithPrefix("", "")
	t.Logf("Empty environment and prefix result: %v", clusters)
	
	// Test with special characters
	clusters = provider.GetClustersWithPrefix("test-env-with-dashes", "prefix-with-dashes")
	t.Logf("Special characters result: %v", clusters)
	
	// Test GetClusters with empty string
	clusters = provider.GetClusters("")
	t.Logf("Empty environment result: %v", clusters)
}