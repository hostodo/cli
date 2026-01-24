package instances

import (
	"fmt"
	"time"

	"github.com/hostodo/hostodo-cli/pkg/api"
	"github.com/hostodo/hostodo-cli/pkg/auth"
	"github.com/hostodo/hostodo-cli/pkg/config"
	"github.com/spf13/cobra"
)

var forceReboot bool

// rebootCmd represents the reboot command
var rebootCmd = &cobra.Command{
	Use:   "reboot <instance-id>",
	Short: "Reboot an instance",
	Long: `Reboot a VPS instance.

This command will gracefully restart the instance. Use --force for immediate restart.

Examples:
  hostodo instances reboot abc123
  hostodo instances reboot abc123 --force`,
	Args: cobra.ExactArgs(1),
	Run:  runReboot,
}

func init() {
	rebootCmd.Flags().BoolVarP(&forceReboot, "force", "f", false, "Force immediate reboot")
}

func runReboot(cmd *cobra.Command, args []string) {
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

	// Reboot instance
	if forceReboot {
		fmt.Printf("Force rebooting instance %s...\n", instanceID)
	} else {
		fmt.Printf("Rebooting instance %s...\n", instanceID)
	}

	err = client.RebootInstance(instanceID)
	if err != nil {
		exitWithError("Failed to reboot instance: %v", err)
	}

	fmt.Println("✓ Instance reboot command sent successfully")
	fmt.Println("  The instance is now rebooting...")
	fmt.Print("\nWaiting for instance to restart")

	// Poll for status (up to 90 seconds)
	sawStopped := false
	for i := 0; i < 90; i++ {
		fmt.Print(".")
		time.Sleep(1 * time.Second)

		status, err := client.GetInstancePowerStatus(instanceID)
		if err == nil {
			if status == "stopped" {
				sawStopped = true
			} else if status == "running" && sawStopped {
				fmt.Println()
				fmt.Println("✓ Instance has rebooted and is now running")
				return
			}
		}
	}

	fmt.Println()
	fmt.Println("⚠ Instance is rebooting (this may take a few moments)")
}
