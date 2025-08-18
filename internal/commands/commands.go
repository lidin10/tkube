package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"tkube/internal/config"
	"tkube/internal/kubectl"
	"tkube/internal/teleport"
)

// Handler handles all tkube commands
type Handler struct {
	configManager  *config.Manager
	teleportClient *teleport.Client
	kubectlClient  *kubectl.Client
	installer      *teleport.TSHInstaller
}

// NewHandler creates a new command handler
func NewHandler(configManager *config.Manager, teleportClient *teleport.Client, kubectlClient *kubectl.Client, installer *teleport.TSHInstaller) *Handler {
	return &Handler{
		configManager:  configManager,
		teleportClient: teleportClient,
		kubectlClient:  kubectlClient,
		installer:      installer,
	}
}

// ConnectToCluster connects to a Kubernetes cluster via Teleport
func (h *Handler) ConnectToCluster(env, cluster string) error {
	config, err := h.configManager.Load()
	if err != nil {
		fmt.Printf("❌ Error loading configuration: %v\n", err)
		fmt.Println("💡 Run 'tkube config path' to see the expected config location")
		return err
	}

	envConfig, exists := config.Environments[env]
	if !exists {
		fmt.Printf("❌ Unknown environment '%s'\n", env)
		fmt.Println()
		fmt.Printf("Available environments: %s\n", strings.Join(h.getEnvironments(), ", "))
		fmt.Println()
		fmt.Println("💡 Tips:")
		fmt.Println("   • Run 'tkube status' to see all configured environments")
		fmt.Println("   • Run 'tkube config show' to see your configuration")
		fmt.Println("   • Use tab completion: tkube <TAB>")
		return fmt.Errorf("unknown environment")
	}

	// Auto-detect tsh version if not set
	if envConfig.TSHVersion == "" {
		versionDetector := teleport.NewVersionDetector()
		version, err := versionDetector.DetectTSHVersion(envConfig.Proxy)
		if err == nil && version != "" {
			// Update configuration with detected version
			if err := h.configManager.UpdateEnvironmentTSHVersion(env, version); err == nil {
				envConfig.TSHVersion = version
				fmt.Printf("🔍 Auto-detected tsh version %s\n", version)
			}
		} else {
			fmt.Printf("⚠️  Could not auto-detect tsh version: %v\n", err)
			fmt.Println("💡 You can manually set the version in your config file")
		}
	}

	// Check if required tsh version is installed
	if envConfig.TSHVersion != "" {
		if !h.installer.IsVersionInstalled(envConfig.TSHVersion) {
			fmt.Printf("📦 tsh v%s not installed - installing...\n", envConfig.TSHVersion)

			// Ask user if they want to install automatically
			if h.promptForInstallation(envConfig.TSHVersion) {
				if err := h.installer.InstallTSH(envConfig.TSHVersion); err != nil {
					fmt.Printf("❌ Installation failed: %v\n", err)
					fmt.Printf("💡 Try: tkube install-tsh %s\n", envConfig.TSHVersion)
					return fmt.Errorf("installation failed")
				}
				fmt.Printf("✅ tsh v%s installed\n", envConfig.TSHVersion)

				// Give filesystem time to sync and verify installation
				time.Sleep(100 * time.Millisecond)
				if !h.installer.IsVersionInstalled(envConfig.TSHVersion) {
					fmt.Printf("⚠️  Installation completed but verification failed\n")
					fmt.Printf("💡 Try running the command again\n")
					return fmt.Errorf("installation verification failed")
				}
			} else {
				fmt.Printf("💡 Run: tkube install-tsh %s\n", envConfig.TSHVersion)
				return fmt.Errorf("required tsh version %s not installed", envConfig.TSHVersion)
			}
		}
	}

	// Check authentication status
	if !h.teleportClient.IsAuthenticatedWithEnv(env, envConfig.Proxy) {
		if config.AutoLogin {
			fmt.Printf("🔐 Authenticating to %s...\n", envConfig.Proxy)
			if err := h.teleportClient.LoginWithEnv(env, envConfig.Proxy); err != nil {
				fmt.Printf("❌ Authentication failed\n")
				fmt.Printf("💡 Try: tsh login --proxy=%s\n", envConfig.Proxy)
				return err
			}
		} else {
			fmt.Printf("❌ Not authenticated to %s\n", envConfig.Proxy)
			fmt.Printf("💡 Run: tsh login --proxy=%s\n", envConfig.Proxy)
			return fmt.Errorf("authentication required")
		}
	}

	// Connect to Kubernetes cluster
	fmt.Printf("🚀 Connecting to %s/%s...\n", env, cluster)
	if err := h.teleportClient.KubeLoginWithEnv(env, envConfig.Proxy, cluster); err != nil {
		fmt.Printf("❌ Connection failed\n")
		fmt.Printf("💡 Check cluster name with: tkube %s <TAB>\n", env)
		return err
	}

	fmt.Printf("✅ Connected to %s/%s\n", env, cluster)
	return nil
}

