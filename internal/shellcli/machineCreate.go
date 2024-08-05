package shellcli

import (
	"github.com/spf13/cobra"
	"log"
)

// commands definitions
var machineCreateCmd = &cobra.Command{
	Use:              "create",
	Short:            "Machine creation",
	Long:             "Machine creation using libvirt and cloud-init or ignition",
	TraverseChildren: true, // ensure local flags do not spread to sub commands

	Run: func(cmd *cobra.Command, args []string) {
		log.Println("Create machine")
	},
}
