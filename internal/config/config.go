package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Environment represents a Teleport environment configuration
type Environment struct {
	Proxy      string `json:"proxy"`
	TSHVersion string `json:"tsh_version,omitempty"`
}

// VersionDetector interface for detecting tsh versions
type VersionDetector interface {
	DetectTSHVersion(proxy string) (string, error)
}

// Config represents the main tkube configuration
type Config struct {
	Environments map[string]Environment `json:"environments"`
	AutoLogin    bool                   `json:"auto_login"`
}

// Manager handles configuration operations
type Manager struct {
	configPath string
}

// NewManager creates a new configuration manager
func NewManager() (*Manager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	configPath := filepath.Join(homeDir, ".tkube", "config.json")
	return &Manager{configPath: configPath}, nil
}

// GetPath returns the configuration file path
func (m *Manager) GetPath() string {
	return m.configPath
}

// Load loads the configuration from file
func (m *Manager) Load() (*Config, error) {
	// Create default config if it doesn't exist
	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		if err := m.createDefault(); err != nil {
			return nil, err
		}
	}

	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// Save saves the configuration to file
func (m *Manager) Save(config *Config) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	configDir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(m.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// createDefault creates a default configuration file
func (m *Manager) createDefault() error {
	defaultConfig := Config{
		Environments: map[string]Environment{
			"prod": {Proxy: "teleport.prod.env:443"},
			"test": {Proxy: "teleport.test.env:443"},
		},
		AutoLogin: true,
	}

	return m.Save(&defaultConfig)
}

// GetEnvironments returns a list of environment names
func (m *Manager) GetEnvironments() ([]string, error) {
	config, err := m.Load()
	if err != nil {
		return nil, err
	}

	var envs []string
	for env := range config.Environments {
		envs = append(envs, env)
	}
	return envs, nil
}

// GetEnvironment returns a specific environment configuration
func (m *Manager) GetEnvironment(name string) (*Environment, error) {
	config, err := m.Load()
	if err != nil {
		return nil, err
	}

	env, exists := config.Environments[name]
	if !exists {
		return nil, fmt.Errorf("environment '%s' not found", name)
	}

	return &env, nil
}

// AddEnvironment adds a new environment to the configuration
func (m *Manager) AddEnvironment(name string, env Environment) error {
	config, err := m.Load()
	if err != nil {
		return err
	}

	config.Environments[name] = env
	return m.Save(config)
}

// RemoveEnvironment removes an environment from the configuration
func (m *Manager) RemoveEnvironment(name string) error {
	config, err := m.Load()
	if err != nil {
		return err
	}

	delete(config.Environments, name)
	return m.Save(config)
}

// UpdateAutoLogin updates the auto-login setting
func (m *Manager) UpdateAutoLogin(autoLogin bool) error {
	config, err := m.Load()
	if err != nil {
		return err
	}

	config.AutoLogin = autoLogin
	return m.Save(config)
}

// UpdateEnvironmentTSHVersion updates the tsh version for a specific environment
func (m *Manager) UpdateEnvironmentTSHVersion(envName, tshVersion string) error {
	config, err := m.Load()
	if err != nil {
		return err
	}

	env, exists := config.Environments[envName]
	if !exists {
		return fmt.Errorf("environment '%s' not found", envName)
	}

	env.TSHVersion = tshVersion
	config.Environments[envName] = env

	return m.Save(config)
}

// AutoDetectAndUpdateTSHVersions automatically detects and updates tsh versions for all environments
func (m *Manager) AutoDetectAndUpdateTSHVersions(detector VersionDetector) (map[string]string, error) {
	config, err := m.Load()
	if err != nil {
		return nil, err
	}

	detectedVersions := make(map[string]string)
	updated := false

	for envName, env := range config.Environments {
		// Skip if version is already set
		if env.TSHVersion != "" {
			detectedVersions[envName] = env.TSHVersion
			continue
		}

		// Try to detect version
		version, err := detector.DetectTSHVersion(env.Proxy)
		if err != nil {
			// Log error but continue with other environments
			fmt.Printf("⚠️  Could not detect tsh version for %s (%s): %v\n", envName, env.Proxy, err)
			continue
		}

		if version != "" {
			// Update environment with detected version
			env.TSHVersion = version
			config.Environments[envName] = env
			detectedVersions[envName] = version
			updated = true
			fmt.Printf("🔍 Auto-detected tsh version %s for environment %s (%s)\n", version, envName, env.Proxy)
		}
	}

	// Save configuration if any updates were made
	if updated {
		if err := m.Save(config); err != nil {
			return nil, fmt.Errorf("failed to save updated configuration: %w", err)
		}
		fmt.Println("✅ Configuration updated with auto-detected tsh versions")
	}

	return detectedVersions, nil
}