// ShowVersion displays version information
func (h *Handler) ShowVersion(version string) {
	fmt.Printf("🚀 tkube version %s\n", version)
	fmt.Println("Enhanced Teleport kubectl wrapper with auto-authentication")
	fmt.Println()

	// Show config status
	configPath := h.configManager.GetPath()
	if _, err := os.Stat(configPath); err == nil {
		config, err := h.configManager.Load()
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

	// Check system tsh (required)
	if cmd := exec.Command("tsh", "version", "--client"); cmd.Run() == nil {
		if output, err := cmd.Output(); err == nil {
			lines := strings.Split(string(output), "\n")
			if len(lines) > 0 {
				versionLine := strings.TrimSpace(lines[0])
				fmt.Printf("  ✅ tsh (system): %s\n", versionLine)
			}
		} else {
			fmt.Println("  ✅ tsh (required): installed")
		}
	} else {
		fmt.Println("  ❌ tsh (system): not found")
		fmt.Println("     Install from: https://goteleport.com/docs/installation/")
	}

	// Check kubectl (optional)
	if h.kubectlClient.IsAvailable() {
		if version, err := h.kubectlClient.CheckVersion(); err == nil {
			fmt.Printf("  ✅ kubectl (optional): %s\n", version)
		} else {
			fmt.Println("  ✅ kubectl (optional): installed")
		}
	} else {
		fmt.Println("  ⚠️  kubectl (optional): not found")
		fmt.Println("     Install from: https://kubernetes.io/docs/tasks/tools/")
	}

	fmt.Println()
	fmt.Println("🔧 Installed tsh versions:")

	// Show all installed tsh versions with paths
	homeDir, err := os.UserHomeDir()
	if err == nil {
		tshBaseDir := filepath.Join(homeDir, ".tkube", "tsh")
		if _, err := os.Stat(tshBaseDir); err == nil {
			// List installed versions
			versions, err := h.teleportClient.GetInstalledTSHVersions()
			if err == nil && len(versions) > 0 {
				for _, version := range versions {
					tshPath := filepath.Join(tshBaseDir, version, "tsh")
					if h.installer.IsVersionInstalled(version) {
						versionInfo := h.installer.GetTSHVersionInfo(tshPath)
						fmt.Printf("  ✅ tsh %s: %s\n", version, tshPath)
						fmt.Printf("      Version: %s\n", versionInfo)
					} else {
						fmt.Printf("  ⚠️  tsh %s: %s (not fully installed)\n", version, tshPath)
					}
				}
			} else {
				fmt.Println("  📁 No custom tsh versions installed")
				fmt.Println("     Use 'tkube install-tsh <version>' to install specific versions")
			}
		} else {
			fmt.Println("  📁 No custom tsh versions directory found")
			fmt.Println("     Use 'tkube install-tsh <version>' to install specific versions")
		}
	}

	fmt.Println()
	fmt.Println("💡 Quick start:")
	fmt.Println("   tkube status              # Check your configuration")
	fmt.Println("   tkube <env> <cluster>     # Connect to a cluster")
	fmt.Println("   tkube completion zsh      # Enable tab completion")
	fmt.Println()
	fmt.Println("🔧 tsh version management:")
	fmt.Println("   tkube tsh-versions        # List installed tsh versions")
	fmt.Println("   tkube install-tsh         # Install specific tsh version")
}

// ShowStatus displays environment status
func (h *Handler) ShowStatus() {
	config, err := h.configManager.Load()
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
		fmt.Printf("   %s\n", h.configManager.GetPath())
		fmt.Println()
		fmt.Println("Example configuration:")
		fmt.Println("  {")
		fmt.Println("    \"environments\": {")
		fmt.Println("      \"prod\": {")
		fmt.Println("        \"proxy\": \"teleport.prod.company.com:443\",")
		fmt.Println("        \"tsh_version\": \"14.0.0\"")
		fmt.Println("      },")
		fmt.Println("      \"test\": {")
		fmt.Println("        \"proxy\": \"teleport.test.company.com:443\",")
		fmt.Println("        \"tsh_version\": \"13.0.0\"")
		fmt.Println("      }")
		fmt.Println("    },")
		fmt.Println("    \"auto_login\": true")
		fmt.Println("  }")
		fmt.Println()
		fmt.Println("💡 Tip: Run 'tkube config show' to see your current configuration")
		return
	}

	fmt.Println("🌍 Available environments and authentication status:")
	fmt.Println()

	for env, envConfig := range config.Environments {
		sessionInfo := h.teleportClient.GetSessionInfo(env, envConfig.Proxy)

		if sessionInfo.IsAuthenticated {
			if sessionInfo.IsExpired {
				fmt.Printf("  \033[33m⏰ %s → %s (expired)\033[0m\n", env, envConfig.Proxy)
			} else if sessionInfo.TimeRemaining != "" {
				// Format time remaining for better readability
				timeStr := h.formatTimeRemaining(sessionInfo.TimeRemaining)

				// Color code based on time remaining
				var color string
				if !strings.Contains(timeStr, "h") {
					// Less than 1 hour - yellow warning
					color = "\033[33m⚠️"
				} else if strings.HasPrefix(timeStr, "1h") || strings.HasPrefix(timeStr, "2h") {
					// 1-2 hours - yellow warning
					color = "\033[33m⚠️"
				} else {
					// More than 2 hours - green
					color = "\033[32m✅"
				}

				fmt.Printf("  %s %s → %s (%s left)\033[0m\n", color, env, envConfig.Proxy, timeStr)
			} else {
				fmt.Printf("  \033[32m✅ %s → %s (authenticated)\033[0m\n", env, envConfig.Proxy)
			}
		} else {
			fmt.Printf("  \033[31m❌ %s → %s (not authenticated)\033[0m\n", env, envConfig.Proxy)
		}

		// Show tsh version information
		if envConfig.TSHVersion != "" {
			if h.installer.IsVersionInstalled(envConfig.TSHVersion) {
				fmt.Printf("      🔧 Using tsh version: %s\n", envConfig.TSHVersion)
			} else {
				fmt.Printf("      ⚠️  Configured tsh version %s is not installed\n", envConfig.TSHVersion)
				fmt.Printf("      💡 Run: tkube install-tsh %s\n", envConfig.TSHVersion)
			}
		} else {
			fmt.Printf("      🔧 Using system tsh\n")
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

// ShowConfig displays the current configuration
func (h *Handler) ShowConfig() {
	config, err := h.configManager.Load()
	if err != nil {
		fmt.Printf("❌ Error loading configuration: %v\n", err)
		return
	}

	fmt.Printf("📍 Configuration file: %s\n", h.configManager.GetPath())
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
		envs, _ := h.configManager.GetEnvironments()
		fmt.Printf("🌍 Found %d environment(s): %s\n",
			len(config.Environments),
			strings.Join(envs, ", "))
	}
}

// ShowConfigPath displays the configuration file path
func (h *Handler) ShowConfigPath() {
	fmt.Printf("📍 Configuration file location:\n")
	fmt.Printf("   %s\n", h.configManager.GetPath())
	fmt.Println()

	if _, err := os.Stat(h.configManager.GetPath()); os.IsNotExist(err) {
		fmt.Println("⚠️  Configuration file does not exist yet.")
		fmt.Println("💡 It will be created automatically when you run tkube for the first time.")
	} else {
		fmt.Println("✅ Configuration file exists.")
		fmt.Println("💡 Run 'tkube config show' to see its contents.")
	}
}

// InstallTSH installs a specific tsh version
func (h *Handler) InstallTSH(version string) error {
	// Check if version is already installed
	if h.installer.IsVersionInstalled(version) {
		fmt.Printf("✅ tsh v%s is already installed\n", version)
		fmt.Println("💡 Use 'tkube tsh-versions' to see all installed versions")
		return nil
	}

	fmt.Printf("📦 Installing tsh v%s...\n", version)

	if err := h.installer.InstallTSH(version); err != nil {
		fmt.Printf("❌ Installation failed: %v\n", err)
		return err
	}

	fmt.Printf("✅ tsh v%s installed successfully\n", version)
	fmt.Println("💡 Use 'tkube tsh-versions' to see all installed versions")
	return nil
}

// AutoInstallTSH automatically installs a specific tsh version
func (h *Handler) AutoInstallTSH(version string) error {
	return h.installer.InstallTSH(version)
}

// ShowTSHVersions displays installed tsh versions
func (h *Handler) ShowTSHVersions() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("❌ Error getting home directory: %v\n", err)
		return
	}

	tshBaseDir := filepath.Join(homeDir, ".tkube", "tsh")

	// Check if tsh directory exists
	if _, err := os.Stat(tshBaseDir); os.IsNotExist(err) {
		fmt.Println("📁 No tsh versions installed yet")
		fmt.Println()
		fmt.Println("💡 To install a tsh version:")
		fmt.Println("   tkube install-tsh <version>")
		fmt.Println()
		fmt.Println("Example:")
		fmt.Println("   tkube install-tsh 14.0.0")
		return
	}

	// Load config to see which environments use which versions
	config, err := h.configManager.Load()
	if err != nil {
		fmt.Printf("⚠️  Warning: Could not load config: %v\n", err)
		fmt.Println()
	}

	fmt.Println("🔧 Installed tsh versions:")
	fmt.Println()

	// List installed versions
	versions, err := h.teleportClient.GetInstalledTSHVersions()
	if err != nil {
		fmt.Printf("❌ Error reading tsh directory: %v\n", err)
		return
	}

	if len(versions) == 0 {
		fmt.Println("   No versions found")
	} else {
		for _, version := range versions {
			tshPath := filepath.Join(tshBaseDir, version, "tsh")
			if h.installer.IsVersionInstalled(version) {
				// Get just the version number from the binary, not the full git info
				cmd := exec.Command(tshPath, "version", "--client")
				output, err := cmd.Output()
				var versionStr string
				if err == nil {
					lines := strings.Split(string(output), "\n")
					if len(lines) > 0 && strings.TrimSpace(lines[0]) != "" {
						// Extract just "Teleport vX.Y.Z" part
						fullVersion := strings.TrimSpace(lines[0])
						if strings.Contains(fullVersion, " git:") {
							versionStr = strings.Split(fullVersion, " git:")[0]
						} else {
							versionStr = fullVersion
						}
					} else {
						versionStr = "Teleport v" + version
					}
				} else {
					versionStr = "Teleport v" + version
				}
				fmt.Printf("   ✅ %s (%s)\n", versionStr, tshPath)
			} else {
				fmt.Printf("   ⚠️  %s: placeholder (not fully installed)\n", version)
			}
		}
	}

	fmt.Println()
	fmt.Println("🌍 Environment version usage:")
	fmt.Println()

	if config != nil {
		for env, envConfig := range config.Environments {
			if envConfig.TSHVersion != "" {
				if h.installer.IsVersionInstalled(envConfig.TSHVersion) {
					fmt.Printf("   ✅ %s → tsh %s\n", env, envConfig.TSHVersion)
				} else {
					fmt.Printf("   ❌ %s → tsh %s (not installed)\n", env, envConfig.TSHVersion)
				}
			} else {
				fmt.Printf("   🔧 %s → system tsh\n", env)
			}
		}
	} else {
		fmt.Println("   Could not load configuration")
	}

	fmt.Println()
	fmt.Println("💡 Commands:")
	fmt.Println("   tkube install-tsh <version>  # Install a specific version")
	fmt.Println("   tkube status                 # Check environment status")
	fmt.Println("   tkube config show            # View configuration")
}

