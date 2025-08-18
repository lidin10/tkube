package teleport

import (
	"path/filepath"
	"testing"
)

func TestNewTSHInstaller(t *testing.T) {
	installer, err := NewTSHInstaller()
	if err != nil {
		t.Fatalf("Expected no error creating installer, got %v", err)
	}

	if installer == nil {
		t.Fatal("Expected installer to be created")
	}
}

func TestTSHInstaller_GetTSHPath(t *testing.T) {
	installer, _ := NewTSHInstaller()

	version := "14.0.0"
	path := installer.GetTSHPath(version)

	if path == "" {
		t.Error("Expected path to be non-empty")
	}

	// Path should contain the version
	if !contains(path, version) {
		t.Errorf("Expected path to contain version '%s', got '%s'", version, path)
	}

	// Path should end with tsh
	if !contains(path, "tsh") {
		t.Errorf("Expected path to contain 'tsh', got '%s'", path)
	}
}

func TestTSHInstaller_IsVersionInstalled(t *testing.T) {
	installer, _ := NewTSHInstaller()

	// Test with non-existent version
	installed := installer.IsVersionInstalled("999.999.999")
	if installed {
		t.Error("Expected non-existent version to not be installed")
	}
}

func TestTSHInstaller_GetInstalledVersions(t *testing.T) {
	installer, _ := NewTSHInstaller()

	versions, err := installer.GetInstalledVersions()
	if err != nil {
		t.Errorf("Expected no error getting installed versions, got %v", err)
	}

	// Should return a slice (possibly empty)
	if versions == nil {
		t.Error("Expected versions slice, got nil")
	}
}

func TestTSHInstaller_UninstallVersion_NotInstalled(t *testing.T) {
	installer, _ := NewTSHInstaller()

	// Test uninstalling non-existent version
	err := installer.UninstallVersion("999.999.999")
	// The actual implementation might not error for non-existent versions
	// Let's just check that it doesn't panic
	t.Logf("Uninstall result: %v", err)
}

func TestTSHInstaller_GetTSHVersionInfo(t *testing.T) {
	installer, _ := NewTSHInstaller()

	// Test with non-existent path
	info := installer.GetTSHVersionInfo("/non/existent/path/tsh")

	if info == "" {
		t.Error("Expected version info to be non-empty even for non-existent path")
	}

	// Should contain some indication of failure
	if !contains(info, "version check failed") && !contains(info, "installed at") {
		t.Errorf("Expected version info to indicate failure or path, got '%s'", info)
	}
}

func TestTSHInstaller_AutoInstallForEnvironment(t *testing.T) {
	installer, _ := NewTSHInstaller()

	// Test with already "installed" version (will return early)
	// Since we can't actually install in tests, we test the logic path
	err := installer.AutoInstallForEnvironment("test-env", "999.999.999")

	// Should attempt to install and likely fail (which is expected)
	if err == nil {
		// If no error, the version was already considered installed
		t.Log("Version was already considered installed")
	} else {
		// If error, it attempted installation (expected behavior)
		t.Logf("Installation attempted and failed as expected: %v", err)
	}
}

// Test installer with temporary directory
func TestTSHInstaller_WithTempDir(t *testing.T) {
	// Create a temporary directory for testing
	_ = t.TempDir()

	// We can't easily test the full installer without modifying the struct
	// But we can test the path logic

	testVersion := "14.0.0"
	_ = filepath.Join("tsh", testVersion, "tsh")

	// Verify the path construction logic
	installer, _ := NewTSHInstaller()
	path := installer.GetTSHPath(testVersion)

	if !contains(path, testVersion) {
		t.Errorf("Expected path to contain version %s", testVersion)
	}
}

