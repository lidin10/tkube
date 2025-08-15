package shell

import (
	"fmt"
	"strings"
	"tkube/internal/config"
	"tkube/internal/teleport"
)

// Provider handles shell completion operations
type Provider struct {
	configManager  *config.Manager
	teleportClient *teleport.Client
}

// CompletionItem represents a completion suggestion with contextual help
type CompletionItem struct {
	Value       string
	Description string
	Category    string
}

// NewProvider creates a new shell completion provider
func NewProvider(configManager *config.Manager, teleportClient *teleport.Client) *Provider {
	return &Provider{
		configManager:  configManager,
		teleportClient: teleportClient,
	}
}

// GetEnvironments returns a list of environment names for completion
func (p *Provider) GetEnvironments() []string {
	envs, err := p.configManager.GetEnvironments()
	if err != nil {
		return nil
	}
	return envs
}

// GetEnvironmentsWithContext returns environment names with contextual information
func (p *Provider) GetEnvironmentsWithContext() []CompletionItem {
	config, err := p.configManager.Load()
	if err != nil {
		return []CompletionItem{
			{
				Value:       "config-error",
				Description: "❌ Error loading configuration",
				Category:    "error",
			},
		}
	}

	if len(config.Environments) == 0 {
		return []CompletionItem{
			{
				Value:       "no-environments",
				Description: "📝 No environments configured. Run 'tkube config add' to add one",
				Category:    "help",
			},
		}
	}

	var items []CompletionItem
	for env, envConfig := range config.Environments {
		// Get authentication status for contextual description
		sessionInfo := p.teleportClient.GetSessionInfo(env, envConfig.Proxy)
		
		var description string
		var category string
		
		if sessionInfo.IsAuthenticated {
			if sessionInfo.IsExpired {
				description = fmt.Sprintf("⏰ %s (session expired)", envConfig.Proxy)
				category = "expired"
			} else if sessionInfo.TimeRemaining != "" {
				timeStr := p.formatTimeRemaining(sessionInfo.TimeRemaining)
				if strings.Contains(timeStr, "h") && !strings.HasPrefix(timeStr, "1h") && !strings.HasPrefix(timeStr, "2h") {
					description = fmt.Sprintf("✅ %s (%s left)", envConfig.Proxy, timeStr)
					category = "authenticated"
				} else {
					description = fmt.Sprintf("⚠️  %s (%s left - expiring soon)", envConfig.Proxy, timeStr)
					category = "expiring"
				}
			} else {
				description = fmt.Sprintf("✅ %s (authenticated)", envConfig.Proxy)
				category = "authenticated"
			}
		} else {
			description = fmt.Sprintf("❌ %s (not authenticated)", envConfig.Proxy)
			category = "unauthenticated"
		}

		items = append(items, CompletionItem{
			Value:       env,
			Description: description,
			Category:    category,
		})
	}

	return items
}

// GetClusters returns a list of cluster names for a given environment
func (p *Provider) GetClusters(env string) []string {
	clusters, err := p.teleportClient.GetClustersForCompletion(env)
	if err != nil {
		return nil
	}
	return clusters
}

