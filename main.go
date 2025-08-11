package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var version = "1.0.0" // Set by build process

type Config struct {
	Environments map[string]string `json:"environments"`
	AutoLogin    bool              `json:"auto_login"`
}

var rootCmd = &cobra.Command{
	Use:   "tkube <environment> <cluster>",
	Short: "🚀 Enhanced Teleport kubectl wrapper with auto-authentication",
	Long: `🚀 tkube - Enhanced Teleport kubectl wrapper

Quickly connect to Kubernetes clusters via Teleport with intelligent 
auto-authentication and cross-shell compatibility.

tkube simplifies your workflow by:
  • Automatically authenticating to Teleport when needed
  • Managing multiple environment configurations
  • Providing smart tab completion for environments and clusters
  • Working seamlessly across bash, zsh, and fish shells

Configuration is stored in ~/.tkube/config.json and created automatically
on first run with example environments.`,
	Example: `  # Connect to a production cluster
  tkube prod my-app-cluster

  # Connect to development environment  
  tkube dev local-cluster

  # Use tab completion to discover clusters
  tkube prod <TAB>

  # Check authentication status across environments
  tkube status`,
	Args: cobra.ExactArgs(2),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			// Complete environments
			return getEnvironments(), cobra.ShellCompDirectiveNoFileComp
		}
		if len(args) == 1 {
			// Complete clusters for the given environment
			return getClusters(args[0]), cobra.ShellCompDirectiveNoFileComp
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return connectToCluster(args[0], args[1])
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version and dependency information",
	Long: `Display tkube version information along with the status of required 
and optional dependencies like tsh (Teleport CLI) and kubectl.

This command helps verify your installation and troubleshoot any 
missing dependencies.`,
	Run: func(cmd *cobra.Command, args []string) {
		showVersion()
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show configured environments and authentication status",
	Long: `Display all configured environments from ~/.tkube/config.json along 
with their Teleport proxy addresses and current authentication status.

This command helps you:
  • See which environments are available
  • Check if you're authenticated to each Teleport proxy
  • Verify your configuration is correct
  • Understand your current auto-login setting`,
	Example: `  # Check status of all environments
  tkube status

  # Typical output shows:
  # ✅ prod → teleport.prod.company.com:443 (authenticated)
  # ❌ test → teleport.test.company.com:443 (not authenticated)`,
	Run: func(cmd *cobra.Command, args []string) {
		showStatus()
	},
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage tkube configuration",
	Long: `Manage your tkube configuration file (~/.tkube/config.json).

The configuration file stores your Teleport environments and settings.
It's automatically created with example values on first run.`,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Long: `Display the current tkube configuration including all environments,
their Teleport proxy addresses, and settings like auto-login.`,
	Run: func(cmd *cobra.Command, args []string) {
		showConfig()
	},
}

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Show configuration file path",
	Long:  `Display the path to your tkube configuration file.`,
	Run: func(cmd *cobra.Command, args []string) {
		showConfigPath()
	},
}

