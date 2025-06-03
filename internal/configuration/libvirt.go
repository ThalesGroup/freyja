package configuration

import (
	"crypto/sha256"
	b64 "encoding/base64"
	"encoding/xml"
	"fmt"
	"freyja/internal"
	"github.com/google/uuid"
	"os"
)

// **********
// DATA MODEL
// **********

type OSArch string

type MemoryUnit string

type VcpuPlacement string

type LibOsInfoId string

type OSBootDev string

type DevicesEmulator string

type DeviceDiskClass string

type DeviceDiskType string

type DeviceDiskDriverName string

type DeviceDiskDriverType string

type DeviceDiskTargetBus string

type DeviceDiskTargetDev string

type DeviceInterfaceType string

type DeviceInterfaceAddressType string

type DeviceInterfaceAddressDomain string

type DeviceInterfaceAddressBus string

type DeviceInterfaceAddressFunction string

type DeviceSerialType string

type DeviceSerialTargetType string

type DeviceSerialTargetModelName string

type DeviceConsoleType string

type DeviceConsoleTargetType string

type NetworkForwardMode string

const (
	X86OSArch OSArch = "x86_64"
)

const (
	KiBMemoryUnit MemoryUnit = "KiB"
	MiBMemoryUnit MemoryUnit = "MiB"
)

const (
	StaticVcpuPlacement VcpuPlacement = "static"
)

const (
	HdOsBootDev OSBootDev = "hd"
)

const (
	QemuX86DevicesEmulator DevicesEmulator = "/usr/bin/qemu-system-x86_64"
)

const (
	DiskDeviceType  DeviceDiskClass = "disk"
	CdromDeviceType DeviceDiskClass = "cdrom"
)

const (
	FileDeviceDiskType DeviceDiskType = "file"
)

const (
	QemuDeviceDiskDriverName DeviceDiskDriverName = "qemu"
)

const (
	QcowDeviceDiskDriverType DeviceDiskDriverType = "qcow2"
	RawDeviceDiskDriverType  DeviceDiskDriverType = "raw"
)

const (
	VirtioDeviceDiskTargetBus DeviceDiskTargetBus = "virtio"
	SataDeviceDiskTargetBus   DeviceDiskTargetBus = "sata"
	IdeDeviceDiskTargetBus    DeviceDiskTargetBus = "ide"
)

const (
	VdaDeviceDiskTargetDev DeviceDiskTargetDev = "vda"
	SdaDeviceDiskTargetDev DeviceDiskTargetDev = "sda"
	HdaDeviceDiskTargetDev DeviceDiskTargetDev = "hda"
	HdbDeviceDiskTargetDev DeviceDiskTargetDev = "hdb"
)

const (
	NetworkDeviceInterfaceType DeviceInterfaceType = "network"
)

const (
	PciDeviceInterfaceAddressType DeviceInterfaceAddressType = "pci"
)

const (
	DefaultDeviceInterfaceAddressDomain DeviceInterfaceAddressDomain = "0x0000"
)

const (
	DefaultDeviceInterfaceAddressBus DeviceInterfaceAddressBus = "0x00"
)

const (
	DefaultDeviceInterfaceAddressFunction DeviceInterfaceAddressFunction = "0x0"
)

const (
	// NetworkForwardModeNat default
	NetworkForwardModeNat NetworkForwardMode = "nat"
	// NetworkForwardModeRoute when user input includes a host's target interface for guests net
	NetworkForwardModeRoute NetworkForwardMode = "route"
)

const (
	PtyDeviceSerialType DeviceSerialType = "pty"
)

const (
	IsaDeviceSerialTargetType DeviceSerialTargetType = "isa-serial"
)

const (
	IsaDeviceSerialTargetModelName DeviceSerialTargetModelName = "isa-serial"
)

const (
	PtyDeviceConsoleType DeviceConsoleType = "pty"
)

const (
	SerialDeviceConsoleTargetType DeviceConsoleType = "serial"
)

const DefaultDomainType string = "kvm"

const DefaultDeviceInterfaceType = string(NetworkDeviceInterfaceType)

