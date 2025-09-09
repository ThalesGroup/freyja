package configuration

import (
	"errors"
	"fmt"
	"freyja/internal"
	"log"
	"net"
	"os"
	"path/filepath"
	"regexp"

	"github.com/spf13/viper"
)

const Version string = "0.1.0-beta-go"

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

const DefaultNetworkName string = "default"

const DefaultNetworkBridgeName string = "virbr0"

const DefaultFilePermissions string = "0600"

const DefaultFileOwner string = "root:root"

const SSHPublicKeySuffix = "_ed25519.pub"

const SSHPrivateKeySuffix = "_ed25519"

var FreyjaWorkspaceDir = filepath.Join(os.Getenv("HOME"), ".freyja")

var FreyjaMachinesWorkspaceDir = filepath.Join(FreyjaWorkspaceDir, "machines")

var FreyjaNetworksWorkspaceDir = filepath.Join(FreyjaWorkspaceDir, "networks")

// FreyjaConfiguration is the base model for freyja configuration parameters
// Example :
// ---
// version: v0.1.0-beta
// networks:
//   - name: ctrlplane
//     cidr: 192.168.123.0/24
//   - name: dataplane
//     cidr: 192.168.124.0/24
//
// machines:
//   - image: "/tmp/CentOS-Stream-GenericCloud-8-20210603.0.x86_64.qcow2" # MANDATORY
//     os: "centos8" # MANDATORY
//     hostname: "vm1" # MANDATORY, MUST NOT contain underscores
//     networks: # MANDATORY, at least one
//   - name: "ctrl-plane"
//     mac: "52:54:02:aa:bb:cc"
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
type FreyjaConfiguration struct {
	Version  string                       `yaml:"version"`
	Machines []FreyjaConfigurationMachine `yaml:"machines"`
	Networks []FreyjaConfigurationNetwork `yaml:"networks,omitempty"`
}

// FreyjaConfigurationMachine is the configuration model for libvirt guest parameters
type FreyjaConfigurationMachine struct {
	// MANDATORY
	Image    string `yaml:"image"`    // Qcow2 image file path on host
	Hostname string `yaml:"hostname"` // domain name in libvirt
	// optional
	Networks []FreyjaConfigurationMachineNetwork `yaml:"networks,omitempty"`
	Users    []FreyjaConfigurationUser           `yaml:"users,omitempty"`
	Storage  uint                                `yaml:"storage"` // GiB
	Memory   uint                                `yaml:"memory"`  // MiB
	Vcpu     uint                                `yaml:"vcpu"`
	Packages []string                            `yaml:"packages,omitempty"`
	Cmd      []string                            `yaml:"cmd,omitempty"`
	Files    []FreyjaConfigurationFile           `yaml:"files,omitempty"`
	Update   bool                                `yaml:"update"`
	Reboot   bool                                `yaml:"reboot"`
}

type FreyjaConfigurationMachineNetwork struct {
	Name string `yaml:"name"`
	Mac  string `yaml:"mac"`
}

type FreyjaConfigurationUser struct {
	Name     string   `yaml:"name"`
	Password string   `yaml:"password"`
	Sudo     bool     `yaml:"sudo"`
	Groups   []string `yaml:"groups,omitempty"`
	Keys     []string `yaml:"keys"`
}

type FreyjaConfigurationFile struct {
	Source      string `yaml:"source"`
	Destination string `yaml:"destination"`
	Permissions string `yaml:"permissions"`
	Owner       string `yaml:"owner"`
}

type FreyjaConfigurationNetwork struct {
	Name string `yaml:"name"`
	// cidr format must be ipv4 as in '255.255.255.255/24'
	CIDR string `yaml:"cidr"`
}

type Configuration interface {
	Validate() error
	BuildFromFile(path string) error
}

// *****************
// CONFIG VALIDATION
// *****************

// Validate checks the whole freyja configuration values for mistakes, misconfigurations, wrong
// formats, etc ...
// Validate is different from Audit
// mistake = wrong configuration value that may cause build configuration issues
func (c *FreyjaConfiguration) Validate() (err error) {
	// verify version
	if err = c.validateVersion(); err != nil {
		return err
	}
	// verify networks
	if err = c.validateNetworks(); err != nil {
		return err
	}
	// verify machines
	if err = c.validateMachines(); err != nil {
		return err
	}

	return nil
}

