package shellcli

import (
	"encoding/xml"
	"fmt"
	"freyja/internal"
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
		domain, err := LibvirtConnexion.DomainLookupByName(infoDomainName)
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

func getDomainXMLDescription(domain libvirt.Domain) (*internal.XMLDomainDescription, error) {
	// DOMAIN
	domainName := domain.Name
	// get domain xml description
	domainDesc, err := LibvirtConnexion.DomainGetXMLDesc(domain, 1)
	if err != nil {
		Logger.Error("Cannot get domain XML description", "domain", domainName)
		return nil, err
	}
	// parse description to xml struct
	var description internal.XMLDomainDescription
	if err := xml.Unmarshal([]byte(domainDesc), &description); err != nil {
		Logger.Error("Cannot unmarshal domain XML description", "domain", domainName)
		return nil, err
	}

	return &description, err
}

func getDomainMacIP(domain libvirt.Domain, targetHostInterface string) (mac string, ip string, err error) {
	// get ip from domain
	// could also use ARP : domIntAddr, err := LibvirtConnexion.DomainInterfaceAddresses(domain, uint32(libvirt.DomainInterfaceAddressesSrcArp), 0)
	domIntAddrs, err := LibvirtConnexion.DomainInterfaceAddresses(domain, uint32(libvirt.DomainInterfaceAddressesSrcLease), 0)
	if err != nil {
		return "", "", fmt.Errorf("cannot find '%s' domain interface from device '%s': %w", domain.Name, targetHostInterface, err)
	}
	for _, inter := range domIntAddrs {
		if inter.Name == targetHostInterface {
			// TODO investigate in case of multiple interfaces per device
			mac = inter.Hwaddr[0]
			ip = inter.Addrs[0].Addr
		}
	}

	// handle empty ip
	if ip == "" {
		ip = "inactive"
	}

	return
}

func getDomainShortDescription(domain libvirt.Domain, description *internal.XMLDomainDescription) (*machineDescription, error) {
	// get networks info from xml struct
	var interfaces []machineInterface
	for _, ifaceData := range description.Devices.Interfaces {
		hostInterface := ifaceData.Target.Device
		mac, ip, err := getDomainMacIP(domain, hostInterface)
		if err != nil {
			Logger.Error("Cannot get domain network information", "domain", domain.Name)
			return nil, err
		}
		iface := &machineInterface{
			HostInterface: hostInterface,
			Network:       ifaceData.Source.Network,
			Mac:           mac,
			IP:            ip,
		}
		interfaces = append(interfaces, *iface)
	}
	// get disks info from xml struct
	var disks []machineDisk
	for _, diskData := range description.Devices.Disks {
		// get each disk info for domain
		if diskData.Device == "disk" {
			allocation, capacity, _, err := LibvirtConnexion.DomainGetBlockInfo(domain, diskData.Source.File, 0)
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
