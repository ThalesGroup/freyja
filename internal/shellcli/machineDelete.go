package shellcli

import (
	"github.com/spf13/cobra"
	"log"
)

// commands definitions
var machineDeleteCmd = &cobra.Command{
	Use:              "delete",
	Short:            "Machine deletion",
	Long:             "Machine deletion using libvirt",
	TraverseChildren: true, // ensure local flags do not spread to sub commands

	Run: func(cmd *cobra.Command, args []string) {
		log.Println("Delete machine")
	},
}
