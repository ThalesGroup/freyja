package shellcli

import (
	"fmt"
	"freyja/internal/configuration"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// RemoteProcNetworkSetAutostart is set to handle the int32 flag value for network autostart
// protocol in libvirt
// source : https://lists.libvirt.org/archives/list/devel@lists.libvirt.org/message/VTZOSYUKTVG3YXGFXOKJS5SLR2VFMMLS/
var RemoteProcNetworkSetAutostart int32 = 48

// commands definitions
var networkCreateCmd = &cobra.Command{
	Use:              "create",
	Short:            "Network creation",
	Long:             "Network creation to attach machines to it",
	TraverseChildren: true, // ensure local flags do not spread to sub commands

	Run: func(cmd *cobra.Command, args []string) {
		Logger.Debug("create networks from configuration file", "config", configurationPath)

		// TODO
		//  - test different usecases for network creation
		//  - also test to ask for info from libvirt
		//  -  test without a default nat configuration, but in my opinion, it needs to be generated
		//     because it will conflict with the 'default' network
		if err := createNetwork(); err != nil {
			Logger.Error("cannot create networks", "reason", err.Error())
			os.Exit(1)
		}
	},
}

func init() {
	// MANDATORY --config, -c
	networkCreateCmd.Flags().StringVarP(&configurationPath, "config", "c", "", "Path to the configuration file to create the networks only.")
	if err := networkCreateCmd.MarkFlagRequired("config"); err != nil {
		log.Panic(err.Error())
	}
	// OPTIONAL --dry-run
	networkCreateCmd.Flags().BoolVarP(&dryRun, "dry-run", "", false, "Generate all config files without creating the machine")
}

func createNetwork() (err error) {
	// build config from path
	var freyjaConfiguration configuration.FreyjaConfiguration
	if err = freyjaConfiguration.BuildFromFile(configurationPath); err != nil {
		return fmt.Errorf("cannot parse configuration file '%s': %w", configurationPath, err)
	}

	// create networks
	for _, network := range freyjaConfiguration.Networks {
		// check if network already exists
		// it prevents to update an existing network used by running machines and avoid
		// unwanted side effects
		foundNet, err := LibvirtConnexion.NetworkLookupByName(network.Name)
		if err != nil {
			// ignore error only if it contains the message that the network does not exist
			// which is the purpose of this command
			if strings.Contains(err.Error(), "not found") {
				// create network directory
				Logger.Debug("create network dir", "network", network.Name, "parent", FreyjaNetworksWorkspaceDir)
				networkDirPath, err := createNetworkDir(&network)
				if err != nil {
					return fmt.Errorf("cannot create network '%s' workspace directory in '%s': %w", networkName, networkDirPath, err)
				}

				// create libvirt network configuration
				// also store the network configs dump path in a map to create them later
				// it is useful in case of a --dry-run generation
				Logger.Debug("create network XML description for lib")
				var xmlDescription []byte
				if xmlDescription, err = configuration.CreateLibvirtNetworkXMLDescription(network); err != nil {
					return fmt.Errorf("cannot create the libvirt XML description for network '%s': %w", network.Name, err)
				}
				filename := fmt.Sprintf("%s%s%s", XMLNetworkDescriptionPrefix, network.Name, XMLNetworkDescriptionSuffix)
				xmlNetworkDescriptionPath := filepath.Join(networkDirPath, filename)
				if err = configuration.DumpNetworkConfig(xmlDescription, xmlNetworkDescriptionPath); err != nil {
					// the xml configuration has been created but cannot be written on disk
					// this is a warning and not an error since it does not prevent the network
					// to be created in libvirt
					Logger.Warn("cannot write libvirt network XML description", "network", network.Name, "path", xmlNetworkDescriptionPath, "reason", err.Error())
				}

				// create the network in libvirt
				if !dryRun {
					Logger.Info("create network", "network", network.Name)
					Logger.Debug("define network from libvirt xml description", "network", network.Name, "path", xmlNetworkDescriptionPath)
					net, err := LibvirtConnexion.NetworkDefineXML(string(xmlDescription))
					if err != nil {
						return fmt.Errorf("cannot define network '%s' in libvirt from xml description: %w", network.Name, err)
					}

					Logger.Debug("create network from libvirt xml description", "network", network.Name, "path", xmlNetworkDescriptionPath)
					if err = LibvirtConnexion.NetworkCreate(net); err != nil {
						return fmt.Errorf("cannot create network '%s' in libvirt: %w", network.Name, err)
					}

					if err = LibvirtConnexion.NetworkSetAutostart(net, RemoteProcNetworkSetAutostart); err != nil {
						Logger.Warn("cannot set network autostart in libvirt", "reason", err.Error(), "network", network.Name)
					}
				} else {
					Logger.Warn("skipped creation", "network", network.Name, "reason", "option --dry-run")
				}

			} else {
				// unwanted error
				return err
			}
		} else if network.Name == foundNet.Name {
			// if error is not nil and network exists, skip this one
			Logger.Warn("skip creation : network already exists", "network", network.Name)
			continue
		}
	}
	return nil
}

// createMachineDir returns the created dir, or an error
func createNetworkDir(network *configuration.FreyjaConfigurationNetwork) (string, error) {
	networkDirPath := filepath.Join(FreyjaNetworksWorkspaceDir, network.Name)
	if _, err := os.Stat(networkDirPath); os.IsNotExist(err) {
		if err := os.MkdirAll(networkDirPath, os.ModePerm); err != nil {
			return "", err
		}
	}
	return networkDirPath, nil
}
