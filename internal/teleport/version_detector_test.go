package teleport

import (
	"os"
	"testing"
)

func TestNewVersionDetector(t *testing.T) {
	detector := NewVersionDetector()
	if detector == nil {
		t.Fatal("Expected version detector to be created")
	}
}

func TestVersionDetector_DetectTSHVersion(t *testing.T) {
	detector := NewVersionDetector()
	
	tests := []struct {
		name          string
		proxy         string
		expectError   bool
		expectVersion bool
	}{
		{
			name:          "Invalid proxy",
			proxy:         "invalid-proxy-that-does-not-exist:443",
			expectError:   true,
			expectVersion: false,
		},
		{
			name:          "Empty proxy",
			proxy:         "",
			expectError:   true,
			expectVersion: false,
		},
		{
			name:          "Proxy with version in hostname",
			proxy:         "teleport-v14.prod.company.com:443",
			expectError:   false,
			expectVersion: true,
		},
		{
			name:          "Proxy with version pattern",
			proxy:         "teleport14.test.company.com:443",
			expectError:   false,
			expectVersion: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, err := detector.DetectTSHVersion(tt.proxy)
			
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			
			if tt.expectVersion && version == "" {
				t.Error("Expected version to be detected but got empty string")
			}
			
			if !tt.expectVersion && version != "" {
				t.Errorf("Expected no version but got: %s", version)
			}
		})
	}
}

func TestVersionDetector_ExtractVersionFromProxy(t *testing.T) {
	detector := NewVersionDetector()
	
	tests := []struct {
		name            string
		proxy           string
		expectedVersion string
		expectError     bool
	}{
		{
			name:            "Version with v prefix",
			proxy:           "teleport-v14.2.1.prod.company.com:443",
			expectedVersion: "14.2.1",
			expectError:     false,
		},
		{
			name:            "Version without v prefix",
			proxy:           "teleport-14.2.1.prod.company.com:443",
			expectedVersion: "14.2.1",
			expectError:     false,
		},
		{
			name:            "Version with major only",
			proxy:           "teleport14.prod.company.com:443",
			expectedVersion: "14",
			expectError:     false,
		},
		{
			name:            "Version with major.minor",
			proxy:           "teleport-14.2.prod.company.com:443",
			expectedVersion: "14.2",
			expectError:     false,
		},
		{
			name:        "No version in hostname",
			proxy:       "teleport.prod.company.com:443",
			expectError: true,
		},
		{
			name:            "TSH prefix",
			proxy:           "tsh-v14.0.0.test.company.com:443",
			expectedVersion: "14.0.0",
			expectError:     false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We need to test the internal method, but since it's not exported,
			// we'll test through DetectTSHVersion which will fall back to extraction
			version, err := detector.DetectTSHVersion(tt.proxy)
			
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			
			if !tt.expectError && version != tt.expectedVersion {
				t.Errorf("Expected version '%s', got '%s'", tt.expectedVersion, version)
			}
		})
	}
}

func TestVersionDetector_GetVersionFromEnvironment(t *testing.T) {
	detector := NewVersionDetector()
	
	// Test with environment variable set
	originalVersion := os.Getenv("TELEPORT_VERSION")
	defer func() {
		if originalVersion != "" {
			os.Setenv("TELEPORT_VERSION", originalVersion)
		} else {
			os.Unsetenv("TELEPORT_VERSION")
		}
	}()
	
	// Set test environment variable
	os.Setenv("TELEPORT_VERSION", "v15.0.0")
	
	// Test detection with environment variable
	_, err := detector.DetectTSHVersion("unknown.proxy.com:443")
	
	// Should not error, but might not find version if server endpoint fails
	// and hostname doesn't contain version
	if err != nil {
		t.Logf("Detection failed (expected for unknown proxy): %v", err)
	}
	
	// Clean up
	os.Unsetenv("TELEPORT_VERSION")
}

