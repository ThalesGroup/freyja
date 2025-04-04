package shellcli

import (
	"fmt"
	"freyja/internal/configuration"
	"github.com/dypflying/go-qcow2lib/qcow2"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"
)

var configurationPath string

var dryRun bool

const BackingImageFormat string = "qcow2"

const BackingImageFilename string = "overlay-image." + BackingImageFormat

const RootImageFileSuffix string = "-root-image." + BackingImageFormat

const XMLMachineDescriptionFilename string = "libvirt-domain.xml"

const XMLNetworkDescriptionPrefix string = "libvirt-network-"
const XMLNetworkDescriptionSuffix string = ".xml"

// MachineNetworkConfig is used as metadata struct to deal with the --dry-run option
// Indeed, when a machine config contains networks, we generate the XML description file for libvirt
// that we dump in the machine's dir for debug, then we register its path and its content inside
// this structure to be created later, during the machine creation.
type MachineNetworkConfig struct {
	Name    string
	Path    string
	Content []byte
}

// commands definitions
var machineCreateCmd = &cobra.Command{
	Use:              "create",
	Short:            "Machine creation",
	Long:             "Machine creation using handler and cloud-init or ignition",
	TraverseChildren: true, // ensure local flags do not spread to sub commands

	Run: func(cmd *cobra.Command, args []string) {
		Logger.Debug("create machines from configuration file", "config", configurationPath)
		// TODO :
		//   - dry-run dont dump xml configuration for networks
		//	 - test different usecases for network and machine creation.
		//		currently, the network creation does not take dhcp conf into account. Fix it.
		//		you can check this by dumping the net-info xml from virsh.
		//   - check that the dry-run option also affect network creation
		//   - test with multiple ssh keys
		//   - create the network if it does not already exists
		// build config from path
		var freyjaConfiguration configuration.FreyjaConfiguration
		if err := freyjaConfiguration.BuildFromFile(configurationPath); err != nil {
			Logger.Error("cannot parse configuration", "configuration", configurationPath, "reason", err.Error())
			os.Exit(1)
		}
		// build cloud init config file
		//var cloudInitData internal.CloudInitUserData
		for _, machine := range freyjaConfiguration.Machines {
			Logger.Info("create", "machine", machine.Hostname)

			// create machine directory
			Logger.Debug("create machine dir", "machine", machine.Hostname, "parent", FreyjaMachinesWorkspaceDir)
			machineDirPath, err := createMachineDir(&machine)
			if err != nil {
				Logger.Error("cannot create machine workspace directory", "machine", machine.Hostname, "reason", err.Error())
				os.Exit(1)
			}

			// create cloud init metadata and user data files
			// YOU MUST name the provision files 'user-data' 'meta-data' !!!!!!!!
			// YOU MUST name the ISO volume 'cidata' !!!!!!
			Logger.Debug("create cloud init user-data and meta-data", "machine", machine.Hostname, "parent", machineDirPath)
			if err := configuration.GenerateCloudInitConfigs(&machine, machineDirPath); err != nil {
				Logger.Error("cannot create machine cloud init configurations", "machine", machine.Hostname, "reason", err.Error())
				os.Exit(1)
			}

			// create cloud-init iso file
			Logger.Debug("create cloud init ISO file", "machine", machine.Hostname, "parent", machineDirPath)
			cloudInitIsoFile, err := configuration.CreateCloudInitIso(&machine, machineDirPath)
			if err != nil {
				Logger.Error("Cannot create machine ISO image file", "machine", machine.Hostname, "reason", err.Error())
				os.Exit(1)
			}

			// Create networks for machines
			// Use the same method as the network create command

			// copy root image to the machine dir
			// !!! NOT SURE IF ROOT IMAGE FILE SHOULD BE COPIED AS WELL
			// basically, no because overlay is made for single machine usage on top of root image
			//rootImageDestinationPath := os.ExpandEnv(filepath.Join(machineDirPath, machine.Hostname+RootImageFileSuffix))
			//rootImageSourcePath := os.ExpandEnv(machine.Image)
			//Logger.Debug("copy machine image file from root", "machine", machine, "root", rootImageSourcePath, "destination", rootImageDestinationPath)
			//if err := internal.CopyFile(rootImageSourcePath, rootImageDestinationPath, 0700); err != nil {
			//	Logger.Error("Cannot copy machine root image file", "machine", machine.Hostname, "reason", fmt.Sprintf("%v", err.Error()))
			//	os.Exit(1)
			//}

			// using : https://github.com/dypflying/go-qcow2lib/blob/main/examples/backing/qcow2_backing.go
			// use 'qemu-img info' to verify it
			rootImageSourcePath := os.ExpandEnv(machine.Image)
			Logger.Debug("create machine image overlay from root image", "machine", machine.Hostname, "parent", machineDirPath, "root", os.ExpandEnv(machine.Image))
			overlayFile, err := createOverlayImage(&machine, rootImageSourcePath, machineDirPath)
			if err != nil {
				Logger.Error("Cannot create machine overlay image file", "machine", machine.Hostname, "reason", fmt.Sprintf("%v", err.Error()))
				os.Exit(1)
			}

			// create the xml description of the libvirt domain from the machine configuration
			// also injects the overlay image file for qemu
			// also injects the cloud init files for startup sequence
			Logger.Debug("create machine's XML libvirt description", "machine", machine.Hostname, "parent", machineDirPath)
			xmlMachineDescription, err := configuration.CreateLibvirtDomainXMLDescription(&machine, overlayFile, cloudInitIsoFile)
			if err != nil {
				Logger.Error("cannot create the libvirt domain XML description from machine configuration", "machine", machine.Hostname, "reason", err.Error())
				os.Exit(1)
			}
			// dump description in machine dir (useful for debug)
			xmlMachineDescriptionPath := filepath.Join(machineDirPath, XMLMachineDescriptionFilename)
			if err := os.WriteFile(xmlMachineDescriptionPath, xmlMachineDescription, 0660); err != nil {
				// the xml configuration has been created but cannot be written on disk
				// this is a warning and not an error since it does not prevent the machine
				// to be created in libvirt
				Logger.Warn("cannot write the libvirt domain XML description in config dir", "machine", machine.Hostname, "path", xmlMachineDescriptionPath, "reason", err.Error())
			}

			// create network configuration
			//xmlNetworkDescriptions, err := GenerateLibvirtNetworksXMLDescriptions(&freyjaConfiguration)
			//if err != nil {
			//	Logger.Error("cannot create the libvirt networks xml descriptions from configuration", "reason", err.Error())
			//	os.Exit(1)
			//}

			// create the machine in libvirt
			if !dryRun {
				// TODO first, create the libvirt networks
				Logger.Debug("create machine's networks")
				// the network is not pushed in libvirt if --dry-run command is used
				//if err = createNetworkFromConfig(&freyjaConfiguration); err != nil {
				//	Logger.Error("cannot create the machine's networks", "machine", machine.Hostname, "reason", err.Error())
				//}

				// second, define the libvirt domain (machine not started yet)
				domain, err := LibvirtConnexion.DomainDefineXML(string(xmlMachineDescription))
				if err != nil {
					Logger.Error("cannot define the machine from libvirt domain XML description", "machine", machine.Hostname, "reason", err.Error())
					os.Exit(1)
				}
				// finally, create the domain (machine startup)
				err = LibvirtConnexion.DomainCreate(domain)
				if err != nil {
					Logger.Error("cannot start the machine", "machine", machine.Hostname, "reason", err.Error())
					os.Exit(1)
				}
			} else {
				Logger.Warn("skipped startup", "machine", machine.Hostname, "reason", "option --dry-run")
			}
		}

	},
}

