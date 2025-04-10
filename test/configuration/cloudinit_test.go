package configuration

import (
	"bytes"
	"errors"
	"freyja/internal"
	"freyja/internal/configuration"
	internalTest "freyja/test"
	"os"
	"path/filepath"
	"testing"
)

const testFileCloudInitDefaultMetadata string = "static/cloudinit_default_metadata.yaml"

const testFileCloudInitDefaultUserData string = "static/cloudinit_default_user_data.yaml"

const testFileCloudInitCompleteMetadataVm1 string = "static/cloudinit_complete_metadata_vm1.yaml"

const testFileCloudInitCompleteMetadataVm2 string = "static/cloudinit_complete_metadata_vm2.yaml"

const testFileCloudInitCompleteUserDataVm1 string = "static/cloudinit_complete_user_data_vm1.yaml"

const testFileCloudInitCompleteUserDataVm2 string = "static/cloudinit_complete_user_data_vm2.yaml"

const cloudInitTestDirname = "cloudinit"

//
// METADATA MODEL
//

// TestBuildMetadataCloudInitConfig test if the built cloud init metadata model, from a machine
// config data model, is valid
func TestBuildMetadataCloudInitConfig(t *testing.T) {
	c := internalTest.BuildCompleteConfig(testFileValidCompleteConfiguration)
	var ci configuration.CloudInitMetadata
	// test for second machine
	// get only the minimum required for machine config
	m := &c.Machines[1]
	if err := ci.Build(m); err != nil {
		t.Errorf("error while building the configuration of machine '%s': '%v", m.Hostname, err)
		t.FailNow()
	}
	if ci.LocalHostname != m.Hostname {
		t.Errorf("Expected local hostname '%s' but got '%s'", m.Hostname, ci.LocalHostname)
		t.FailNow()
	}

	if ci.InstanceID != m.Hostname {
		t.Errorf("Expected instance id '%s' but got '%s'", m.Hostname, ci.InstanceID)
		t.FailNow()
	}
}

//
// USER DATA MODEL
//

// TestBuildUserDataCloudInitDefaultConfig checks the cloud init user data model for values
// automatically set or by default
func TestBuildUserDataCloudInitDefaultConfig(t *testing.T) {
	c := internalTest.BuildConfig(testFileValidDefaultConfiguration)
	var ci configuration.CloudInitUserData
	// test for second machine
	// get only the minimum required for machine config
	if err := ci.Build(&c.Machines[0]); err != nil {
		t.Errorf("Could not build cloud init configuration from complete config: %v", err)
		t.FailNow()
	}
	// hostname
	if ci.Hostname == "" {
		t.Errorf("Hostname is empty instead of not empty")
		t.Fail()
	}
	// output
	if ci.Output.All != configuration.CloudInitUserDataOutputAllString {
		t.Errorf("'All' value expected is '%s' but got '%s'", configuration.CloudInitUserDataOutputAllString, ci.Output.All)
		t.Fail()
	}
	// users
	if len(ci.Users) != 1 {
		t.Errorf("Expected no users but got '%d'", len(ci.Users))
		t.FailNow()
	}
	expectedUName := configuration.DefaultUserName
	u := ci.Users[0]
	if u.Name != expectedUName {
		t.Errorf("Expected user name '%s' but got '%s'", configuration.DefaultUserName, u.Name)
		t.Fail()
	}
	if u.Sudo != nil {
		t.Errorf("Expected nil user sudo value but got '%s'", u.Sudo)
		t.Fail()
	}
	if u.LockPasswd {
		t.Errorf("Expected user lockpasswd 'false' but got 'true'")
		t.Fail()
	}
	if u.Shell != configuration.CloudInitUserDataUserShellString {
		t.Errorf("Expected user sudo value '%s' but got '%s'", configuration.CloudInitUserDataUserShellString, u.Shell)
		t.Fail()
	}
	if u.Passwd != configuration.DefaultUserPassword {
		t.Errorf("Expected password '%s' but got '%s'", configuration.DefaultUserPassword, u.Passwd)
		t.Fail()
	}
	if u.Groups != "" {
		t.Errorf("Expected no user groups but got '%s'", u.Groups)
		t.Fail()
	}
	if u.SshAuthorizedKeys != nil {
		t.Errorf("Expected no user ssh keys but got '%v'", u.SshAuthorizedKeys)
		t.Fail()
	}
	// package update
	if ci.PackageUpdate {
		t.Errorf("Expected package_update 'false' but got 'true'")
		t.Fail()
	}
	// package upgrade
	if ci.PackageUpgrade {
		t.Errorf("Expected package_upgrade 'false' but got 'true'")
		t.Fail()
	}
	// packages
	if ci.Packages != nil {
		t.Errorf("Expected no packages but got '%v'", ci.Packages)
		t.Fail()
	}
	// files
	if len(ci.WriteFiles) != 0 {
		t.Errorf("Expected no files but got '%d'", len(ci.WriteFiles))
		t.FailNow()
	}
	// commands
	if ci.RunCmd != nil {
		t.Errorf("Expected no run cmd but got '%d'", len(ci.RunCmd))
		t.FailNow()
	}
	// reboot
	if ci.PowerState != nil {
		t.Errorf("Expected no power state mode but got '%v'", ci.PowerState)
		t.Fail()
	}
}