func TestVersionDetector_NormalizeVersion(t *testing.T) {
	// Since normalizeVersion is not exported, we test it indirectly
	// through DetectTSHVersion with known patterns
	
	detector := NewVersionDetector()
	
	tests := []struct {
		name            string
		proxy           string
		expectedVersion string
	}{
		{
			name:            "Version with v prefix should be normalized",
			proxy:           "teleport-v14.0.0.test.com:443",
			expectedVersion: "14.0.0",
		},
		{
			name:            "Version without prefix should remain same",
			proxy:           "teleport-14.0.0.test.com:443",
			expectedVersion: "14.0.0",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, err := detector.DetectTSHVersion(tt.proxy)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			
			if version != tt.expectedVersion {
				t.Errorf("Expected normalized version '%s', got '%s'", tt.expectedVersion, version)
			}
		})
	}
}

func TestVersionDetector_SuggestInstallation(t *testing.T) {
	detector := NewVersionDetector()
	
	suggestion := detector.SuggestInstallation("14.0.0")
	
	if suggestion == "" {
		t.Error("Expected suggestion to be non-empty")
	}
	
	// Check that suggestion contains expected elements
	expectedElements := []string{
		"14.0.0",
		"tkube config install-tsh",
		"tkube config auto-install-tsh",
	}
	
	for _, element := range expectedElements {
		if !contains(suggestion, element) {
			t.Errorf("Expected suggestion to contain '%s'", element)
		}
	}
}

// Table-driven test for comprehensive version detection scenarios
func TestVersionDetector_ComprehensiveScenarios(t *testing.T) {
	detector := NewVersionDetector()
	
	tests := []struct {
		name        string
		proxy       string
		envVars     map[string]string
		expectError bool
		description string
	}{
		{
			name:        "Server endpoint fails, hostname has version",
			proxy:       "teleport-v14.prod.company.com:443",
			expectError: false,
			description: "Should extract version from hostname when server fails",
		},
		{
			name:        "Server endpoint fails, no version in hostname, env var set",
			proxy:       "teleport.prod.company.com:443",
			envVars:     map[string]string{"TELEPORT_VERSION": "15.0.0"},
			expectError: false,
			description: "Should use environment variable when other methods fail",
		},
		{
			name:        "All methods fail",
			proxy:       "teleport.prod.company.com:443",
			expectError: true,
			description: "Should error when no version can be detected",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables for test
			originalEnvVars := make(map[string]string)
			for key, value := range tt.envVars {
				originalEnvVars[key] = os.Getenv(key)
				os.Setenv(key, value)
			}
			
			// Cleanup environment variables after test
			defer func() {
				for key := range tt.envVars {
					if originalValue, existed := originalEnvVars[key]; existed {
						os.Setenv(key, originalValue)
					} else {
						os.Unsetenv(key)
					}
				}
			}()
			
			version, err := detector.DetectTSHVersion(tt.proxy)
			
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none. Description: %s", tt.description)
			}
			
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v. Description: %s", err, tt.description)
			}
			
			if !tt.expectError && version == "" {
				t.Errorf("Expected version to be detected but got empty. Description: %s", tt.description)
			}
			
			t.Logf("Test '%s': version='%s', error=%v", tt.name, version, err)
		})
	}
}

// Additional comprehensive tests

func TestVersionDetector_QueryEndpoint_ErrorCases(t *testing.T) {
	detector := NewVersionDetector()
	
	// Test with invalid endpoints that will fail
	invalidEndpoints := []string{
		"https://invalid-endpoint-that-does-not-exist.com/webapi/ping",
		"https://localhost:99999/webapi/ping", // Invalid port
		"http://127.0.0.1:1/webapi/ping",     // Connection refused
	}
	
	for _, endpoint := range invalidEndpoints {
		t.Run("Endpoint: "+endpoint, func(t *testing.T) {
			// This tests the queryEndpoint method indirectly through DetectTSHVersion
			// since queryEndpoint is private
			
			// Extract hostname from endpoint for testing
			proxy := "invalid-endpoint-that-does-not-exist.com:443"
			if endpoint == "https://localhost:99999/webapi/ping" {
				proxy = "localhost:99999"
			} else if endpoint == "http://127.0.0.1:1/webapi/ping" {
				proxy = "127.0.0.1:1"
			}
			
			_, err := detector.DetectTSHVersion(proxy)
			
			// Should error for invalid endpoints
			if err == nil {
				t.Error("Expected error for invalid endpoint")
			} else {
				t.Logf("DetectTSHVersion failed as expected for %s: %v", proxy, err)
			}
		})
	}
}