// ValidateVersion audits the version configuration
func (c *FreyjaConfiguration) validateVersion() error {
	if c.Version != Version {
		return errors.New(fmt.Sprintf("wrong version format : current version is '%s' but found '%s'", Version, c.Version))
	}
	return nil
}

// validateNetworks validate the network configurations in section 'networks' at the config's root
func (c *FreyjaConfiguration) validateNetworks() error {
	for _, network := range c.Networks {
		if network.Name == "" {
			return fmt.Errorf("missing network name")
		}
		if network.CIDR == "" {
			return fmt.Errorf("missing network's CIDR")
		}
		ip, netName, err := net.ParseCIDR(network.CIDR)
		if err != nil {
			return fmt.Errorf("wrong network's CIDR '%s': %w", network.CIDR, err)
		}
		if ip == nil || netName == nil {
			return fmt.Errorf("cannot retrieve IP or Network values from network's cidr value '%s'", network.CIDR)
		}

	}
	return nil
}

func (c *FreyjaConfiguration) validateMachines() (err error) {
	if len(c.Machines) == 0 {
		return errors.New("configure at least one machine but found 0")
	}
	for _, machine := range c.Machines {
		if machine.Hostname == "" {
			return errors.New("missing mandatory machine hostname")
		}
		if machine.Image == "" {
			return errors.New("missing mandatory machine image")
		}
		if !internal.FileExists(machine.Image) {
			return fmt.Errorf("image file not found : '%s'", machine.Image)
		}
		if len(machine.Networks) != 0 {
			// verify networks
			for _, network := range machine.Networks {
				err = network.validateMachineNetwork()
				if err != nil {
					return &internal.ConfigurationError{Message: err.Error()}
				}
			}
		}
		// verify files
		for _, file := range machine.Files {
			err = file.validateMachineFiles()
			if err != nil {
				return &internal.ConfigurationError{Message: err.Error()}
			}
		}
	}
	return nil
}

// validateMachineNetwork audits the network configuration including
//   - the name of the network
//   - the format of the mac address
func (cn *FreyjaConfigurationMachineNetwork) validateMachineNetwork() error {
	// network name
	if cn.Name == "" {
		return errors.New("network name is empty")
	}
	// network mac address
	if cn.Mac != "" {
		macRegex := "^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$"
		match, err := regexp.MatchString(macRegex, cn.Mac)
		if err != nil {
			return fmt.Errorf("cannot verify pattern matching for string '%s': %w", cn.Mac, err)
		}
		if !match {
			return errors.New(fmt.Sprintf("wrong mac format : regex is '%s' but found value '%s'", macRegex, cn.Mac))
		}
	}
	return nil
}

// validateMachineFiles validate the machine's configuration for file injection in the filesystem
func (cf *FreyjaConfigurationFile) validateMachineFiles() error {
	// validate source
	if cf.Source == "" {
		return errors.New("machine file source config is mandatory but found empty")
	}
	if cf.Destination == "" {
		return errors.New("machine file destination config is mandatory but found empty")
	}
	// validate permissions
	if cf.Permissions != "" {
		pattern := "^[0-7]{4}$"
		match, err := regexp.MatchString(pattern, cf.Permissions)
		if err != nil {
			return fmt.Errorf("cannot verify pattern matching for string '%s'. Should be '^[0-7]{4}$'. Reason : %w", cf.Permissions, err)
		}
		if !match {
			return errors.New(fmt.Sprintf("wrong file permissions : value '%s' does not match pattern '%s'", cf.Permissions, pattern))
		}
	}
	// validate owner
	if cf.Owner != "" {
		pattern := "^[a-zA-Z0-9_]*:[a-zA-Z0-9_]*$"
		match, err := regexp.MatchString(pattern, cf.Owner)
		if err != nil {
			return fmt.Errorf("cannot verify pattern matching for string '%s'. Should be '^[a-zA-Z0-9_]*:[a-zA-Z0-9_]*$'. Reason : %w", cf.Owner, err)
		}
		if !match {
			return errors.New(fmt.Sprintf("wrong file owner : value '%s' does not match pattern '%s'", cf.Owner, pattern))
		}
	}
	return nil
}

// *****
// AUDIT
// *****

