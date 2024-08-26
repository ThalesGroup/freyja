package shellcli

import (
	"encoding/xml"
	"fmt"
	"freyja/internal"
	"github.com/digitalocean/go-libvirt"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"log"
)

type shortNetworkDescription struct {
	Name   string               `yaml:"name"`
	Mode   string               `yaml:"mode"`
	Bridge string               `yaml:"bridge"`
	Mac    string               `yaml:"mac"`
	IP     string               `yaml:"ip"`
	Dhcp   shortDHCPDescription `yaml:"dhcp"`
}

type shortDHCPDescription struct {
	Start string `yaml:"start"`
	End   string `yaml:"end"`
}

var infoNetworkName string

// commands definitions
var networkInfoCmd = &cobra.Command{
	Use:              "info",
	Short:            "Print a network information",
	Long:             "Print a network information using libvirt",
	TraverseChildren: true, // ensure local flags do not spread to sub commands

	Run: func(cmd *cobra.Command, args []string) {
		// generate short yaml description
		info, err := getNetworkShortDescription(infoNetworkName)
		if err != nil {
			log.Panicf("Cannot get network '%s' info: %v", infoNetworkName, err)
		}
		output, err := yaml.Marshal(info)
		if err != nil {
			log.Panicf("Cannot parse network '%s' info in yaml format: %v", infoNetworkName, err)
		}
		fmt.Print(string(output))
	},
}

func init() {
	networkInfoCmd.Flags().StringVarP(&infoNetworkName, "name", "n", "", "Name of the network to describe.")
	if err := networkInfoCmd.MarkFlagRequired("name"); err != nil {
		log.Panic(err.Error())
	}
}

func getNetworkShortDescription(networkName string) (*shortNetworkDescription, error) {
	// get network from name
	network, err := LibvirtConnexion.NetworkLookupByName(networkName)
	if err != nil {
		Logger.Error("Cannot lookup network by name using Qemu connexion", "network", network.Name)
		return nil, err
	}
	// get xml description from network
	xmlDescription, err := LibvirtConnexion.NetworkGetXMLDesc(network, uint32(libvirt.NetworkXMLInactive))
	if err != nil {
		Logger.Error("Cannot get network XML description", "network", network.Name)
		return nil, err
	}
	var description internal.XMLNetworkDescription
	err = xml.Unmarshal([]byte(xmlDescription), &description)
	if err != nil {
		Logger.Error("Cannot unmarshal domain XML description", "domain", infoNetworkName)
		return nil, err
	}
	// generate short description
	return &shortNetworkDescription{
		Name:   description.Name,
		Mode:   description.Forward.Mode,
		Bridge: description.Bridge.Name,
		Mac:    description.Mac.Address,
		IP:     description.Ip.Address,
		Dhcp: shortDHCPDescription{
			Start: description.Ip.Dhcp.Range.Start,
			End:   description.Ip.Dhcp.Range.End,
		},
	}, nil
}