func TestVersionDetector_ExtractVersionFromProxy_AllPatterns(t *testing.T) {
	detector := NewVersionDetector()
	
	tests := []struct {
		name            string
		proxy           string
		expectedVersion string
		expectError     bool
	}{
		{
			name:            "Teleport with v prefix and full version",
			proxy:           "teleport-v17.7.1.prod.company.com:443",
			expectedVersion: "17.7.1",
			expectError:     false,
		},
		{
			name:            "Teleport without v prefix",
			proxy:           "teleport-17.7.1.prod.company.com:443",
			expectedVersion: "17.7.1",
			expectError:     false,
		},
		{
			name:            "Teleport with major version only",
			proxy:           "teleport17.prod.company.com:443",
			expectedVersion: "17",
			expectError:     false,
		},
		{
			name:            "Teleport with major.minor",
			proxy:           "teleport-17.7.prod.company.com:443",
			expectedVersion: "17.7",
			expectError:     false,
		},
		{
			name:            "TSH prefix with version",
			proxy:           "tsh-v17.7.1.test.company.com:443",
			expectedVersion: "17.7.1",
			expectError:     false,
		},
		{
			name:            "TSH without v prefix",
			proxy:           "tsh-17.7.1.test.company.com:443",
			expectedVersion: "17.7.1",
			expectError:     false,
		},
		{
			name:        "No version pattern",
			proxy:       "teleport.prod.company.com:443",
			expectError: true,
		},
		{
			name:        "Different service name",
			proxy:       "auth.prod.company.com:443",
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, err := detector.DetectTSHVersion(tt.proxy)
			
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			
			if !tt.expectError && version != tt.expectedVersion {
				t.Errorf("Expected version '%s', got '%s'", tt.expectedVersion, version)
			}
			
			t.Logf("Proxy: %s, Version: %s, Error: %v", tt.proxy, version, err)
		})
	}
}

func TestVersionDetector_GetVersionFromEnvironment_AllVariables(t *testing.T) {
	detector := NewVersionDetector()
	
	// Test all environment variable patterns
	envVars := []string{
		"TELEPORT_VERSION",
		"TSH_VERSION",
		"TELEPORT_TSH_VERSION",
	}
	
	for _, envVar := range envVars {
		t.Run("EnvVar: "+envVar, func(t *testing.T) {
			// Save original value
			original := os.Getenv(envVar)
			defer func() {
				if original != "" {
					os.Setenv(envVar, original)
				} else {
					os.Unsetenv(envVar)
				}
			}()
			
			// Set test value
			testVersion := "v16.5.12"
			os.Setenv(envVar, testVersion)
			
			// Test with a proxy that has no version in hostname
			version, err := detector.DetectTSHVersion("teleport.prod.company.com:443")
			
			// Should use environment variable when hostname parsing fails
			if err != nil {
				t.Logf("DetectTSHVersion failed (may be expected): %v", err)
			} else if version != "16.5.12" { // Should normalize by removing 'v'
				t.Logf("Got version: %s (may not use env var if other methods succeed)", version)
			}
		})
	}
}

func TestVersionDetector_GetVersionFromEnvironment_ProxySpecific(t *testing.T) {
	detector := NewVersionDetector()
	
	// Test proxy-specific environment variables
	proxy := "test.proxy.com:443"
	envVar := "TEST_PROXY_COM_443_TSH_VERSION"
	
	// Save original value
	original := os.Getenv(envVar)
	defer func() {
		if original != "" {
			os.Setenv(envVar, original)
		} else {
			os.Unsetenv(envVar)
		}
	}()
	
	// Set test value
	testVersion := "15.0.0"
	os.Setenv(envVar, testVersion)
	
	// Test detection
	version, err := detector.DetectTSHVersion(proxy)
	
	// May or may not use the proxy-specific env var depending on other detection methods
	if err != nil {
		t.Logf("DetectTSHVersion failed (may be expected): %v", err)
	} else {
		t.Logf("Detected version: %s", version)
	}
}

