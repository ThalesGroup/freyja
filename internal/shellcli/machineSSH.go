package shellcli

import (
	"fmt"
	"log"
	"os"

	"github.com/digitalocean/go-libvirt"
	"github.com/spf13/cobra"
)

var consoleDomainName string

// commands definitions
var machineConsoleCmd = &cobra.Command{
	Use:              "console",
	Short:            "Open machine console",
	Long:             "Open libvirt console mode for a domain that has a console device",
	TraverseChildren: true, // ensure local flags do not spread to sub commands

	Run: func(cmd *cobra.Command, args []string) {
		// domains list
		Logger.Info("Opening machine console", "machine", consoleDomainName)

		if err := OpenConsole(consoleDomainName); err != nil {
			Logger.Error("Failed to open console", "domain", consoleDomainName, "reason", err.Error())
			os.Exit(1)
		}
	},
}

func init() {
	machineConsoleCmd.Flags().StringVarP(&consoleDomainName, "name", "n", "", "Name of the machine to describe.")
	if err := machineConsoleCmd.MarkFlagRequired("name"); err != nil {
		log.Panic(err.Error())
	}
}

func OpenConsole(domainName string) (err error) {
	// get domain by name
	domain, err := LibvirtConnexion.DomainLookupByName(domainName)
	if err != nil {
		return fmt.Errorf("failed to lookup domain '%s': %w", domainName, err)
	}

	// look for console device to use
	xmlDescription, err := getDomainXMLDescription(domain)
	if err != nil {
		return fmt.Errorf("failed to get XML description for domain '%s': %w", domainName, err)
	}
	consoleDeviceName := xmlDescription.Devices.Console[0].Alias.Name

	devName := libvirt.OptString{consoleDeviceName}
	if err := LibvirtConnexion.DomainOpenConsole(domain, devName, os.Stdout, uint32(libvirt.DomainConsoleSafe)); err != nil {
		return fmt.Errorf("failed to open console '%s' for domain '%s': %w", consoleDeviceName, domainName, err)
	}
	return nil
}
