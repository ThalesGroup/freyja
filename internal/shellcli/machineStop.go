package shellcli

import (
	"github.com/spf13/cobra"
	"log"
)

// commands definitions
var machineStopCmd = &cobra.Command{
	Use:              "stop",
	Short:            "Machine shutdown",
	Long:             "Machine shutdown using handler",
	TraverseChildren: true, // ensure local flags do not spread to sub commands

	Run: func(cmd *cobra.Command, args []string) {
		log.Println("Shutdown machine")
	},
}
