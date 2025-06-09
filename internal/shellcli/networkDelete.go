package shellcli

import (
	"fmt"
	"freyja/internal"
	"freyja/internal/configuration"
	"github.com/spf13/cobra"
	"os"
)

var networkNames []string

// networkDeleteCmd delete a list of networks from names provided in input or from a Freyja config
// file
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
		} else if networkNames != nil {
			err = deleteNetworksByName(networkNames)
			if err != nil {
				Logger.Error("cannot delete networks", "reason", err.Error())
			}

		}

	},
}

func init() {
	// names are used to delete the networks provided by name
	networkDeleteCmd.Flags().StringArrayVarP(&networkNames, "name", "n", nil, "Name of the network to delete. Repeat this flag to delete multiple networks.")
	// configuration is used to delete all the network contains in this configuration
	networkDeleteCmd.Flags().StringVarP(&configurationPath, "config", "c", "", "Path to the configuration file to delete all the networks described in it.")

	if &networkNames == nil && &configurationPath == nil {
		Logger.Error("canceled: you must provide at least '-n' or '-c' arguments")
		os.Exit(1)
	}
}

// deleteNetworksByConf gets the configuration file in input and delete all the network set inside
// using Libvirt
func deleteNetworksByConf() (err error) {
	// build config from a path
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
	return deleteNetworksByName(networksToDelete)
}

// deleteNetworksByName takes a list of network names and delete them in Libvirt
func deleteNetworksByName(names []string) (err error) {
	Logger.Info("delete networks", "names", names)

	if internal.AskUserYesNoConfirmation() {
		for _, name := range names {
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

			// delete config dirs
			networkDirPath := GetLibvirtNetworkDir(name)
			if err = os.RemoveAll(networkDirPath); err != nil {
				return fmt.Errorf("cannot remove network directory '%s': %w", networkDirPath, err)
			}

			Logger.Info("Network deleted", "network", name)
		}
	} else {
		Logger.Info("Canceled")
	}
	return nil
}