// TestBuildUserDataCloudInitDefaultConfig checks the cloud init user data model all the values
// that can be set in a configuration
func TestBuildUserDataCloudInitCompleteConfig(t *testing.T) {
	c := internalTest.BuildCompleteConfig(testFileValidCompleteConfiguration)
	var ci configuration.CloudInitUserData
	// test for first machine
	if err := ci.Build(&c.Machines[0]); err != nil {
		t.Errorf("Could not build cloud init configuration from complete config: %v", err)
		t.FailNow()
	}
	// users
	if len(ci.Users) != 2 {
		t.Errorf("Expected 2 users but got '%d'", len(ci.Users))
		t.FailNow()
	}
	expectedU1Name := "sam"
	u1 := ci.Users[0]
	if u1.Name != expectedU1Name {
		t.Errorf("Expected user name '%s' but got '%s'", expectedU1Name, u1.Name)
		t.Fail()
	}
	if !compareOrderedStringSlices(u1.Sudo, configuration.GetCloudInitUserDataUserSudoStringConst()) {
		t.Errorf("Expected user sudo value '%s' but got '%s'", configuration.CloudInitUserDataUserSudoString, u1.Sudo)
		t.Fail()
	}
	if u1.LockPasswd {
		t.Errorf("Expected user lockpasswd 'false' but got 'true'")
		t.Fail()
	}
	expectedU1Passwd := "$6$6LEpjaxLaT/pu5$wwHsyMlZ2JpHObVJBKGbZUmR5oJ4GocH0zRQYKAuWEwq9ifG4N3Vi/E3ZXTj1bK.QQrOmttA7zIZUIEBaU6Yx."
	if u1.Passwd != expectedU1Passwd {
		t.Errorf("Expected user name '%s' but got '%s'", expectedU1Passwd, u1.Passwd)
		t.Fail()
	}
	expectedU1Groups := "group1,group2,sudo"
	if u1.Groups != expectedU1Groups {
		t.Errorf("Expected user groups '%s' but got '%s'", expectedU1Groups, u1.Groups)
		t.Fail()
	}
	expectedU1SshKeys := []string{"key", "key"}
	if !compareOrderedStringSlices(u1.SshAuthorizedKeys, expectedU1SshKeys) {
		t.Errorf("Expected user ssh authorized keys '%v' but got '%v'", expectedU1SshKeys, u1.SshAuthorizedKeys)
		t.Fail()
	}
	// package update
	if !ci.PackageUpdate {
		t.Errorf("Expected package_update 'true' but got 'false'")
		t.Fail()
	}
	// package upgrade
	if !ci.PackageUpgrade {
		t.Errorf("Expected package_upgrade 'true' but got 'false'")
		t.Fail()
	}
	// packages
	expectedPackages := []string{"curl", "net-tools"}
	if !compareOrderedStringSlices(expectedPackages, ci.Packages) {
		t.Errorf("Expected packages '%v' but got '%v'", expectedPackages, ci.Packages)
		t.Fail()
	}
	// files
	if len(ci.WriteFiles) != 2 {
		t.Errorf("Expected 2 files but got '%d'", len(ci.WriteFiles))
		t.FailNow()
	}
	f1 := ci.WriteFiles[0]
	if f1.Encoding != configuration.CloudInitUserDataFilesEncoding {
		t.Errorf("Expected file encoding '%s' but got '%s'", configuration.CloudInitUserDataFilesEncoding, f1.Encoding)
		t.Fail()
	}
	expectedF1Content := internal.EncodeB64Bytes([]byte(internalTest.ExpectedHelloFileContent))
	if f1.Content != expectedF1Content {
		t.Errorf("Wrong file 1 content")
		t.Fail()
	}
	expectedF1Path := "/root/hello.txt"
	if f1.Path != expectedF1Path {
		t.Errorf("Expected file path '%s' but got '%s'", expectedF1Path, f1.Path)
		t.Fail()
	}
	expectedF1Permissions := "0700"
	if f1.Permissions != expectedF1Permissions {
		t.Errorf("Expected file permissions '%s' but got '%s'", expectedF1Permissions, f1.Permissions)
		t.Fail()
	}
	expectedF1Owner := "root:freyja"
	if f1.Owner != expectedF1Owner {
		t.Errorf("Expected file owner '%s' but got '%s'", expectedF1Owner, f1.Owner)
		t.Fail()
	}
	// commands
	expectedCommands := []string{"echo 'hello world !' > /tmp/test.txt", "cat /tmp/test.txt"}
	if !compareOrderedStringSlices(ci.RunCmd, expectedCommands) {
		t.Errorf("Expected run cmd '%v' but got '%v'", expectedCommands, ci.RunCmd)
		t.Fail()
	}
	// reboot
	if ci.PowerState == nil {
		t.Errorf("Expected power state but got none")
		t.FailNow()
	}
	ps := ci.PowerState
	expectedPsMode := "reboot"
	if ps.Mode != expectedPsMode {
		t.Errorf("Expected power state mode '%s' but got '%s'", expectedPsMode, ps.Mode)
		t.Fail()
	}
	expectedPsMessage := "First reboot"
	if ps.Message != expectedPsMessage {
		t.Errorf("Expected power state message '%s' but got '%s'", expectedPsMessage, ps.Message)
		t.Fail()
	}
	expectedPsTimeout := uint(30)
	if ps.Timeout != expectedPsTimeout {
		t.Errorf("Expected power state timeout '%d' but got '%d'", expectedPsTimeout, ps.Timeout)
		t.Fail()
	}
	if !ps.Condition {
		t.Errorf("Expected power state condition 'true' but got 'false'")
		t.Fail()
	}

}

