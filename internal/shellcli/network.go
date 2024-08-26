package shellcli

import (
	"github.com/spf13/cobra"
)

// NetworkCmd
// commands definitions
var networkCmd = &cobra.Command{
	Use:              "network",
	Short:            "Virtual network management",
	Long:             "Virtual network management to create, start, stop or delete networks.",
	TraverseChildren: true, // ensure local flags do not spread to sub commands

	Run: func(cmd *cobra.Command, args []string) {},
}

func init() {
	networkCmd.AddCommand(networkCreateCmd)
	networkCmd.AddCommand(networkListCmd)
	networkCmd.AddCommand(networkDeleteCmd)
	networkCmd.AddCommand(networkInfoCmd)
}
