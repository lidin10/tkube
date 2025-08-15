package kubectl

import (
	"os/exec"
	"strings"
	"testing"
)

func TestNewClient(t *testing.T) {
	client := NewClient()
	if client == nil {
		t.Fatal("Expected client to be created")
	}
}

func TestClient_IsAvailable(t *testing.T) {
	client := NewClient()
	
	// This test depends on whether kubectl is installed
	// We'll test both scenarios
	available := client.IsAvailable()
	
	// Verify the result matches actual kubectl availability
	cmd := exec.Command("kubectl", "version", "--client", "--short")
	err := cmd.Run()
	expectedAvailable := err == nil
	
	if available != expectedAvailable {
		t.Errorf("Expected IsAvailable() to return %v, got %v", expectedAvailable, available)
	}
}

func TestClient_CheckVersion(t *testing.T) {
	client := NewClient()
	
	if !client.IsAvailable() {
		t.Skip("kubectl not available, skipping version check test")
	}
	
	version, err := client.CheckVersion()
	if err != nil {
		t.Errorf("Expected no error when kubectl is available, got %v", err)
	}
	
	if version == "" {
		t.Error("Expected version string to be non-empty")
	}
}

func TestClient_CheckVersion_NotAvailable(t *testing.T) {
	client := NewClient()
	
	if client.IsAvailable() {
		t.Skip("kubectl is available, cannot test unavailable scenario")
	}
	
	_, err := client.CheckVersion()
	if err == nil {
		t.Error("Expected error when kubectl is not available")
	}
}

func TestClient_GetContexts(t *testing.T) {
	client := NewClient()
	
	if !client.IsAvailable() {
		t.Skip("kubectl not available, skipping contexts test")
	}
	
	contexts, err := client.GetContexts()
	
	// Even if no contexts are configured, this should not error
	// It should return an empty slice
	if err != nil {
		// Only fail if it's not a "no configuration file" error
		if !strings.Contains(err.Error(), "no configuration file") &&
		   !strings.Contains(err.Error(), "no contexts") {
			t.Errorf("Unexpected error getting contexts: %v", err)
		}
	}
	
	// Contexts should be a slice (possibly empty)
	if contexts == nil {
		t.Error("Expected contexts slice, got nil")
	}
}

func TestClient_GetContext(t *testing.T) {
	client := NewClient()
	
	if !client.IsAvailable() {
		t.Skip("kubectl not available, skipping current context test")
	}
	
	context, err := client.GetContext()
	
	// This might error if no context is set, which is valid
	if err != nil {
		// Check if it's an expected error
		if !strings.Contains(err.Error(), "no configuration file") &&
		   !strings.Contains(err.Error(), "current-context is not set") {
			t.Errorf("Unexpected error getting current context: %v", err)
		}
	} else {
		// If no error, context should be a string (possibly empty)
		if context == "" {
			t.Log("Current context is empty (no context set)")
		}
	}
}

func TestClient_SetContext(t *testing.T) {
	client := NewClient()
	
	if !client.IsAvailable() {
		t.Skip("kubectl not available, skipping set context test")
	}
	
	// Test with a non-existent context (should error)
	err := client.SetContext("non-existent-context-12345")
	if err == nil {
		t.Error("Expected error when setting non-existent context")
	}
}

func TestClient_TestConnection(t *testing.T) {
	client := NewClient()
	
	if !client.IsAvailable() {
		t.Skip("kubectl not available, skipping connection test")
	}
	
	// This will likely error unless connected to a cluster
	err := client.TestConnection()
	
	// We don't fail the test if there's no cluster connection
	// This is expected in most test environments
	if err != nil {
		t.Logf("Connection test failed (expected in test environment): %v", err)
	}
}

