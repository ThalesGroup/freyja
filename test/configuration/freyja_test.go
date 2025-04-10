package configuration

import (
	"freyja/internal/configuration"
	internalTest "freyja/test"
	"log"
	"testing"
)

const FreyjaUnitTestConfigDir string = internalTest.FreyjaUnitTestDir + "/freyja-config"

// TestFileEmptyConfiguration is used to test empty configuration (missing required values)
const testFileEmptyConfiguration string = "static/empty_conf.yaml"

// TestFileValidDefaultConfiguration is used to test minimal required values and default values
const testFileValidDefaultConfiguration string = "static/default_conf.yaml"

// TestFileValidDefaultFilesConfiguration is used to test minimal required values and default values
const testFileValidDefaultFilesConfiguration string = "static/default_conf_files.yaml"

// TestFileValidCompleteConfiguration is used to test all the possible values in a configuration
const testFileValidCompleteConfiguration string = "static/complete_conf.yaml"

func compareOrderedStringSlices(slice1 []string, slice2 []string) bool {
	if len(slice1) != len(slice2) {
		return false
	}
	for i := range slice1 {
		if slice1[i] != slice2[i] {
			return false
		}
	}
	return true
}

// replaceFirstConfNetwork takes the first network configuration in config and replace it by the
// given one. Useful for the unit tests.
func replaceFirstConfNetwork(c *configuration.FreyjaConfiguration, n *configuration.FreyjaConfigurationMachineNetwork) {
	c.Machines[0].Networks[0] = *n
}

// replaceFirstConfUser takes the first user configuration in config and replace it by the
// given one. Useful for the unit tests.
func replaceFirstConfUser(c *configuration.FreyjaConfiguration, u *configuration.FreyjaConfigurationUser) {
	c.Machines[0].Users[0] = *u
}

// replaceFirstConfFile takes the first file configuration in config and replace it by the
// given one. Useful for the unit tests.
func replaceFirstConfFile(c *configuration.FreyjaConfiguration, f *configuration.FreyjaConfigurationFile) {
	c.Machines[0].Files[0] = *f
}

func TestValidate(t *testing.T) {
	c := internalTest.BuildCompleteConfig(testFileValidCompleteConfiguration)
	testValidateVersion(t, c)
	testValidateNetworks(t, c)
	testValidateMachineNetwork(t, c)
	testValidateMachineUser(t, c)
	testValidateMachineFiles(t, c)
}

func testValidateVersion(t *testing.T, c *configuration.FreyjaConfiguration) {
	// invalid
	values := []string{"1", "beta", ""}
	for _, value := range values {
		c.Version = value
		if err := c.Validate(); err == nil {
			t.Logf("Version valid instead of invalid for value: %s", value)
			t.Fail()
		}
	}
	// valid
	values = []string{"v0.1.1", "0.1", "1.0", "0.1.0-beta"}
	for _, value := range values {
		c.Version = value
		if err := c.Validate(); err != nil {
			t.Logf("Version invalid instead of valid for value: %s", value)
			t.Fail()
		}
	}
}

func testValidateNetworks(t *testing.T, c *configuration.FreyjaConfiguration) {
	for i, network := range c.Networks {
		network.Name = ""
		c.Networks[i] = network
		if err := c.Validate(); err == nil {
			t.Logf("network missing name did not raised an error")
			t.Fail()
		}
		network.Name = "valid" // make it valid again

		invalidValues := []string{"", "aosidfiabjk", "10.11.12.1", "12.12.12.12.12/24"}
		for _, value := range invalidValues {
			network.CIDR = value
			c.Networks[i] = network
			if err := c.Validate(); err == nil {
				t.Logf("invalid dhcp start of range did not raised an error for value: %s", value)
				t.Fail()
			}

		}
		network.CIDR = "192.168.123.0/24" // make it valid again
		c.Networks[i] = network
	}
}

