package shellcli

import (
	"github.com/spf13/cobra"
)

// MachineCmd
// commands definitions
var MachineCmd = &cobra.Command{
	Use:              "machine",
	Short:            "Virtual machine management",
	Long:             "Virtual machine management to create, start, stop or delete machines.",
	TraverseChildren: true, // ensure local flags do not spread to sub commands

	Run: func(cmd *cobra.Command, args []string) {},
}

func init() {
	MachineCmd.AddCommand(machineCreateCmd)
	MachineCmd.AddCommand(machineStartCmd)
	MachineCmd.AddCommand(machineStopCmd)
	MachineCmd.AddCommand(machineRestartCmd)
	MachineCmd.AddCommand(machineDeleteCmd)
	MachineCmd.AddCommand(machineListCmd)
	MachineCmd.AddCommand(machineInfoCmd)
}
