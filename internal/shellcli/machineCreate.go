package shellcli

import (
	"freyja/internal"
	"github.com/kdomanski/iso9660"
	"github.com/spf13/cobra"
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
			machineDirPath, err := createMachineDir(&machine)
			if err != nil {
				Logger.Error("cannot create machine workspace directory", "machine", machine.Hostname, "reason", err)
				os.Exit(1)
			}
			// create cloud init metadata and user data files
			if err := internal.GenerateCloudInitConfigs(&machine, machineDirPath); err != nil {
				Logger.Error("cannot create machine cloud init configurations", "machine", machine.Hostname, "reason", err)
				os.Exit(1)
			}
			// create iso from user inputs
			doit := false
			if doit {
				// TODO :
				//  - try to use genisoimage or similar in order to create the iso image from the cloud init file:
				//  for example, 'xorriso' or 'genisoimage' are two utilities to generate ISO images from cloud init files, according to the RFC9660 specifications
				//    https://github.com/kdomanski/iso9660
				//cloudconfig.New()
				isoWriter, err := iso9660.NewWriter()
				if err != nil {
					Logger.Error("cannot create the iso9660 writer for the image", "image", machine.Hostname, "reason", err)
					os.Exit(1)
				}
				// add cloud init metadata
				// todo: provide the cloud init metadata file path
				fm, err := os.Open("/metadata-file")
				if err != nil {
					Logger.Error("")
					os.Exit(1)
				}
				defer fm.Close()
				if err = isoWriter.AddFile(fm, "meta-data"); err != nil {
					Logger.Error("")
					os.Exit(1)
				}
				// add cloud init user data
				// todo: provide the cloud init user data file path
				fu, err := os.Open("/userdata-file")
				if err != nil {
					Logger.Error("")
					os.Exit(1)
				}
				defer fu.Close()
				if err = isoWriter.AddFile(fu, "user-data"); err != nil {
					Logger.Error("")
					os.Exit(1)
				}
				// write iso on filesystem
				// todo: provide the iso output file path
				outputFile, err := os.OpenFile("/home/user/output.iso", os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
				if err != nil {
					Logger.Error("")
					os.Exit(1)
				}
				// todo: provide a dedicated volume id for this iso image
				if err = isoWriter.WriteTo(outputFile, "volume id"); err != nil {
					Logger.Error("")
					os.Exit(1)
				}
				if err = outputFile.Close(); err != nil {
					Logger.Error("")
					os.Exit(1)
				}
			}

		}

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