// Audit is different from Validate
// Audit verifies if the environment and host conditions are met to create the machine.
// For example, it verifies if provision files exist on host, if the ssh keys in the configuration
// already exist or have been created before the machine instantiation, etc ...
// Thus, the purpose of this function is to be called later than Validate, after that the
// configurations and all the required environment has been generated.
func (cm *FreyjaConfigurationMachine) Audit() (err error) {

	// image
	absImage, err := filepath.Abs(os.ExpandEnv(cm.Image))
	if !internal.FileExists(absImage) {
		return &internal.ConfigurationError{Message: fmt.Sprintf("Machine Image File '%s' of machine '%s' not found", cm.Image, cm.Hostname)}
	}
	// users
	for _, user := range cm.Users {
		if err = user.AuditUser(); err != nil {
			return &internal.ConfigurationError{Message: err.Error()}
		}
	}
	// files
	for _, file := range cm.Files {
		absPath, err := filepath.Abs(os.ExpandEnv(file.Source))
		if err != nil {
			return &internal.ConfigurationError{Message: err.Error()}
		}
		if !internal.FileExists(absPath) {
			return &internal.ConfigurationError{Message: fmt.Sprintf("file's source '%s' of machine '%s' does not exists", file.Source, cm.Hostname)}
		}
	}

	return nil
}

func (cu *FreyjaConfigurationUser) AuditUser() (err error) {
	if len(cu.Keys) == 0 {
		return fmt.Errorf("user '%s' keys is mandatory but found empty", cu.Name)
	} else {
		for _, key := range cu.Keys {
			absPath, err := internal.GetAbsPath(key)
			if err != nil {
				return fmt.Errorf("cannot retrieve absolute path of file '%s' for user '%s'", key, cu.Name)
			}
			if !internal.FileExists(absPath) {
				return fmt.Errorf("user '%s' key file '%s' does not exists", cu.Name, key)
			}
		}
	}
	return nil
}

// **************
// BUILDER CONFIG
// **************

// BuildFromFile generate the configuration from a file
func (c *FreyjaConfiguration) BuildFromFile(path string) error {
	viper.SetConfigType("yaml")
	// load file
	absPath, err := filepath.Abs(path)
	if err != nil {
		log.Printf("Cannot get the absolut path of file '%s'", path)
		return err
	}
	viper.AddConfigPath(absPath)
	viper.SetConfigFile(absPath)
	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			log.Printf("config file not found in %s", absPath)
			return err
		}
		return err
	}
	if err := viper.Unmarshal(&c); err != nil {
		return err
	}
	// set default values if not configured
	// also resolve absolute path values if env vars are set instead
	if err = c.SetValues(); err != nil {
		return err
	}
	// custom user configuration audit
	if err := c.Validate(); err != nil {
		return err
	}
	return nil
}

// **********************
// DEFAULT VALUES SETTERS
// **********************

// SetValues set values to parameters that have not been configured but are still required
// for libvirt
func (c *FreyjaConfiguration) SetValues() (err error) {
	// the user is not the user of the machines in the config
	// it relates to the user that launched freyja
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot get user's home directory: %w", err)
	}
	userSSHDir := filepath.Join(userHomeDir, ".ssh")
	for i, machine := range c.Machines {
		// resolve the image path
		machine.Image = os.ExpandEnv(machine.Image)
		// default user
		if len(machine.Users) == 0 {
			users := make([]FreyjaConfigurationUser, 1)
			users[0] = FreyjaConfigurationUser{
				Name:     DefaultUserName,
				Password: DefaultUserPassword,
				Sudo:     false,
				// injection of a public ssh key path in the config.
				// this path is a symlink to the real location of the public key, because its
				// content will be uploaded in the authorized keys file of the machine.
				// the key must be located in the machine's dir.
				// the key pair is not yet created.
				// the symlinks of the private and the public key must be created in the user's
				// home directory.
				// the symlinks don't exist yet
				Keys: []string{filepath.Join(userSSHDir, machine.Hostname+"_"+DefaultUserName+SSHPublicKeySuffix)},
			}
			machine.Users = users
		} else {
			for j, user := range machine.Users {
				if user.Name == "" {
					user.Name = DefaultUserName
				}
				if user.Password == "" {
					user.Password = DefaultUserPassword
				}
				// if no ssh key is configured for the user, we inject of a public ssh key path in
				// the config.
				// this path is a symlink to the real location of the public key, because its
				// content will be uploaded in the authorized keys file of the machine.
				// the key must be located in the machine's dir.
				// the key pair is not yet created.
				// the symlinks of the private and the public key must be created in the user's
				// home directory.
				// the symlinks don't exist yet
				if len(user.Keys) == 0 {
					user.Keys = []string{filepath.Join(userSSHDir, machine.Hostname+"_"+user.Name+SSHPublicKeySuffix)}
				} else {
					for k, key := range user.Keys {
						user.Keys[k] = os.ExpandEnv(key)
					}
				}
				machine.Users[j] = user
			}
		}
		// default network
		if len(machine.Networks) == 0 {
			networks := make([]FreyjaConfigurationMachineNetwork, 1)
			networks[0] = FreyjaConfigurationMachineNetwork{
				Name: DefaultNetworkName,
				//Interface: DefaultInterfaceName,
			}
			machine.Networks = networks
		} else {
			for _, network := range machine.Networks {
				if network.Name == "" {
					network.Name = DefaultNetworkName
				}
			}
		}
		// default storage
		if machine.Storage == 0 {
			// default '20' GiB
			machine.Storage = DefaultMachineStorage
		}
		if machine.Memory == 0 {
			// default '4096' MiB
			machine.Memory = DefaultMachineMemory
		}
		if machine.Vcpu == 0 {
			// default '1' vcpu
			machine.Vcpu = DefaultMachineVcpu
		}
		if len(machine.Files) != 0 {
			for j, file := range machine.Files {
				// expand env variables in the source path
				file.Source = os.ExpandEnv(file.Source)
				if file.Permissions == "" {
					file.Permissions = DefaultFilePermissions
				}
				if file.Owner == "" {
					file.Owner = DefaultFileOwner
				}
				machine.Files[j] = file
			}
		}

		c.Machines[i] = machine
	}
	return nil
}

