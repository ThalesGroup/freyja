package shellcli

import (
	"freyja/internal"
	"github.com/spf13/cobra"
	"log"
	"os"
	"strings"
)

var startDomainName string

// commands definitions
var machineStartCmd = &cobra.Command{
	Use:              "start",
	Short:            "Machine startup",
	Long:             "Machine startup using handler",
	TraverseChildren: true, // ensure local flags do not spread to sub commands

	Run: func(cmd *cobra.Command, args []string) {
		// logger
		Logger = internal.InitLogger()
		// execute
		machineStart(startDomainName)
	},
}

func init() {
	machineStartCmd.Flags().StringVarP(&startDomainName, "name", "n", "", "Name of the machine to start.")
	if err := machineStartCmd.MarkFlagRequired("name"); err != nil {
		log.Panic(err.Error())
	}
}

func machineStart(domainName string) {
	domain, _ := LibvirtConnexion.DomainLookupByName(domainName)
	if err := LibvirtConnexion.DomainCreate(domain); err != nil {
		if strings.Contains(err.Error(), "already running") {
			Logger.Warn("Machine is already running", "name", domainName)
			os.Exit(0)
		} else if strings.Contains(err.Error(), "not found") {
			Logger.Error("Machine not found", "name", domainName)
			os.Exit(1)
		} else {
			log.Panicf("Could not start the machine: %s. Reason: %v", domainName, err)
		}
	}
	Logger.Info("Start machine", "name", domainName)
}