const DefaultNetworkForwardMode = string(NetworkForwardModeNat)

const DefaultInterfaceSourceBridge string = "virbr0"

const DefaultInterfaceAddressType string = string(PciDeviceInterfaceAddressType)

const DefaultInterfaceModelType string = "virtio"

const DefaultInterfaceSourceNetwork string = "default"

const DefaultOsType string = "hvm"

var DefaultDeviceInterface = XMLDomainDescriptionDevicesInterface{
	Type: DefaultDeviceInterfaceType,
	//Mac:     nil,
	Source: &XMLDomainDescriptionDevicesInterfaceSource{
		//Bridge:  DefaultInterfaceSourceBridge,
		Network: DefaultInterfaceSourceNetwork,
	},
	Model: &XMLDomainDescriptionDevicesInterfaceModel{
		Type: DefaultInterfaceModelType,
	},
}

// XMLDomainDescription
// Assuming you have the XML description already obtained from the domain object
// Parse the XML description to extract network interface information
// Here's an example of a struct that can be used to unmarshal the network interface information
//
// <domain type="kvm">
//
//	<name>debian12-manual</name>
//	<uuid>8220de9b-b004-4339-b770-cf8e312c5cb2</uuid>
//	<memory unit="MiB">4096</memory>
//	<vcpu placement="static">1</vcpu>
//	<os>
//	    <type arch='x86_64'>hvm</type>
//	    <boot dev='hd'/>
//	</os>
//	<devices>
//	    <emulator>/usr/bin/qemu-system-x86_64</emulator>
//	    <disk type="file" device="disk">
//	        <driver name="qemu" type="qcow2"/>
//	        <source file="/home/kaio/Projects/thales/freyja/test/manual/debian-12-generic-amd64.qcow2"/>
//	        <target dev="hda" bus="ide"/>
//	    </disk>
//	    <disk type="file" device="cdrom">
//	        <driver name="qemu" type="raw"/>
//	        <source file="/home/kaio/Projects/thales/freyja/test/manual/debian12-cloud-init.iso"/>
//	        <target dev="hdb" bus="ide"/>
//	        <readonly/>
//	    </disk>
//	    <interface type="network">
//	        <mac address="52:54:00:17:49:b7"/>
//	        <source network="default"/>
//	    </interface>
//	    <console type='pty'>
//	        <target type='serial'/>
//	    </console>
//	</devices>
//
// </domain>
type XMLDomainDescription struct {
	// root
	XMLName  xml.Name                      `xml:"domain"`
	Type     string                        `xml:"type,attr"`
	Name     string                        `xml:"name"`
	UUID     string                        `xml:"uuid"`
	Vcpu     *XMLDomainDescriptionVcpu     `xml:"vcpu"`
	Memory   *XMLDomainDescriptionMemory   `xml:"memory"`
	OS       *XMLDomainDescriptionOS       `xml:"os"`
	Metadata *XMLDomainDescriptionMetadata `xml:"metadata,omitempty"`
	Devices  *XMLDomainDescriptionDevices  `xml:"devices"`
}

// XMLDomainDescriptionVcpu
//
//	<vcpu placement="static">1</vcpu>
type XMLDomainDescriptionVcpu struct {
	XMLName   xml.Name `xml:"vcpu"`
	Placement string   `xml:"placement,attr"`
	Value     uint64   `xml:",chardata"`
}

// XMLDomainDescriptionMemory
// <memory unit="MiB">4096</memory>
type XMLDomainDescriptionMemory struct {
	XMLName xml.Name `xml:"memory"`
	Unit    string   `xml:"unit,attr"`
	Value   uint64   `xml:",chardata"`
}

// XMLDomainDescriptionOS
//
// <os>
//
//	<type arch="x86_64">hvm</type>
//	<boot dev="hd"/>
//
// </os>
type XMLDomainDescriptionOS struct {
	XMLName xml.Name                    `xml:"os"`
	Type    *XMLDomainDescriptionOSType `xml:"type"`
	Boot    *XMLDomainDescriptionOSBoot `xml:"boot"`
}

