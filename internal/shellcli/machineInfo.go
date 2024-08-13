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

// XMLDomainDescription
// Assuming you have the XML description already obtained from the domain object
// Parse the XML description to extract network interface information
// Here's an example of a struct that can be used to unmarshal the network interface information
//
// <?xml version="1.0" encoding="UTF-8" standalone="no"?>
// <domain id="1" type="kvm">
//
//	<name>debian12</name>
//	<uuid>716f0ab7-6382-4503-bdc3-0d5bc1765277</uuid>
//	<description>debian12</description>
//	<metadata>
//	    <libosinfo:libosinfo xmlns:libosinfo="http://libosinfo.org/xmlns/libvirt/domain/1.0">
//	        <libosinfo:os id="http://debian.org/debian/12"/>
//	    </libosinfo:libosinfo>
//	</metadata>
//	<memory unit="KiB">4194304</memory>
//	<currentMemory unit="KiB">4194304</currentMemory>
//	<vcpu placement="static">2</vcpu>
//	<os>
//	    <type arch="x86_64" machine="pc-q35-7.2">hvm</type>
//	    <boot dev="hd"/>
//	</os>
//	<devices>
//	    <emulator>/usr/bin/qemu-system-x86_64</emulator>
//	    <disk device="disk" type="file">
//	        <driver name="qemu" type="qcow2"/>
//	        <source file="/home/kaio/freyja-workspace/build/debian12/debian12_vdisk.debian12" index="2"/>
//	        <backingStore index="3" type="file">
//	            <format type="qcow2"/>
//	            <source file="/home/kaio/Images/debian12"/>
//	            <backingStore/>
//	        </backingStore>
//	        <target bus="virtio" dev="vda"/>
//	        <alias name="virtio-disk0"/>
//	        <address bus="0x04" domain="0x0000" function="0x0" slot="0x00" type="pci"/>
//	    </disk>
//	    <disk device="cdrom" type="file">
//	        <driver name="qemu" type="raw"/>
//	        <source file="/home/kaio/freyja-workspace/build/debian12/debian12_cloud_init.iso" index="1"/>
//	        <backingStore/>
//	        <target bus="sata" dev="sda"/>
//	        <readonly/>
//	        <alias name="sata0-0-0"/>
//	        <address bus="0" controller="0" target="0" type="drive" unit="0"/>
//	    </disk>
//	    <interface type="network">
//	        <mac address="52:54:00:25:77:0d"/>
//	        <source bridge="virbr0" network="default" portid="5b2b65a8-8c46-4109-9117-38e4bbef3cd6"/>
//	        <target dev="vnet0"/>
//	        <model type="virtio"/>
//	        <alias name="net0"/>
//	        <address bus="0x01" domain="0x0000" function="0x0" slot="0x00" type="pci"/>
//	    </interface>
//	</devices>
//
// </domain>
type xMLDomainDescription struct {
	// root
	XMLName    xml.Name `xml:"domain"`
	DomainType string   `xml:"type,attr"`
	DomainID   int      `xml:"id,attr"`
	Name       string   `xml:"name"`
	UUID       string   `xml:"uuid"`
	Memory     uint64   `xml:"memory"` //KiB
	Vcpu       uint     `xml:"vcpu"`
	OS         struct {
		Type struct {
			Arch string `xml:"arch,attr"`
		} `xml:"type"`
	} `xml:"os"`
	Metadata struct {
		LibOsInfo struct {
			XMLName xml.Name `xml:"libosinfo"`
			Os      struct {
				ID string `xml:"id,attr"`
			} `xml:"os"`
		} `xml:"libosinfo"`
	} `xml:"metadata"`
	Devices []struct {
		Disks      []disk            `xml:"disk"`
		Interfaces []domainInterface `xml:"interface"`
	} `xml:"devices"`
}

type disk struct {
	XMLName xml.Name `xml:"disk"`
	Device  string   `xml:"device,attr"`
	Type    string   `xml:"type,attr"`
	Source  struct {
		File string `xml:"file,attr"`
	} `xml:"source"`
	BackingStore struct {
		Type   string `xml:"type,attr"`
		Format struct {
			Type string `xml:"type,attr"`
		} `xml:"format"`
		Source struct {
			File string `xml:"file,attr"`
		} `xml:"source"`
	} `xml:"backingStore"`
	Target struct {
		Bus    string `xml:"bus,attr"`
		Device string `xml:"dev,attr"`
	} `xml:"target"`
}