func TestClient_GetClusterInfo(t *testing.T) {
	client := NewClient()
	
	if !client.IsAvailable() {
		t.Skip("kubectl not available, skipping cluster info test")
	}
	
	// This will likely error unless connected to a cluster
	info, err := client.GetClusterInfo()
	
	// We don't fail the test if there's no cluster connection
	if err != nil {
		t.Logf("Cluster info failed (expected in test environment): %v", err)
	} else {
		// If successful, info should be non-empty
		if info == "" {
			t.Error("Expected cluster info to be non-empty when successful")
		}
	}
}

// Table-driven test for various kubectl scenarios
func TestClient_Operations(t *testing.T) {
	tests := []struct {
		name           string
		operation      string
		expectError    bool
		skipIfNoKubectl bool
	}{
		{
			name:           "Check version",
			operation:      "version",
			expectError:    false,
			skipIfNoKubectl: true,
		},
		{
			name:           "Get contexts",
			operation:      "contexts",
			expectError:    false,
			skipIfNoKubectl: true,
		},
		{
			name:           "Get current context",
			operation:      "current-context",
			expectError:    false, // May error, but that's expected
			skipIfNoKubectl: true,
		},
	}
	
	client := NewClient()
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipIfNoKubectl && !client.IsAvailable() {
				t.Skip("kubectl not available, skipping test")
			}
			
			var err error
			
			switch tt.operation {
			case "version":
				_, err = client.CheckVersion()
			case "contexts":
				_, err = client.GetContexts()
			case "current-context":
				_, err = client.GetContext()
			}
			
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			
			// For operations that might legitimately error (like no config),
			// we don't fail the test, just log
			if err != nil {
				t.Logf("Operation %s returned error (may be expected): %v", tt.operation, err)
			}
		})
	}
}

// Additional comprehensive tests

func TestClient_GetContext_Available(t *testing.T) {
	client := NewClient()
	
	if !client.IsAvailable() {
		t.Skip("kubectl not available, skipping context test")
	}
	
	// Try to get current context
	context, err := client.GetContext()
	
	// This might error if no context is set, which is valid
	if err != nil {
		t.Logf("GetContext returned error (may be expected): %v", err)
	} else {
		t.Logf("Current context: %s", context)
	}
}

func TestClient_GetContexts_Available(t *testing.T) {
	client := NewClient()
	
	if !client.IsAvailable() {
		t.Skip("kubectl not available, skipping contexts test")
	}
	
	// Try to get all contexts
	contexts, err := client.GetContexts()
	
	// This might error if no config file exists
	if err != nil {
		t.Logf("GetContexts returned error (may be expected): %v", err)
	} else {
		t.Logf("Found %d contexts", len(contexts))
	}
}

func TestClient_SetContext_Available(t *testing.T) {
	client := NewClient()
	
	if !client.IsAvailable() {
		t.Skip("kubectl not available, skipping set context test")
	}
	
	// Try to set a non-existent context (should error)
	err := client.SetContext("non-existent-context-test-12345")
	if err == nil {
		t.Error("Expected error when setting non-existent context")
	} else {
		t.Logf("SetContext returned expected error: %v", err)
	}
}

func TestClient_TestConnection_Available(t *testing.T) {
	client := NewClient()
	
	if !client.IsAvailable() {
		t.Skip("kubectl not available, skipping connection test")
	}
	
	// Try to test connection
	err := client.TestConnection()
	
	// This will likely error unless connected to a cluster
	if err != nil {
		t.Logf("TestConnection returned error (expected in test environment): %v", err)
	} else {
		t.Log("TestConnection succeeded")
	}
}

func TestClient_GetClusterInfo_Available(t *testing.T) {
	client := NewClient()
	
	if !client.IsAvailable() {
		t.Skip("kubectl not available, skipping cluster info test")
	}
	
	// Try to get cluster info
	info, err := client.GetClusterInfo()
	
	// This will likely error unless connected to a cluster
	if err != nil {
		t.Logf("GetClusterInfo returned error (expected in test environment): %v", err)
	} else {
		t.Logf("Cluster info: %s", info)
		if info == "" {
			t.Error("Expected cluster info to be non-empty when successful")
		}
	}
}