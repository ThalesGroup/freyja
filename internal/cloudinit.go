package internal

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
	"path/filepath"
	"strings"
)

type CloudInitConfiguration interface {
	Build(machine *ConfigurationMachine) error
	Write(directory string) error
}

func writeCloudInitConfig(directory string, filename string, ci interface{}) error {
	err := os.MkdirAll(directory, os.ModePerm)
	if err != nil {
		return fmt.Errorf("could not create cloud init into parent directory '%s' for file '%s': %w", directory, filename, err)
	}
	metadata, err := yaml.Marshal(&ci)
	if err != nil {
		return fmt.Errorf("could not parse cloud init config file '%s' into bytes: %w", filename, err)
	}
	path := filepath.Join(directory, filename)
	if err := os.WriteFile(path, metadata, os.ModePerm); err != nil {
		return fmt.Errorf("could not write cloud init config for machine '%s' into file '%s': %w", filename, path, err)
	}
	return nil
}

//
// METADATA
//

const CloudInitMetadataFileSuffix = "_metadata.clinit"

const CloudInitUserDataFileSuffix = "_user_data.clinit"

func (ci *CloudInitMetadata) Build(machine *ConfigurationMachine) error {
	ci.InstanceID = machine.Hostname
	ci.LocalHostname = machine.Hostname
	return nil
}

func (ci *CloudInitMetadata) Write(directory string) error {
	filename := fmt.Sprintf("%s%s", ci.LocalHostname, CloudInitMetadataFileSuffix)
	return writeCloudInitConfig(directory, filename, &ci)
}

// USER DATA

// Build generate the configuration from a file
func (ci *CloudInitUserData) Build(machine *ConfigurationMachine) error {
	// set cloud-init values according to configuration inputs
	// hostname
	ci.Hostname = machine.Hostname
	// cloud init output
	ci.Output = &CloudInitUserDataOutput{
		All: CloudInitUserDataOutputAllString,
	}
	// users
	ci.Users = make([]CloudInitUserDataUser, len(machine.Users))
	for i, u := range machine.Users {
		ciu := &ci.Users[i]
		ciu.Name = u.Name
		ciu.Shell = CloudInitUserDataUserShellString
		ciu.Passwd = u.Password
		ciu.Groups = strings.Join(u.Groups, ",")
		if u.Sudo {
			// inject sudo list of parameters
			ciu.Sudo = GetCloudInitUserDataUserSudoStringConst()
			// inject sudo in groups
			ciu.Groups = ciu.Groups + ",sudo"
		}
		// key content must be read and injected in cloud init user data
		if u.Keys != nil {
			keys := make([]string, len(u.Keys))
			for j, key := range u.Keys {
				content, err := os.ReadFile(key)
				if err != nil {
					return err
				}
				keys[j] = string(content)
			}
			ciu.SshAuthorizedKeys = keys
		}
	}
	// update & upgrade
	ci.PackageUpdate = machine.Update
	ci.PackageUpgrade = machine.Update
	// packages
	ci.Packages = machine.Packages
	// files
	if machine.Files != nil {

	}
	ci.WriteFiles = make([]CloudInitUserDataFiles, len(machine.Files))
	for i, f := range machine.Files {
		cif := &ci.WriteFiles[i]
		cif.Encoding = CloudInitUserDataFilesEncoding
		// file content
		contentBytes, err := os.ReadFile(f.Source)
		if err != nil {
			return err
		}
		cif.Content = EncodeB64Bytes(contentBytes)
		// path
		cif.Path = f.Destination
		cif.Permissions = f.Permissions
		cif.Owner = f.Owner
	}
	// commands at boot
	ci.RunCmd = machine.Cmd
	// power state for reboot after first boot
	if machine.Reboot {
		ci.PowerState = &CloudInitUserDataPowerState{
			Mode:      CloudInitUserPowerStateMode,
			Message:   CloudInitUserPowerStateMessage,
			Timeout:   CloudInitUserPowerStateTimeout,
			Condition: CloudInitUserPowerStateCondition,
		}
	}

	return nil
}

func (ci *CloudInitUserData) Write(directory string) error {
	filename := fmt.Sprintf("%s%s", ci.Hostname, CloudInitUserDataFileSuffix)
	return writeCloudInitConfig(directory, filename, &ci)
}
