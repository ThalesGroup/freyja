package shellcli

import (
	"github.com/spf13/cobra"
	"log"
)

// commands definitions
var machineRestartCmd = &cobra.Command{
	Use:              "restart",
	Short:            "Machine re-initialisation",
	Long:             "Machine re-initialisation using handler",
	TraverseChildren: true, // ensure local flags do not spread to sub commands

	Run: func(cmd *cobra.Command, args []string) {
		log.Println("Delete machine")
	},
}