type XMLDomainDescriptionOSType struct {
	XMLName xml.Name `xml:"type"`
	Arch    string   `xml:"arch,attr"`
	Type    string   `xml:",chardata"`
}

type XMLDomainDescriptionOSBoot struct {
	XMLName xml.Name `xml:"boot"`
	Dev     string   `xml:"dev,attr"`
}

// XMLDomainDescriptionMetadata
//
//	<metadata>
//	    <libosinfo:libosinfo xmlns:libosinfo="http://libosinfo.org/xmlns/libvirt/domain/1.0">
//	        <libosinfo:os id="http://debian.org/debian/12"/>
//	    </libosinfo:libosinfo>
//	</metadata>
type XMLDomainDescriptionMetadata struct {
	XMLName   xml.Name                               `xml:"metadata"`
	LibOsInfo *XMLDomainDescriptionMetadataLibOsInfo `xml:"libosinfo"`
}

type XMLDomainDescriptionMetadataLibOsInfo struct {
	XMLName xml.Name                                 `xml:"libosinfo"`
	Os      *XMLDomainDescriptionMetadataLibOsInfoOs `xml:"os"`
}

type XMLDomainDescriptionMetadataLibOsInfoOs struct {
	XMLName xml.Name `xml:"os"`
	ID      string   `xml:"id,attr"`
}

// XMLDomainDescriptionDevices
//
//	<devices>
//	    <emulator>/usr/bin/qemu-system-x86_64</emulator>
//	    <disk type="file" device="disk">
//	        <driver name="qemu" type="qcow2"/>
//	        <source file="/home/kaio/Projects/thales/freyja/test/manual/debian-12-generic-amd64.qcow2"/>
//	        <target dev="hda" bus="ide"/>
//	    </disk>
//	    <disk type="file" device="cdrom">
//	        <driver name="qemu" type="raw"/>
//	        <source file="/home/kaio/Projects/thales/freyja/test/manual/debian12-cloud-init.iso"/>
//	        <target dev="hdb" bus="ide"/>
//	        <readonly/>
//	    </disk>
//	    <interface type="network">
//	        <mac address="52:54:00:17:49:b7"/>
//	        <source network="default"/>
//	    </interface>
//	    <console type='pty'>
//	        <target type='serial'/>
//	    </console>
//	</devices>
type XMLDomainDescriptionDevices struct {
	XMLName xml.Name `xml:"devices"`
	// no xml name because it is a list
	Emulator   string                                 `xml:"emulator"`
	Disks      []XMLDomainDescriptionDevicesDisk      `xml:"disk"`
	Interfaces []XMLDomainDescriptionDevicesInterface `xml:"interface"`
	Console    []XMLDomainDescriptionDevicesConsole   `xml:"console,omitempty"`
}

// XMLDomainDescriptionDevicesDisk
// TODO check the target, alias and address of the disk
type XMLDomainDescriptionDevicesDisk struct {
	XMLName      xml.Name                                     `xml:"disk"`
	Driver       *XMLDomainDescriptionDevicesDiskDriver       `xml:"driver"`
	Device       string                                       `xml:"device,attr"`
	Type         string                                       `xml:"type,attr"`
	Source       *XMLDomainDescriptionDevicesDiskSource       `xml:"source"`
	BackingStore *XMLDomainDescriptionDevicesDiskBackingStore `xml:"backingStore,omitempty"`
	Target       *XMLDomainDescriptionDevicesDiskTarget       `xml:"target"`
	Address      *XMLDomainDescriptionDevicesDiskAddress      `xml:"address,omitempty"`
}

type XMLDomainDescriptionDevicesDiskDriver struct {
	XMLName xml.Name `xml:"driver"`
	Name    string   `xml:"name,attr"`
	Type    string   `xml:"type,attr"`
}

type XMLDomainDescriptionDevicesDiskSource struct {
	XMLName xml.Name `xml:"source"`
	File    string   `xml:"file,attr"`
}

type XMLDomainDescriptionDevicesDiskBackingStore struct {
	XMLName xml.Name                                           `xml:"backingStore"`
	Type    string                                             `xml:"type,attr"`
	Format  *XMLDomainDescriptionDevicesDiskBackingStoreFormat `xml:"format"`
	Source  *XMLDomainDescriptionDevicesDiskBackingStoreSource `xml:"source"`
}