// getEnvironments returns a list of environment names
func (h *Handler) getEnvironments() []string {
	envs, err := h.configManager.GetEnvironments()
	if err != nil {
		return nil
	}
	return envs
}

// AutoDetectVersions automatically detects and updates tsh versions for all environments
func (h *Handler) AutoDetectVersions() {
	fmt.Println("🔍 Auto-detecting tsh versions for all environments...")
	fmt.Println()

	versionDetector := teleport.NewVersionDetector()
	detectedVersions, err := h.configManager.AutoDetectAndUpdateTSHVersions(versionDetector)
	if err != nil {
		fmt.Printf("❌ Error during auto-detection: %v\n", err)
		return
	}

	if len(detectedVersions) == 0 {
		fmt.Println("📁 No new versions detected or all environments already have versions set")
		return
	}

	fmt.Println("✅ Auto-detection completed!")
	fmt.Println()
	fmt.Println("📋 Detected versions:")
	for env, version := range detectedVersions {
		fmt.Printf("  • %s: tsh %s\n", env, version)
	}
	fmt.Println()
	fmt.Println("💡 Next steps:")
	fmt.Println("   • Check if required versions are installed: tkube tsh-versions")
	fmt.Println("   • Install missing versions: tkube install-tsh <version>")
	fmt.Println("   • Connect to clusters: tkube <env> <cluster>")
}

