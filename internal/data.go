package internal

import "encoding/xml"

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

type DeviceSerialType string

type DeviceSerialTargetType string

type DeviceSerialTargetModelName string

type DeviceConsoleType string

type DeviceConsoleTargetType string

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

const DefaultInterfaceSourceBridge string = "virbr0"

const DefaultInterfaceModelType string = "virtio"

const DefaultInterfaceSourceNetwork string = "default"

const DefaultOsType string = "hvm"

// DefaultUserName = freyja
const DefaultUserName string = "freyja"

// DefaultUserPassword = master
const DefaultUserPassword string = "$6$GM./aNJikL/g$AR2c35i1QIaimKo/zOC/1Qg2JO65ysPPjv/leWBcgBXaxNV3V8IcgJVeTzt4VHWzcja66zsBnR1iyYtO2DPme/"

// DefaultMachineStorage = 20 GiB
const DefaultMachineStorage uint = 20

// DefaultMachineMemory = 4096 MiB
const DefaultMachineMemory uint = 4096

// DefaultMachineVcpu = 1 vcpu
const DefaultMachineVcpu uint = 1

var DefaultDeviceInterface XMLDomainDescriptionDevicesInterface = XMLDomainDescriptionDevicesInterface{
	Type: DefaultDeviceInterfaceType,
	// TODO : can the mac address remain nil ?
	//Mac:     nil,
	Source: &XMLDomainDescriptionDevicesInterfaceSource{
		Bridge:  DefaultInterfaceSourceBridge,
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
	BackingStore *XMLDomainDescriptionDevicesDiskBackingStore `xml:"backingStore"`
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
//	    <address bus="0x01" domain="0x0000" function="0x0" slot="0x00" type="pci"/>
//	</interface>
type XMLDomainDescriptionDevicesInterface struct {
	XMLName xml.Name                                    `xml:"interface"`
	Type    string                                      `xml:"type,attr"`
	Mac     *XMLDomainDescriptionDevicesInterfaceMac    `xml:"mac"`
	Source  *XMLDomainDescriptionDevicesInterfaceSource `xml:"source"`
	Target  *XMLDomainDescriptionDevicesInterfaceTarget `xml:"target"`
	Model   *XMLDomainDescriptionDevicesInterfaceModel  `xml:"model"`
}

type XMLDomainDescriptionDevicesInterfaceMac struct {
	XMLName xml.Name `xml:"mac"`
	Address string   `xml:"address,attr"`
}

type XMLDomainDescriptionDevicesInterfaceSource struct {
	XMLName xml.Name `xml:"source"`
	Bridge  string   `xml:"bridge,attr"`
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
// CREATE EXAMPLE :
// <network>
//
//	<name>default</name>
//	<bridge name="virbr0"/>
//	<forward mode="nat"/>
//	<ip address="192.168.122.1" netmask="255.255.255.0">
//	  <dhcp>
//	    <range start="192.168.122.2" end="192.168.122.254"/>
//	  </dhcp>
//	</ip>
//	<ip family="ipv6" address="2001:db8:ca2:2::1" prefix="64"/>
//
// </network>
// INFO EXAMPLE :
// <network>
//
//	<name>default</name>
//	<uuid>39d20ff2-296f-4bc5-b7c7-0ea755ab76f3</uuid>
//	<forward mode='nat'>
//	  <nat>
//	    <port start='1024' end='65535'/>
//	  </nat>
//	</forward>
//	<bridge name='virbr0' stp='on' delay='0'/>
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
	UUID    string                        `xml:"uuid"`
	Forward *XMLNetworkDescriptionForward `xml:"forward"`
	Bridge  *XMLNetworkDescriptionBridge  `xml:"bridge"`
	Mac     *XMLNetworkDescriptionMac     `xml:"mac"`
	Ip      *XMLNetworkDescriptionIp      `xml:"ip"`
}

type XMLNetworkDescriptionForward struct {
	XMLName xml.Name `xml:"forward"`
	Mode    string   `xml:"mode,attr"`
}

type XMLNetworkDescriptionBridge struct {
	XMLName xml.Name `xml:"bridge"`
	Name    string   `xml:"name,attr"`
}

type XMLNetworkDescriptionMac struct {
	XMLName xml.Name `xml:"mac"`
	Address string   `xml:"address,attr"`
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

// ConfigurationData is the base model for freyja configuration parameters
// Example :
// ---
// version: v0.1.0-beta
// machines:
//   - image: "/tmp/CentOS-Stream-GenericCloud-8-20210603.0.x86_64.qcow2" # MANDATORY
//     os: "centos8" # MANDATORY
//     hostname: "vm1" # MANDATORY, MUST NOT contain underscores
//     networks: # MANDATORY, at least one
//   - name: "ctrl-plane"
//     mac: "52:54:02:aa:bb:cc"
//     interface: "vnet0"
//   - name: "data-plane"
//     mac: "52:54:02:aa:bb:cd"
//     users: # MANDATORY
//   - name: "sam" # MANDATORY
//     password: "$6$6LEpjaxLaT/pu5$wwHsyMlZ2JpHObVJBKGbZUmR5oJ4GocH0zRQYKAuWEwq9ifG4N3Vi/E3ZXTj1bK.QQrOmttA7zIZUIEBaU6Yx." # MANDATORY, here 'master'
//     keys: # Optional, default '$HOME/.ssh/id_rsa.pub'
//   - "/tmp/freyja-unit-test/config/sam.pub"
//   - "/tmp/freyja-unit-test/config/ext.pub"
//     groups: ["sudo", "freyja"]
//   - name: "frodo" # MANDATORY
//     password: "$6$6LEpjaxLaT/pu5$wwHsyMlZ2JpHObVJBKGbZUmR5oJ4GocH0zRQYKAuWEwq9ifG4N3Vi/E3ZXTj1bK.QQrOmttA7zIZUIEBaU6Yx." # MANDATORY, here 'master'ub"
//     storage: 100 # Optional, default '30'
//     memory: 8192 # Optional, default '4096'
//     vcpu: 4 # Optional, default '2'
//     packages: [ "curl", "net-tools" ]
//     cmd:
//   - "echo 'hello world !' > /tmp/test.txt"
//   - "cat /tmp/test.txt"
//     files:
//   - source: "/tmp/freyja-unit-test/config/hello.txt"
//     destination: "/root/hello.txt"
//     permissions : "0700"
//     owner: "root:freyja"
//   - source: "/tmp/freyja-unit-test/config/world.txt"
//     destination: "/home/sam/world.txt"
type ConfigurationData struct {
	Version  string                 `yaml:"version"`
	Machines []ConfigurationMachine `yaml:"machines"`
}

// ConfigurationMachine is the configuration model for libvirt guest parameters
type ConfigurationMachine struct {
	// MANDATORY
	Image    string `yaml:"image"`    // Qcow2 image file path on host
	Os       string `yaml:"os"`       // os type in libosinfo
	Hostname string `yaml:"hostname"` // domain name in libvirt
	// optional
	Networks []ConfigurationNetwork `yaml:"networks"`
	Users    []ConfigurationUser    `yaml:"users"`
	Storage  uint                   `yaml:"storage"` // GiB
	Memory   uint                   `yaml:"memory"`  // MiB
	Vcpu     uint                   `yaml:"vcpu"`
	Packages []string               `yaml:"packages"`
	Cmd      []string               `yaml:"cmd"`
	Files    []ConfigurationFile    `yaml:"files"`
	Update   bool                   `yaml:"update"`
	Reboot   bool                   `yaml:"reboot"`
}

type ConfigurationNetwork struct {
	Name      string `yaml:"name"`
	Mac       string `yaml:"mac"`
	Interface string `yaml:"interface"`
}

type ConfigurationUser struct {
	Name     string   `yaml:"name"`
	Password string   `yaml:"password"`
	Sudo     bool     `yaml:"sudo"`
	Groups   []string `yaml:"groups"`
	Keys     []string `yaml:"keys"`
}

type ConfigurationFile struct {
	Source      string `yaml:"source"`
	Destination string `yaml:"destination"`
	Permissions string `yaml:"permissions"`
	Owner       string `yaml:"owner"`
}

// CloudInitMetadata is the user metadata configuration specifications
type CloudInitMetadata struct {
	InstanceID    string `yaml:"instance"`       // = machine hostname
	LocalHostname string `yaml:"local-hostname"` // = machine hostname
}

// CloudInitUserData is the user configuration specification with cloud-init specifications.
// The format used here is YAML.
// This format is mandatory to generate the ISO9660 disk image file for machine provisioning.
// Compliant with cloud-init version 24.2
// Example from https://cloudinit.readthedocs.io/en/latest/reference/examples.html :
// Specs from https://cloudinit.readthedocs.io/en/latest/reference/modules.html
//
// #cloud-config
// hostname: debian12
// output:
//
//	all: ">> /var/log/cloud-init.log"
//
// users:
//   - name: "freyja"
//     sudo: [ 'ALL=(ALL) NOPASSWD:ALL' ]
//     lock_passwd: false
//     shell: /bin/bash
//     passwd: "$6$GM./aNJikL/g$AR2c35i1QIaimKo/zOC/1Qg2JO65ysPPjv/leWBcgBXaxNV3V8IcgJVeTzt4VHWzcja66zsBnR1iyYtO2DPme/"
//     groups: sudo
//     ssh_authorized_keys:
//   - ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAILISxfJd/91TY9DH97/Y6t2zejV8p0x7L4Ygjy45iMPp kaio@kaio-host
//
// ssh_authorized_keys:
//   - ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAILISxfJd/91TY9DH97/Y6t2zejV8p0x7L4Ygjy45iMPp kaio@kaio-host
//
// package_update: False
// package_upgrade: False
// packages:
//   - vim
//   - git
//
// write_files:
//   - content: aGVsbG8gd29ybGQhCg==
//     encoding: base64
//     path: /home/freyja/hello.txt
//     permissions: 0760
//     owner: freyja:freyja
//
// runcmd:
//   - systemctl stop network && systemctl start network
//
// # if reboot needed
// power_state:
//
//	mode: reboot
//	message: First reboot
//	timeout: 30
//	condition: True
type CloudInitUserData struct {
	Hostname       string                       `yaml:"hostname"` // MANDATORY
	Output         *CloudInitUserDataOutput     `yaml:"output"`
	Users          []CloudInitUserDataUser      `yaml:"users"`                 // MANDATORY ??
	PackageUpdate  bool                         `yaml:"package_update"`        // default : false
	PackageUpgrade bool                         `yaml:"package_upgrade"`       // default : false
	Packages       []string                     `yaml:"packages,omitempty"`    // default: empty
	WriteFiles     []CloudInitUserDataFiles     `yaml:"write_files,omitempty"` // default: empty
	RunCmd         []string                     `yaml:"runcmd,omitempty"`      // default: empty
	PowerState     *CloudInitUserDataPowerState `yaml:"power_state,omitempty"` // default: nil
}

const CloudInitUserDataHeader string = "#cloud-config\n"

const CloudInitUserDataOutputAllString = ">> /var/log/cloud-init.log"

type CloudInitUserDataOutput struct {
	All string `yaml:"all"`
}

const CloudInitUserDataUserSudoString string = "ALL=(ALL) NOPASSWD:ALL"

func GetCloudInitUserDataUserSudoStringConst() []string {
	return []string{CloudInitUserDataUserSudoString}
}

const CloudInitUserDataUserShellString = "/bin/bash"

// CloudInitUserDataUser
// name: "freyja"
// sudo: [ 'ALL=(ALL) NOPASSWD:ALL' ]
// lock_passwd: false
// shell: /bin/bash
// passwd: "$6$GM./aNJikL/g$AR2c35i1QIaimKo/zOC/1Qg2JO65ysPPjv/leWBcgBXaxNV3V8IcgJVeTzt4VHWzcja66zsBnR1iyYtO2DPme/"
// groups: sudo
// ssh_authorized_keys:
// - ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAILISxfJd/91TY9DH97/Y6t2zejV8p0x7L4Ygjy45iMPp kaio@kaio-host
type CloudInitUserDataUser struct {
	Name              string   `yaml:"name"`           // MANDATORY
	Sudo              []string `yaml:"sudo,omitempty"` // example if sudo : [ 'ALL=(ALL) NOPASSWD:ALL' ]
	LockPasswd        bool     `yaml:"lock_passwd"`    // default: false
	Shell             string   `yaml:"shell"`          // default: /bin/bash
	Passwd            string   `yaml:"passwd,flow"`
	Groups            string   `yaml:"groups,omitempty"` // comma-separated string, ex: freyja, libvirt, sudo
	SshAuthorizedKeys []string `yaml:"ssh_authorized_keys,omitempty"`
}

const CloudInitUserDataFilesEncoding = "base64"

// CloudInitUserDataFiles
// content: aGVsbG8gd29ybGQhCg==
// encoding: base64
// path: /home/freyja/hello.txt
// permissions: 0760
// owner: freyja:freyja
type CloudInitUserDataFiles struct {
	Content     string `yaml:"content"`
	Encoding    string `yaml:"encoding"` // = base64
	Path        string `yaml:"path"`
	Permissions string `yaml:"permissions,omitempty"`
	Owner       string `yaml:"owner,omitempty"`
}

const CloudInitUserPowerStateMode string = "reboot"
const CloudInitUserPowerStateMessage string = "First reboot"
const CloudInitUserPowerStateTimeout = uint(30)
const CloudInitUserPowerStateCondition bool = true

// CloudInitUserDataPowerState if reboot is needed. all are default values
//
//	mode: reboot
//	message: First reboot
//	timeout: 30
//	condition: True
type CloudInitUserDataPowerState struct {
	Mode      string `yaml:"mode"`      // = reboot
	Message   string `yaml:"message"`   // = First reboot
	Timeout   uint   `yaml:"timeout"`   // = 30 (seconds)
	Condition bool   `yaml:"condition"` // = True
}