func init() {
	// Add subcommands to config
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configPathCmd)
	
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(configCmd)
	
	// Add completion command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion scripts",
		Long: `Generate shell completion scripts for tkube.

Shell completion provides intelligent tab completion for:
  • Environment names (from your config)
  • Cluster names (fetched live from Teleport)
  • Command names and flags

Once installed, you can use tab completion like:
  tkube <TAB>           # Shows: prod, test, dev, help, status, version
  tkube prod <TAB>      # Shows: cluster1, cluster2, cluster3...

INSTALLATION INSTRUCTIONS:

Bash:
  # Load for current session
  source <(tkube completion bash)

  # Install permanently
  # Linux:
  tkube completion bash > /etc/bash_completion.d/tkube
  # macOS with Homebrew:
  tkube completion bash > /usr/local/etc/bash_completion.d/tkube

Zsh:
  # Load for current session  
  source <(tkube completion zsh)

  # Install permanently
  tkube completion zsh > "${fpath[1]}/_tkube"
  # Then restart your shell

Fish:
  # Load for current session
  tkube completion fish | source

  # Install permanently
  tkube completion fish > ~/.config/fish/completions/tkube.fish

PowerShell:
  # Load for current session
  tkube completion powershell | Out-String | Invoke-Expression

  # Install permanently
  tkube completion powershell > tkube.ps1
  # Then source from your PowerShell profile
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		Run: func(cmd *cobra.Command, args []string) {
			switch args[0] {
			case "bash":
				cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			}
		},
	})
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func connectToCluster(env, cluster string) error {
	config, err := loadConfig()
	if err != nil {
		fmt.Printf("❌ Error loading configuration: %v\n", err)
		fmt.Println("💡 Run 'tkube config path' to see the expected config location")
		return err
	}

	proxy, exists := config.Environments[env]
	if !exists {
		fmt.Printf("❌ Unknown environment '%s'\n", env)
		fmt.Println()
		fmt.Printf("Available environments: %s\n", strings.Join(getEnvironments(), ", "))
		fmt.Println()
		fmt.Println("💡 Tips:")
		fmt.Println("   • Run 'tkube status' to see all configured environments")
		fmt.Println("   • Run 'tkube config show' to see your configuration")
		fmt.Println("   • Use tab completion: tkube <TAB>")
		return fmt.Errorf("unknown environment")
	}

	// Check authentication status
	if !isAuthenticated(proxy) {
		if config.AutoLogin {
			fmt.Printf("🔐 Not authenticated to %s, attempting login...\n", proxy)
			if err := teleportLogin(proxy); err != nil {
				fmt.Printf("❌ Failed to authenticate to %s\n", proxy)
				fmt.Println()
				fmt.Println("💡 Troubleshooting:")
				fmt.Printf("   • Verify proxy address: %s\n", proxy)
				fmt.Println("   • Check your network connection")
				fmt.Println("   • Ensure you have valid Teleport credentials")
				fmt.Printf("   • Try manual login: tsh login --proxy=%s\n", proxy)
				return err
			}
			fmt.Printf("✅ Successfully authenticated to %s\n", proxy)
		} else {
			fmt.Printf("❌ Not authenticated to %s\n", proxy)
			fmt.Println()
			fmt.Println("🔐 Authentication required:")
			fmt.Printf("   tsh login --proxy=%s\n", proxy)
			fmt.Println()
			fmt.Println("💡 Or enable auto-login by setting 'auto_login': true in your config")
			fmt.Printf("   Config file: %s\n", getConfigPath())
			return fmt.Errorf("authentication required")
		}
	}

	// Connect to Kubernetes cluster
	fmt.Printf("🚀 Connecting to cluster '%s' in '%s' environment...\n", cluster, env)
	if err := kubeLogin(proxy, cluster); err != nil {
		fmt.Printf("❌ Failed to connect to cluster '%s'\n", cluster)
		fmt.Println()
		fmt.Println("💡 Troubleshooting:")
		fmt.Println("   • Verify cluster name is correct")
		fmt.Printf("   • Check available clusters: tsh --proxy=%s kube ls\n", proxy)
		fmt.Println("   • Ensure you have access to this cluster")
		fmt.Println("   • Use tab completion: tkube " + env + " <TAB>")
		return err
	}
	
	fmt.Printf("✅ Successfully connected to cluster '%s'!\n", cluster)
	fmt.Println()
	fmt.Println("💡 You can now use kubectl to interact with your cluster")
	return nil
}

func getEnvironments() []string {
	config, err := loadConfig()
	if err != nil {
		return nil
	}

	var envs []string
	for env := range config.Environments {
		envs = append(envs, env)
	}
	return envs
}

func getClusters(env string) []string {
	config, err := loadConfig()
	if err != nil {
		return nil
	}

	proxy, exists := config.Environments[env]
	if !exists {
		return nil
	}

	// Try to get clusters from tsh
	cmd := exec.Command("tsh", "--proxy="+proxy, "kube", "ls", "--format=json")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	// Parse JSON output to extract cluster names
	var clusters []map[string]interface{}
	if err := json.Unmarshal(output, &clusters); err != nil {
		return nil
	}

	var clusterNames []string
	for _, cluster := range clusters {
		if name, ok := cluster["kube_cluster_name"].(string); ok {
			clusterNames = append(clusterNames, name)
		}
	}

	return clusterNames
}

func loadConfig() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(homeDir, ".tkube", "config.json")
	
	// Create default config if it doesn't exist
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := createDefaultConfig(configPath); err != nil {
			return nil, err
		}
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func createDefaultConfig(configPath string) error {
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	defaultConfig := Config{
		Environments: map[string]string{
			"prod": "teleport.prod.company.com:443",
			"test": "teleport.test.company.com:443",
			"dev":  "teleport.dev.company.com:443",
		},
		AutoLogin: true,
	}

	data, err := json.MarshalIndent(defaultConfig, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

func isAuthenticated(proxy string) bool {
	cmd := exec.Command("tsh", "status", "--proxy="+proxy)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}

	outputStr := string(output)
	return strings.Contains(outputStr, "logged in") || strings.Contains(outputStr, "Valid until")
}

func teleportLogin(proxy string) error {
	cmd := exec.Command("tsh", "login", "--proxy="+proxy)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func kubeLogin(proxy, cluster string) error {
	cmd := exec.Command("tsh", "--proxy="+proxy, "kube", "login", cluster)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func showVersion() {
	fmt.Printf("🚀 tkube version %s\n", version)
	fmt.Println("Enhanced Teleport kubectl wrapper with auto-authentication")
	fmt.Println()
	
	// Show config status
	configPath := getConfigPath()
	if _, err := os.Stat(configPath); err == nil {
		config, err := loadConfig()
		if err == nil {
			fmt.Printf("📍 Config: %s (%d environments)\n", configPath, len(config.Environments))
		} else {
			fmt.Printf("📍 Config: %s (error loading)\n", configPath)
		}
	} else {
		fmt.Printf("📍 Config: %s (not created yet)\n", configPath)
	}
	fmt.Println()
	
	fmt.Println("🔧 Dependencies:")
	
	// Check tsh (required)
	if cmd := exec.Command("tsh", "version", "--client"); cmd.Run() == nil {
		if output, err := cmd.Output(); err == nil {
			lines := strings.Split(string(output), "\n")
			if len(lines) > 0 {
				versionLine := strings.TrimSpace(lines[0])
				fmt.Printf("  ✅ tsh (required): %s\n", versionLine)
			}
		} else {
			fmt.Println("  ✅ tsh (required): installed")
		}
	} else {
		fmt.Println("  ❌ tsh (required): not found")
		fmt.Println("     Install from: https://goteleport.com/docs/installation/")
	}
	
	// Check kubectl (optional)
	if cmd := exec.Command("kubectl", "version", "--client", "--short"); cmd.Run() == nil {
		if output, err := cmd.Output(); err == nil {
			versionStr := strings.TrimSpace(string(output))
			if strings.Contains(versionStr, "Client Version:") {
				parts := strings.Fields(versionStr)
				if len(parts) >= 3 {
					fmt.Printf("  ✅ kubectl (optional): %s\n", parts[2])
				}
			} else {
				fmt.Println("  ✅ kubectl (optional): installed")
			}
		} else {
			fmt.Println("  ✅ kubectl (optional): installed")
		}
	} else {
		fmt.Println("  ⚠️  kubectl (optional): not found")
		fmt.Println("     Install from: https://kubernetes.io/docs/tasks/tools/")
	}
	
	fmt.Println()
	fmt.Println("💡 Quick start:")
	fmt.Println("   tkube status          # Check your configuration")
	fmt.Println("   tkube <env> <cluster> # Connect to a cluster")
	fmt.Println("   tkube completion zsh  # Enable tab completion")
}

func showStatus() {
	config, err := loadConfig()
	if err != nil {
		fmt.Printf("❌ Error loading configuration: %v\n", err)
		fmt.Println()
		fmt.Println("💡 Tip: Run 'tkube config path' to see the config file location")
		return
	}

	if len(config.Environments) == 0 {
		fmt.Println("❌ No environments configured")
		fmt.Println()
		fmt.Println("📝 Configure environments in your config file:")
		fmt.Printf("   %s\n", getConfigPath())
		fmt.Println()
		fmt.Println("Example configuration:")
		fmt.Println("  {")
		fmt.Println("    \"environments\": {")
		fmt.Println("      \"prod\": \"teleport.prod.company.com:443\",")
		fmt.Println("      \"test\": \"teleport.test.company.com:443\"")
		fmt.Println("    },")
		fmt.Println("    \"auto_login\": true")
		fmt.Println("  }")
		fmt.Println()
		fmt.Println("💡 Tip: Run 'tkube config show' to see your current configuration")
		return
	}

	fmt.Println("🌍 Available environments and authentication status:")
	fmt.Println()

	for env, proxy := range config.Environments {
		if isAuthenticated(proxy) {
			fmt.Printf("  \033[32m✅ %s → %s (authenticated)\033[0m\n", env, proxy)
		} else {
			fmt.Printf("  \033[31m❌ %s → %s (not authenticated)\033[0m\n", env, proxy)
		}
	}

	fmt.Println()
	if config.AutoLogin {
		fmt.Println("🔐 Auto-login: \033[32menabled\033[0m")
	} else {
		fmt.Println("🔐 Auto-login: \033[33mdisabled\033[0m")
	}
	fmt.Println()
	fmt.Println("💡 Usage: tkube <environment> <cluster>")
	fmt.Println("💡 Tab completion: tkube <TAB> to see environments, tkube prod <TAB> to see clusters")
}

func showConfig() {
	config, err := loadConfig()
	if err != nil {
		fmt.Printf("❌ Error loading configuration: %v\n", err)
		return
	}

	fmt.Printf("📍 Configuration file: %s\n", getConfigPath())
	fmt.Println()

	// Pretty print the configuration
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		fmt.Printf("❌ Error formatting configuration: %v\n", err)
		return
	}

	fmt.Println("📄 Current configuration:")
	fmt.Println(string(data))
	fmt.Println()
	
	if len(config.Environments) > 0 {
		fmt.Printf("🌍 Found %d environment(s): %s\n", 
			len(config.Environments), 
			strings.Join(getEnvironments(), ", "))
	}
}

func showConfigPath() {
	fmt.Printf("📍 Configuration file location:\n")
	fmt.Printf("   %s\n", getConfigPath())
	fmt.Println()
	
	if _, err := os.Stat(getConfigPath()); os.IsNotExist(err) {
		fmt.Println("⚠️  Configuration file does not exist yet.")
		fmt.Println("💡 It will be created automatically when you run tkube for the first time.")
	} else {
		fmt.Println("✅ Configuration file exists.")
		fmt.Println("💡 Run 'tkube config show' to see its contents.")
	}
}

func getConfigPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".tkube", "config.json")
}