package shellcli

import (
	"encoding/xml"
	"freyja/internal"
	"github.com/spf13/cobra"
	"log"
	"os"
	"strings"
)

const networkTemplatePath string = "templates/network.xml.tmpl"

type NetworkData struct {
	Name      string
	UUID      string
	Interface string
}

var networkName string

// commands definitions
var networkCreateCmd = &cobra.Command{
	Use:              "create",
	Short:            "Network creation",
	Long:             "Network creation using libvirt",
	TraverseChildren: true, // ensure local flags do not spread to sub commands

	Run: func(cmd *cobra.Command, args []string) {
		// TODO :
		//  to create a network with existing routed interfaces on host :
		//  https://libvirt.org/formatnetwork.html#routed-network-config

		config := NetworkCreateConfig(networkName)
		configXMLBytes, err := xml.Marshal(config)
		if err != nil {
			Logger.Error("cannot parse network config", "network", networkName, "config", config, "reason", err)
			os.Exit(1)
		}
		configStr := string(configXMLBytes)

		// create the network in libvirt
		_, err = LibvirtConnexion.NetworkCreateXML(configStr)
		if err != nil {
			if strings.Contains(err.Error(), "already exists") {
				Logger.Warn("Network already exists", "network", networkName)
				os.Exit(0)
			}
			Logger.Error("Cannot create the network using libvirt", "config", configStr, "reason", err)
			os.Exit(1)
		}

		// define the network in libvirt
		net, err := LibvirtConnexion.NetworkDefineXML(configStr)
		if err != nil {
			Logger.Warn("Network created but could not define it in libvirt", "network", networkName, "reason", err)
		}

		// set autostart
		if err := LibvirtConnexion.NetworkSetAutostart(net, 0); err != nil {
			Logger.Warn("Network created and defined but could not set autostart in libvirt", "network", networkName, "reason", err)
		}
		Logger.Info("Network created", "name", networkName)

	},
}

func init() {
	networkCreateCmd.Flags().StringVarP(&networkName, "name", "n", "", "Name of the network to create")
	if err := networkCreateCmd.MarkFlagRequired("name"); err != nil {
		log.Panic(err.Error())
	}
}

func NetworkCreateConfig(networkName string) internal.XMLNetworkDescription {
	configForward := internal.XMLNetworkDescriptionForward{
		Mode: "bridge",
	}
	configBridge := internal.XMLNetworkDescriptionBridge{
		Name: "virbr0",
	}
	return internal.XMLNetworkDescription{
		Name:    networkName,
		UUID:    internal.GenerateUUID(),
		Forward: configForward,
		Bridge:  configBridge,
	}
}
