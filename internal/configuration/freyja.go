package configuration

import (
	"errors"
	"fmt"
	"freyja/internal"
	"github.com/spf13/viper"
	"log"
	"path/filepath"
	"regexp"
)

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

const DefaultInterfaceName string = "virbr0"

// FreyjaConfiguration is the base model for freyja configuration parameters
// Example :
// ---
// version: v0.1.0-beta
// networks:
//   - name: ctrl-plane
//     dhcp:
//     start: 192.168.123.3
//     end: 192.168.123.254
//   - name: data-plane
//     dhcp:
//     start: 192.168.124.3
//     end: 192.168.124.254
//
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
type FreyjaConfiguration struct {
	Version  string                       `yaml:"version"`
	Machines []FreyjaConfigurationMachine `yaml:"machines"`
	Networks []FreyjaConfigurationNetwork `yaml:"networks,omitempty"`
}

// FreyjaConfigurationMachine is the configuration model for libvirt guest parameters
type FreyjaConfigurationMachine struct {
	// MANDATORY
	Image    string `yaml:"image"`    // Qcow2 image file path on host
	Os       string `yaml:"os"`       // os type in libosinfo
	Hostname string `yaml:"hostname"` // domain name in libvirt
	// optional
	Networks []FreyjaConfigurationMachineNetwork `yaml:"networks,omitempty"`
	Users    []FreyjaConfigurationUser           `yaml:"users,omitempty"`
	Storage  uint                                `yaml:"storage"` // GiB
	Memory   uint                                `yaml:"memory"`  // MiB
	Vcpu     uint                                `yaml:"vcpu"`
	Packages []string                            `yaml:"packages"`
	Cmd      []string                            `yaml:"cmd"`
	Files    []FreyjaConfigurationFile           `yaml:"files"`
	Update   bool                                `yaml:"update"`
	Reboot   bool                                `yaml:"reboot"`
}

type FreyjaConfigurationMachineNetwork struct {
	Name      string `yaml:"name"`
	Mac       string `yaml:"mac"`
	Interface string `yaml:"interface"`
}

type FreyjaConfigurationUser struct {
	Name     string   `yaml:"name"`
	Password string   `yaml:"password"`
	Sudo     bool     `yaml:"sudo"`
	Groups   []string `yaml:"groups"`
	Keys     []string `yaml:"keys"`
}

type FreyjaConfigurationFile struct {
	Source      string `yaml:"source"`
	Destination string `yaml:"destination"`
	Permissions string `yaml:"permissions"`
	Owner       string `yaml:"owner"`
}

type FreyjaConfigurationNetwork struct {
	Name string                         `yaml:"name"`
	Dhcp FreyjaConfigurationNetworkDHCP `yaml:"dhcp,omitempty"`
}

type FreyjaConfigurationNetworkDHCP struct {
	// DHCP range
	Start string `yaml:"start"`
	End   string `yaml:"end"`
}

type Configuration interface {
	Validate() error
	BuildFromFile(path string) error
}

// Validate audits the whole freyja configuration for content mistakes
// mistake = configuration value that may cause further issues during machine creation in libvirt
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
	if len(c.Machines) == 0 {
		return errors.New("configure at least one machine but found 0")
	}
	for _, machine := range c.Machines {
		if len(machine.Networks) != 0 {
			// verify networks
			for _, network := range machine.Networks {
				err = network.validateMachineNetwork()
				if err != nil {
					return &internal.ConfigurationError{Message: err.Error()}
				}
			}
			// verify users
			for _, user := range machine.Users {
				err = user.validateUser()
				if err != nil {
					return &internal.ConfigurationError{Message: err.Error()}
				}
			}
			// verify files
			for _, file := range machine.Files {
				err = file.validateFiles()
				if err != nil {
					return &internal.ConfigurationError{Message: err.Error()}
				}
			}
		}
	}
	return nil
}

// ValidateVersion audits the version configuration
func (c *FreyjaConfiguration) validateVersion() error {
	if c.Version == "" {
		return errors.New("missing version value")
	}
	regex := "^[a-z]*[0-9]+\\.[0-9]+.*$"
	match, err := regexp.MatchString(regex, c.Version)
	if err != nil {
		return fmt.Errorf("cannot verify pattern matching for string '%s': %w", c.Version, err)
	}
	if !match {
		return errors.New(fmt.Sprintf("wrong version format : regex is '%s' but found value '%s'", regex, c.Version))
	}
	return nil
}

func (c *FreyjaConfiguration) validateNetworks() error {
	for _, network := range c.Networks {
		if network.Name == "" {
			return fmt.Errorf("missing network name")
		}
		pNetworkDhcp := &network.Dhcp
		if pNetworkDhcp == nil {
			return fmt.Errorf("missing DHCP configuration for network '%s'", network.Name)
		}
		if network.Dhcp.Start == "" {
			return fmt.Errorf("missing DHCP start of range for network '%s'", network.Name)
		}
		if network.Dhcp.End == "" {
			return fmt.Errorf("missing DHCP end of range for network '%s'", network.Name)
		}
	}
	return nil
}

// ValidateNetwork audits the network configuration including
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

// ValidateUser audits the user configuration including
//   - the path of the keys on the host
func (cu *FreyjaConfigurationUser) validateUser() error {
	if len(cu.Keys) > 0 {
		for _, key := range cu.Keys {
			// check if the key path's file exists
			if !internal.FileExists(key) {
				return fmt.Errorf("user key file '%s' does not exists", key)
			}
		}
	}
	return nil
}

func (cf *FreyjaConfigurationFile) validateFiles() error {
	// validate source
	if !internal.FileExists(cf.Source) {
		return errors.New(fmt.Sprintf("configuration not found : '%s' does not exist", cf.Source))
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
	c.setDefaultValues()
	// custom user configuration audit
	if err := c.Validate(); err != nil {
		return err
	}
	return nil
}

// setDefaultValues set values to parameters that have not been configured but are still required
// for libvirt
func (c *FreyjaConfiguration) setDefaultValues() {
	for i, machine := range c.Machines {
		// default user
		if len(machine.Users) == 0 {
			users := make([]FreyjaConfigurationUser, 1)
			users[0] = FreyjaConfigurationUser{
				Name:     DefaultUserName,
				Password: DefaultUserPassword,
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
		//if len(machine.Files) != 0 {
		//	for j, file := range machine.Files {
		//		if file.Permissions == "" {
		//			file.Permissions = DefaultFilePermissions
		//		}
		//		if file.Owner == "" {
		//			file.Owner = DefaultFileOwner
		//		}
		//		machine.Files[j] = file
		//	}
		//}

		c.Machines[i] = machine
	}
}

func (c *FreyjaConfiguration) setUsers() {
	//machines := make([]FreyjaConfigurationMachine, len(c.Machines))
	for i, machine := range c.Machines {
		users := make([]FreyjaConfigurationUser, 1)
		users[0] = FreyjaConfigurationUser{
			Name:     DefaultUserName,
			Password: DefaultUserPassword,
		}
		machine.Users = users
		c.Machines[i] = machine
	}
}
