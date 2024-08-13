package shellcli

import (
	"github.com/spf13/cobra"
	"log"
)

var configuration string

// commands definitions
var machineCreateCmd = &cobra.Command{
	Use:              "create",
	Short:            "Machine creation",
	Long:             "Machine creation using handler and cloud-init or ignition",
	TraverseChildren: true, // ensure local flags do not spread to sub commands

	Run: func(cmd *cobra.Command, args []string) {},
}

func init() {
	machineCreateCmd.Flags().StringVarP(&startDomainName, "config", "c", "", "Path to the configuration file to create the machines and the networks.")
	if err := machineCreateCmd.MarkFlagRequired("config"); err != nil {
		log.Panic(err.Error())
	}
}

func machineCreate(configuration string) {
	// TODO create a metadata to know which domain uses what network
}