func testValidateMachineNetwork(t *testing.T, c *configuration.FreyjaConfiguration) {
	configurationNetwork := c.Machines[0].Networks[0]
	// invalid name value
	// error should be raised here
	configurationNetwork.Name = ""
	replaceFirstConfNetwork(c, &configurationNetwork)
	err := c.Validate()
	if err == nil {
		t.Logf("Name valid instead of invalid because empty")
		t.Fail()
	}
	// make the name valid for further tests
	configurationNetwork.Name = "test"
	replaceFirstConfNetwork(c, &configurationNetwork)
	// invalid mac addresses
	// errors should be raised here
	values := []string{"001b:63:84:45:e6", "01:63:84:45:a", "001b638445e6", "xx:1b:63:84:45:e6"}
	for _, value := range values {
		configurationNetwork.Mac = value
		replaceFirstConfNetwork(c, &configurationNetwork)
		err = c.Validate()
		if err == nil {
			t.Logf("Mac address valid instead of invalid for value: %s", value)
			t.Fail()
		}
	}
	// valid mac addresses
	values = []string{"00:1b:63:84:45:e6", "00-1B-63-84-45-E6"}
	for _, value := range values {
		configurationNetwork.Mac = value
		replaceFirstConfNetwork(c, &configurationNetwork)
		err = c.Validate()
		if err != nil {
			t.Logf("Mac address invalid instead of valid for value: %s", value)
			t.Fail()
		}
	}

}

func testValidateMachineUser(t *testing.T, c *configuration.FreyjaConfiguration) {
	configurationUser := c.Machines[0].Users[0]
	tempFile := internalTest.WriteTempTestFile("test-valid-user-key.pub", "config", []byte("test"))
	// invalid
	configurationUser.Keys = append(configurationUser.Keys, "dumb")
	replaceFirstConfUser(c, &configurationUser)
	if err := c.Validate(); err == nil {
		t.Logf("User valid instead of invalid for key values: %v", configurationUser.Keys)
		t.FailNow()
	}
	// valid
	configurationUser.Keys = []string{tempFile, "/dev/null"}
	replaceFirstConfUser(c, &configurationUser)
	if err := c.Validate(); err != nil {
		t.Logf("User invalid instead of valid for key values: %v", configurationUser.Keys)
		t.FailNow()
	}
}

func testValidateMachineFiles(t *testing.T, c *configuration.FreyjaConfiguration) {
	configurationFile := c.Machines[0].Files[0]

	// SOURCE
	// invalid values
	values := []string{"dumb", "/dumb/dumber.txt"}
	for _, value := range values {
		configurationFile.Source = value
		replaceFirstConfFile(c, &configurationFile)
		if err := c.Validate(); err == nil {
			t.Logf("Source file valid instead of invalid for value: %s", value)
			t.Fail()
		}
	}
	// valid values
	tempFile := internalTest.WriteTempTestFile("test-valid-file-source.txt", "config", []byte("test"))
	values = []string{tempFile, "/dev/null"}
	for _, value := range values {
		configurationFile.Source = value
		replaceFirstConfFile(c, &configurationFile)
		if err := c.Validate(); err != nil {
			t.Logf("Owner invalid instead of valid for value: %s", value)
			t.Fail()
		}
	}

	// PERMISSIONS
	// invalid values
	values = []string{"root", "0t00", "0*00", "000", "00000"}
	for _, value := range values {
		configurationFile.Permissions = value
		replaceFirstConfFile(c, &configurationFile)
		if err := c.Validate(); err == nil {
			t.Logf("file permissions valid instead of invalid for value: %s", value)
			t.Fail()
		}
	}
	// valid values
	valid := []string{"0010", "0777"}
	for _, value := range valid {
		configurationFile.Permissions = value
		replaceFirstConfFile(c, &configurationFile)
		if err := c.Validate(); err != nil {
			t.Logf("file permissions invalid instead of valid for value: %s", value)
			t.Fail()
		}
	}

	// OWNER
	// invalid values
	values = []string{"root", "r**t:r**t"}
	for _, value := range values {
		configurationFile.Owner = value
		replaceFirstConfFile(c, &configurationFile)
		if err := c.Validate(); err == nil {
			t.Logf("file owner invalid instead of valid for value: %s", value)
			t.Fail()
		}
	}
	// valid values
	values = []string{"root:root", "r00t:r00t"}
	for _, value := range values {
		configurationFile.Owner = value
		replaceFirstConfFile(c, &configurationFile)
		if err := c.Validate(); err != nil {
			t.Logf("file owner valid instead of invalid for value: %s", value)
			t.Fail()
		}
	}

}

