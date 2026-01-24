package cmd

import (
	"fmt"
	"os"

	"github.com/hostodo/hostodo-cli/cmd/auth"
	"github.com/hostodo/hostodo-cli/cmd/instances"
	"github.com/spf13/cobra"
)

var (
	// Version information (will be set during build)
	Version = "0.2.0"
	cfgFile string
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "hostodo",
	Short: "Hostodo CLI - Manage your VPS instances from the command line",
	Long: `Hostodo CLI is a beautiful, interactive command-line interface for managing
your Hostodo VPS instances.

Features:
  - Interactive TUI with Bubble Tea
  - List and manage instances
  - Control instance power (start/stop/reboot)
  - Multiple output formats (interactive, JSON, simple table)
  - Secure credential storage in system keychain

Authentication:
  hostodo login                    # Authenticate with your account
  hostodo logout                   # Sign out
  hostodo whoami                   # Show current user

Instance Management:
  hostodo instances list           # List all your instances
  hostodo instances get <id>       # Get details about an instance
  hostodo instances start <id>     # Start an instance
  hostodo instances stop <id>      # Stop an instance
  hostodo instances reboot <id>    # Reboot an instance`,
	Version: Version,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Add subcommands
	rootCmd.AddCommand(auth.AuthCmd)
	rootCmd.AddCommand(instances.InstancesCmd)

	// Root-level aliases for common auth commands
	rootCmd.AddCommand(loginAliasCmd)
	rootCmd.AddCommand(logoutAliasCmd)
	rootCmd.AddCommand(whoamiAliasCmd)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.hostodo/config.json)")
	rootCmd.PersistentFlags().String("api-url", "", "API URL (default is https://console.hostodo.com or $HOSTODO_API_URL)")
}

// loginAliasCmd is a convenience alias for 'auth login'
var loginAliasCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Hostodo (alias for 'auth login')",
	Long: `Authenticate with your Hostodo account using device flow.

This is a convenience alias for 'hostodo auth login'.

Example:
  hostodo login`,
	Run: func(cmd *cobra.Command, args []string) {
		auth.AuthCmd.SetArgs([]string{"login"})
		auth.AuthCmd.Execute()
	},
}

// logoutAliasCmd is a convenience alias for 'auth logout'
var logoutAliasCmd = &cobra.Command{
	Use:   "logout",
	Short: "Sign out from Hostodo (alias for 'auth logout')",
	Long: `Sign out from your Hostodo account.

This is a convenience alias for 'hostodo auth logout'.

Example:
  hostodo logout`,
	Run: func(cmd *cobra.Command, args []string) {
		auth.AuthCmd.SetArgs([]string{"logout"})
		auth.AuthCmd.Execute()
	},
}

// whoamiAliasCmd is a convenience alias for 'auth whoami'
var whoamiAliasCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Display current logged-in user (alias for 'auth whoami')",
	Long: `Show information about the currently authenticated user.

This is a convenience alias for 'hostodo auth whoami'.

Example:
  hostodo whoami`,
	Run: func(cmd *cobra.Command, args []string) {
		auth.AuthCmd.SetArgs([]string{"whoami"})
		auth.AuthCmd.Execute()
	},
}

func initConfig() {
	// Configuration is loaded on-demand by each command
	// This allows flexibility for different authentication states
}

// checkAuth verifies that the user is authenticated
func checkAuth() error {
	// This will be called by commands that require authentication
	// Implemented in each command as needed
	return nil
}

func exitWithError(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: "+msg+"\n", args...)
	os.Exit(1)
}