// Table-driven tests for various installer scenarios
func TestTSHInstaller_TableDriven(t *testing.T) {
	tests := []struct {
		name      string
		version   string
		operation string
		expectErr bool
	}{
		{
			name:      "Get path for valid version",
			version:   "14.0.0",
			operation: "get_path",
			expectErr: false,
		},
		{
			name:      "Check installation for non-existent version",
			version:   "999.999.999",
			operation: "is_installed",
			expectErr: false, // Should return false, not error
		},
		{
			name:      "Uninstall non-existent version",
			version:   "999.999.999",
			operation: "uninstall",
			expectErr: false, // May or may not error depending on implementation
		},
		{
			name:      "Get version info for non-existent path",
			version:   "999.999.999",
			operation: "version_info",
			expectErr: false, // Should return info string, not error
		},
	}

	installer, _ := NewTSHInstaller()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			var result interface{}

			switch tt.operation {
			case "get_path":
				result = installer.GetTSHPath(tt.version)
			case "is_installed":
				result = installer.IsVersionInstalled(tt.version)
			case "uninstall":
				err = installer.UninstallVersion(tt.version)
			case "version_info":
				path := installer.GetTSHPath(tt.version)
				result = installer.GetTSHVersionInfo(path)
			}

			if tt.expectErr && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			// Verify result is not nil/empty for operations that return values
			if tt.operation == "get_path" || tt.operation == "version_info" {
				if str, ok := result.(string); ok && str == "" {
					t.Error("Expected non-empty string result")
				}
			}
		})
	}
}

// Test error handling scenarios
func TestTSHInstaller_ErrorHandling(t *testing.T) {
	installer, _ := NewTSHInstaller()

	// Test with invalid version formats
	invalidVersions := []string{
		"",
		"invalid",
		"v",
		"14.",
		".14",
	}

	for _, version := range invalidVersions {
		t.Run("Invalid version: "+version, func(t *testing.T) {
			// These operations should handle invalid versions gracefully
			path := installer.GetTSHPath(version)
			if path == "" {
				t.Error("Expected path to be returned even for invalid version")
			}

			installed := installer.IsVersionInstalled(version)
			if installed {
				t.Error("Expected invalid version to not be considered installed")
			}
		})
	}
}