// GetClustersWithContext returns cluster names with contextual information
func (p *Provider) GetClustersWithContext(env string) []CompletionItem {
	envConfig, err := p.configManager.GetEnvironment(env)
	if err != nil {
		return []CompletionItem{
			{
				Value:       "env-not-found",
				Description: fmt.Sprintf("❌ Environment '%s' not found", env),
				Category:    "error",
			},
		}
	}

	// Check tsh version status
	requiredVersion := envConfig.TSHVersion
	if requiredVersion != "" {
		installer, _ := teleport.NewTSHInstaller()
		if !installer.IsVersionInstalled(requiredVersion) {
			return []CompletionItem{
				{
					Value:       "tsh-not-installed",
					Description: fmt.Sprintf("📦 tsh v%s not installed. Run: tkube install-tsh %s", requiredVersion, requiredVersion),
					Category:    "missing-dependency",
				},
			}
		}
	}

	// Check authentication status
	sessionInfo := p.teleportClient.GetSessionInfo(env, envConfig.Proxy)
	if !sessionInfo.IsAuthenticated || sessionInfo.IsExpired {
		return []CompletionItem{
			{
				Value:       "not-authenticated",
				Description: fmt.Sprintf("🔐 Not authenticated to %s. Run: tkube %s <cluster> to authenticate", envConfig.Proxy, env),
				Category:    "authentication-required",
			},
		}
	}

	// Get clusters
	clusters, err := p.teleportClient.GetClustersForCompletion(env)
	if err != nil {
		return []CompletionItem{
			{
				Value:       "cluster-fetch-error",
				Description: fmt.Sprintf("⚠️  Failed to fetch clusters from %s", envConfig.Proxy),
				Category:    "error",
			},
		}
	}

	// Handle special status messages
	if len(clusters) == 1 && (strings.HasPrefix(clusters[0], "📦") ||
		strings.HasPrefix(clusters[0], "❌") ||
		strings.HasPrefix(clusters[0], "⚠️") ||
		strings.HasPrefix(clusters[0], "🔐") ||
		strings.HasPrefix(clusters[0], "ℹ️")) {
		return []CompletionItem{
			{
				Value:       "status-message",
				Description: clusters[0],
				Category:    "status",
			},
		}
	}

	if len(clusters) == 0 {
		return []CompletionItem{
			{
				Value:       "no-clusters",
				Description: fmt.Sprintf("ℹ️  No clusters available in environment '%s'", env),
				Category:    "info",
			},
		}
	}

	// Convert clusters to completion items with descriptions
	var items []CompletionItem
	for _, cluster := range clusters {
		description := fmt.Sprintf("🚀 Connect to %s/%s", env, cluster)
		
		// Add contextual information based on session time remaining
		if sessionInfo.TimeRemaining != "" {
			timeStr := p.formatTimeRemaining(sessionInfo.TimeRemaining)
			if !strings.Contains(timeStr, "h") || strings.HasPrefix(timeStr, "1h") || strings.HasPrefix(timeStr, "2h") {
				description += fmt.Sprintf(" (session expires in %s)", timeStr)
			}
		}

		items = append(items, CompletionItem{
			Value:       cluster,
			Description: description,
			Category:    "cluster",
		})
	}

	return items
}

// GetClustersWithPrefix returns a list of cluster names that match the given prefix
func (p *Provider) GetClustersWithPrefix(env, prefix string) []string {
	clusters, err := p.teleportClient.GetClustersForCompletion(env)
	if err != nil {
		return nil
	}

	// If there's only one result and it's an error/info message, return it regardless of prefix
	if len(clusters) == 1 && (strings.HasPrefix(clusters[0], "📦") ||
		strings.HasPrefix(clusters[0], "❌") ||
		strings.HasPrefix(clusters[0], "⚠️") ||
		strings.HasPrefix(clusters[0], "🔐") ||
		strings.HasPrefix(clusters[0], "ℹ️")) {
		return clusters
	}

	if prefix == "" {
		return clusters
	}

	var filtered []string
	for _, cluster := range clusters {
		if strings.HasPrefix(cluster, prefix) {
			filtered = append(filtered, cluster)
		}
	}

	return filtered
}

// GetCommands returns a list of available commands for completion
func (p *Provider) GetCommands() []string {
	return []string{
		"status",
		"version",
		"config",
		"completion",
		"install-tsh",
		"auto-install-tsh",
		"tsh-versions",
	}
}

