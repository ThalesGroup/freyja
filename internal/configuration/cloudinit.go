package configuration

import (
	"fmt"
	"freyja/internal"
	"github.com/kdomanski/iso9660"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"strings"
)

const ISOCloudInitFilename string = "cloud-init.iso"

// **********
// DATA MODEL
// **********

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
	SshAuthorizedKeys []string `yaml:"ssh_authorized_keys,omitempty,flow"`
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

// **************
// IMPLEMENTATION
// **************

type CloudInitConfiguration interface {
	Build(machine *FreyjaConfigurationMachine) error
	Write(directory string) error
}

const CloudInitMetadataFileName = "meta-data"

const CloudInitUserDataFileName = "user-data"

func (ci *CloudInitMetadata) Build(machine *FreyjaConfigurationMachine) error {
	ci.InstanceID = machine.Hostname
	ci.LocalHostname = machine.Hostname
	return nil
}

func (ci *CloudInitMetadata) Write(directory string) (err error) {
	var data []byte
	if data, err = yaml.Marshal(ci); err != nil {
		return fmt.Errorf("could not parse cloud init meta data in '%s': %w", directory, err)
	}
	return writeCloudInitConfig(directory, CloudInitMetadataFileName, data)
}

// USER DATA

// Build generate the configuration from a file
func (ci *CloudInitUserData) Build(machine *FreyjaConfigurationMachine) error {
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

		if u.Sudo {
			// inject sudo list of parameters
			ciu.Sudo = GetCloudInitUserDataUserSudoStringConst()
			// inject sudo in groups
			u.Groups = append(u.Groups, "sudo")
		}
		ciu.Groups = strings.Join(u.Groups, ",")
		// key content must be read and injected in cloud init user data
		if u.Keys != nil {
			keys := make([]string, len(u.Keys))
			for j, key := range u.Keys {
				resolvedKeyPath := os.ExpandEnv(key)
				content, err := os.ReadFile(resolvedKeyPath)
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
		cif.Content = internal.EncodeB64Bytes(contentBytes)
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

// marshal is a custom function to parse the cloud init user data to a yaml file data
// this is necessary for YAML v3 package
// YAML v3 package struggles to handle double quoting strings and we need to enforce them for few
// config fields
func (ci *CloudInitUserData) marshal() (data []byte, err error) {

	// Convert the input to a YAML node
	var rootNode yaml.Node
	if err = rootNode.Encode(ci); err != nil {
		return nil, err
	}

	// Running through the yaml nodes to customize the output
	// Root node
	for i, node := range rootNode.Content {
		// "users" map
		if node.Value == "users" {
			usersNode := rootNode.Content[i+1]
			// for each user
			for _, userNode := range usersNode.Content {
				var passwdValueNode *yaml.Node
				// for each config of a user
				for j, userField := range userNode.Content {
					// finding 'passwd' config
					// apply double quote
					if userField.Value == "passwd" {
						passwdValueNode = userNode.Content[j+1]
						passwdValueNode.Style = yaml.DoubleQuotedStyle
					}
					// finding 'groups' config
					// apply double quote
					if userField.Value == "groups" {
						passwdValueNode = userNode.Content[j+1]
						passwdValueNode.Style = yaml.DoubleQuotedStyle
					}
				}
			}
		}
	}

	// Marshal the modified node back to YAML
	return yaml.Marshal(&rootNode)
}

func (ci *CloudInitUserData) Write(directory string) (err error) {
	var data []byte
	if data, err = ci.marshal(); err != nil {
		return fmt.Errorf("could not parse cloud init user data in '%s': %w", directory, err)
	}
	return writeCloudInitConfig(directory, CloudInitUserDataFileName, data)
}

//
// GENERATE
//

// writeCloudInitConfig writes both user data and metadata cloud init files into the given output
// directory.
// The array of bytes 'data' is the serialization result of the cloud init configuration.
func writeCloudInitConfig(directory string, filename string, data []byte) (err error) {
	// create the config dir if it does exist yet
	err = os.MkdirAll(directory, os.ModePerm)
	if err != nil {
		return fmt.Errorf("could not create cloud init into parent directory '%s' for file '%s': %w", directory, filename, err)
	}
	// re-create the file to inject data
	path := filepath.Join(directory, filename)
	_, err = internal.RemoveIfExists(path)
	if err != nil {
		return fmt.Errorf("could not remove file '%s': %w", path, err)
	}
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("cannot create cloud init file '%s' : %w", path, err)
	}
	// inject the mandatory '#cloud-config' string at the beginning of the file
	if _, err = file.WriteString(CloudInitUserDataHeader); err != nil {
		return fmt.Errorf("could not write cloud init config for machine '%s' into file '%s': %w", filename, path, err)
	}
	// inject the cloud init user data
	if _, err = file.Write(data); err != nil {
		return fmt.Errorf("could not write cloud init config for machine '%s' into file '%s': %w", filename, path, err)
	}
	return nil
}

func GenerateCloudInitConfigs(machine *FreyjaConfigurationMachine, outputDir string) error {
	// metadata generation
	var cm CloudInitMetadata
	if err := cm.Build(machine); err != nil {
		msg := fmt.Sprintf("cannot build cloud init metadata for machine '%s'", machine.Hostname)
		return fmt.Errorf(msg, err)
	}
	if err := cm.Write(outputDir); err != nil {
		msg := fmt.Sprintf("cannot write cloud init metadata file for machine '%s'", machine.Hostname)
		return fmt.Errorf(msg, err)
	}
	// user data generation
	var cu CloudInitUserData
	if err := cu.Build(machine); err != nil {
		msg := fmt.Sprintf("cannot build cloud init user data for machine '%s'", machine.Hostname)
		return fmt.Errorf(msg, err)
	}
	if err := cu.Write(outputDir); err != nil {
		msg := fmt.Sprintf("cannot write cloud init user data file for machine '%s'", machine.Hostname)
		return fmt.Errorf(msg, err)
	}
	return nil
}

//
// ISO
//

// CreateCloudInitIso creates the ISO file following the iso9660 standard.
// Implementation using kdomanski/iso9660 : https://github.com/kdomanski/iso9660
// it is also used in terraform for proxmox : https://github.com/Telmate/terraform-provider-proxmox/blob/186ec3f23bf4a62fcad35f6292fa1350b8e1183b/proxmox/resource_cloud_init_disk.go#L77-L122
// YOU MUST name the provision files 'user-data' 'meta-data' !!!!!!!!
// YOU MUST name the ISO volume 'cidata' !!!!!!
func CreateCloudInitIso(machine *FreyjaConfigurationMachine, machineDir string) (string, error) {
	isoWriter, err := iso9660.NewWriter()
	if err != nil {
		return "", fmt.Errorf("cannot create the iso9660 writer for the machine '%s' : %w", machine.Hostname, err)
	}
	// add cloud init metadata in ISO
	metadataPath := filepath.Join(machineDir, CloudInitMetadataFileName)
	fm, err := os.Open(metadataPath)
	if err != nil {
		return "", fmt.Errorf("cannot open the cloud init metadata file in '%s' : %w", metadataPath, err)
	}
	defer fm.Close()
	// YOU MUST name the provision files 'user-data' 'meta-data' !!!!!!!!
	if err = isoWriter.AddFile(fm, "meta-data"); err != nil {
		return "", fmt.Errorf("cannot add the cloud init metadata file in ISO for the machine '%s' : %w", machine.Hostname, err)
	}
	// add cloud init user data in ISO
	userdataPath := filepath.Join(machineDir, CloudInitUserDataFileName)
	fu, err := os.Open(userdataPath)
	if err != nil {
		return "", fmt.Errorf("cannot open the cloud init user data file in '%s' : %w", userdataPath, err)
	}
	defer fu.Close()
	// YOU MUST name the provision files 'user-data' 'meta-data' !!!!!!!!
	if err = isoWriter.AddFile(fu, "user-data"); err != nil {
		return "", fmt.Errorf("cannot add the cloud init user data file in ISO for the machine '%s' : %w", machine.Hostname, err)
	}
	// delete pre-existing ISO file for provision update

	// write iso on filesystem
	isoOutputPath := filepath.Join(machineDir, ISOCloudInitFilename)
	_, err = internal.RemoveIfExists(isoOutputPath)
	if err != nil {
		return "", fmt.Errorf("cannot replace the cloud init ISO file in '%s' : %w", isoOutputPath, err)
	}
	outputFile, err := os.OpenFile(isoOutputPath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0664)
	if err != nil {
		return "", fmt.Errorf("cannot open the ISO image file path '%s' : %w", isoOutputPath, err)
	}
	// calculate the iso file ID
	// YOU MUST name the ISO volume 'cidata' !!!!!!
	if err = isoWriter.WriteTo(outputFile, "cidata"); err != nil {
		return "", fmt.Errorf("cannot write the ISO image file in '%s' : %w", isoOutputPath, err)
	}
	if err = outputFile.Close(); err != nil {
		return "", err
	}

	return isoOutputPath, nil
}