type XMLDomainDescriptionDevicesDiskBackingStoreFormat struct {
	XMLName xml.Name `xml:"format"`
	Type    string   `xml:"type,attr"`
}

type XMLDomainDescriptionDevicesDiskBackingStoreSource struct {
	XMLName xml.Name `xml:"source"`
	File    string   `xml:"file,attr"`
}

type XMLDomainDescriptionDevicesDiskTarget struct {
	XMLName xml.Name `xml:"target"`
	Bus     string   `xml:"bus,attr"`
	Device  string   `xml:"dev,attr"`
}

// XMLDomainDescriptionDevicesDiskAddress not mandatory
type XMLDomainDescriptionDevicesDiskAddress struct {
	XMLName    xml.Name `xml:"address"`
	Type       string   `xml:"type,attr"`
	Controller string   `xml:"controller,attr"`
	Bus        string   `xml:"bus,attr"`
	Target     string   `xml:"target,attr"`
	Unit       string   `xml:"unit,attr"`
}

// XMLDomainDescriptionDevicesInterface
//
//	<interface type="network">
//	    <mac address="52:54:00:25:77:0d"/>
//	    <source bridge="virbr0" network="default" portid="5b2b65a8-8c46-4109-9117-38e4bbef3cd6"/>
//	    <target dev="vnet0"/>
//	    <model type="virtio"/>
//	    <alias name="net0"/>
//	    <address slot="0x02" type="pci"/>
//	</interface>
//
// !! WARNING : the address entry is very important to push in libvirt because it defines the order
// of the interfaces mounted in the VM.
// For example, slot="0x02" will define the "enp0s2" interface,
// the slot="0x03" the interface enp0s3, and so on ...
type XMLDomainDescriptionDevicesInterface struct {
	XMLName xml.Name                                     `xml:"interface"`
	Type    string                                       `xml:"type,attr"`
	Mac     *XMLDomainDescriptionDevicesInterfaceMac     `xml:"mac"`
	Source  *XMLDomainDescriptionDevicesInterfaceSource  `xml:"source"`
	Target  *XMLDomainDescriptionDevicesInterfaceTarget  `xml:"target"`
	Model   *XMLDomainDescriptionDevicesInterfaceModel   `xml:"model"`
	Address *XMLDomainDescriptionDevicesInterfaceAddress `xml:"address"`
}

type XMLDomainDescriptionDevicesInterfaceMac struct {
	XMLName xml.Name `xml:"mac"`
	Address string   `xml:"address,attr"`
}

type XMLDomainDescriptionDevicesInterfaceSource struct {
	XMLName xml.Name `xml:"source"`
	Bridge  string   `xml:"bridge,attr,omitempty"`
	Network string   `xml:"network,attr"`
}

type XMLDomainDescriptionDevicesInterfaceTarget struct {
	XMLName xml.Name `xml:"target"`
	Device  string   `xml:"dev,attr"`
}

type XMLDomainDescriptionDevicesInterfaceModel struct {
	XMLName xml.Name `xml:"model"`
	Type    string   `xml:"type,attr"`
}

// XMLDomainDescriptionDevicesInterfaceAddress
// <address type='pci' domain='0x0000' bus='0x01' slot='0x01' function='0x0'/>
type XMLDomainDescriptionDevicesInterfaceAddress struct {
	XMLName  xml.Name `xml:"address"`
	Type     string   `xml:"type,attr"`
	Domain   string   `xml:"domain,attr"`
	Bus      string   `xml:"bus,attr"`
	Slot     string   `xml:"slot,attr"`
	Function string   `xml:"function,attr"`
}

// XMLDomainDescriptionDevicesConsole
//
//	<console type='pty'>
//	    <target type='serial'/>
//	</console>
type XMLDomainDescriptionDevicesConsole struct {
	XMLName xml.Name                                  `xml:"console"`
	Type    string                                    `xml:"type,attr"`
	Target  *XMLDomainDescriptionDevicesConsoleTarget `xml:"target"`
}

