package shellcli

import (
	"github.com/spf13/cobra"
	"log"
	"os"
	"strings"
)

var stopDomainName string

// commands definitions
var machineStopCmd = &cobra.Command{
	Use:              "stop",
	Short:            "Machine shutdown",
	Long:             "Machine shutdown using handler",
	TraverseChildren: true, // ensure local flags do not spread to sub commands

	Run: func(cmd *cobra.Command, args []string) {
		domain, _ := LibvirtConnexion.DomainLookupByName(stopDomainName)
		if err := LibvirtConnexion.DomainShutdown(domain); err != nil {
			if strings.Contains(err.Error(), "not running") {
				Logger.Warn("Machine is already stopped", "name", stopDomainName)
				os.Exit(0)
			} else if strings.Contains(err.Error(), "not found") {
				Logger.Error("Machine not found", "name", stopDomainName)
				os.Exit(1)
			} else {
				log.Panicf("Could not stop the machine: %s. Reason: %v", stopDomainName, err)
			}
		}
		Logger.Info("Stop machine", "name", stopDomainName)
	},
}

func init() {
	machineStopCmd.Flags().StringVarP(&stopDomainName, "name", "n", "", "Name of the machine to stop.")
	if err := machineStopCmd.MarkFlagRequired("name"); err != nil {
		log.Panic(err.Error())
	}
}
