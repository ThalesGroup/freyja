package shellcli

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"log"
	"os"
	"strings"
)

var domainName string

// commands definitions
var machineStartCmd = &cobra.Command{
	Use:              "start",
	Short:            "Machine startup",
	Long:             "Machine startup using handler",
	TraverseChildren: true, // ensure local flags do not spread to sub commands

	Run: func(cmd *cobra.Command, args []string) {
		// logger
		setLogger()
		// execute
		machineStart(domainName)
	},
}

func init() {
	machineStartCmd.Flags().StringVarP(&domainName, "name", "n", "", "Name of the machine to start.")
	if err := machineStartCmd.MarkFlagRequired("name"); err != nil {
		log.Panic(err.Error())
	}
}

func machineStart(domainName string) {
	domain, _ := LibvirtConnexion.DomainLookupByName(domainName)
	err := LibvirtConnexion.DomainCreate(domain)
	if err != nil {
		if strings.Contains(err.Error(), "already running") {
			Logger.Warn("Skip : machine is already running", "name", domainName)
			os.Exit(0)
		} else if strings.Contains(err.Error(), "not found") {
			Logger.Error("Machine not found", "name", domainName)
			os.Exit(1)
		} else {
			log.Panic("Could not start the machine", "name", domainName, "error", err)
		}
	}
	Logger.Info("Starting machine",
		zap.String("name", domainName))
}
