package shellcli

import (
	"github.com/spf13/cobra"
	"log"
)

// commands definitions
var networkCreateCmd = &cobra.Command{
	Use:              "create",
	Short:            "Virtual network creation",
	Long:             "Virtual network creation using libvirt",
	TraverseChildren: true, // ensure local flags do not spread to sub commands

	Run: func(cmd *cobra.Command, args []string) {
		log.Println("Create network")
	},
}