func TestBuildEmptyConfiguration(t *testing.T) {
	//var c internal.Configuration
	var c configuration.FreyjaConfiguration
	if err := c.BuildFromFile(testFileEmptyConfiguration); err == nil {
		t.Logf("empty configuration valid instead of invalid: %v", c)
		t.Fail()
	}
}

// TestBuildDefaultConfiguration only checks if minimum required values are set and that default values work
func TestBuildDefaultConfiguration(t *testing.T) {
	// build config
	var c configuration.FreyjaConfiguration
	if err := c.BuildFromFile(testFileValidDefaultConfiguration); err != nil {
		log.Printf("Cannot build configuration from file '%s': %v", testFileValidDefaultConfiguration, err)
		t.Fail()
	}
	// test config
	expectedVersion := "v0.1.0-beta"
	if c.Version != expectedVersion {
		t.Logf("expected version '%s' but got '%s'", expectedVersion, c.Version)
		t.Fail()
	}
	if len(c.Machines) != 1 {
		t.Logf("expected 1 machine %d", len(c.Machines))
		t.Fail()
	}
	// test machine default values
	m := c.Machines[0]
	// mandatory values
	expectedImage := "/tmp/debian-12-generic-amd64.qcow2"
	if m.Image != expectedImage {
		t.Logf("expected image '%s' but got '%s'", expectedImage, m.Image)
		t.Fail()
	}
	expectedOs := "debian12"
	if m.Os != expectedOs {
		t.Logf("expected OS '%s' but got '%s'", expectedOs, m.Os)
		t.Fail()
	}
	expectedHostname := "vm1"
	if m.Hostname != expectedHostname {
		t.Logf("expected OS '%s' but got '%s'", expectedHostname, m.Hostname)
		t.Fail()
	}
	// default values
	// Networks : empty by default
	if len(m.Networks) != 0 {
		t.Logf("wrong networks value, expected empty but got '%v'", m.Networks)
		t.Fail()
	}
	// Users : freyja:master by default
	if len(m.Users) != 1 {
		t.Logf("expected only one user but got '%d'", len(m.Users))
		t.Fail()
	}
	if m.Users[0].Name != configuration.DefaultUserName {
		t.Logf("expected username '%s' but got '%s'", configuration.DefaultUserName, m.Users[0].Name)
		t.Fail()
	}
	if m.Users[0].Password != configuration.DefaultUserPassword {
		t.Logf("expected password '%s' but got '%s'", configuration.DefaultUserPassword, m.Users[0].Password)
		t.Fail()
	}
	// Storage
	if m.Storage != configuration.DefaultMachineStorage {
		t.Logf("expected storage '%d' but got '%d'", configuration.DefaultMachineStorage, m.Storage)
		t.Fail()
	}
	// memory
	if m.Memory != configuration.DefaultMachineMemory {
		t.Logf("expected memory '%d' but got '%d'", configuration.DefaultMachineMemory, m.Memory)
		t.Fail()
	}
	// vcpu
	if m.Vcpu != configuration.DefaultMachineVcpu {
		t.Logf("expected vcpu '%d' but got '%d'", configuration.DefaultMachineVcpu, m.Vcpu)
		t.Fail()
	}
	// packages : empty by default
	if len(m.Packages) != 0 {
		t.Logf("wrong packages value, expected empty but got '%v'", m.Networks)
		t.Fail()
	}
	// files : empty by default
	if len(m.Files) != 0 {
		t.Logf("wrong files value, expected empty but got '%v'", m.Networks)
		t.Fail()
	}
	// update : false by default
	if m.Update {
		t.Logf("wrong update value, expected false but got true")
		t.Fail()
	}
	// reboot : false by default
	if m.Reboot {
		t.Logf("wrong update value, expected false but got true")
		t.Fail()
	}

}

