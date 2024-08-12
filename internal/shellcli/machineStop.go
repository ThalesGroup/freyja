package shellcli

import (
	"freyja/internal"
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
		// logger
		internal.InitLogger()
		// execute
		machineStop(stopDomainName)
	},
}

func init() {
	machineStopCmd.Flags().StringVarP(&stopDomainName, "name", "n", "", "Name of the machine to stop.")
	if err := machineStopCmd.MarkFlagRequired("name"); err != nil {
		log.Panic(err.Error())
	}
}

func machineStop(domainName string) {
	domain, _ := LibvirtConnexion.DomainLookupByName(domainName)
	if err := LibvirtConnexion.DomainShutdown(domain); err != nil {
		if strings.Contains(err.Error(), "not running") {
			Logger.Warn("Machine is already stopped", "name", domainName)
			os.Exit(0)
		} else if strings.Contains(err.Error(), "not found") {
			Logger.Error("Machine not found", "name", domainName)
			os.Exit(1)
		} else {
			log.Panicf("Could not stop the machine: %s. Reason: %v", domainName, err)
		}
	}
	Logger.Info("Stop machine", "name", domainName)
}
