package shellcli

import (
	"crypto/ed25519"
	"encoding/pem"
	"errors"
	"fmt"
	"freyja/internal"
	"freyja/internal/configuration"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/dypflying/go-qcow2lib/qcow2"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
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

	// NETWORK LIBVIRT CONF
	// the networks must exist in libvirt before the machine descriptions are
	// generated, because the machine interfaces plug into the host bridges that
	// libvirt allocates for these networks
	xmlNetworkDescriptions, err := BuildLibvirtNetworksDescription(freyjaConfiguration)
	if err != nil {
		Logger.Error("cannot build libvirt network descriptions", "reason", err.Error())
		os.Exit(1)
	}
	if !dryRun {
		Logger.Debug("create machines' networks")
		if err := CreateNetworksInLibvirt(xmlNetworkDescriptions); err != nil {
			if alreadyExistsError, ok := errors.AsType[*internal.NetworkAlreadyExistsError](err); ok {
				Logger.Warn("skipped network creation: already exists", "network", alreadyExistsError.Network)
			} else {
				Logger.Error("cannot create networks in Libvirt from XML descriptions", "reason", err.Error())
				os.Exit(1)
			}

		}
	} else {
		Logger.Warn("skipped networks creation", "reason", "option --dry-run")
	}

	// NETWORK HOST BRIDGES
	// resolved and audited for every machine before creating anything, so that a host
	// missing the qemu-bridge-helper setup fails fast instead of leaving half-created machines
	warnBridgeHelperNotSetuid()
	machinesNetworkBridges := make([]map[string]string, len(freyjaConfiguration.Machines))
	var usedBridges []string
	for i, machine := range freyjaConfiguration.Machines {
		networkBridges, err := ResolveNetworkBridges(machine)
		if err != nil {
			Logger.Error("cannot resolve the host bridges of the machine's networks", "machine", machine.Hostname, "reason", err.Error())
			os.Exit(1)
		}
		machinesNetworkBridges[i] = networkBridges
		for _, bridgeName := range networkBridges {
			usedBridges = append(usedBridges, bridgeName)
		}
	}
	if err := auditBridgeHelperSetup(BridgeHelperConfPath, usedBridges); err != nil {
		fix := fmt.Sprintf("see the 'Rootless setup' section of the installation requirements, or allow the bridges '%s' and '%s0' to '%s%d' in %s",
			configuration.DefaultNetworkBridgeName, configuration.FreyjaBridgePrefix, configuration.FreyjaBridgePrefix, configuration.MaxFreyjaBridges-1, BridgeHelperConfPath)
		if !dryRun {
			Logger.Error("the host is not set up to let the machines plug into the host bridges", "reason", err.Error(), "fix (as admin)", fix)
			os.Exit(1)
		}
		Logger.Warn("the host is not set up to let the machines plug into the host bridges", "reason", err.Error(), "fix (as admin)", fix)
	}

	var createdMachines []string
	for i, machine := range freyjaConfiguration.Machines {
		Logger.Info("create", "machine", machine.Hostname)
		networkBridges := machinesNetworkBridges[i]

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
		xmlMachineDescription, err := BuildLibvirtMachineDescription(machine, cloudInitIsoFile, overlayFile, networkBridges)
		if err != nil {
			Logger.Error("cannot build libvirt machine description", "machine", machine.Hostname, "reason", err.Error())
			os.Exit(1)
		}

		// CREATE & START
		if !dryRun {
			// first, define the libvirt domain (machine not started yet)
			domain, err := LibvirtDomainConnexion.DomainDefineXML(string(xmlMachineDescription))
			if err != nil {
				Logger.Error("cannot define the machine in libvirt from domain XML description", "machine", machine.Hostname, "reason", err.Error())
				os.Exit(1)
			}
			// second, create the domain (machine startup)
			err = LibvirtDomainConnexion.DomainCreate(domain)
			if err != nil {
				Logger.Error("cannot start the machine", "machine", machine.Hostname, "reason", err.Error())
				// undefine the machine so that a new 'create' does not collide with it;
				// its dir is kept for debugging and is reused by the next 'create'
				if undefineErr := LibvirtDomainConnexion.DomainUndefine(domain); undefineErr != nil {
					Logger.Error("cannot undefine the machine that failed to start", "machine", machine.Hostname, "reason", undefineErr.Error(),
						"fix", fmt.Sprintf("virsh -c qemu:///session undefine %s", machine.Hostname))
				} else {
					Logger.Info("undefined the machine that failed to start, its dir is kept for debugging", "machine", machine.Hostname, "dir", machine.GetMachineDir())
				}
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

	// Default SSH keys generation.
	// Prior to this step, during configuration build phase, if no SSH key is provided for a user
	// in the configuration, a public ssh key path is automatically injected in this user
	// configuration.
	// If it is the case, we must create the key pair since it is not created yet.
	for _, user := range machine.Users {
		sshPubKeyIsDefault, err := configuration.IsSSHPublicKeyDefault(machine.Hostname, user.Name, user.SSHPublicKeys[0])
		if err != nil {
			return "", "", fmt.Errorf("cannot verify if key '%s' is a default generated key: %w", user.SSHPublicKeys[0], err)
		}
		if sshPubKeyIsDefault {
			Logger.Info("generates default SSH keys in machine dir", "machine", machine.Hostname, "user", user.Name, "dir", machineDirPath)
			if err := GenerateSSHKeyPair(machine.Hostname, user.Name); err != nil {
				Logger.Error("cannot create machine SSH key pair for the given user", "machine", machine.Hostname, "user", user.Name, "reason", err.Error())
				os.Exit(1)
			}
		}
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



	return
}

func BuildLibvirtMachineDescription(machine configuration.FreyjaConfigurationMachine, cloudInitIsoPath string, imageOverlayPath string, networkBridges map[string]string) (xmlMachineDescription []byte, err error) {
	machineDirPath := machine.GetMachineDir()
	// create the xml description of the libvirt domain from the machine configuration
	// also injects the overlay image file for qemu
	// also injects the cloud init files for startup sequence
	// also injects the host bridges of the machine's networks for the interfaces
	Logger.Debug("create machine's XML libvirt description", "machine", machine.Hostname, "parent", machineDirPath)
	xmlMachineDescription, err = configuration.CreateLibvirtDomainXMLDescription(&machine, imageOverlayPath, cloudInitIsoPath, networkBridges)
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

// ResolveNetworkBridges maps each network of the machine to the host bridge that
// the system libvirt daemon allocated for it. The machine runs in the unprivileged
// session daemon, which cannot resolve the system networks by name, so its
// interfaces plug directly into the networks' host bridges.
func ResolveNetworkBridges(machine configuration.FreyjaConfigurationMachine) (map[string]string, error) {
	networkBridges := make(map[string]string, len(machine.Networks))
	for _, network := range machine.Networks {
		bridgeName, err := GetNetworkBridgeName(network.Name)
		if err != nil {
			if dryRun {
				// the network may legitimately not exist yet, since its creation is
				// skipped by the --dry-run option
				Logger.Warn("cannot resolve the network's host bridge", "machine", machine.Hostname, "network", network.Name, "reason", err.Error())
				networkBridges[network.Name] = unresolvedBridgeName
				continue
			}
			return nil, fmt.Errorf("cannot resolve the host bridge of network '%s' for machine '%s': %w", network.Name, machine.Hostname, err)
		}
		networkBridges[network.Name] = bridgeName
	}
	return networkBridges, nil
}

// BridgeHelperConfPath is the ACL file of qemu-bridge-helper: it must allow every
// host bridge that the unprivileged machines plug their interfaces into
const BridgeHelperConfPath = "/etc/qemu/bridge.conf"

// bridgeHelperPaths are the usual locations of qemu-bridge-helper, per distro family
var bridgeHelperPaths = []string{"/usr/lib/qemu/qemu-bridge-helper", "/usr/libexec/qemu-bridge-helper"}

// unresolvedBridgeName stands for a network bridge that cannot be resolved because
// the network creation was skipped by the --dry-run option
const unresolvedBridgeName = "dry-run-unresolved"

// maxBridgeHelperACLIncludeDepth stops 'include' loops in the qemu-bridge-helper ACL files
const maxBridgeHelperACLIncludeDepth = 10

// warnBridgeHelperNotSetuid warns when qemu-bridge-helper is not setuid. It is only a
// warning since some distros grant the helper its privileges another way (e.g. file
// capabilities), which freyja cannot check reliably.
func warnBridgeHelperNotSetuid() {
	for _, helperPath := range bridgeHelperPaths {
		info, err := os.Stat(helperPath)
		if err != nil {
			continue
		}
		if info.Mode()&os.ModeSetuid == 0 {
			Logger.Warn("qemu-bridge-helper is not setuid: the machines may not plug into the host bridges",
				"helper", helperPath,
				"fix (as admin)", fmt.Sprintf("dpkg-statoverride --update --add root libvirt 4750 %s", helperPath))
		}
		return
	}
}

// auditBridgeHelperSetup checks that the qemu-bridge-helper ACL file at aclPath lets the
// machines plug into the given host bridges. The machines run unprivileged in the session
// daemon and rely on this setuid helper to plug into the bridges of the system daemon's
// networks. The ACL file requires root once per host, so freyja only reports the problem
// for the admin instead of fixing it.
// It returns an error when the ACL file is missing or does not allow one of the bridges.
// An ACL file that the user cannot read (but the setuid helper can) is only warned about.
func auditBridgeHelperSetup(aclPath string, bridges []string) error {
	acl, err := parseBridgeHelperACL(aclPath)
	if errors.Is(err, os.ErrPermission) {
		Logger.Warn("cannot verify that the qemu-bridge-helper ACL file allows the machines' host bridges", "path", aclPath, "reason", err.Error())
		return nil
	} else if err != nil {
		return fmt.Errorf("invalid qemu-bridge-helper ACL file: %w", err)
	}

	var notAllowed []string
	checked := make(map[string]bool)
	for _, bridgeName := range bridges {
		if bridgeName == unresolvedBridgeName || checked[bridgeName] {
			continue
		}
		checked[bridgeName] = true
		if !acl.allows(bridgeName) {
			notAllowed = append(notAllowed, bridgeName)
		}
	}
	if len(notAllowed) > 0 {
		return fmt.Errorf("the qemu-bridge-helper ACL file '%s' does not allow the host bridges %v (an 'allow' rule must match them and no 'deny' rule)", aclPath, notAllowed)
	}
	return nil
}

// bridgeHelperACL mirrors the access rules of qemu-bridge-helper: a bridge is allowed
// when an 'allow' rule matches it and no 'deny' rule does, whatever the rules order
type bridgeHelperACL struct {
	allowAll bool
	denyAll  bool
	allowed  map[string]bool
	denied   map[string]bool
}

func (acl *bridgeHelperACL) allows(bridgeName string) bool {
	isAllowed := acl.allowAll || acl.allowed[bridgeName]
	isDenied := acl.denyAll || acl.denied[bridgeName]
	return isAllowed && !isDenied
}

// parseBridgeHelperACL reads a qemu-bridge-helper ACL file, following its 'include' rules
func parseBridgeHelperACL(path string) (*bridgeHelperACL, error) {
	acl := &bridgeHelperACL{allowed: make(map[string]bool), denied: make(map[string]bool)}
	if err := acl.parseFile(path, 0); err != nil {
		return nil, err
	}
	return acl, nil
}

// parseFile adds the rules of an ACL file. The syntax is one rule per line among
// 'allow <bridge|all>', 'deny <bridge|all>' and 'include <file>', '#' starting a comment line.
// Unknown rules are errors, like for qemu-bridge-helper.
func (acl *bridgeHelperACL) parseFile(path string, depth int) error {
	if depth > maxBridgeHelperACLIncludeDepth {
		return fmt.Errorf("too many nested 'include' rules, stopped at '%s'", path)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	for lineNumber, line := range strings.Split(string(content), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 || strings.HasPrefix(fields[0], "#") {
			continue
		}
		if len(fields) != 2 {
			return fmt.Errorf("'%s' line %d: expected '<allow|deny|include> <argument>', got '%s'", path, lineNumber+1, strings.TrimSpace(line))
		}
		rule, argument := fields[0], fields[1]
		switch rule {
		case "allow":
			if argument == "all" {
				acl.allowAll = true
			} else {
				acl.allowed[argument] = true
			}
		case "deny":
			if argument == "all" {
				acl.denyAll = true
			} else {
				acl.denied[argument] = true
			}
		case "include":
			if err := acl.parseFile(argument, depth+1); err != nil {
				return err
			}
		default:
			return fmt.Errorf("'%s' line %d: unknown rule '%s'", path, lineNumber+1, rule)
		}
	}
	return nil
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

// SSHPrivateKeyFileMode is required by OpenSSH, which refuses private keys readable by others
const SSHPrivateKeyFileMode os.FileMode = 0600

const SSHPublicKeyFileMode os.FileMode = 0644

// GenerateSSHKeyPair creates an ed25519 SSH key pair in OpenSSH format for a given user,
// inside the config directory of the given machine
func GenerateSSHKeyPair(machineName string, username string) error {
	comment := fmt.Sprintf("%s@%s", username, machineName)
	privateKey, publicKey, err := marshalED25519KeyPair(comment)
	if err != nil {
		return fmt.Errorf("cannot generate ed25519 key pair for machine '%s' and user '%s': %w", machineName, username, err)
	}
	privateKeyPath, publicKeyPath, err := configuration.GetSSHKeysPaths(machineName, username)
	if err != nil {
		return fmt.Errorf("cannot generate default ssh key paths: %w", err)
	}
	if err := writeFileWithMode(privateKeyPath, privateKey, SSHPrivateKeyFileMode); err != nil {
		return fmt.Errorf("cannot write ed25519 private key for machine '%s' and user '%s': %w", machineName, username, err)
	}
	if err := writeFileWithMode(publicKeyPath, publicKey, SSHPublicKeyFileMode); err != nil {
		return fmt.Errorf("cannot write ed25519 public key for machine '%s' and user '%s': %w", machineName, username, err)
	}

	return nil
}

// marshalED25519KeyPair generates an ed25519 key pair and encodes it in OpenSSH formats:
// the private key as an "OPENSSH PRIVATE KEY" PEM block, and the public key as an
// authorized_keys line ("ssh-ed25519 <base64> <comment>\n")
func marshalED25519KeyPair(comment string) (privateKeyPEM []byte, publicKeyLine []byte, err error) {
	// nil reader: ed25519 uses crypto/rand
	rawPublicKey, rawPrivateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot generate ed25519 key: %w", err)
	}

	pemBlock, err := ssh.MarshalPrivateKey(rawPrivateKey, comment)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot encode private key in OpenSSH format: %w", err)
	}
	privateKeyPEM = pem.EncodeToMemory(pemBlock)

	sshPublicKey, err := ssh.NewPublicKey(rawPublicKey)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot encode public key in OpenSSH format: %w", err)
	}
	// MarshalAuthorizedKey ends with "\n" and has no comment
	authorizedKey := strings.TrimSuffix(string(ssh.MarshalAuthorizedKey(sshPublicKey)), "\n")
	publicKeyLine = []byte(authorizedKey + " " + comment + "\n")

	return privateKeyPEM, publicKeyLine, nil
}

// writeFileWithMode writes data to path and enforces mode, even if the file already exists
// (os.WriteFile only applies the mode on creation)
func writeFileWithMode(path string, data []byte, mode os.FileMode) error {
	if err := os.WriteFile(path, data, mode); err != nil {
		return err
	}
	return os.Chmod(path, mode)
}
