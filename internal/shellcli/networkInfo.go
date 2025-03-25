package shellcli

import (
	"encoding/xml"
	"fmt"
	"freyja/internal/configuration"
	"github.com/digitalocean/go-libvirt"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"log"
	"os"
)

type NetworkDescription struct {
	Name     string           `yaml:"name"`
	Mode     string           `yaml:"mode"`
	Bridge   string           `yaml:"bridge"`
	Gateway  string           `yaml:"gateway"`
	Netmask  string           `yaml:"netmask"`
	Dhcp     DHCPDescription  `yaml:"dhcp"`
	Machines []NetworkMachine `yaml:"machines,omitempty"`
}

type DHCPDescription struct {
	Start string `yaml:"start"`
	End   string `yaml:"end"`
}

type NetworkMachine struct {
	Name string `yaml:"name"`
	Mac  string `yaml:"mac"`
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
		info, err := getNetworkDescription(infoNetworkName)
		if err != nil {
			Logger.Error("cannot get network info", "network", infoNetworkName, "reason", err.Error())
			os.Exit(1)
		}
		output, err := yaml.Marshal(info)
		if err != nil {
			Logger.Error("cannot parse network info", "network", infoNetworkName, "reason", err.Error())
			os.Exit(1)
		}
		fmt.Print(string(output))
	},
}

func init() {
	// mandatory
	networkInfoCmd.Flags().StringVarP(&infoNetworkName, "name", "n", "", "Name of the network to describe.")
	if err := networkInfoCmd.MarkFlagRequired("name"); err != nil {
		log.Panic(err.Error())
	}
}

// getNetworkDescription provides the description of a libvirt network
// do not confuse it with a domain interface
func getNetworkDescription(networkName string) (netDesc *NetworkDescription, err error) {
	// get network from name
	network, err := LibvirtConnexion.NetworkLookupByName(networkName)
	if err != nil {
		Logger.Error("Cannot lookup network by name using Qemu connexion", "network", networkName)
		return nil, err
	}

	// get xml description from network
	xmlDescription, err := LibvirtConnexion.NetworkGetXMLDesc(network, uint32(libvirt.NetworkXMLInactive))
	if err != nil {
		Logger.Error("Cannot get network XML description", "network", network.Name)
		return nil, err
	}
	var description configuration.XMLNetworkDescription
	err = xml.Unmarshal([]byte(xmlDescription), &description)
	if err != nil {
		Logger.Error("Cannot unmarshal domain XML description", "domain", infoNetworkName)
		return nil, err
	}

	// generate short description
	netDesc = &NetworkDescription{
		Name:   description.Name,
		Bridge: description.Bridge.Name,
	}
	if description.Forward != nil {
		if description.Forward.Mode == "" {
			netDesc.Mode = configuration.DefaultNetworkForwardMode
		} else {
			netDesc.Mode = description.Forward.Mode
		}
	} else {
		netDesc.Mode = configuration.DefaultNetworkForwardMode
	}
	if description.Ip != nil {
		netDesc.Gateway = description.Ip.Address
		netDesc.Netmask = description.Ip.Netmask
		netDesc.Dhcp = DHCPDescription{
			Start: description.Ip.Dhcp.Range.Start,
			End:   description.Ip.Dhcp.Range.End,
		}
	}

	// get the machines using this network
	machines, err := getNetworkDomains(networkName)
	if err != nil {
		return nil, fmt.Errorf("cannot list all the domains using network '%s': %w", networkName, err)
	}
	netDesc.Machines = machines

	return
}

// getNetworkDomains list all the domains using the given network
func getNetworkDomains(networkName string) (machines []NetworkMachine, err error) {
	// get domains
	domains, err := getDomainsList()
	if err != nil {
		return nil, fmt.Errorf("cannot get list of domains for network '%s': %w", networkName, err)
	}

	// filter domains for this network
	for _, domain := range domains {
		// get domain description
		xmlDescription, err := getDomainXMLDescription(domain)
		if err != nil {
			return nil, fmt.Errorf("cannot get domain's '%s' xml description: %w", domain.Name, err)
		}
		// in this description, get interfaces descriptions and filter by network name
		for _, iface := range xmlDescription.Devices.Interfaces {
			if iface.Source.Network == networkName {
				machines = append(machines, NetworkMachine{
					Name: xmlDescription.Name,
					Mac:  iface.Mac.Address,
				})
			}
		}
	}

	return
}
