package shellcli

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"
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

		domain, _ := LibvirtConnexion.DomainLookupByName(domainName)
		err := LibvirtConnexion.DomainCreate(domain)
		if err != nil {
			if strings.Contains(err.Error(), "already running") {
				Logger.Warn("Skip : machine is already running",
					zap.String("machine", domainName))
				os.Exit(0)
			} else if strings.Contains(err.Error(), "not found") {
				Logger.Error("Machine not found",
					zap.String("machine", domainName))
				os.Exit(1)
			} else {
				Logger.Panic("Could not start the domain",
					zap.String("machine", domain.Name),
					zap.Error(err))
				os.Exit(1)
			}
		}
		Logger.Info("Starting domain",
			zap.String("domain", domainName))
	},
}

func init() {
	machineStartCmd.Flags().StringVarP(&domainName, "name", "n", "", "Name of the machine to start.")
}