//
// CONFIG GENERATION (WRITING)
//

type expectedFilesPath struct {
	expectedMetadata, expectedUserdata string
}

func getCloudInitTestDirPath() string {
	return filepath.Join(internalTest.FreyjaUnitTestDir, cloudInitTestDirname)
}

// TestGenerateCloudInitConfigs verifies that the configs are properly built and written
// all the other specs are already tested in the other tests
func TestGenerateCloudInitConfigs(t *testing.T) {

	// default and minimal config
	expectedDefaultFiles := make(map[string]expectedFilesPath)
	expectedDefaultFiles["vm1"] = expectedFilesPath{
		expectedMetadata: testFileCloudInitDefaultMetadata,
		expectedUserdata: testFileCloudInitDefaultUserData,
	}
	c := internalTest.BuildConfig(testFileValidDefaultConfiguration)
	testGeneratedDefaultCloudInitConfigs(t, "TestGenerateDefaultCloudInitConfigs", c, expectedDefaultFiles)

	// complete configuration
	expectedCompleteFiles := make(map[string]expectedFilesPath)
	expectedCompleteFiles["vm1"] = expectedFilesPath{
		expectedMetadata: testFileCloudInitCompleteMetadataVm1,
		expectedUserdata: testFileCloudInitCompleteUserDataVm1,
	}
	expectedCompleteFiles["vm2"] = expectedFilesPath{
		expectedMetadata: testFileCloudInitCompleteMetadataVm2,
		expectedUserdata: testFileCloudInitCompleteUserDataVm2,
	}
	c = internalTest.BuildCompleteConfig(testFileValidCompleteConfiguration)
	testGeneratedDefaultCloudInitConfigs(t, "TestGenerateCompleteCloudInitConfigs", c, expectedCompleteFiles)
}

