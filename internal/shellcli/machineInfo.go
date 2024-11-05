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

type shortDescription struct {
	Hostname     string            `yaml:"hostname"`
	OSInfo       string            `yaml:"osInfo"`
	Architecture string            `yaml:"architecture"`
	Memory       string            `yaml:"memory"` // "<amount> MB"
	Vcpu         uint              `yaml:"vcpu"`
	Interfaces   []shortInterfaces `yaml:"interfaces"`
	Disks        []shortDisk
}

type shortInterfaces struct {
	TargetDevice string                  `yaml:"targetDevice"`
	Network      shortNetworkDescription `yaml:"network"`
}

type shortDisk struct {
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

func getDomainShortDescription(domain libvirt.Domain, description *internal.XMLDomainDescription) (*shortDescription, error) {
	// get networks info from xml struct
	var interfaces []shortInterfaces
	for _, ifaceData := range description.Devices.Interfaces {
		networkInfo, err := getNetworkShortDescription(ifaceData.Source.Network)
		if err != nil {
			Logger.Error("Cannot get domain network information", "domain", domain.Name, "network", ifaceData.Source.Network)
			return nil, err
		}
		iface := &shortInterfaces{
			TargetDevice: ifaceData.Target.Device,
			Network:      *networkInfo,
		}
		interfaces = append(interfaces, *iface)
	}
	// get disks info from xml struct
	var disks []shortDisk
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
			d := &shortDisk{
				Device:     diskData.Target.Device,
				Capacity:   humanCapacity,
				Allocation: humanAllocation,
			}
			disks = append(disks, *d)
		}

	}

	humanMemory := fmt.Sprintf("%.3f GB", internal.KibToGiB(description.Memory.Value))
	return &shortDescription{
		Hostname:     description.Name,
		OSInfo:       description.Metadata.LibOsInfo.Os.ID,
		Architecture: description.OS.Type.Arch,
		Memory:       humanMemory,
		Vcpu:         description.Vcpu,
		Interfaces:   interfaces,
		Disks:        disks,
	}, nil
}
