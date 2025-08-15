package main

import (
	"os"
	"tkube/internal/commands"
	"tkube/internal/config"
	"tkube/internal/kubectl"
	"tkube/internal/shell"
	"tkube/internal/teleport"

	"github.com/spf13/cobra"
)

var version = "1.2.0" // Set by build process

func main() {
	// Initialize dependencies
	configManager, err := config.NewManager()
	if err != nil {
		os.Exit(1)
	}

	teleportClient, err := teleport.NewClient(configManager)
	if err != nil {
		os.Exit(1)
	}
	kubectlClient := kubectl.NewClient()
	installer, err := teleport.NewTSHInstaller()
	if err != nil {
		os.Exit(1)
	}
	shellProvider := shell.NewProvider(configManager, teleportClient)
	commandHandler := commands.NewHandler(configManager, teleportClient, kubectlClient, installer)

	// Create root command
	rootCmd := &cobra.Command{
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
				return shellProvider.GetEnvironments(), cobra.ShellCompDirectiveNoFileComp
			}
			if len(args) == 1 {
				// Complete clusters for the given environment with prefix filtering
				return shellProvider.GetClustersWithPrefix(args[0], toComplete), cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return commandHandler.ConnectToCluster(args[0], args[1])
		},
	}

	// Create version command
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Show version and dependency information",
		Long: `Display tkube version information along with the status of required 
and optional dependencies like tsh (Teleport CLI) and kubectl.

This command helps verify your installation and troubleshoot any 
missing dependencies.`,
		Run: func(cmd *cobra.Command, args []string) {
			commandHandler.ShowVersion()
		},
	}

	// Create status command
	statusCmd := &cobra.Command{
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
			commandHandler.ShowStatus()
		},
	}

	// Create config command
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage tkube configuration",
		Long: `Manage your tkube configuration file (~/.tkube/config.json).

The configuration file stores your Teleport environments and settings.
It's automatically created with example values on first run.`,
	}

	// Create config subcommands
	configShowCmd := &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		Long: `Display the current tkube configuration including all environments,
their Teleport proxy addresses, and settings like auto-login.`,
		Run: func(cmd *cobra.Command, args []string) {
			commandHandler.ShowConfig()
		},
	}

	configPathCmd := &cobra.Command{
		Use:   "path",
		Short: "Show configuration file path",
		Long:  `Display the path to your tkube configuration file.`,
		Run: func(cmd *cobra.Command, args []string) {
			commandHandler.ShowConfigPath()
		},
	}

	installTSHCmd := &cobra.Command{
		Use:   "install-tsh [version]",
		Short: "Install a specific version of tsh",
		Long: `Install a specific version of tsh for use with tkube.

This command downloads and installs the specified tsh version to ~/.tkube/tsh/[version]/.
You can then configure environments to use specific tsh versions in your config.json.

Example configuration:
  {
    "environments": {
      "prod": {
        "proxy": "teleport.prod.company.com:443",
        "tsh_version": "16.4.0"
      },
      "test": {
        "proxy": "teleport.test.company.com:443",
        "tsh_version": "17.7.1"
      }
    }
  }`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return commandHandler.InstallTSH(args[0])
		},
	}

	tshVersionsCmd := &cobra.Command{
		Use:   "tsh-versions",
		Short: "List installed tsh versions",
		Long: `List all installed tsh versions and show which environments use them.

This command helps you:
  • See which tsh versions are installed
  • Check which environments are configured to use specific versions
  • Verify version compatibility`,
		Run: func(cmd *cobra.Command, args []string) {
			commandHandler.ShowTSHVersions()
		},
	}

	autoDetectVersionsCmd := &cobra.Command{
		Use:   "auto-detect-versions",
		Short: "Automatically detect and update tsh versions for all environments",
		Long: `Automatically detect the required tsh version for each environment by:
  • Querying Teleport servers for version information
  • Extracting version from proxy hostnames
  • Checking environment variables
  • Updating configuration with detected versions

This command is useful for:
  • Initial setup of new environments
  • Keeping versions up to date
  • Troubleshooting version compatibility issues`,
		Run: func(cmd *cobra.Command, args []string) {
			commandHandler.AutoDetectVersions()
		},
	}

	// Add subcommands to config (only configuration-related commands)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configPathCmd)

	// Add commands to root
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(tshVersionsCmd)
	rootCmd.AddCommand(installTSHCmd)
	rootCmd.AddCommand(autoDetectVersionsCmd)

	// Enable completion command
	rootCmd.CompletionOptions.DisableDefaultCmd = false

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

	// Execute root command
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