type domainInterface struct {
	XMLName xml.Name `xml:"interface"`
	Type    string   `xml:"type,attr"`
	Mac     struct {
		Address string `xml:"address,attr"`
	} `xml:"mac"`
	Source struct {
		Bridge  string `xml:"bridge,attr"`
		Network string `xml:"network,attr"`
	} `xml:"source"`
	Target struct {
		Device string `xml:"dev,attr"`
	} `xml:"target"`
}

type shortDescription struct {
	Hostname     string           `yaml:"hostname"`
	OSInfo       string           `yaml:"osInfo"`
	Architecture string           `yaml:"architecture"`
	Memory       uint64           `yaml:"memory"`
	Vcpu         uint             `yaml:"vcpu"`
	Interfaces   []shortInterface `yaml:"interfaces"`
	Disks        []shortDisk
}

type shortInterface struct {
	Network      string `yaml:"network"`
	MacAddress   string `yaml:"macAddress"`
	Bridge       string `yaml:"bridge"`
	TargetDevice string `yaml:"targetDevice"`
}

type shortDisk struct {
	Device     string `yaml:"device"`
	Capacity   uint64 `yaml:"capacity"`
	Allocation uint64 `yaml:"allocation"`
}

var infoDomainName string

// commands definitions
var machineInfoCmd = &cobra.Command{
	Use:              "info",
	Short:            "Print a machine information",
	Long:             "Print a machine information using libvirt",
	TraverseChildren: true, // ensure local flags do not spread to sub commands

	Run: func(cmd *cobra.Command, args []string) {
		// logger
		Logger = internal.InitPrettyLogger()

		// domains list
		info, err := getDomainInfo(infoDomainName)
		if err != nil {
			Logger.Error("Cannot get domain info", "domain", infoDomainName)
			log.Panic(err)
		}
		output, err := yaml.Marshal(info)
		if err != nil {
			log.Panic("Cannot parse domain info in yaml format: ", err)
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

func getDomainInfo(domainName string) (*shortDescription, error) {
	// DOMAIN
	domain, err := LibvirtConnexion.DomainLookupByName(domainName)
	if err != nil {
		log.Panic(err)
	}
	// get domain xml description
	domainDesc, err := LibvirtConnexion.DomainGetXMLDesc(domain, 1)
	if err != nil {
		Logger.Error("Cannot get domain XML description", "domain", domain.Name)
		return nil, err
	}
	// parse description to xml struct
	var description xMLDomainDescription
	err = xml.Unmarshal([]byte(domainDesc), &description)
	if err != nil {
		log.Panic("Failed to unmarshal XML:", err)
	}
	// get networks info from xml struct
	var interfaces []shortInterface
	for _, ifaceData := range description.Devices[0].Interfaces {
		iface := &shortInterface{
			Network:      ifaceData.Source.Network,
			MacAddress:   ifaceData.Mac.Address,
			Bridge:       ifaceData.Source.Bridge,
			TargetDevice: ifaceData.Target.Device,
		}
		interfaces = append(interfaces, *iface)
	}
	// get disks info from xml struct
	var disks []shortDisk
	for _, diskData := range description.Devices[0].Disks {
		// get each disk info for domain
		if diskData.Device == "disk" {
			allocation, capacity, _, err := LibvirtConnexion.DomainGetBlockInfo(domain, diskData.Source.File, 0)
			if err != nil {
				log.Panic(err)
			}
			d := &shortDisk{
				Device:     diskData.Target.Device,
				Capacity:   capacity,
				Allocation: allocation,
			}
			disks = append(disks, *d)
		}

	}

	return &shortDescription{
		Hostname:     description.Name,
		OSInfo:       description.Metadata.LibOsInfo.Os.ID,
		Architecture: description.OS.Type.Arch,
		Memory:       description.Memory,
		Vcpu:         description.Vcpu,
		Interfaces:   interfaces,
		Disks:        disks,
	}, nil

}

// print using a map as input a generate json or yaml output
func printMachinesInfo(domains []libvirt.Domain) {
	// Set any additional field the developer may add
	//fields := make(map[string]interface{}, r.NumAttrs())
	//r.Attrs(func(a slog.Attr) bool {
	//	fields[a.Key] = a.Value.Any()
	//	return true
	//})
	//b, err := yaml.Marshal(fields)
	//if err != nil {
	//	// error if the values cannot be parsed in json
	//	return err
	//}
}
