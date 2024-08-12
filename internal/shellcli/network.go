package shellcli

import (
	"github.com/spf13/cobra"
)

// NetworkCmd
// commands definitions
var NetworkCmd = &cobra.Command{
	Use:              "network",
	Short:            "Virtual network management",
	Long:             "Virtual network management to create, start, stop or delete networks.",
	TraverseChildren: true, // ensure local flags do not spread to sub commands

	Run: func(cmd *cobra.Command, args []string) {},
}

func init() {
	NetworkCmd.AddCommand(networkCreateCmd)
	NetworkCmd.AddCommand(networkListCmd)
	NetworkCmd.AddCommand(networkDeleteCmd)
}