type XMLDomainDescriptionDevicesConsoleTarget struct {
	XMLName xml.Name `xml:"target"`
	Type    string   `xml:"type,attr"`
}

// XMLNetworkDescription example
// https://libvirt.org/formatnetwork.html#routed-network-config
// name is mandatory and unique
// uuid is optional
// bridge is  mandatory and should start with 'virbr'
// domain is optional to define the dns
// forward is optional
// ip ?
//
// WARNINGS:
//   - For networks with a forward mode of bridge, private, vepa, and passthrough, it is assumed that
//     the host has any necessary DNS and DHCP services already setup outside the scope of libvirt.
//
// EXAMPLE :
// <network>
//
//	<name>default</name>
//	<uuid>39d20ff2-296f-4bc5-b7c7-0ea755ab76f3</uuid>
//	<bridge name='virbr0'/>
//	<forward mode="nat"/>
//	<mac address='52:54:00:78:b0:16'/>
//	<ip address='192.168.122.1' netmask='255.255.255.0'>
//	  <dhcp>
//	    <range start='192.168.122.2' end='192.168.122.254'/>
//	  </dhcp>
//	</ip>
//
// </network>
type XMLNetworkDescription struct {
	XMLName xml.Name                      `xml:"network"`
	Name    string                        `xml:"name"`
	UUID    string                        `xml:"uuid,omitempty"`
	Forward *XMLNetworkDescriptionForward `xml:"forward,omitempty"`
	Bridge  *XMLNetworkDescriptionBridge  `xml:"bridge"`
	Mac     *XMLNetworkDescriptionMac     `xml:"mac,omitempty"`
	Ip      *XMLNetworkDescriptionIp      `xml:"ip,omitempty"`
}

type XMLNetworkDescriptionForward struct {
	XMLName xml.Name `xml:"forward"`
	Mode    string   `xml:"mode,attr"`
	Dev     string   `xml:"dev,attr,omitempty"`
}

type XMLNetworkDescriptionBridge struct {
	XMLName xml.Name `xml:"bridge"`
	Name    string   `xml:"name,attr"`
}

type XMLNetworkDescriptionMac struct {
	XMLName xml.Name `xml:"mac"`
	Address string   `xml:"address,attr,omitempty"`
}

type XMLNetworkDescriptionIp struct {
	XMLName xml.Name                     `xml:"ip"`
	Address string                       `xml:"address,attr"`
	Netmask string                       `xml:"netmask,attr"`
	Dhcp    *XMLNetworkDescriptionIpDhcp `xml:"dhcp"`
}

type XMLNetworkDescriptionIpDhcp struct {
	XMLName xml.Name                          `xml:"dhcp"`
	Range   *XMLNetworkDescriptionIpDhcpRange `xml:"range"`
}

type XMLNetworkDescriptionIpDhcpRange struct {
	XMLName xml.Name `xml:"range"`
	Start   string   `xml:"start,attr"`
	End     string   `xml:"end,attr"`
}

// **************
// IMPLEMENTATION
// **************

func generateMachineUUID(machineName string) (string, error) {
	h := sha256.New()
	_, err := h.Write([]byte(machineName))
	if err != nil {
		return "", fmt.Errorf("cannot generate a hash based on the machine hostname '%s': %w", machineName, err)
	}
	sum := h.Sum(nil)
	mID := b64.StdEncoding.EncodeToString(sum)[:16]
	// UUID is generated from the ID
	mUUIDRaw, err := uuid.FromBytes(sum[:16])
	if err != nil {
		return "", fmt.Errorf("cannot create machine UUID based on its ID '%s': %w", mID, err)
	}
	return fmt.Sprintf("%v", mUUIDRaw), nil
}

