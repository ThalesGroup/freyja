package shellcli

import (
	"fmt"
	"freyja/internal"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"log"
	"os"
	"path/filepath"
)

var configurationPath string

// commands definitions
var machineCreateCmd = &cobra.Command{
	Use:              "create",
	Short:            "Machine creation",
	Long:             "Machine creation using handler and cloud-init or ignition",
	TraverseChildren: true, // ensure local flags do not spread to sub commands

	Run: func(cmd *cobra.Command, args []string) {
		// build config from path
		var configurationData internal.ConfigurationData
		if err := configurationData.BuildFromFile(configurationPath); err != nil {
			Logger.Error("cannot parse configuration", "configuration", configurationPath, "reason", err)
			os.Exit(1)
		}
		// build cloud init config file
		var cloudInitData internal.CloudInitUserData
		for _, machine := range configurationData.Machines {
			if err := cloudInitData.Build(&machine); err != nil {
				Logger.Error("cannot build cloud init specs from machine configuration", "configuration", machine, "reason", err)
				os.Exit(1)
			}
			// create machine directory
			machinePath, err := createMachineDir(&machine)
			if err != nil {
				Logger.Error("cannot create machine workspace directory", "machine", machine.Hostname, "reason", err)
				os.Exit(1)
			}
			// cloud init specs dump YAML file
			cloudInitFilePath := filepath.Join(machinePath, fmt.Sprintf("%s.clinit", machine.Hostname))
			b, err := yaml.Marshal(cloudInitData)
			if err != nil {
				Logger.Error("cannot parse cloud init data into YAML", "data", cloudInitData, "reason", err)
				os.Exit(1)
			}
			if err := os.WriteFile(cloudInitFilePath, b, os.ModePerm); err != nil {
				Logger.Error("cannot write cloud init YAML specs on filesystem", "path", cloudInitFilePath, "reason", err)
				os.Exit(1)
			}

		}

		// TODO :
		//   - build and create the cloud-init from model, using the go module for cloud init
		//   - build and create the vm configuration from model
		//   - create the domain using libvirt with the cloud init iso file for boot provisioning
		//cloudconfig.New()
	},
}

func init() {
	machineCreateCmd.Flags().StringVarP(&configurationPath, "config", "c", "", "Path to the configuration file to create the machines and the networks.")
	if err := machineCreateCmd.MarkFlagRequired("config"); err != nil {
		log.Panic(err.Error())
	}
}

// createMachineDir returns the created dir, or an error
func createMachineDir(machine *internal.ConfigurationMachine) (string, error) {
	machineDirPath := filepath.Join(FreyjaMachinesWorkspaceDir, machine.Hostname)
	if _, err := os.Stat(machineDirPath); os.IsNotExist(err) {
		if err := os.MkdirAll(machineDirPath, os.ModePerm); err != nil {
			return "", err
		}
	}
	return machineDirPath, nil
}

func machineCreate(configurationPath string) (*internal.XMLDomainDescription, error) {
	//config, err := internal.BuildConfiguration(configurationPath)
	//if err != nil {
	//	return nil, err
	//}
	//
	return nil, nil

	// TODO create a metadata to know which domain uses what network
}
