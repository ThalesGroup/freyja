package internal

import (
	"bytes"
	"embed"
	"fmt"
	"freyja/internal"
	"os"
	"path/filepath"
	"testing"
)

// testFilesDir
//
//go:embed static/*
var testFilesCloudInitDir embed.FS

// testFileEmptyConfiguration is used to test empty configuration (missing required values)
const testFileCloudInitDefaultMetadata string = "static/cloudinit_default_metadata.yaml"

// testFileEmptyConfiguration is used to test empty configuration (missing required values)
const testFileCloudInitDefaultUserData string = "static/cloudinit_default_user_data.yaml"

const testFileCloudInitCompleteUserData string = "static/cloudinit_complete_user_data.yaml"

//
// UTILS
//

// testWriteCloudConfig compares two cloud init files (user data or metadata) : an expected one
// with a generated one from a machine config data model
func writeAndCompareCloudInitConfig(t *testing.T, expectedFilePath string, testedFilename string, cloudInitType internal.CloudInitConfiguration, machineConfig *internal.ConfigurationMachine) {
	expectedContent, err := os.ReadFile(expectedFilePath)
	if err != nil {
		t.Errorf("could not read unit test file '%s', reason: %v", expectedFilePath, err)
		t.FailNow()
	}

	// test for second machine
	// get only the minimum required for machine config
	if err := cloudInitType.Build(machineConfig); err != nil {
		t.Errorf("Could not build cloud init configuration from complete config: %v", err)
		t.FailNow()
	}

	dirPath := filepath.Join(FreyjaUnitTestDir, "cloudinit")
	if err := cloudInitType.Write(dirPath); err != nil {
		t.Errorf("could not write cloud init config in '%s', reason: %v", dirPath, err)
		t.FailNow()
	}

	path := filepath.Join(dirPath, testedFilename)
	testContent, err := os.ReadFile(path)
	if err != nil {
		t.Errorf("could not read cloud init file '%s', reason: %v", path, err)
		t.FailNow()
	}

	if !bytes.Equal(expectedContent, testContent) {
		t.Errorf("wrong content, expected '%s' but got '%s'", string(expectedContent), string(testContent))
		t.FailNow()
	}
}

//
// METADATA
//

// TestBuildMetadataCloudInitConfig test if the built cloud init metadata model, from a machine
// config data model, is valid
func TestBuildMetadataCloudInitConfig(t *testing.T) {
	c := BuildCompleteConfig()
	var ci internal.CloudInitMetadata
	// test for second machine
	// get only the minimum required for machine config
	m := &c.Machines[1]
	ci.Build(m)
	if ci.LocalHostname != m.Hostname {
		t.Errorf("Expected local hostname '%s' but got '%s'", m.Hostname, ci.LocalHostname)
		t.FailNow()
	}

	if ci.InstanceID != m.Hostname {
		t.Errorf("Expected instance id '%s' but got '%s'", m.Hostname, ci.InstanceID)
		t.FailNow()
	}
}

// TestWriteMetadataCloudInitConfig test if the written cloud init metadata file is valid
func TestWriteMetadataCloudInitConfig(t *testing.T) {
	c := BuildCompleteConfig()
	m := &c.Machines[1]
	testedFilename := fmt.Sprintf("%s%s", m.Hostname, internal.CloudInitMetadataFileSuffix)
	var ci internal.CloudInitMetadata
	writeAndCompareCloudInitConfig(t, testFileCloudInitDefaultMetadata, testedFilename, &ci, m)
}

//
// USER DATA
//

// TestBuildUserDataCloudInitDefaultConfig checks the cloud init user data model for values
// automatically set or by default
func TestBuildUserDataCloudInitDefaultConfig(t *testing.T) {
	c := BuildDefaultConfig()
	var ci internal.CloudInitUserData
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
	if ci.Output.All != internal.CloudInitUserDataOutputAllString {
		t.Errorf("'All' value expected is '%s' but got '%s'", internal.CloudInitUserDataOutputAllString, ci.Output.All)
		t.Fail()
	}
	// users
	if len(ci.Users) != 1 {
		t.Errorf("Expected no users but got '%d'", len(ci.Users))
		t.FailNow()
	}
	expectedUName := internal.DefaultUserName
	u := ci.Users[0]
	if u.Name != expectedUName {
		t.Errorf("Expected user name '%s' but got '%s'", internal.DefaultUserName, u.Name)
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
	if u.Shell != internal.CloudInitUserDataUserShellString {
		t.Errorf("Expected user sudo value '%s' but got '%s'", internal.CloudInitUserDataUserShellString, u.Shell)
		t.Fail()
	}
	if u.Passwd != internal.DefaultUserPassword {
		t.Errorf("Expected password '%s' but got '%s'", internal.DefaultUserPassword, u.Passwd)
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
	c := BuildCompleteConfig()
	var ci internal.CloudInitUserData
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
	if !compareOrderedStringSlices(u1.Sudo, internal.GetCloudInitUserDataUserSudoStringConst()) {
		t.Errorf("Expected user sudo value '%s' but got '%s'", internal.CloudInitUserDataUserSudoString, u1.Sudo)
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
	expectedU1Groups := "group1,group2"
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
	if f1.Encoding != internal.CloudInitUserDataFilesEncoding {
		t.Errorf("Expected file encoding '%s' but got '%s'", internal.CloudInitUserDataFilesEncoding, f1.Encoding)
		t.Fail()
	}
	expectedF1Content := internal.EncodeB64Bytes([]byte(ExpectedHelloFileContent))
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

// TestWriteMetadataCloudInitConfig test if the written cloud init user data file is valid
func TestWriteUserDataCloudInitDefaultConfig(t *testing.T) {
	c := BuildDefaultConfig()
	m := &c.Machines[0]
	testedFilename := fmt.Sprintf("%s%s", m.Hostname, internal.CloudInitUserDataFileSuffix)
	var ci internal.CloudInitUserData
	writeAndCompareCloudInitConfig(t, testFileCloudInitDefaultUserData, testedFilename, &ci, m)
}

// TestWriteUserDataCloudInitCompleteConfig test if the written cloud init user data file is valid
func TestWriteUserDataCloudInitCompleteConfig(t *testing.T) {
	c := BuildCompleteConfig()
	m := &c.Machines[0]
	testedFilename := fmt.Sprintf("%s%s", m.Hostname, internal.CloudInitUserDataFileSuffix)
	var ci internal.CloudInitUserData
	writeAndCompareCloudInitConfig(t, testFileCloudInitCompleteUserData, testedFilename, &ci, m)
}
