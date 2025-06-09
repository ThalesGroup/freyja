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

// networkCreateCmd creates a new network in libvirt by generating its configuration in xml format,
// defines it in libvirt then create it
var networkCreateCmd = &cobra.Command{
	Use:              "create",
	Short:            "Network creation",
	Long:             "Network creation to attach machines to it",
	TraverseChildren: true, // ensure local flags do not spread to sub commands

	Run: func(cmd *cobra.Command, args []string) {
		Logger.Debug("create networks from configuration file", "config", configurationPath)

		// build config from path
		var freyjaConfiguration configuration.FreyjaConfiguration
		if err := freyjaConfiguration.BuildFromFile(configurationPath); err != nil {
			Logger.Error("cannot parse configuration", "configuration", configurationPath, "reason", err.Error())
			os.Exit(1)
		}

		xmlDescriptions, err := GenerateLibvirtNetworksXMLDescriptions(&freyjaConfiguration, FreyjaNetworksWorkspaceDir)
		if err != nil {
			Logger.Error("cannot generate networks XML descriptions for Libvirt", "reason", err.Error())
			os.Exit(1)
		}

		if !dryRun {
			if err := CreateNetworksInLibvirt(xmlDescriptions); err != nil {
				Logger.Error("cannot create networks in Libvirt from XML descriptions", "reason", err.Error())
				os.Exit(1)
			}
		} else {
			Logger.Warn("skipped networks creation in libvirt", "reason", "option --dry-run")
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

func GetLibvirtNetworkDir(networkName string) (path string) {
	return filepath.Join(FreyjaNetworksWorkspaceDir, networkName)
}

func GetLibvirtNetworkDescriptionPath(networkName string) (path string) {
	networkDir := GetLibvirtNetworkDir(networkName)
	filename := fmt.Sprintf("%s%s%s", XMLNetworkDescriptionPrefix, networkName, XMLNetworkDescriptionSuffix)
	return filepath.Join(networkDir, filename)
}

// GenerateLibvirtNetworksXMLDescriptions create the network configurations for libvirt and dump
// them on disk in xml format, inside their dedicated directory.
// Returns the dumped configurations.
func GenerateLibvirtNetworksXMLDescriptions(config *configuration.FreyjaConfiguration, networksDir string) (xmlDescriptions map[string][]byte, err error) {
	// create networks
	xmlDescriptions = make(map[string][]byte, len(config.Networks))
	for _, network := range config.Networks {
		// check first if a network already exists in libvirt
		// it prevents updating and overwrite an existing network config that is already used by
		// running machines
		foundNet, err := LibvirtConnexion.NetworkLookupByName(network.Name)
		if err != nil {
			// ignore error only if it contains the message that the network does not exist
			// which is the purpose of this command
			if strings.Contains(err.Error(), "not found") {

				// create network directory
				Logger.Debug("create network dir", "network", network.Name, "parent", networksDir)
				xmlNetworkDescriptionPath := GetLibvirtNetworkDescriptionPath(network.Name)
				networkDir := filepath.Dir(xmlNetworkDescriptionPath)
				if _, err := os.Stat(networkDir); os.IsNotExist(err) {
					if err := os.MkdirAll(networkDir, os.ModePerm); err != nil {
						return nil, err
					}
				}

				// create libvirt network configuration
				// also store the network configs dump a path in a map to create them later
				// it is useful in case of a --dry-run generation
				Logger.Debug("create network XML description for libvirt", "network", network.Name)
				var xmlDescription []byte
				if xmlDescription, err = configuration.CreateLibvirtNetworkXMLDescription(network); err != nil {
					return nil, fmt.Errorf("cannot create the libvirt XML description for network '%s': %w", network.Name, err)
				}

				// dump configuration on disk
				if err = configuration.DumpNetworkConfig(xmlDescription, xmlNetworkDescriptionPath); err != nil {
					// the xml configuration has been created but cannot be written on disk
					// this is a warning and not an error since it does not prevent the network
					// to be created in libvirt
					Logger.Warn("cannot write libvirt network XML description", "network", network.Name, "path", xmlNetworkDescriptionPath, "reason", err.Error())
				}

				// add it in the result
				xmlDescriptions[network.Name] = xmlDescription

			} else {
				// otherwise, the error is unexpected
				return nil, err
			}
		} else if network.Name == foundNet.Name {
			// if error is not nil and network exists, we do not overwrite its configuration file,
			// and we will not create it in libvirt, excluding it from the returned configs.
			// In this case, we do not abort the command execution.
			// If machines are created, they will boot on top of the existing networks.
			// This behavior can be discussed
			Logger.Warn("network already exists", "network", network.Name)
		}
	}
	return xmlDescriptions, nil
}

// CreateNetworksInLibvirt get the Libvirt's XML description (configuration) of a network and
// define it then create it in libvirt.
func CreateNetworksInLibvirt(xmlDescriptions map[string][]byte) error {

	for name, desc := range xmlDescriptions {
		Logger.Debug("create network in Libvirt from xml descriptions", "network", name)

		net, err := LibvirtConnexion.NetworkDefineXML(string(desc))
		if err != nil {
			return fmt.Errorf("cannot define network in libvirt from xml description: %w", err)
		}

		if err = LibvirtConnexion.NetworkCreate(net); err != nil {
			return fmt.Errorf("cannot create network in libvirt: %w", err)
		}

		if err = LibvirtConnexion.NetworkSetAutostart(net, RemoteProcNetworkSetAutostart); err != nil {
			Logger.Warn("cannot set network autostart in libvirt. continue.", "reason", err.Error())
		}
	}
	return nil
}