// TestBuildDefaultFilesConfig only checks if minimum required values for files are set and that default values work
func TestBuildDefaultFilesConfig(t *testing.T) {
	// build config
	var c configuration.FreyjaConfiguration
	if err := c.BuildFromFile(testFileValidDefaultFilesConfiguration); err != nil {
		log.Printf("Cannot build configuration from file '%s': %v", testFileValidDefaultFilesConfiguration, err)
		t.Fail()
	}
	// test config
	for _, m := range c.Machines {
		lf := len(m.Files)
		if lf == 0 {
			t.Logf("files config should not be empty but found 0")
			t.Fail()
		}
	}
}

// TestBuildCompleteConfig checks all the values that can be set in a configuration
func TestBuildCompleteConfig(t *testing.T) {
	c := internalTest.BuildCompleteConfig(testFileValidCompleteConfiguration)
	// test config
	// VERSION, HOSTNAME, OS AND IMAGE VALUES ARE ALREADY TESTED IN THE DEFAULT CONFIG TEST
	if len(c.Machines) != 2 {
		t.Logf("expected 2 machines but got %d", len(c.Machines))
		t.Fail()
	}
	m1 := c.Machines[0]
	// test networks
	if len(m1.Networks) != 2 {
		t.Logf("expected 2 networks but got %d", len(m1.Networks))
		t.Fail()
	}
	n1 := m1.Networks[0]
	if n1.Name != "ctrl-plane" {
		t.Logf("expected network name 'ctrl-plane' but got '%s'", n1.Name)
		t.Fail()
	}
	if n1.Mac != "52:54:02:aa:bb:cc" {
		t.Logf("expected network mac '52:54:02:aa:bb:cc' but got '%s'", n1.Mac)
		t.Fail()
	}
	if n1.Interface != "virbr0" {
		t.Logf("expected network interface 'vnet0' but got '%s'", n1.Interface)
		t.Fail()
	}
	n2 := m1.Networks[1]
	if n2.Name != "data-plane" {
		t.Logf("expected network name 'data-plane' but got '%s'", n2.Name)
		t.Fail()
	}
	if n2.Mac != "52:54:02:aa:bb:cd" {
		t.Logf("expected network mac '52:54:02:aa:bb:cd' but got '%s'", n2.Mac)
		t.Fail()
	}
	// test users
	if len(m1.Users) != 2 {
		t.Logf("expected 2 users but got %d", len(m1.Users))
		t.Fail()
	}
	u1 := m1.Users[0]
	if u1.Name != "sam" {
		t.Logf("expected user name 'sam' but got '%s'", u1.Name)
		t.Fail()
	}
	u1ExpectedPass := "$6$6LEpjaxLaT/pu5$wwHsyMlZ2JpHObVJBKGbZUmR5oJ4GocH0zRQYKAuWEwq9ifG4N3Vi/E3ZXTj1bK.QQrOmttA7zIZUIEBaU6Yx."
	if u1.Password != u1ExpectedPass {
		t.Logf("expected user pass '%s' but got '%s'", u1ExpectedPass, u1.Password)
		t.Fail()
	}
	// user ssh keys
	u1ExpectedKeys := []string{internalTest.FreyjaUnitTestDirCommon + "/sam.pub", internalTest.FreyjaUnitTestDirCommon + "/ext.pub"}
	if !compareOrderedStringSlices(u1ExpectedKeys, u1.Keys) {
		t.Logf("expected user keys '%v' but got '%v'", u1ExpectedKeys, u1.Keys)
		t.Fail()
	}
	// user groups
	u1ExpectedGroups := []string{"group1", "group2"}
	if !compareOrderedStringSlices(u1ExpectedGroups, u1.Groups) {
		t.Logf("expected user groups '%v' but got '%v'", u1ExpectedGroups, u1.Groups)
		t.Fail()
	}
	u2 := m1.Users[1]
	if u2.Name != "frodo" {
		t.Logf("expected user name 'frodo' but got '%s'", u2.Name)
		t.Fail()
	}
	// test storage, memory, vcpu
	if m1.Storage != 100 {
		t.Logf("expected storage '100' but got '%d'", m1.Storage)
		t.Fail()
	}
	if m1.Memory != 8192 {
		t.Logf("expected Memory '100' 8192 got '%d'", m1.Memory)
		t.Fail()
	}
	if m1.Vcpu != 4 {
		t.Logf("expected Vcpu '4' but got '%d'", m1.Vcpu)
		t.Fail()
	}
	// packages
	m1ExpectedPackages := []string{"curl", "net-tools"}
	if !compareOrderedStringSlices(m1ExpectedPackages, m1.Packages) {
		t.Logf("expected packages '%v' but got '%v'", m1ExpectedPackages, m1.Packages)
		t.Fail()
	}
	// commands
	m1ExpectedCmd := []string{"echo 'hello world !' > /tmp/test.txt", "cat /tmp/test.txt"}
	if !compareOrderedStringSlices(m1ExpectedCmd, m1.Cmd) {
		t.Logf("expected commands '%v' but got '%v'", m1ExpectedCmd, m1.Cmd)
		t.Fail()
	}
	// files
	if len(m1.Files) != 2 {
		t.Logf("expected 2 files but got %d", len(m1.Files))
		t.Fail()
	}
	f1 := m1.Files[0]
	if f1.Source != internalTest.FreyjaUnitTestDirCommon+"/hello.txt" {
		t.Logf("expected file source '/tmp/freyja-unit-test/config/hello.txt' but got '%s'", f1.Source)
		t.Fail()
	}
	if f1.Destination != "/root/hello.txt" {
		t.Logf("expected file Destination '/root/hello.txt' but got '%s'", f1.Destination)
		t.Fail()
	}
	if f1.Permissions != "0700" {
		t.Logf("expected file Permissions '0700' but got '%s'", f1.Permissions)
		t.Fail()
	}
	if f1.Owner != "root:freyja" {
		t.Logf("expected file Destination 'root:freyja' but got '%s'", f1.Owner)
		t.Fail()
	}
	f2 := m1.Files[1]
	if f2.Source != internalTest.FreyjaUnitTestDirCommon+"/world.txt" {
		t.Logf("expected file source '/tmp/freyja-unit-test/config/world.txt' but got '%s'", f2.Source)
		t.Fail()
	}
	// update
	if !m1.Update {
		t.Logf("expected update 'true' but got 'false'")
		t.Fail()
	}
	// reboot
	if !m1.Reboot {
		t.Logf("expected reboot 'true' but got 'false'")
		t.Fail()
	}
	// just testing mandatory values of machine 2 to make sure that the 2 machines are considered
	m2 := c.Machines[1]
	expectedImage := "/tmp/ubuntu-20.04-LTS-20210603.0.x86_64.qcow2"
	if m2.Image != expectedImage {
		t.Logf("expected image '%s' but got '%s'", expectedImage, m2.Image)
		t.Fail()
	}
	expectedOs := "ubuntu20"
	if m2.Os != expectedOs {
		t.Logf("expected OS '%s' but got '%s'", expectedOs, m2.Os)
		t.Fail()
	}
	expectedHostname := "vm2"
	if m2.Hostname != expectedHostname {
		t.Logf("expected OS '%s' but got '%s'", expectedHostname, m2.Hostname)
		t.Fail()
	}
}