func init() {
	// MANDATORY --config, -c
	machineCreateCmd.Flags().StringVarP(&configurationPath, "config", "c", "", "Path to the configuration file to create the machines and the networks.")
	if err := machineCreateCmd.MarkFlagRequired("config"); err != nil {
		log.Panic(err.Error())
	}
	// OPTIONAL --dry-run
	machineCreateCmd.Flags().BoolVarP(&dryRun, "dry-run", "", false, "Generate all config files without creating the machine")
}

// getMachineDirByName builds the machine directory path from its configuration
func getMachineDirByName(machineName string) string {
	return filepath.Join(FreyjaMachinesWorkspaceDir, machineName)
}

// createMachineDir returns the created dir, or an error
func createMachineDir(machine *configuration.FreyjaConfigurationMachine) (string, error) {
	machineDirPath := filepath.Join(FreyjaMachinesWorkspaceDir, machine.Hostname)
	if _, err := os.Stat(machineDirPath); os.IsNotExist(err) {
		if err := os.MkdirAll(machineDirPath, os.ModePerm); err != nil {
			return "", err
		}
	}
	return machineDirPath, nil
}

func createOverlayImage(machine *configuration.FreyjaConfigurationMachine, rootImagePath string, machineDir string) (string, error) {
	// using : https://github.com/dypflying/go-qcow2lib/blob/main/examples/backing/qcow2_backing.go
	opts := make(map[string]any)
	backingFile, err := filepath.Abs(rootImagePath)
	if err != nil {
		return "", fmt.Errorf("cannot read base image file '%s' : %w", machine.Image, err)
	}
	overlayFile := filepath.Join(machineDir, BackingImageFilename)
	opts[qcow2.OPT_SIZE] = machine.Storage << 30 //qcow2 file's size is 1g
	opts[qcow2.OPT_FMT] = BackingImageFormat     //qcow2 format
	opts[qcow2.OPT_SUBCLUSTER] = true            //enable sub-cluster
	opts[qcow2.OPT_BACKING] = backingFile
	if err := qcow2.Blk_Create(overlayFile, opts); err != nil {
		return "", fmt.Errorf("failed to create overlay qcow2 file '%s' : %w", overlayFile, err)
	}

	return overlayFile, nil
}