func CreateLibvirtDomainXMLDescription(cm *FreyjaConfigurationMachine, overlayFile string, cloudInitIsoFile string) ([]byte, error) {
	// TODO: how to generate the unique id ?
	//  - use '1', '2', ... and store the existing ids in metadata to make sure to not reuse an existing one ?
	//  - use a uuid ?
	//  - use a hash of the hostname ? -> choice 1
	mName := cm.Hostname
	// ID is the fnv hash of the machine hostname
	mUUID, err := generateMachineUUID(mName)
	if err != nil {
		return nil, fmt.Errorf("failed to generate machine UUID: %w", err)
	}
	// memory
	mMemory := XMLDomainDescriptionMemory{
		Unit:  string(MiBMemoryUnit),
		Value: uint64(cm.Memory),
	}
	// cpu
	//mVcpu := cm.Vcpu
	mVcpu := XMLDomainDescriptionVcpu{
		Placement: string(StaticVcpuPlacement),
		Value:     uint64(cm.Vcpu),
	}
	// OS description
	Arch := string(X86OSArch)
	mOSType := XMLDomainDescriptionOSType{
		Type: DefaultOsType,
		Arch: Arch,
	}
	mOSBoot := XMLDomainDescriptionOSBoot{
		Dev: string(HdOsBootDev),
	}
	mOS := XMLDomainDescriptionOS{
		Type: &mOSType,
		Boot: &mOSBoot,
	}
	// metadata
	// TODO : find a way to get the proper os type from golang
	//  for now, no metadata for os info part

	// live image device
	//    <disk device="disk" type="file">
	//      <driver name="qemu" type="qcow2"/>
	//      <source file="/home/user/.freyja/machines/vm1/overlay-image.qcow2"/>
	//      <backingStore type="file">
	//        <format type="qcow2"/>
	//        <source file="/tmp/debian-12-generic-amd64.qcow2"/>
	//      </backingStore>
	//      <target bus="ide" dev="hda"/>
	//    </disk>
	//    <disk device="cdrom" type="file">
	//      <driver name="qemu" type="raw"/>
	//      <source file="/home/user/.freyja/machines/vm1/cloud-init.iso"/>
	//      <target bus="ide" dev="hdb"/>
	//    </disk>
	//
	// For the first boot, we directly create a backing storage as the root image and use
	// the overlay file as the source image of the machine.
	// It acts like we already did a snapshot.
	// It allows multiple machines to use the same root image and boot on their own disk overlay.
	liveImageDevice := XMLDomainDescriptionDevicesDisk{
		Device: string(DiskDeviceType),
		Type:   string(FileDeviceDiskType),
		Driver: &XMLDomainDescriptionDevicesDiskDriver{
			Name: string(QemuDeviceDiskDriverName),
			Type: string(QcowDeviceDiskDriverType),
		},
		// !!! WARNING : The order of disk image chain in libvirt is :
		//    <disk>
		//      <source file="snapshot2"/>
		//      <backingStore>
		//        <source file="snapshot1"/>
		//        <backingStore>
		//          <source file="root-image"/>
		//        </backingStore>
		//      </backingStore>
		//    </disk>
		Source: &XMLDomainDescriptionDevicesDiskSource{
			File: os.ExpandEnv(overlayFile),
		},
		BackingStore: &XMLDomainDescriptionDevicesDiskBackingStore{
			Type: string(FileDeviceDiskType),
			Format: &XMLDomainDescriptionDevicesDiskBackingStoreFormat{
				Type: string(QcowDeviceDiskDriverType),
			},
			Source: &XMLDomainDescriptionDevicesDiskBackingStoreSource{
				File: cm.Image,
			},
		},
		Target: &XMLDomainDescriptionDevicesDiskTarget{
			// Only works for 'ide' bus type
			Bus:    string(IdeDeviceDiskTargetBus),
			Device: string(HdaDeviceDiskTargetDev),
		},
	}

	// cloud init raw iso device
	//	    <disk device="cdrom" type="file">
	//	        <driver name="qemu" type="raw"/>
	//	        <source file="/home/kaio/freyja-workspace/build/debian12/debian12_cloud_init.iso" index="1"/>
	//	        <backingStore/>
	//	        <target bus="sata" dev="sda"/>
	//	        <readonly/>
	//	        <alias name="sata0-0-0"/>
	//	        <address bus="0" controller="0" target="0" type="drive" unit="0"/>
	//	    </disk>
	cloudInitIsoDevice := XMLDomainDescriptionDevicesDisk{
		Device: string(CdromDeviceType),
		Type:   string(FileDeviceDiskType),
		Driver: &XMLDomainDescriptionDevicesDiskDriver{
			Name: string(QemuDeviceDiskDriverName),
			Type: string(RawDeviceDiskDriverType),
		},
		Source: &XMLDomainDescriptionDevicesDiskSource{
			File: cloudInitIsoFile,
		},
		Target: &XMLDomainDescriptionDevicesDiskTarget{
			// Only works for 'ide' bus type
			Bus:    string(IdeDeviceDiskTargetBus),
			Device: string(HdbDeviceDiskTargetDev),
		},
	}

	// network device
	//        <interface type="network">
	//            <mac address="52:54:00:17:49:b7"/>
	//            <source network="default"/>
	//            //<address type='pci' domain='0x0000' bus='0x01' slot='0x01' function='0x0'/>
	//        </interface>
	var networkInterfaceDevices []XMLDomainDescriptionDevicesInterface
	if len(cm.Networks) > 0 {
		// user defined networks
		networkInterfaceDevices = make([]XMLDomainDescriptionDevicesInterface, len(cm.Networks))
		for i, network := range cm.Networks {
			networkSlot, err := cm.GetLibvirtNetworkAddressSlot(network.Name)
			if err != nil {
				return nil, fmt.Errorf("cannot get libvirt network slot for network '%s': %w", network.Name, err)
			}
			networkInterfaceDevice := XMLDomainDescriptionDevicesInterface{
				Type: string(NetworkDeviceInterfaceType),
				Source: &XMLDomainDescriptionDevicesInterfaceSource{
					//Bridge:  DefaultInterfaceSourceBridge,
					Network: network.Name,
				},
				Address: &XMLDomainDescriptionDevicesInterfaceAddress{
					Type: DefaultInterfaceAddressType,
					// WARNING 1
					// it is very important to sort the interface slots for the pci address
					// indeed, it allows us to control the name of the interface in the machine, which is
					// crucial for cloud init to use them in the configuration.
					// For example, if the slot is 0x02, the interface will be called enp0s2, if the slot
					// is 0x03, the interface will be called enp0s3, and so on ...
					// here, slots 0 and 1 are reserved by libvirt.
					// We have to use slots 2 and above
					// WARNING 2
					// Qemu addresses are expressed in format '0000:01:00.0' which correspond to :
					// <address type='pci' domain='0x0000' bus='0x01' slot='0x00' function='0x0'/>
					// with the same order
					// WARNING 3
					// Some addresses must not be used to avoid conflict with hardcoded ones in Qemu
					// For the x86_64 architecture's I440FX-based machine types the following devices are hard coded into QEMU and can't be moved or eliminated:
					//0000:00:00.0	Host bridge
					//0000:00:01.0	ISA bridge
					//0000:00:01.1	primary IDE controller
					//0000:00:01.2	PIIX3 USB controller
					//0000:00:01.3	ACPI (power management) and SMBus controller
					//0000:00:02.0  primary video card
					// For the x86_64 architecture's Q35-based machine types the following devices are hard coded into QEMU and can't be moved or eliminated:
					//0000:00:00.0	Host bridge
					//0000:00:1f.2	primary SATA controller
					//0000:00:1f.0	ISA bridge
					//0000:00:1f.3	SMBus
					//0000:00:1a.0	USB2 controller 1
					//0000:00:1b.0	ICH9 sound chip
					//0000:00:1d.0	USB2 controller 2
					// WARNING 4
					// to avoid any collision with addresses, we will use addresses starting from
					// 0000:01:01.0 for machine interfaces, resulting in the configuration :
					// <address type='pci' domain='0x0000' bus='0x01' slot='0x01' function='0x0'/>
					// and we will only increase the slot number per interface
					Domain:   string(DefaultDeviceInterfaceAddressDomain),
					Bus:      string(DefaultDeviceInterfaceAddressBus),
					Slot:     networkSlot,
					Function: string(DefaultDeviceInterfaceAddressFunction),
				},
				//Target: XMLDomainDescriptionDevicesInterfaceTarget{}, // provide if user conf specifies a host interface
				//Target: nil,
				//Model: &XMLDomainDescriptionDevicesInterfaceModel{
				//	Type: DefaultInterfaceModelType,
				//},
			}
			if network.Mac != "" {
				networkInterfaceDevice.Mac = &XMLDomainDescriptionDevicesInterfaceMac{
					Address: network.Mac,
				}
			}
			networkInterfaceDevices[i] = networkInterfaceDevice
		}
	} else {
		// if no network define in user conf input, stick with the default one
		networkInterfaceDevices = []XMLDomainDescriptionDevicesInterface{DefaultDeviceInterface}
	}

	// console device for graphical debug
	//	        <console type='pty'>
	//	          <target type='serial' port='0'/>
	//	        </console>
	consoleDevice := XMLDomainDescriptionDevicesConsole{
		Type: string(PtyDeviceConsoleType),
		Target: &XMLDomainDescriptionDevicesConsoleTarget{
			Type: string(SerialDeviceConsoleTargetType),
		},
	}

	// xml
	deviceDisks := []XMLDomainDescriptionDevicesDisk{liveImageDevice, cloudInitIsoDevice}
	xmlDescription := XMLDomainDescription{
		Type:   DefaultDomainType,
		Name:   mName,
		UUID:   mUUID,
		Vcpu:   &mVcpu,
		Memory: &mMemory,
		OS:     &mOS,
		Devices: &XMLDomainDescriptionDevices{
			Emulator:   string(QemuX86DevicesEmulator),
			Disks:      deviceDisks,
			Interfaces: networkInterfaceDevices,
			Console:    []XMLDomainDescriptionDevicesConsole{consoleDevice},
		},
	}
	return xml.Marshal(xmlDescription)
}