// promptForInstallation asks the user if they want to install the required tsh version
func (h *Handler) promptForInstallation(version string) bool {
	fmt.Printf("🔧 Would you like to automatically install tsh version %s? (Y/n): ", version)

	// Read user input
	var response string
	fmt.Scanln(&response)

	// Normalize response
	response = strings.ToLower(strings.TrimSpace(response))

	// Default to yes if no input or "y", "yes"
	if response == "" || response == "y" || response == "yes" {
		return true
	}

	// Return false for "n", "no", or any other input
	return false
}

// formatTimeRemaining formats time duration for better readability
func (h *Handler) formatTimeRemaining(timeStr string) string {
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

	// For very short times (less than 1 hour), show a warning color
	if !strings.Contains(timeStr, "h") {
		// This is less than an hour, could be just minutes
		return timeStr
	}

	return timeStr
}
// AddEnvironmentInteractive adds a new environment interactively
func (h *Handler) AddEnvironmentInteractive() error {
	// TODO: Implement interactive environment management
	fmt.Println("❌ Interactive environment management not yet implemented")
	fmt.Println("💡 Please edit your config file manually: tkube config path")
	return fmt.Errorf("interactive management not implemented")
}

// EditEnvironmentInteractive edits an existing environment interactively
func (h *Handler) EditEnvironmentInteractive(name string) error {
	// TODO: Implement interactive environment management
	fmt.Println("❌ Interactive environment management not yet implemented")
	fmt.Println("💡 Please edit your config file manually: tkube config path")
	return fmt.Errorf("interactive management not implemented")
}