// GetCommandsWithContext returns commands with contextual descriptions
func (p *Provider) GetCommandsWithContext() []CompletionItem {
	config, err := p.configManager.Load()
	
	var items []CompletionItem
	
	// Status command with dynamic description
	statusDesc := "📊 Show configured environments and authentication status"
	if err == nil && len(config.Environments) > 0 {
		authCount := 0
		for env, envConfig := range config.Environments {
			if p.teleportClient.CheckAuthenticationStatus(env, envConfig.Proxy) {
				authCount++
			}
		}
		statusDesc = fmt.Sprintf("📊 Show status (%d/%d environments authenticated)", authCount, len(config.Environments))
	} else if err == nil && len(config.Environments) == 0 {
		statusDesc = "📊 Show status (no environments configured)"
	}
	
	items = append(items, CompletionItem{
		Value:       "status",
		Description: statusDesc,
		Category:    "info",
	})

	// Version command
	items = append(items, CompletionItem{
		Value:       "version",
		Description: "🚀 Show tkube version and dependency information",
		Category:    "info",
	})

	// Config command with dynamic description
	configDesc := "⚙️  Manage tkube configuration"
	if err == nil {
		configDesc = fmt.Sprintf("⚙️  Manage configuration (%d environments)", len(config.Environments))
	} else {
		configDesc = "⚙️  Manage configuration (config file not found)"
	}
	
	items = append(items, CompletionItem{
		Value:       "config",
		Description: configDesc,
		Category:    "config",
	})

	// Completion command
	items = append(items, CompletionItem{
		Value:       "completion",
		Description: "🔧 Generate shell completion scripts (bash, zsh, fish, powershell)",
		Category:    "setup",
	})

	// Install-tsh command with dynamic description
	installer, _ := teleport.NewTSHInstaller()
	installedVersions, _ := installer.GetInstalledVersions()
	installDesc := "📦 Install a specific tsh version"
	if len(installedVersions) > 0 {
		installDesc = fmt.Sprintf("📦 Install tsh version (%d versions installed)", len(installedVersions))
	}
	
	items = append(items, CompletionItem{
		Value:       "install-tsh",
		Description: installDesc,
		Category:    "setup",
	})

	// TSH versions command
	versionDesc := "🔧 List installed tsh versions"
	if len(installedVersions) > 0 {
		versionDesc = fmt.Sprintf("🔧 List tsh versions (%d installed)", len(installedVersions))
	} else {
		versionDesc = "🔧 List tsh versions (none installed)"
	}
	
	items = append(items, CompletionItem{
		Value:       "tsh-versions",
		Description: versionDesc,
		Category:    "info",
	})

	return items
}

// GetConfigSubcommands returns a list of config subcommands
func (p *Provider) GetConfigSubcommands() []string {
	return []string{
		"show",
		"path",
		"add",
		"edit",
		"remove",
		"validate",
	}
}

// GetConfigSubcommandswithContext returns config subcommands with contextual descriptions
func (p *Provider) GetConfigSubcommandsWithContext() []CompletionItem {
	config, err := p.configManager.Load()
	
	var items []CompletionItem

	// Show command
	showDesc := "📄 Display current configuration"
	if err == nil {
		showDesc = fmt.Sprintf("📄 Show configuration (%d environments)", len(config.Environments))
	} else {
		showDesc = "📄 Show configuration (config file not found)"
	}
	
	items = append(items, CompletionItem{
		Value:       "show",
		Description: showDesc,
		Category:    "view",
	})

	// Path command
	items = append(items, CompletionItem{
		Value:       "path",
		Description: "📍 Show configuration file location",
		Category:    "view",
	})

	// Add command
	items = append(items, CompletionItem{
		Value:       "add",
		Description: "➕ Interactively add a new environment",
		Category:    "modify",
	})

	// Edit command with dynamic description
	editDesc := "✏️  Interactively edit an existing environment"
	if err == nil && len(config.Environments) > 0 {
		envNames := make([]string, 0, len(config.Environments))
		for env := range config.Environments {
			envNames = append(envNames, env)
		}
		editDesc = fmt.Sprintf("✏️  Edit environment (available: %s)", strings.Join(envNames, ", "))
	} else if err == nil && len(config.Environments) == 0 {
		editDesc = "✏️  Edit environment (no environments to edit)"
	}
	
	items = append(items, CompletionItem{
		Value:       "edit",
		Description: editDesc,
		Category:    "modify",
	})

	// Remove command with dynamic description
	removeDesc := "🗑️  Interactively remove an environment"
	if err == nil && len(config.Environments) > 0 {
		removeDesc = fmt.Sprintf("🗑️  Remove environment (%d available)", len(config.Environments))
	} else if err == nil && len(config.Environments) == 0 {
		removeDesc = "🗑️  Remove environment (no environments to remove)"
	}
	
	items = append(items, CompletionItem{
		Value:       "remove",
		Description: removeDesc,
		Category:    "modify",
	})

	// Validate command
	validateDesc := "✅ Validate configuration for common issues"
	if err != nil {
		validateDesc = "✅ Validate configuration (config file has errors)"
	}
	
	items = append(items, CompletionItem{
		Value:       "validate",
		Description: validateDesc,
		Category:    "check",
	})

	return items
}

