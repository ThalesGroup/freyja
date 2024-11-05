package internal

import (
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"log"
	"path/filepath"
	"regexp"
)

type Configuration interface {
	Validate() error
	BuildFromFile(path string) error
}

// Validate audits the whole freyja configuration for content mistakes
// mistake = configuration value that may cause further issues during machine creation in libvirt
func (c *ConfigurationData) Validate() error {
	// verify version
	err := c.validateVersion()
	if err != nil {
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
				err = network.validateNetwork()
				if err != nil {
					return &ConfigurationError{Message: err.Error()}
				}
			}
			// verify users
			for _, user := range machine.Users {
				err = user.validateUser()
				if err != nil {
					return &ConfigurationError{Message: err.Error()}
				}
			}
			// verify files
			for _, file := range machine.Files {
				err = file.validateFiles()
				if err != nil {
					return &ConfigurationError{Message: err.Error()}
				}
			}
		}
	}
	return nil
}

// ValidateVersion audits the version configuration
func (c *ConfigurationData) validateVersion() error {
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

// ValidateNetwork audits the network configuration including
//   - the name of the network
//   - the format of the mac address
func (cn *ConfigurationNetwork) validateNetwork() error {
	if cn.Name == "" {
		return errors.New("network name is empty")
	}
	macRegex := "^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$"
	match, err := regexp.MatchString(macRegex, cn.Mac)
	if err != nil {
		return fmt.Errorf("cannot verify pattern matching for string '%s': %w", cn.Mac, err)
	}
	if !match {
		return errors.New(fmt.Sprintf("wrong mac format : regex is '%s' but found value '%s'", macRegex, cn.Mac))
	}
	return nil
}

// ValidateUser audits the user configuration including
//   - the path of the keys on the host
func (cu *ConfigurationUser) validateUser() error {
	if len(cu.Keys) > 0 {
		for _, key := range cu.Keys {
			// check if the key path's file exists
			if err := CheckIfFileExists(key); err != nil {
				return fmt.Errorf("user key file '%s' does not exists: %w", key, err)
			}
		}
	}
	return nil
}

func (cf *ConfigurationFile) validateFiles() error {
	// validate source
	if err := CheckIfFileExists(cf.Source); err != nil {
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
func (c *ConfigurationData) BuildFromFile(path string) error {
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
func (c *ConfigurationData) setDefaultValues() {
	for i, machine := range c.Machines {
		// default user
		if len(machine.Users) == 0 {
			users := make([]ConfigurationUser, 1)
			users[0] = ConfigurationUser{
				Name:     DefaultUserName,
				Password: DefaultUserPassword,
			}
			machine.Users = users
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

func (c *ConfigurationData) setUsers() {
	//machines := make([]ConfigurationMachine, len(c.Machines))
	for i, machine := range c.Machines {
		users := make([]ConfigurationUser, 1)
		users[0] = ConfigurationUser{
			Name:     DefaultUserName,
			Password: DefaultUserPassword,
		}
		machine.Users = users
		c.Machines[i] = machine
	}
}