// RemoveEnvironmentInteractive removes an environment interactively
func (h *Handler) RemoveEnvironmentInteractive(name string) error {
	// TODO: Implement interactive environment management
	fmt.Println("❌ Interactive environment management not yet implemented")
	fmt.Println("💡 Please edit your config file manually: tkube config path")
	return fmt.Errorf("interactive management not implemented")
}

// Logout logs out from Teleport environments
func (h *Handler) Logout(env string) error {
	config, err := h.configManager.Load()
	if err != nil {
		fmt.Printf("❌ Error loading configuration: %v\n", err)
		return err
	}

	if env == "" {
		// Logout from all environments
		fmt.Println("🔓 Logging out from all environments...")
		
		for envName, envConfig := range config.Environments {
			fmt.Printf("🔓 Logging out from %s (%s)...\n", envName, envConfig.Proxy)
			if err := h.teleportClient.LogoutWithEnv(envName, envConfig.Proxy); err != nil {
				fmt.Printf("⚠️  Failed to logout from %s: %v\n", envName, err)
			} else {
				fmt.Printf("✅ Logged out from %s\n", envName)
			}
		}
		
		fmt.Println("✅ Logout from all environments completed")
		return nil
	}

	// Logout from specific environment
	envConfig, exists := config.Environments[env]
	if !exists {
		fmt.Printf("❌ Unknown environment '%s'\n", env)
		fmt.Printf("Available environments: %s\n", strings.Join(h.getEnvironments(), ", "))
		return fmt.Errorf("unknown environment")
	}

	fmt.Printf("🔓 Logging out from %s (%s)...\n", env, envConfig.Proxy)
	if err := h.teleportClient.LogoutWithEnv(env, envConfig.Proxy); err != nil {
		fmt.Printf("❌ Failed to logout from %s: %v\n", env, err)
		return err
	}

	fmt.Printf("✅ Logged out from %s\n", env)
	return nil
}

// ValidateConfiguration validates the current configuration
func (h *Handler) ValidateConfiguration() error {
	// TODO: Implement configuration validation
	fmt.Println("❌ Configuration validation not yet implemented")
	fmt.Println("💡 Please check your config file manually: tkube config show")
	return fmt.Errorf("validation not implemented")
}