// CreateLibvirtNetworkXMLDescription
// to create a network with existing routed interfaces on host :
// https://libvirt.org/formatnetwork.html#routed-network-config
// Minimalist possible configuration is :
// <network>
//
//	<name>default</name>
//
// </network>
//
// A more detailed configuration is :
// <network>
//
//	<name>default</name>
//	<forward mode="nat"/>
//	<ip address="192.168.122.1" netmask="255.255.255.0">
//	  <dhcp>
//	    <range start="192.168.122.3" end="192.168.122.254"/>
//	  </dhcp>
//	</ip>
//
// </network>
func CreateLibvirtNetworkXMLDescription(networkConfiguration FreyjaConfigurationNetwork) (data []byte, err error) {
	gateway, netmask, dhcpStart, dhcpEnd, err := internal.CalculateSubnetInfo(networkConfiguration.CIDR)
	xmlNetworkDescriptionIp := &XMLNetworkDescriptionIp{
		Address: gateway,
		Netmask: netmask,
		Dhcp: &XMLNetworkDescriptionIpDhcp{
			Range: &XMLNetworkDescriptionIpDhcpRange{
				Start: dhcpStart,
				End:   dhcpEnd,
			},
		},
	}

	xmlNetworkDescription := XMLNetworkDescription{
		Name: networkConfiguration.Name,
		Forward: &XMLNetworkDescriptionForward{
			Mode: string(NetworkForwardModeNat),
		},
		Ip: xmlNetworkDescriptionIp,
	}

	//UUID:    internal.GenerateUUID(),

	// set mac address only if provided
	// otherwise, libvirt will deliver one
	//if networkConfiguration.Mac != "" {
	//	xmlNetworkDescription.Mac = &XMLNetworkDescriptionMac{
	//		Address: networkConfiguration.Mac,
	//	}
	//}

	return xml.Marshal(xmlNetworkDescription)
}

// DumpNetworkConfig writes the xml network description for libvirt on disk
func DumpNetworkConfig(xmlNetworkDescription []byte, path string) (err error) {
	// re-create the file to inject data
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("cannot create XML network description'%s' : %w", path, err)
	}
	// write the xml description to location
	if _, err = file.Write(xmlNetworkDescription); err != nil {
		return fmt.Errorf("could not write XML network description to '%s': %w", path, err)
	}
	return nil
}
