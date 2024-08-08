package shellcli

import (
	"github.com/spf13/cobra"
	"log"
)

// commands definitions
var networkDeleteCmd = &cobra.Command{
	Use:              "delete",
	Short:            "Virtual network deletion",
	Long:             "Virtual network deletion using handler",
	TraverseChildren: true, // ensure local flags do not spread to sub commands

	Run: func(cmd *cobra.Command, args []string) {
		log.Println("Delete network")
	},
}
