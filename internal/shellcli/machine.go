package shellcli

import (
	"github.com/spf13/cobra"
)

// commands definitions
var machineCmd = &cobra.Command{
	Use:              "machine",
	Short:            "Virtual machine management",
	Long:             "Virtual machine management to create, start, stop or delete machines.",
	TraverseChildren: true, // ensure local flags do not spread to sub commands

	Run: func(cmd *cobra.Command, args []string) {},
}

func init() {
	machineCmd.AddCommand(machineCreateCmd)
	machineCmd.AddCommand(machineStartCmd)
	machineCmd.AddCommand(machineStopCmd)
	machineCmd.AddCommand(machineRestartCmd)
	machineCmd.AddCommand(machineDeleteCmd)
	machineCmd.AddCommand(machineListCmd)
}
