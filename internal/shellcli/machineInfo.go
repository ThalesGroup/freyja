package shellcli

import (
	"encoding/xml"
	"fmt"
	"freyja/internal"
	"freyja/internal/configuration"
	"github.com/digitalocean/go-libvirt"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"log"
	"os"
)

type machineDescription struct {
	Hostname     string             `yaml:"hostname"`
	OSInfo       string             `yaml:"osInfo"`
	Architecture string             `yaml:"architecture"`
	Memory       string             `yaml:"memory"` // "<amount> MB"
	Vcpu         uint               `yaml:"vcpu"`
	Interfaces   []machineInterface `yaml:"interfaces"`
	Disks        []machineDisk
}

type machineInterface struct {
	HostInterface string `yaml:"hostInterface"`
	Network       string `yaml:"network"`
	Mac           string `yaml:"mac"`
	IP            string `yaml:"ip"`
}

type machineDisk struct {
	Device     string `yaml:"device"`
	Capacity   string `yaml:"capacity"`   // "<amount> GB"
	Allocation string `yaml:"allocation"` // "<amount> GB"
}

var infoDomainName string

// commands definitions
var machineInfoCmd = &cobra.Command{
	Use:              "info",
	Short:            "Print a machine information",
	Long:             "Print a machine information using libvirt",
	TraverseChildren: true, // ensure local flags do not spread to sub commands

	Run: func(cmd *cobra.Command, args []string) {

		// get domain by name
		domain, err := LibvirtDomainConnexion.DomainLookupByName(infoDomainName)
		if err != nil {
			Logger.Error("Cannot lookup domain from qemu connexion", "domain", infoDomainName, "reason", err)
			os.Exit(1)
		}

		// get xml description from domain
		xmlDescription, err := getDomainXMLDescription(domain)
		if err != nil {
			Logger.Error("Cannot get domain's XML description", "domain", infoDomainName, "reason", err)
			os.Exit(1)
		}

		// generate short YAML description
		info, err := getDomainShortDescription(domain, xmlDescription)
		output, err := yaml.Marshal(info)
		if err != nil {
			Logger.Error("Cannot parse domain info in yaml format", "domain", infoDomainName, "reason", err)
			os.Exit(1)
		}
		fmt.Print(string(output))

	},
}

func init() {
	machineInfoCmd.Flags().StringVarP(&infoDomainName, "name", "n", "", "Name of the machine to describe.")
	if err := machineInfoCmd.MarkFlagRequired("name"); err != nil {
		log.Panic(err.Error())
	}
}

func getDomainXMLDescription(domain libvirt.Domain) (*configuration.XMLDomainDescription, error) {
	// DOMAIN
	domainName := domain.Name
	// get domain xml description
	domainDesc, err := LibvirtDomainConnexion.DomainGetXMLDesc(domain, 1)
	if err != nil {
		Logger.Error("Cannot get domain XML description", "domain", domainName)
		return nil, err
	}
	// parse description to xml struct
	var description configuration.XMLDomainDescription
	if err := xml.Unmarshal([]byte(domainDesc), &description); err != nil {
		Logger.Error("Cannot unmarshal domain XML description", "domain", domainName)
		return nil, err
	}

	return &description, err
}

// getDomainIPByMac searches the IP of a machine interface in the DHCP leases of
// the system daemon's networks, by mac address.
// The machines run in the session daemon but their interfaces plug into the host
// bridges of the system daemon's networks, so the DHCP servers delivering their
// leases belong to the system daemon.
func getDomainIPByMac(mac string) (ip string, err error) {
	networks, _, err := LibvirtNetworkConnexion.ConnectListAllNetworks(1, libvirt.ConnectListNetworksActive)
	if err != nil {
		return "", fmt.Errorf("cannot list the system daemon's networks: %w", err)
	}
	for _, network := range networks {
		leases, _, err := LibvirtNetworkConnexion.NetworkGetDhcpLeases(network, []string{mac}, 1, 0)
		if err != nil {
			// the network may not have a DHCP server (no lease API): skip it
			continue
		}
		for _, lease := range leases {
			return lease.Ipaddr, nil
		}
	}
	return "", nil
}

func getDomainShortDescription(domain libvirt.Domain, description *configuration.XMLDomainDescription) (short *machineDescription, err error) {
	// get networks info from xml struct
	var interfaces []machineInterface
	for _, descriptionIface := range description.Devices.Interfaces {
		var hostInterface, ip string
		// if machine is up, we can retrieve the host interface from libvirt
		if descriptionIface.Target != nil {
			hostInterface = descriptionIface.Target.Device
		} else {
			hostInterface = "not up: machine down"
		}
		// the IP is searched in the DHCP leases of the system daemon's networks
		var mac string
		if descriptionIface.Mac != nil {
			mac = descriptionIface.Mac.Address
			ip, err = getDomainIPByMac(mac)
			if err != nil {
				Logger.Error("Cannot get domain network information", "domain", domain.Name)
				return nil, err
			}
		}
		if ip == "" {
			ip = "inactive"
		}
		// bridge interfaces do not carry the network name: display the host bridge
		network := descriptionIface.Source.Network
		if network == "" {
			network = descriptionIface.Source.Bridge
		}
		machineIface := &machineInterface{
			HostInterface: hostInterface,
			Network:       network,
			Mac:           mac,
			IP:            ip,
		}
		interfaces = append(interfaces, *machineIface)
	}

	// get disks info from xml struct
	var disks []machineDisk
	for _, diskData := range description.Devices.Disks {
		// get each disk info for domain
		if diskData.Device == "disk" {
			allocation, capacity, _, err := LibvirtDomainConnexion.DomainGetBlockInfo(domain, diskData.Source.File, 0)
			if err != nil {
				Logger.Error("Cannot get domain disk information", "domain", domain.Name)
				return nil, err
			}
			humanCapacity := fmt.Sprintf("%.3f GB", internal.BytesToGiB(capacity))
			humanAllocation := fmt.Sprintf("%.3f GB", internal.BytesToGiB(allocation))
			d := &machineDisk{
				Device:     diskData.Target.Device,
				Capacity:   humanCapacity,
				Allocation: humanAllocation,
			}
			disks = append(disks, *d)
		}

	}

	// os info
	var osInfo string
	if description.Metadata == nil {
		osInfo = "generic"
	} else {
		osInfo = description.Metadata.LibOsInfo.Os.ID
	}

	// memory
	humanMemory := fmt.Sprintf("%.3f GB", internal.KibToGiB(description.Memory.Value))

	// summary
	return &machineDescription{
		Hostname:     description.Name,
		OSInfo:       osInfo,
		Architecture: description.OS.Type.Arch,
		Memory:       humanMemory,
		Vcpu:         uint(description.Vcpu.Value),
		Interfaces:   interfaces,
		Disks:        disks,
	}, nil
}