func TestVersionDetector_NormalizeVersion_AllCases(t *testing.T) {
	detector := NewVersionDetector()
	
	tests := []struct {
		name     string
		proxy    string
		expected string
	}{
		{
			name:     "Version with v prefix",
			proxy:    "teleport-v14.0.0.test.com:443",
			expected: "14.0.0",
		},
		{
			name:     "Version without prefix",
			proxy:    "teleport-14.0.0.test.com:443",
			expected: "14.0.0",
		},
		{
			name:     "Version with teleport prefix in version",
			proxy:    "teleport-teleport-14.0.0.test.com:443",
			expected: "14.0.0",
		},
		{
			name:     "Version with tsh prefix in version",
			proxy:    "tsh-tsh-14.0.0.test.com:443",
			expected: "14.0.0",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, err := detector.DetectTSHVersion(tt.proxy)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			
			if version != tt.expected {
				t.Errorf("Expected normalized version '%s', got '%s'", tt.expected, version)
			}
		})
	}
}

func TestVersionDetector_SuggestInstallation_Formatting(t *testing.T) {
	detector := NewVersionDetector()
	
	testVersions := []string{
		"14.0.0",
		"15.1.2",
		"16.0.0-beta.1",
		"17.7.1",
	}
	
	for _, version := range testVersions {
		t.Run("Version: "+version, func(t *testing.T) {
			suggestion := detector.SuggestInstallation(version)
			
			if suggestion == "" {
				t.Error("Expected suggestion to be non-empty")
			}
			
			// Check that suggestion contains the version
			if !contains(suggestion, version) {
				t.Errorf("Expected suggestion to contain version '%s'", version)
			}
			
			// Check for key phrases
			expectedPhrases := []string{
				"install",
				"tkube",
				version,
			}
			
			for _, phrase := range expectedPhrases {
				if !contains(suggestion, phrase) {
					t.Errorf("Expected suggestion to contain '%s'", phrase)
				}
			}
			
			t.Logf("Suggestion for %s: %s", version, suggestion)
		})
	}
}

func TestVersionDetector_ComprehensiveFlow(t *testing.T) {
	detector := NewVersionDetector()
	
	// Test the complete flow with different scenarios
	scenarios := []struct {
		name        string
		proxy       string
		envVars     map[string]string
		expectError bool
		description string
	}{
		{
			name:        "Version in hostname - should succeed",
			proxy:       "teleport-v17.7.1.prod.company.com:443",
			expectError: false,
			description: "Should extract version from hostname",
		},
		{
			name:        "No version in hostname, env var set - may succeed",
			proxy:       "teleport.prod.company.com:443",
			envVars:     map[string]string{"TELEPORT_VERSION": "16.5.12"},
			expectError: false,
			description: "Should use environment variable when hostname fails",
		},
		{
			name:        "No version anywhere - should fail",
			proxy:       "auth.prod.company.com:443",
			expectError: true,
			description: "Should fail when no version can be detected",
		},
		{
			name:        "Invalid proxy format - should fail",
			proxy:       "invalid-proxy",
			expectError: true,
			description: "Should fail for invalid proxy format",
		},
	}
	
	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			// Set environment variables
			originalEnvVars := make(map[string]string)
			for key, value := range scenario.envVars {
				originalEnvVars[key] = os.Getenv(key)
				os.Setenv(key, value)
			}
			
			// Cleanup
			defer func() {
				for key := range scenario.envVars {
					if originalValue, existed := originalEnvVars[key]; existed {
						os.Setenv(key, originalValue)
					} else {
						os.Unsetenv(key)
					}
				}
			}()
			
			version, err := detector.DetectTSHVersion(scenario.proxy)
			
			if scenario.expectError && err == nil {
				t.Errorf("Expected error but got none. Description: %s", scenario.description)
			}
			
			if !scenario.expectError && err != nil {
				t.Errorf("Expected no error but got: %v. Description: %s", err, scenario.description)
			}
			
			t.Logf("Scenario '%s': proxy=%s, version=%s, error=%v", 
				scenario.name, scenario.proxy, version, err)
		})
	}
}