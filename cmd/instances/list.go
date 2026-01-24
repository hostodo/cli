package instances

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hostodo/hostodo-cli/pkg/api"
	"github.com/hostodo/hostodo-cli/pkg/auth"
	"github.com/hostodo/hostodo-cli/pkg/config"
	"github.com/hostodo/hostodo-cli/pkg/ui"
	"github.com/spf13/cobra"
)

var (
	jsonOutput    bool
	simpleOutput  bool
	detailsOutput bool
	limit         int
	offset        int
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all your instances",
	Long: `List all your VPS instances with various output formats.

Output Formats:
  • Interactive TUI (default) - Beautiful, scrollable table with keyboard navigation
  • JSON (--json)              - JSON format for scripting and automation
  • Simple (--simple)          - Static ASCII table for quick viewing
  • Details (--details)        - Detailed view with all information

Examples:
  hostodo instances list                    # Interactive TUI
  hostodo instances list --json             # JSON output
  hostodo instances list --simple           # Simple table
  hostodo instances list --details          # Detailed view
  hostodo instances list --limit 50         # Show 50 instances`,
	Run: runList,
}

func init() {
	listCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	listCmd.Flags().BoolVar(&simpleOutput, "simple", false, "Output as simple table")
	listCmd.Flags().BoolVar(&detailsOutput, "details", false, "Show detailed information")
	listCmd.Flags().IntVar(&limit, "limit", 100, "Maximum number of instances to fetch")
	listCmd.Flags().IntVar(&offset, "offset", 0, "Offset for pagination")
}

func runList(cmd *cobra.Command, args []string) {
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

	// Fetch instances
	instancesResp, err := client.ListInstances(limit, offset)
	if err != nil {
		exitWithError("Failed to fetch instances: %v", err)
	}

	if len(instancesResp.Results) == 0 {
		fmt.Println("No instances found.")
		fmt.Println("\nYou don't have any VPS instances yet.")
		fmt.Println("Visit https://console.hostodo.com to deploy your first instance!")
		return
	}

	// Display based on output format
	if jsonOutput {
		// JSON output
		output, err := ui.FormatInstancesJSON(instancesResp.Results)
		if err != nil {
			exitWithError("Failed to format JSON: %v", err)
		}
		fmt.Println(output)
	} else if simpleOutput {
		// Simple table output
		fmt.Println() // Add spacing
		output := ui.FormatInstancesSimpleTable(instancesResp.Results)
		fmt.Println(output)
		fmt.Printf("\nTotal: %d instances\n", instancesResp.Count)
	} else if detailsOutput {
		// Detailed output
		fmt.Println() // Add spacing
		output := ui.FormatInstancesDetailedTable(instancesResp.Results)
		fmt.Println(output)
		fmt.Printf("\nTotal: %d instances\n", instancesResp.Count)
	} else {
		// Interactive TUI (default)
		p := tea.NewProgram(ui.NewTableModel(instancesResp.Results))
		if _, err := p.Run(); err != nil {
			exitWithError("Failed to run interactive table: %v", err)
		}
	}
}

func exitWithError(msg string, args ...interface{}) {
	fmt.Printf("Error: "+msg+"\n", args...)
	return
}
