package instances

import (
	"fmt"
	"time"

	"github.com/hostodo/hostodo-cli/pkg/api"
	"github.com/hostodo/hostodo-cli/pkg/auth"
	"github.com/hostodo/hostodo-cli/pkg/config"
	"github.com/spf13/cobra"
)

var forceStop bool

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop <instance-id>",
	Short: "Stop a running instance",
	Long: `Stop a running VPS instance.

This command will gracefully shut down the instance. Use --force for immediate shutdown.

Examples:
  hostodo instances stop abc123
  hostodo instances stop abc123 --force`,
	Args: cobra.ExactArgs(1),
	Run:  runStop,
}

func init() {
	stopCmd.Flags().BoolVarP(&forceStop, "force", "f", false, "Force immediate shutdown")
}

func runStop(cmd *cobra.Command, args []string) {
	instanceID := args[0]

	// Load config
	cfg, err := config.Load()
	if err != nil {
		exitWithError("Failed to load config: %v", err)
	}

	// Check authentication
	if !auth.IsAuthenticated() {
		exitWithError("You are not logged in. Please run: hostodo login")
	}

	// Create API client
	client, err := api.NewClient(cfg)
	if err != nil {
		exitWithError("Failed to create API client: %v", err)
	}

	// Stop instance
	if forceStop {
		fmt.Printf("Force stopping instance %s...\n", instanceID)
	} else {
		fmt.Printf("Stopping instance %s...\n", instanceID)
	}

	err = client.StopInstance(instanceID)
	if err != nil {
		exitWithError("Failed to stop instance: %v", err)
	}

	fmt.Println("✓ Instance stop command sent successfully")
	fmt.Println("  The instance is now shutting down...")
	fmt.Print("\nWaiting for instance to stop")

	// Poll for status (up to 60 seconds)
	for i := 0; i < 60; i++ {
		fmt.Print(".")
		time.Sleep(1 * time.Second)

		status, err := client.GetInstancePowerStatus(instanceID)
		if err == nil && status == "stopped" {
			fmt.Println()
			fmt.Println("✓ Instance is now stopped")
			return
		}
	}

	fmt.Println()
	fmt.Println("⚠ Instance is stopping (this may take a few moments)")
}
