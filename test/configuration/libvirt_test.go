package configuration

import (
	"encoding/xml"
	"freyja/internal/configuration"
	"freyja/internal/shellcli"
	internalTest "freyja/test"
	"path"
	"testing"
)

const FreyjaUnitTestUtilsDir = internalTest.FreyjaUnitTestDir + "/libvirt"

// **************************
// DOMAIN CONFIGURATION TEST
// **************************

const testValidCompleteConfiguration string = "../configuration/static/complete_conf.yaml"

func TestCreateLibvirtDomainXMLDescription(t *testing.T) {
	// build freyja config
	config := internalTest.BuildConfig(testValidCompleteConfiguration)
	// build libvirt XML domain description from freyja config
	// the first machine sets all the available values in configuration
	machine := config.Machines[0]
	overlayTestFile := path.Join(FreyjaUnitTestUtilsDir, shellcli.BackingImageFilename)
	cloudInitTestFile := path.Join(FreyjaUnitTestUtilsDir, "cloud-init.iso")
	contentBytes, err := configuration.CreateLibvirtDomainXMLDescription(&machine, overlayTestFile, cloudInitTestFile)
	if err != nil {
		t.Fatal(err)
	}
	// unmarshal xml config
	var description configuration.XMLDomainDescription
	if err := xml.Unmarshal(contentBytes, &description); err != nil {
		t.Errorf("cannot unmarshall xml description")
		t.FailNow()
	}
	// test domain
	if description.Type != string(configuration.DefaultDomainType) {
		t.Errorf("domain Type should be %s", configuration.DefaultDomainType)
		t.Fail()
	}
	// test name
	if description.Name != "vm1" {
		t.Errorf("domain name should be vm1")
		t.Fail()
	}
	// test vcpus
	if description.Vcpu.Placement != string(configuration.DefaultDomainVcpuPlacement) {
		t.Errorf("domain vcpu Placement should be %s", configuration.DefaultDomainVcpuPlacement)
		t.Fail()
	}
	if description.Vcpu.Value != 4 {
		t.Errorf("domain vcpu should be 4")
		t.Fail()
	}
	// test memory
	if description.Memory.Unit != string(configuration.DefaultDomainMemoryUnit) {
		t.Errorf("domain memory should be %s", configuration.DefaultDomainMemoryUnit)
		t.Fail()
	}
	if description.Memory.Value != 8192 {
		t.Errorf("domain memory should be 8192")
		t.Fail()
	}
	// test OS description
	if description.OS.Type.Arch != string(configuration.DefaultOsArch) {
		t.Errorf("domain os arch should be %s", configuration.DefaultOsArch)
		t.Fail()
	}
	if description.OS.Type.Type != string(configuration.DefaultOsType) {
		t.Errorf("domain os type should be %s", configuration.DefaultOsType)
		t.Fail()
	}
	// test devices
	if description.Devices.Emulator != string(configuration.DefaultDevicesEmulator) {

		t.Errorf("devices emulator should be %s", configuration.DefaultDevicesEmulator)
		t.Fail()
	}
	diskDevices := description.Devices.Disks
	if len(diskDevices) != 2 {
		t.Errorf("number of disk devices should be 2")
		t.FailNow()
	}
	// test live image device
	liveImageDevice := diskDevices[0]
	if liveImageDevice.Device != string(configuration.DefaultDiskDeviceClass) {
		t.Errorf("live image device class should be %s", configuration.DefaultDiskDeviceClass)
		t.Fail()
	}
	if liveImageDevice.Type != string(configuration.DefaultDeviceDiskType) {
		t.Errorf("live image device type should be %s", configuration.DefaultDeviceDiskType)
		t.Fail()
	}
	liveImageDeviceDriver := liveImageDevice.Driver
	if liveImageDeviceDriver.Name != string(configuration.DeviceDiskDriverNameQemu) {
		t.Errorf("live image driver name should be %s", configuration.DeviceDiskDriverNameQemu)
		t.Fail()
	}
	if liveImageDeviceDriver.Type != string(configuration.DeviceDiskDriverTypeQcow) {
		t.Errorf("live image driver type should be %s", configuration.DeviceDiskDriverTypeQcow)
		t.Fail()
	}
	if liveImageDevice.Source.File != overlayTestFile {
		t.Errorf("live image source file should be %s", overlayTestFile)
		t.Fail()
	}
	liveImageDeviceBackingStore := liveImageDevice.BackingStore
	if liveImageDeviceBackingStore.Type != string(configuration.DefaultDeviceDiskType) {
		t.Errorf("live image backing store type should be %s", configuration.DefaultDeviceDiskType)
		t.Fail()
	}
	if liveImageDeviceBackingStore.Format.Type != string(configuration.DeviceDiskDriverTypeQcow) {
		t.Errorf("live image backing store type should be %s", configuration.DeviceDiskDriverTypeQcow)
		t.Fail()
	}
	if liveImageDeviceBackingStore.Source.File != machine.Image {
		t.Errorf("live image source file should be %s", machine.Image)
		t.Fail()
	}
	if liveImageDevice.Target.Bus != string(configuration.DeviceDiskTargetBusIde) {
		t.Errorf("live image target bus should be %s", configuration.DeviceDiskTargetBusIde)
		t.Fail()
	}
	if liveImageDevice.Target.Device != string(configuration.DeviceDiskTargetDevHda) {
		t.Errorf("live image target device should be %s", configuration.DeviceDiskTargetDevHda)
		t.Fail()
	}
	// test cloud init iso cd live
	cloudInitCDDevice := diskDevices[1]
	if cloudInitCDDevice.Device != string(configuration.DeviceDiskClassCdrom) {
		t.Errorf("cloud-init device should be %s", configuration.DeviceDiskClassCdrom)
		t.Fail()
	}
	if cloudInitCDDevice.Type != string(configuration.DefaultDeviceDiskType) {
		t.Errorf("cloud-init device type should be %s", configuration.DefaultDeviceDiskType)
		t.Fail()
	}
	if cloudInitCDDevice.Driver.Name != string(configuration.DeviceDiskDriverNameQemu) {
		t.Errorf("cloud-init device driver name should be %s", configuration.DeviceDiskDriverNameQemu)
		t.Fail()
	}
	if cloudInitCDDevice.Driver.Type != string(configuration.DeviceDiskDriverTypeRaw) {
		t.Errorf("cloud-init device driver type should be %s", configuration.DeviceDiskDriverTypeRaw)
		t.Fail()
	}
	if cloudInitCDDevice.Source.File != cloudInitTestFile {
		t.Errorf("cloud-init device source file should be %s", cloudInitTestFile)
		t.Fail()
	}
	if cloudInitCDDevice.Target.Bus != string(configuration.DeviceDiskTargetBusIde) {
		t.Errorf("cloud-init target bus should be %s", configuration.DeviceDiskTargetBusIde)
		t.Fail()
	}
	if cloudInitCDDevice.Target.Device != string(configuration.DeviceDiskTargetDevHdb) {
		t.Errorf("cloud-init target device should be %s", configuration.DeviceDiskTargetDevHdb)
		t.Fail()
	}
	// test network interfaces
	interfaceDevices := description.Devices.Interfaces
	if len(interfaceDevices) != 2 {
		t.Errorf("number of interfaces should be 2")
		t.FailNow()
	}
	// test ctrl plane network interface's domain configuration
	ctrlNetworkConfig := machine.Networks[0]
	ctrlPlaneInterfaceDevice := interfaceDevices[0]
	if ctrlPlaneInterfaceDevice.Type != string(configuration.DeviceInterfaceTypeNetwork) {
		t.Errorf("ctrl plane interface type should be %s", configuration.DeviceInterfaceTypeNetwork)
		t.Fail()
	}
	if ctrlPlaneInterfaceDevice.Mac.Address != ctrlNetworkConfig.Mac {
		t.Errorf("ctrl plane mac address should be %s", ctrlNetworkConfig.Mac)
		t.Fail()
	}
	if ctrlPlaneInterfaceDevice.Source.Network != ctrlNetworkConfig.Name {
		t.Errorf("ctrl plane network name should be %s", ctrlNetworkConfig.Name)
		t.Fail()
	}
	ctrlPlaneInterfaceDeviceAddress := ctrlPlaneInterfaceDevice.Address
	if ctrlPlaneInterfaceDeviceAddress.Type != string(configuration.DeviceInterfaceAddressTypePci) {
		t.Errorf("ctrl plane address type should be %s", configuration.DeviceInterfaceAddressTypePci)
		t.Fail()
	}
	if ctrlPlaneInterfaceDeviceAddress.Domain != configuration.DefaultDeviceInterfaceAddressDomain {
		t.Errorf("ctrl plane address domain should be %s", configuration.DefaultDeviceInterfaceAddressDomain)
		t.Fail()
	}
	if ctrlPlaneInterfaceDeviceAddress.Bus != configuration.DefaultDeviceInterfaceAddressBus {
		t.Errorf("ctrl plane address bus should be %s", configuration.DefaultDeviceInterfaceAddressBus)
		t.Fail()
	}
	if ctrlPlaneInterfaceDeviceAddress.Slot != "0x02" {
		// here we harden the verification of the value on purpose
		t.Errorf("ctrl plane address bus should be 0x02")
		t.Fail()
	}
	if ctrlPlaneInterfaceDeviceAddress.Function != configuration.DefaultDeviceInterfaceAddressFunction {
		t.Errorf("ctrl plane address function should be %s", configuration.DefaultDeviceInterfaceAddressFunction)
		t.Fail()
	}
	// test data plane network interface's domain configuration
	dataNetworkConfig := machine.Networks[1]
	dataPlaneInterfaceDevice := interfaceDevices[1]
	if dataPlaneInterfaceDevice.Type != string(configuration.DeviceInterfaceTypeNetwork) {
		t.Errorf("ctrl plane interface type should be %s", configuration.DeviceInterfaceTypeNetwork)
		t.Fail()
	}
	if dataPlaneInterfaceDevice.Mac.Address != dataNetworkConfig.Mac {
		t.Errorf("ctrl plane mac address should be %s", dataNetworkConfig.Mac)
		t.Fail()
	}
	if dataPlaneInterfaceDevice.Source.Network != dataNetworkConfig.Name {
		t.Errorf("ctrl plane network name should be %s", dataNetworkConfig.Name)
		t.Fail()
	}
	dataPlaneInterfaceDeviceAddress := dataPlaneInterfaceDevice.Address
	if dataPlaneInterfaceDeviceAddress.Type != string(configuration.DeviceInterfaceAddressTypePci) {
		t.Errorf("ctrl plane address type should be %s", configuration.DeviceInterfaceAddressTypePci)
		t.Fail()
	}
	if dataPlaneInterfaceDeviceAddress.Domain != configuration.DefaultDeviceInterfaceAddressDomain {
		t.Errorf("ctrl plane address domain should be %s", configuration.DefaultDeviceInterfaceAddressDomain)
		t.Fail()
	}
	if dataPlaneInterfaceDeviceAddress.Bus != configuration.DefaultDeviceInterfaceAddressBus {
		t.Errorf("ctrl plane address bus should be %s", configuration.DefaultDeviceInterfaceAddressBus)
		t.Fail()
	}
	if dataPlaneInterfaceDeviceAddress.Slot != "0x03" {
		// here we harden the verification of the value on purpose
		t.Errorf("ctrl plane address bus should be 0x03")
		t.Fail()
	}
	if dataPlaneInterfaceDeviceAddress.Function != configuration.DefaultDeviceInterfaceAddressFunction {
		t.Errorf("ctrl plane address function should be %s", configuration.DefaultDeviceInterfaceAddressFunction)
		t.Fail()
	}
	// test console pty device
	consoleDevice := description.Devices.Console[0]
	if consoleDevice.Type != string(configuration.DeviceConsoleTypePty) {
		t.Errorf("console device type should be %s", configuration.DeviceConsoleTypePty)
		t.Fail()
	}
	if consoleDevice.Target.Type != string(configuration.DeviceConsoleTargetTypeSerial) {
		t.Errorf("console device target type should be %s", configuration.DeviceConsoleTargetTypeSerial)
		t.Fail()
	}

}