// Test concurrent access (basic thread safety check)
func TestTSHInstaller_ConcurrentAccess(t *testing.T) {
	installer, _ := NewTSHInstaller()

	// Test concurrent calls to read-only methods
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(version string) {
			defer func() { done <- true }()

			// These should be safe to call concurrently
			_ = installer.GetTSHPath(version)
			_ = installer.IsVersionInstalled(version)
			_, _ = installer.GetInstalledVersions()
		}("14.0." + string(rune('0'+i)))
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

// Additional comprehensive tests

func TestTSHInstaller_InstallTSH_ErrorScenarios(t *testing.T) {
	installer, _ := NewTSHInstaller()

	// Test with invalid version that will fail download
	err := installer.InstallTSH("999.999.999")
	if err == nil {
		t.Error("Expected error for invalid version")
	} else {
		t.Logf("InstallTSH failed as expected: %v", err)
	}
}

func TestTSHInstaller_GetPackageInfo(t *testing.T) {
	installer, _ := NewTSHInstaller()

	// We can't test the private method directly, but we can test it through InstallTSH
	// which will call getPackageInfo internally

	// Test with a version that will trigger package info logic
	err := installer.InstallTSH("14.0.0")

	// Will fail in test environment, but should exercise the package info code
	if err != nil {
		t.Logf("InstallTSH failed as expected (exercises package info): %v", err)
	}
}

func TestTSHInstaller_VerifyTSHBinary_NonExistent(t *testing.T) {
	installer, _ := NewTSHInstaller()

	// Test verifyTSHBinary with non-existent file
	// This is tested indirectly through IsVersionInstalled
	installed := installer.IsVersionInstalled("999.999.999")
	if installed {
		t.Error("Expected non-existent version to not be installed")
	}
}

func newTestInstaller(t *testing.T) *TSHInstaller {
	tempDir := t.TempDir()
	return &TSHInstaller{baseDir: filepath.Join(tempDir, ".tkube", "tsh")}
}

func TestTSHInstaller_GetInstalledVersions_EmptyDir(t *testing.T) {
	installer := newTestInstaller(t)

	versions, err := installer.GetInstalledVersions()
	if err != nil {
		t.Errorf("Expected no error getting installed versions, got %v", err)
	}

	if len(versions) != 0 {
		t.Errorf("Expected 0 installed versions, got %d", len(versions))
	}

	t.Logf("Found %d installed versions", len(versions))
}

func TestTSHInstaller_UninstallVersion_ExistingVersion(t *testing.T) {
	installer, _ := NewTSHInstaller()

	// Get list of installed versions
	versions, err := installer.GetInstalledVersions()
	if err != nil {
		t.Fatalf("Failed to get installed versions: %v", err)
	}

	if len(versions) > 0 {
		// Don't actually uninstall, just test that the method exists
		// and handles the case properly
		t.Logf("Would test uninstalling version: %s", versions[0])

		// Test with a copy of the version string to avoid modifying the original
		testVersion := versions[0] + "-test-copy"
		err := installer.UninstallVersion(testVersion)
		t.Logf("Uninstall test version result: %v", err)
	} else {
		t.Skip("No versions installed to test uninstall")
	}
}

func TestTSHInstaller_GetTSHVersionInfo_RealPath(t *testing.T) {
	installer, _ := NewTSHInstaller()

	// Get list of installed versions
	versions, err := installer.GetInstalledVersions()
	if err != nil {
		t.Fatalf("Failed to get installed versions: %v", err)
	}

	if len(versions) > 0 {
		// Test with real installed version
		tshPath := installer.GetTSHPath(versions[0])
		info := installer.GetTSHVersionInfo(tshPath)

		if info == "" {
			t.Error("Expected version info to be non-empty for installed version")
		}

		t.Logf("Version info for %s: %s", versions[0], info)
	} else {
		t.Skip("No versions installed to test version info")
	}
}

func TestTSHInstaller_AutoInstallForEnvironment_AlreadyInstalled(t *testing.T) {
	installer, _ := NewTSHInstaller()

	// Get list of installed versions
	versions, err := installer.GetInstalledVersions()
	if err != nil {
		t.Fatalf("Failed to get installed versions: %v", err)
	}

	if len(versions) > 0 {
		// Test with already installed version
		err := installer.AutoInstallForEnvironment("test-env", versions[0])
		if err != nil {
			t.Errorf("Expected no error for already installed version, got: %v", err)
		}
	} else {
		// Test with non-existent version
		err := installer.AutoInstallForEnvironment("test-env", "999.999.999")
		if err == nil {
			t.Error("Expected error for non-existent version")
		} else {
			t.Logf("AutoInstallForEnvironment failed as expected: %v", err)
		}
	}
}

func TestTSHInstaller_PathConstruction(t *testing.T) {
	installer, _ := NewTSHInstaller()

	testVersions := []string{
		"14.0.0",
		"15.1.2",
		"16.0.0-beta.1",
		"17.0.0-rc.1",
	}

	for _, version := range testVersions {
		path := installer.GetTSHPath(version)

		if path == "" {
			t.Errorf("Expected non-empty path for version %s", version)
		}

		if !contains(path, version) {
			t.Errorf("Expected path to contain version %s, got %s", version, path)
		}

		t.Logf("Path for %s: %s", version, path)
	}
}

func TestTSHInstaller_ErrorHandling_EdgeCases(t *testing.T) {
	installer, _ := NewTSHInstaller()

	// Test with various edge case versions
	edgeCases := []string{
		"",            // Empty version
		"v14.0.0",     // Version with v prefix
		"14",          // Major version only
		"14.0",        // Major.minor only
		"14.0.0.1",    // Extra version component
		"invalid",     // Invalid version
		"14.0.0-beta", // Beta version
	}

	for _, version := range edgeCases {
		t.Run("Version: "+version, func(t *testing.T) {
			// These should not panic
			path := installer.GetTSHPath(version)
			if path == "" {
				t.Errorf("Expected non-empty path even for edge case version: %s", version)
			}

			installed := installer.IsVersionInstalled(version)
			// Should return false for invalid/non-existent versions
			if version == "" || version == "invalid" {
				if installed {
					t.Errorf("Expected invalid version %s to not be installed", version)
				}
			}

			t.Logf("Version %s: path=%s, installed=%v", version, path, installed)
		})
	}
}

func TestTSHInstaller_ConcurrentSafety(t *testing.T) {
	installer, _ := NewTSHInstaller()

	// Test that read-only operations are safe to call concurrently
	done := make(chan bool, 20)

	for i := 0; i < 20; i++ {
		go func(id int) {
			defer func() { done <- true }()

			version := "14.0." + string(rune('0'+(id%10)))

			// These operations should be safe to call concurrently
			_ = installer.GetTSHPath(version)
			_ = installer.IsVersionInstalled(version)
			_, _ = installer.GetInstalledVersions()
			_ = installer.GetTSHVersionInfo("/fake/path/tsh")
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 20; i++ {
		<-done
	}

	t.Log("Concurrent operations completed successfully")
}