// testGeneratedDefaultCloudInitConfigs test the generated cloud init files
// To test the content, this method takes the first machine config only
func testGeneratedDefaultCloudInitConfigs(t *testing.T, testDirName string, config *configuration.FreyjaConfiguration, expected map[string]expectedFilesPath) {
	var err error
	for _, machine := range config.Machines {
		testDir := filepath.Join(getCloudInitTestDirPath(), testDirName, machine.Hostname)
		if err = configuration.GenerateCloudInitConfigs(&machine, testDir); err != nil {
			t.Errorf("cannot generate cloud init configs for machine '%s', reason: %v", machine.Hostname, err)
			t.FailNow()
		}

		// test if files have been created
		resultMetadataFilePath := filepath.Join(testDir, configuration.CloudInitMetadataFileName)
		resultUserdataFilePath := filepath.Join(testDir, configuration.CloudInitUserDataFileName)
		if _, err = os.Stat(resultMetadataFilePath); errors.Is(err, os.ErrNotExist) {
			t.Errorf("cloud init metadata file not found in '%s'", testDir)
			t.FailNow()
		}
		if _, err = os.Stat(resultUserdataFilePath); errors.Is(err, os.ErrNotExist) {
			t.Errorf("cloud init user data file not found in '%s'", testDir)
			t.FailNow()
		}

		// test files content
		// metadata
		var expectedMetadataRaw []byte
		expectedMetadataFilePath := expected[machine.Hostname].expectedMetadata
		if expectedMetadataRaw, err = os.ReadFile(expectedMetadataFilePath); err != nil {
			t.Errorf("cannot read expected matadata file for comparison in '%s': %v", expectedMetadataFilePath, err)
			t.FailNow()
		}
		var resultMetadataRaw []byte
		if resultMetadataRaw, err = os.ReadFile(resultMetadataFilePath); err != nil {
			t.Errorf("cannot read result matadata file for comparison in '%s': %v", resultMetadataFilePath, err)
			t.FailNow()
		}
		if !bytes.Equal(expectedMetadataRaw, resultMetadataRaw) {
			t.Errorf("expected and result metadata content do not match for '%s'", resultMetadataFilePath)
			t.Fail()
		}
		// user data
		var expectedUserdataRaw []byte
		expectedUserdataFilePath := expected[machine.Hostname].expectedUserdata
		if expectedUserdataRaw, err = os.ReadFile(expectedUserdataFilePath); err != nil {
			t.Errorf("cannot read expected matadata file for comparison in '%s': %v", expectedUserdataFilePath, err)
			t.FailNow()
		}
		var resultUserdataRaw []byte
		if resultUserdataRaw, err = os.ReadFile(resultUserdataFilePath); err != nil {
			t.Errorf("cannot read result matadata file for comparison in '%s': %v", resultUserdataFilePath, err)
			t.FailNow()
		}
		if !bytes.Equal(expectedUserdataRaw, resultUserdataRaw) {
			t.Errorf("expected and result user data content do not match for '%s'", resultUserdataFilePath)
			t.Fail()
		}
	}

}
