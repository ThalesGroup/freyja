package shellcli

import (
	"fmt"
	"freyja/internal"
	"freyja/internal/configuration"
	"github.com/digitalocean/go-libvirt"
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

// #network create -c test/configuration/static/network_complete_conf.yaml --dry-run
// deleteNetworksByName takes a list of network names and delete them in Libvirt
func deleteNetworksByName(names []string) (err error) {
	Logger.Info("delete networks", "names", names)

	var deletedNetworks []string
	if internal.AskUserYesNoConfirmation() {
		for _, name := range names {
			network, errD := LibvirtConnexion.NetworkLookupByName(name)
			if errD != nil {
				Logger.Warn("skipped network deletion in libvirt", "network", name, "reason", errD.Error())
			} else if errD = deleteNetworkInLibvirt(network); errD != nil {
				// delete in libvirt
				Logger.Error("cannot delete network in libvirt", "network", name, "reason", errD.Error())
			}

			// delete config dirs
			networkDirPath := GetLibvirtNetworkDir(name)
			errRemoveDir := os.RemoveAll(networkDirPath)
			if errRemoveDir != nil {
				Logger.Error("cannot remove network directory", "path", networkDirPath, "reason", errRemoveDir.Error())
			} else {
				Logger.Info("removed network directory", "path", networkDirPath)
			}

		}
	} else {
		Logger.Info("Canceled")
		return nil
	}

	Logger.Info("Networks deleted", "networks", deletedNetworks)
	return err
}

// deleteNetworkInLibvirt execute 'destroy' and 'undefine' operations in libvirt for a network
func deleteNetworkInLibvirt(network libvirt.Network) (err error) {
	// destroy
	if err = LibvirtConnexion.NetworkDestroy(network); err != nil {
		return fmt.Errorf("cannot destroy network '%s' in libvirt: %v", network.Name, err.Error())
	} else if err = LibvirtConnexion.NetworkUndefine(network); err != nil {
		// undefine
		return fmt.Errorf("cannot undefine network '%s' in libvirt: %v", network.Name, err.Error())
	}
	return nil
}