// *****
// UTILS
// *****

// GetMachineDir builds the machine directory path from its configuration
func (cm *FreyjaConfigurationMachine) GetMachineDir() string {
	return filepath.Join(FreyjaMachinesWorkspaceDir, cm.Hostname)
}

func GetMachineDirByName(hostname string) (dir string, err error) {
	dir = filepath.Join(FreyjaMachinesWorkspaceDir, hostname)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return "", fmt.Errorf("machine dir '%s' does not exists", dir)
	}
	return dir, nil
}

// CreateMachineDir returns the created dir, or an error
func (cm *FreyjaConfigurationMachine) CreateMachineDir() (string, error) {
	machineDirPath := filepath.Join(FreyjaMachinesWorkspaceDir, cm.Hostname)
	if _, err := os.Stat(machineDirPath); os.IsNotExist(err) {
		if err := os.MkdirAll(machineDirPath, os.ModePerm); err != nil {
			return "", err
		}
	}
	return machineDirPath, nil
}

// GetLibvirtNetworkAddressSlot will search into the machine networks and will return the slot
// address in libvirt format such as '0x02'.
// The two last digits of the address are calculated as follows : hexa(rank in machine networks' list + 2)
func (cm *FreyjaConfigurationMachine) GetLibvirtNetworkAddressSlot(networkName string) (slot string, err error) {
	for i, network := range cm.Networks {
		if network.Name == networkName {
			return fmt.Sprintf("0x%02x", i+2), nil
		}
	}
	return "", fmt.Errorf("cannot find '%s' in machine networks", networkName)
}

// GetCloudInitInterfaceName will search into the machine networks and will return the interface
// name for cloud init provisioning. This is also the name that the interface will have within the
// machine once it is started, such as 'enp0s2'.
// The two last digit of the address is equal to the rank in machine networks' list + 2
func (cm *FreyjaConfigurationMachine) GetCloudInitInterfaceName(networkName string) (slot string, err error) {
	for i, network := range cm.Networks {
		if network.Name == networkName {
			return fmt.Sprintf("enp0s%d", i+2), nil
		}
	}
	return "", fmt.Errorf("cannot find '%s' in machine networks", networkName)
}

// GetNetworkConfigByName look for the network configuration at freyja config file's root with a
// given name
func GetNetworkConfigByName(name string, networks []FreyjaConfigurationNetwork) (config *FreyjaConfigurationNetwork, err error) {
	for _, network := range networks {
		if network.Name == name {
			return &network, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("network config '%s' not found", name))
}

// GetSSHKeysPaths generates the path of a machine's user ssh keys
func GetSSHKeysPaths(machineDir string, username string) (privateKeyPath string, publicKeyPath string) {
	return filepath.Join(machineDir, username+SSHPublicKeySuffix), filepath.Join(machineDir, username+SSHPrivateKeySuffix)
}
