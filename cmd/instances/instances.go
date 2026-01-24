package instances

import (
	"github.com/spf13/cobra"
)

// InstancesCmd represents the instances command
var InstancesCmd = &cobra.Command{
	Use:   "instances",
	Short: "Manage your VPS instances",
	Long: `Manage your Hostodo VPS instances.

Available commands:
  list        List all instances
  get         Get details about a specific instance
  start       Start a stopped instance
  stop        Stop a running instance
  reboot      Reboot an instance

Examples:
  hostodo instances list
  hostodo instances list --json
  hostodo instances list --simple
  hostodo instances list --details
  hostodo instances get <instance-id>
  hostodo instances start <instance-id>
  hostodo instances stop <instance-id>
  hostodo instances reboot <instance-id>`,
}

func init() {
	// Add subcommands
	InstancesCmd.AddCommand(listCmd)
	InstancesCmd.AddCommand(getCmd)
	InstancesCmd.AddCommand(startCmd)
	InstancesCmd.AddCommand(stopCmd)
	InstancesCmd.AddCommand(rebootCmd)
}