// GetCompletionShells returns a list of supported completion shells
func (p *Provider) GetCompletionShells() []string {
	return []string{
		"bash",
		"zsh",
		"fish",
		"powershell",
	}
}

// GetCompletionShellsWithContext returns completion shells with contextual descriptions
func (p *Provider) GetCompletionShellsWithContext() []CompletionItem {
	return []CompletionItem{
		{
			Value:       "bash",
			Description: "🐚 Generate Bash completion script",
			Category:    "shell",
		},
		{
			Value:       "zsh",
			Description: "🐚 Generate Zsh completion script (recommended for macOS)",
			Category:    "shell",
		},
		{
			Value:       "fish",
			Description: "🐚 Generate Fish completion script",
			Category:    "shell",
		},
		{
			Value:       "powershell",
			Description: "🐚 Generate PowerShell completion script",
			Category:    "shell",
		},
	}
}

// GetSystemStatus returns overall system status for contextual help
func (p *Provider) GetSystemStatus() map[string]interface{} {
	status := make(map[string]interface{})
	
	// Check configuration
	config, err := p.configManager.Load()
	if err != nil {
		status["config_error"] = err.Error()
		status["environments_count"] = 0
	} else {
		status["environments_count"] = len(config.Environments)
		status["auto_login"] = config.AutoLogin
		
		// Count authenticated environments
		authCount := 0
		expiredCount := 0
		for env, envConfig := range config.Environments {
			sessionInfo := p.teleportClient.GetSessionInfo(env, envConfig.Proxy)
			if sessionInfo.IsAuthenticated {
				if sessionInfo.IsExpired {
					expiredCount++
				} else {
					authCount++
				}
			}
		}
		status["authenticated_count"] = authCount
		status["expired_count"] = expiredCount
	}
	
	// Check tsh installations
	installer, _ := teleport.NewTSHInstaller()
	installedVersions, _ := installer.GetInstalledVersions()
	status["tsh_versions_installed"] = len(installedVersions)
	
	return status
}

// formatTimeRemaining formats time duration for better readability
func (p *Provider) formatTimeRemaining(timeStr string) string {
	// Remove seconds for cleaner display: "11h16m0s" -> "11h16m"
	if strings.HasSuffix(timeStr, "0s") {
		timeStr = strings.TrimSuffix(timeStr, "0s")
	} else if strings.Contains(timeStr, "s") {
		// If there are seconds, remove them: "11h16m30s" -> "11h16m"
		if idx := strings.LastIndex(timeStr, "m"); idx > 0 {
			if nextIdx := strings.Index(timeStr[idx:], "s"); nextIdx > 0 {
				timeStr = timeStr[:idx+1]
			}
		}
	}
	return timeStr
}
