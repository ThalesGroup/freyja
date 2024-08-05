package shellcli

import (
	"github.com/spf13/cobra"
	"log"
)

// commands definitions
var machineStartCmd = &cobra.Command{
	Use:              "start",
	Short:            "Machine startup",
	Long:             "Machine startup using libvirt",
	TraverseChildren: true, // ensure local flags do not spread to sub commands

	Run: func(cmd *cobra.Command, args []string) {
		log.Println("Start machine")
	},
}
