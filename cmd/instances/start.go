package instances

import (
	"fmt"
	"time"

	"github.com/hostodo/hostodo-cli/pkg/api"
	"github.com/hostodo/hostodo-cli/pkg/auth"
	"github.com/hostodo/hostodo-cli/pkg/config"
	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start <instance-id>",
	Short: "Start a stopped instance",
	Long: `Start a stopped VPS instance.

This command will power on the instance. The instance must be in a stopped state.

Examples:
  hostodo instances start abc123`,
	Args: cobra.ExactArgs(1),
	Run:  runStart,
}

func runStart(cmd *cobra.Command, args []string) {
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

	// Start instance
	fmt.Printf("Starting instance %s...\n", instanceID)
	err = client.StartInstance(instanceID)
	if err != nil {
		exitWithError("Failed to start instance: %v", err)
	}

	fmt.Println("✓ Instance start command sent successfully")
	fmt.Println("  The instance is now booting up...")
	fmt.Print("\nWaiting for instance to start")

	// Poll for status (up to 30 seconds)
	for i := 0; i < 30; i++ {
		fmt.Print(".")
		time.Sleep(1 * time.Second)

		status, err := client.GetInstancePowerStatus(instanceID)
		if err == nil && status == "running" {
			fmt.Println()
			fmt.Println("✓ Instance is now running")
			return
		}
	}

	fmt.Println()
	fmt.Println("⚠ Instance is starting (this may take a few moments)")
}
