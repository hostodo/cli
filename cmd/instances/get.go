package instances

import (
	"fmt"

	"github.com/hostodo/hostodo-cli/pkg/api"
	"github.com/hostodo/hostodo-cli/pkg/auth"
	"github.com/hostodo/hostodo-cli/pkg/config"
	"github.com/hostodo/hostodo-cli/pkg/ui"
	"github.com/spf13/cobra"
)

var getJSONOutput bool

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get <instance-id>",
	Short: "Get details about a specific instance",
	Long: `Get detailed information about a specific VPS instance.

Displays comprehensive information including:
  • Basic information (ID, hostname, status)
  • Network configuration (IPs, MAC address)
  • Resource allocation (RAM, CPU, Disk, Bandwidth)
  • Plan and template details
  • Billing information
  • Timeline (created, updated)

Examples:
  hostodo instances get abc123
  hostodo instances get abc123 --json`,
	Args: cobra.ExactArgs(1),
	Run:  runGet,
}

func init() {
	getCmd.Flags().BoolVar(&getJSONOutput, "json", false, "Output as JSON")
}

func runGet(cmd *cobra.Command, args []string) {
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

	// Fetch instance
	instance, err := client.GetInstance(instanceID)
	if err != nil {
		exitWithError("Failed to fetch instance: %v", err)
	}

	// Get power status
	powerStatus, err := client.GetInstancePowerStatus(instanceID)
	if err == nil {
		instance.PowerStatus = powerStatus
	}

	// Display
	if getJSONOutput {
		output, err := ui.FormatInstancesJSON([]api.Instance{*instance})
		if err != nil {
			exitWithError("Failed to format JSON: %v", err)
		}
		fmt.Println(output)
	} else {
		fmt.Println(ui.FormatInstanceDetail(instance))
	}
}
