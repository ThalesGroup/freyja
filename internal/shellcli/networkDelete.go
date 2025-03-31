package shellcli

import (
	"fmt"
	"freyja/internal"
	"freyja/internal/configuration"
	"github.com/spf13/cobra"
	"os"
)

var networkName string

// commands definitions
var networkDeleteCmd = &cobra.Command{
	Use:              "delete",
	Short:            "Virtual network deletion",
	Long:             "Virtual network deletion using handler",
	TraverseChildren: true, // ensure local flags do not spread to sub commands

	Run: func(cmd *cobra.Command, args []string) {
		Logger.Warn("deleting networks that are still in use may affect machines !")
		var err error
		if configurationPath != "" {
			err = deleteNetworksByConf()
		} else if networkName != "" {
			Logger.Info("delete network", "name", networkName)
			err = deleteNetworkByName(networkName)
		}

		if err != nil {
			Logger.Error("cannot delete networks", "reason", err.Error())
		}

	},
}

func init() {
	// names are used to delete the networks provided by name
	networkDeleteCmd.Flags().StringVarP(&networkName, "name", "n", "", "Name of the network to delete.")
	// configuration is used to delete all the network contains in this configuration
	networkDeleteCmd.Flags().StringVarP(&configurationPath, "config", "c", "", "Path to the configuration file to delete all the networks described in it.")

	if &networkName == nil && &configurationPath == nil {
		Logger.Error("canceled: you must provide at least '-n' or '-c' arguments")
		os.Exit(1)
	}
}

func deleteNetworksByConf() (err error) {
	// build config from path
	var freyjaConfiguration configuration.FreyjaConfiguration
	if err = freyjaConfiguration.BuildFromFile(configurationPath); err != nil {
		return fmt.Errorf("cannot parse configuration file '%s': %w", configurationPath, err)
	}

	// delete networks
	// assemble the list to create and to log for user information
	networksToDelete := make([]string, len(freyjaConfiguration.Networks))
	for i, network := range freyjaConfiguration.Networks {
		networksToDelete[i] = network.Name
	}
	// delete network one by one after confirmation
	Logger.Info("delete networks", "names", networksToDelete)
	for _, name := range networksToDelete {
		if err = deleteNetworkByName(name); err != nil {
			return err
		}
	}
	return nil
}

func deleteNetworkByName(name string) (err error) {

	if internal.AskUserYesNoConfirmation() {
		// find
		network, err := LibvirtConnexion.NetworkLookupByName(name)
		if err != nil {
			return fmt.Errorf("cannot find network '%s': %w", name, err)
		}

		// destroy
		if err := LibvirtConnexion.NetworkDestroy(network); err != nil {
			return fmt.Errorf("cannot destroy network '%s': %w", name, err)
		}

		// undefine
		if err := LibvirtConnexion.NetworkUndefine(network); err != nil {
			return fmt.Errorf("cannot undefine network '%s': %w", name, err)
		}

		Logger.Info("Network deleted", "network", name)
	} else {
		Logger.Info("Canceled")
	}
	return nil
}