// **************************
// NETWORK CONFIGURATION TEST
// **************************

// TestValidDefaultNetworkConfiguration is used to test the default values in a configuration
const testValidCompleteNetworkConfiguration string = "../configuration/static/network_complete_conf.yaml"

func TestCreateLibvirtNetworkXMLDescription(t *testing.T) {
	config := internalTest.BuildConfig(testValidCompleteNetworkConfiguration)
	// <network>
	//  <name>ctrlplane</name>
	//  <ip address="192.168.123.1" netmask="255.255.255.0">
	//    <dhcp>
	//      <range start="192.168.123.2" end="192.168.123.254"/>
	//    </dhcp>
	//  </ip>
	//</network>
	network := config.Networks[0]
	contentBytes, err := configuration.CreateLibvirtNetworkXMLDescription(network)
	if err != nil {
		t.Fatal(err)
	}
	var description configuration.XMLNetworkDescription
	if err := xml.Unmarshal(contentBytes, &description); err != nil {
		t.Errorf("cannot unmarshall xml description")
		t.FailNow()
	}

	if description.Name != network.Name {
		t.Errorf("expected network name '%s' but got '%s'", network.Name, description.Name)
		t.Fail()
	}

	expectedGateway := "192.168.123.1"
	if description.Ip.Address != expectedGateway {
		t.Errorf("expected network gateway '%s' but got '%s'", description.Ip.Address, expectedGateway)
		t.Fail()
	}

	expectedNetmask := "255.255.255.0"
	if description.Ip.Netmask != expectedNetmask {
		t.Errorf("expected netmask '%s' but got '%s'", description.Ip.Address, expectedNetmask)
		t.Fail()
	}

	expectedDhcpStart := "192.168.123.2"
	if description.Ip.Dhcp.Range.Start != expectedDhcpStart {
		t.Errorf("expected dhcp start '%s' but got '%s'", description.Ip.Dhcp.Range.Start, expectedDhcpStart)
		t.Fail()
	}

	expectedDhcpEnd := "192.168.123.254"
	if description.Ip.Dhcp.Range.End != expectedDhcpEnd {
		t.Errorf("expected dhcp start '%s' but got '%s'", description.Ip.Dhcp.Range.End, expectedDhcpEnd)
		t.Fail()
	}

}
