package shellcli

import (
	"crypto/ed25519"
	"errors"
	"fmt"
	"freyja/internal"
	"freyja/internal/configuration"
	"log"
	"os"
	"path/filepath"

	"github.com/dypflying/go-qcow2lib/qcow2"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
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
		createMachine(configurationPath)

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

// createMachine is the main function used to generate the machines config files and start them in
// libvirt
func createMachine(configurationPath string) {
	Logger.Debug("create machines from configuration file", "config", configurationPath)
	// TODO :
	//   - implement the command 'freyja ssh -n vm1' to avoid 'freyja machine info -n vm1' to get ip beforehand
	//		tutorial : https://www.pixelstech.net/article/1699714722-guide-to-implement-an-ssh-client-using-golang
	//	   the key configuration must be reviewed :
	//			- rename the conf 'keys' as 'sshPublicKeys'
	//			- for now, ssh public keys are configured per user as a list.
	//	     		It means a user can give multiple public keys.
	//				It is good because one user might support connexions from multiple remote hosts.
	//				These public keys are uploaded in the home dir of this user in the guest.
	//				if no public keys are provided, freyja generates a pair automatically in the machine dir
	//				if the user doesn't provide a key, inject the public key path in the SSHPublicKey conf
	//			- private keys are not precised in configuration.
	//			  until now, the private keys were configured in '.ssh/config'
	//		      in golang's ssh package, the pair has to be provided.
	//            so we need a way to find the private key from the public key
	//			  this is feasible only if freyja auto generates the pair. Otherwise it is up to the
	//			  user to configure it in .ssh/config.
	//			  in other words, we need a new '--private-key' optional flag for freyja ssh command
	//			  if not provided, freyja is looking into the machine dir and search the keys by name,
	//				  using the nomenclature of the auto generated keys.
	//	 - implements features included in python version
	//   - customize errors, especially for freyja configuration, cloud init and libvirt error. Check conditions if strings.contain() to replace them with if errors.As(err, &myCustomErr)
	// build config from path
	var freyjaConfiguration configuration.FreyjaConfiguration

	// BUILD CONFIG
	if err := freyjaConfiguration.BuildFromFile(configurationPath); err != nil {
		Logger.Error("cannot parse configuration", "configuration", configurationPath, "reason", err.Error())
		os.Exit(1)
	}

	var createdMachines []string
	for _, machine := range freyjaConfiguration.Machines {
		Logger.Info("create", "machine", machine.Hostname)

		// BUILD PROVISION FILES
		cloudInitIsoFile, overlayFile, err := BuildProvisioningFiles(machine)
		if err != nil {
			Logger.Error("cannot build provisioning files", "machine", machine.Hostname, "reason", err.Error())
			os.Exit(1)
		}

		// AUDIT PROVISIONING
		if err = machine.Audit(); err != nil {
			Logger.Error("machine provision files audit failed", "machine", machine.Hostname, "reason", err.Error())
			os.Exit(1)
		}

		// DOMAIN LIBVIRT CONF
		xmlMachineDescription, err := BuildLibvirtMachineDescription(machine, cloudInitIsoFile, overlayFile)
		if err != nil {
			Logger.Error("cannot build libvirt machine description", "machine", machine.Hostname, "reason", err.Error())
			os.Exit(1)
		}

		// NETWORK LIBVIRT CONF
		xmlNetworkDescriptions, err := BuildLibvirtNetworksDescription(freyjaConfiguration)
		if err != nil {
			Logger.Error("cannot build libvirt network descriptions", "machine", machine.Hostname, "reason", err.Error())
			os.Exit(1)
		}

		// create the default ssh keys if needed for each user without an ssh key config
		//for _, user := range machine.Users {
		//	if err := GenerateSSHKeyPair(machineDirPath, user.Name); err != nil {
		//		Logger.Error("cannot create machine SSH key pair for the given user", "machine", machine.Hostname, "user", user.Name, "reason", err.Error())
		//		os.Exit(1)
		//	}
		//}

		// CREATE & START
		if !dryRun {
			// first, create the networks for machines, if any
			Logger.Debug("create machine's networks")
			if err := CreateNetworksInLibvirt(xmlNetworkDescriptions); err != nil {
				var alreadyExistsError *internal.NetworkAlreadyExistsError
				if errors.As(err, &alreadyExistsError) {
					Logger.Warn("skipped network creation: already exists", "network", alreadyExistsError.Network)
				} else {
					Logger.Error("cannot create networks in Libvirt from XML descriptions", "reason", err.Error())
					os.Exit(1)
				}

			}

			// second, define the libvirt domain (machine not started yet)
			domain, err := LibvirtConnexion.DomainDefineXML(string(xmlMachineDescription))
			if err != nil {
				Logger.Error("cannot define the machine in libvirt from domain XML description", "machine", machine.Hostname, "reason", err.Error())
				os.Exit(1)
			}
			// finally, create the domain (machine startup)
			err = LibvirtConnexion.DomainCreate(domain)
			if err != nil {
				Logger.Error("cannot start the machine", "machine", machine.Hostname, "reason", err.Error())
				os.Exit(1)
			}
			createdMachines = append(createdMachines, machine.Hostname)
		} else {
			Logger.Warn("skipped startup", "machine", machine.Hostname, "reason", "option --dry-run")
		}
	}

	Logger.Info("created", "machines", createdMachines)
}

// BuildProvisioningFiles builds all the necessary files in the machine configuration directory
// inside
// TODO create the ssh keys by default
func BuildProvisioningFiles(machine configuration.FreyjaConfigurationMachine) (cloudInitISOPath string, imageOverlayPath string, err error) {
	// create machine directory
	Logger.Debug("create machine dir", "machine", machine.Hostname, "parent", configuration.FreyjaMachinesWorkspaceDir)
	machineDirPath, err := machine.CreateMachineDir()
	if err != nil {
		return "", "", fmt.Errorf("cannot create machine workspace directory for machine '%s': %w", machine.Hostname, err)
	}

	// dump freyja config with injected values in the machine's dir
	// the config will contain the current machine config only
	// very useful for debugging or duplicating the machine
	machineConfigDump := configuration.FreyjaConfiguration{
		Version:  configuration.Version,
		Machines: []configuration.FreyjaConfigurationMachine{machine},
	}
	machineConfigDumpBytes, err := yaml.Marshal(machineConfigDump)
	if err != nil {
		return "", "", fmt.Errorf("cannot marshalize freyja '%s' configuration dump: %w", machine.Hostname, err)
	}
	machineConfigDumpPath := filepath.Join(machine.GetMachineDir(), "freyja-"+machine.Hostname+"-config-dump.yaml")
	if err := os.WriteFile(machineConfigDumpPath, machineConfigDumpBytes, 0660); err != nil {
		return "", "", fmt.Errorf("cannot write freyja '%s' configuration dump: %w", machine.Hostname, err)
	}

	// create cloud init metadata and user data files
	// YOU MUST name the provision files 'user-data' 'meta-data' !!!!!!!!
	// YOU MUST name the ISO volume 'cidata' !!!!!!
	Logger.Debug("create cloud init user-data and meta-data", "machine", machine.Hostname, "parent", machineDirPath)
	if err := configuration.GenerateCloudInitConfigs(&machine, machineDirPath); err != nil {
		return "", "", fmt.Errorf("cannot create cloud init configurations for machine '%s': %w", machine.Hostname, err)
	}

	// create cloud-init iso file
	Logger.Debug("create cloud init ISO file", "machine", machine.Hostname, "parent", machineDirPath)
	cloudInitISOPath, err = configuration.CreateCloudInitIso(&machine, machineDirPath)
	if err != nil {
		return "", "", fmt.Errorf("cannot create cloud-init ISO file for machine '%s': %w", machine.Hostname, err)
	}

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
	imageOverlayPath, err = createOverlayImage(&machine, rootImageSourcePath, machineDirPath)
	if err != nil {
		return "", "", fmt.Errorf("cannot create overlay image file for machine '%s': %w", machine.Hostname, err)
	}

	// Default SSH keys generation.
	// Prior to this step, during configuration build phase, if no SSH key is provided for a user
	// in the configuration, a public ssh key path is automatically injected in this user
	// configuration.
	// If it is the case, we must create the key pair since it is not created yet.
	for _, user := range machine.Users {
		if configuration.IsSSHPublicKeyDefault(machineDirPath, user.Name, user.SSHPublicKeys[0]) {
			if err := GenerateSSHKeyPair(machineDirPath, user.Name); err != nil {
				Logger.Error("cannot create machine SSH key pair for the given user", "machine", machine.Hostname, "user", user.Name, "reason", err.Error())
				os.Exit(1)
			}
		}
	}

	return
}

func BuildLibvirtMachineDescription(machine configuration.FreyjaConfigurationMachine, cloudInitIsoPath string, imageOverlayPath string) (xmlMachineDescription []byte, err error) {
	machineDirPath := machine.GetMachineDir()
	// create the xml description of the libvirt domain from the machine configuration
	// also injects the overlay image file for qemu
	// also injects the cloud init files for startup sequence
	Logger.Debug("create machine's XML libvirt description", "machine", machine.Hostname, "parent", machineDirPath)
	xmlMachineDescription, err = configuration.CreateLibvirtDomainXMLDescription(&machine, imageOverlayPath, cloudInitIsoPath)
	if err != nil {
		return nil, fmt.Errorf("cannot create the libvirt domain XML description from configuration for machine '%s': %w", machine.Hostname, err)
	}

	// dump description in machine dir (useful for debug)
	xmlMachineDescriptionPath := filepath.Join(machineDirPath, XMLMachineDescriptionFilename)
	if err := os.WriteFile(xmlMachineDescriptionPath, xmlMachineDescription, 0660); err != nil {
		// the xml configuration has been created but cannot be written on disk
		// this is a warning and not an error since it does not prevent the machine
		// to be created in libvirt
		Logger.Warn("cannot write the libvirt domain XML description in config dir", "machine", machine.Hostname, "path", xmlMachineDescriptionPath, "reason", err.Error())
	}

	return

}

func BuildLibvirtNetworksDescription(freyjaConfiguration configuration.FreyjaConfiguration) (xmlNetworkDescriptions map[string][]byte, err error) {
	// create network configuration
	xmlNetworkDescriptions, err = GenerateLibvirtNetworksXMLDescriptions(&freyjaConfiguration, configuration.FreyjaNetworksWorkspaceDir)
	if err != nil {
		Logger.Error("cannot create the libvirt networks xml descriptions from configuration", "reason", err.Error())
		os.Exit(1)
	}

	return
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

// GenerateSSHKeyPair create an ed25519 for a given user, inside the given machine config directory
func GenerateSSHKeyPair(dir string, username string) (err error) {
	// create ssh key pair with ed25519 algorithm
	// we use nil io.Reader to let ed25519 package use crypto/rand by itself
	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		return fmt.Errorf("cannot generate ed25519 public/private key for machine '%s' and user '%s': %w", dir, username, err)
	}
	publicKeyPath, privateKeyPath := configuration.GetSSHKeysPaths(dir, username)
	if err := os.WriteFile(publicKeyPath, publicKey, 0660); err != nil {
		return fmt.Errorf("cannot write ed25519 public key for machine '%s' and user '%s': %w", dir, username, err)
	}
	if err := os.WriteFile(privateKeyPath, privateKey, 0660); err != nil {
		return fmt.Errorf("cannot write ed25519 private key: for machine '%s' and user '%s': %w", dir, username, err)
	}

	return nil
}
