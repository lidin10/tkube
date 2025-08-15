package shell

import (
	"strings"
	"tkube/internal/config"
	"tkube/internal/teleport"
)

// Provider handles shell completion operations
type Provider struct {
	configManager  *config.Manager
	teleportClient *teleport.Client
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

// GetClusters returns a list of cluster names for a given environment
func (p *Provider) GetClusters(env string) []string {
	clusters, err := p.teleportClient.GetClustersForCompletion(env)
	if err != nil {
		return nil
	}
	return clusters
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

// GetConfigSubcommands returns a list of config subcommands
func (p *Provider) GetConfigSubcommands() []string {
	return []string{
		"show",
		"path",
		"install-tsh",
		"auto-install-tsh",
		"tsh-versions",
	}
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
