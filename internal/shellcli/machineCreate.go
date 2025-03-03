package shellcli

import (
	"crypto/sha256"
	b64 "encoding/base64"
	"encoding/xml"
	"fmt"
	"freyja/internal"
	"github.com/dypflying/go-qcow2lib/qcow2"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"
)

var configurationPath string

var dryRun bool

const BackingImageFormat string = "qcow2"

const BackingImageFileSuffix string = "-overlay-image." + BackingImageFormat

const RootImageFileSuffix string = "-root-image." + BackingImageFormat

const XMLMachineDescriptionSuffix string = "-libvirt-conf.xml"

// commands definitions
var machineCreateCmd = &cobra.Command{
	Use:              "create",
	Short:            "Machine creation",
	Long:             "Machine creation using handler and cloud-init or ignition",
	TraverseChildren: true, // ensure local flags do not spread to sub commands

	Run: func(cmd *cobra.Command, args []string) {
		Logger.Debug("create machines from configuration file", "config", configurationPath)
		// TODO :
		//   - test with multiple ssh keys
		//   - create the network if it does not already exists
		// build config from path
		var configurationData internal.ConfigurationData
		if err := configurationData.BuildFromFile(configurationPath); err != nil {
			Logger.Error("cannot parse configuration", "configuration", configurationPath, "reason", err.Error())
			os.Exit(1)
		}
		// build cloud init config file
		//var cloudInitData internal.CloudInitUserData
		for _, machine := range configurationData.Machines {
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
			if err := internal.GenerateCloudInitConfigs(&machine, machineDirPath); err != nil {
				Logger.Error("cannot create machine cloud init configurations", "machine", machine.Hostname, "reason", err.Error())
				os.Exit(1)
			}

			// create cloud-init iso file
			Logger.Debug("create cloud ISO file", "machine", machine.Hostname, "parent", machineDirPath)
			cloudInitIsoFile, err := internal.CreateCloudInitIso(&machine, machineDirPath)
			if err != nil {
				Logger.Error("Cannot create machine ISO image file", "machine", machine.Hostname, "reason", err.Error())
				os.Exit(1)
			}

			// copy root image to the machine dir
			// !!! NOT SURE IF ROOT IMAGE FILE SHOULD BE COPIED AS WELL
			// basically, no because overlay is made for single machine usage on top of root image
			rootImageDestinationPath := os.ExpandEnv(filepath.Join(machineDirPath, machine.Hostname+RootImageFileSuffix))
			rootImageSourcePath := os.ExpandEnv(machine.Image)
			Logger.Debug("copy machine image file from root", "machine", machine, "root", rootImageSourcePath, "destination", rootImageDestinationPath)
			if err := internal.CopyFile(rootImageSourcePath, rootImageDestinationPath, 0700); err != nil {
				Logger.Error("Cannot copy machine root image file", "machine", machine.Hostname, "reason", fmt.Sprintf("%v", err.Error()))
				os.Exit(1)
			}

			// using : https://github.com/dypflying/go-qcow2lib/blob/main/examples/backing/qcow2_backing.go
			// use 'qemu-img info' to verify it
			//rootImageDestinationPath := os.ExpandEnv(filepath.Join(machineDirPath, machine.Hostname+RootImageFileSuffix))
			Logger.Debug("create machine image overlay from root image", "machine", machine.Hostname, "parent", machineDirPath, "root", os.ExpandEnv(machine.Image))
			overlayFile, err := createOverlayImage(&machine, rootImageDestinationPath, machineDirPath)
			if err != nil {
				Logger.Error("Cannot create machine overlay image file", "machine", machine.Hostname, "reason", fmt.Sprintf("%v", err.Error()))
				os.Exit(1)
			}

			// create the xml description of the libvirt domain from the machine configuration
			// also injects the overlay image file for qemu
			// also injects the cloud init files for startup sequence
			Logger.Debug("create machine's XML libvirt description", "machine", machine.Hostname, "parent", machineDirPath)
			xmlMachineDescription, err := createLibvirtDomainXMLDescription(&machine, overlayFile, cloudInitIsoFile)
			if err != nil {
				Logger.Error("cannot create the libvirt domain XML description from machine configuration", "machine", machine.Hostname, "reason", fmt.Sprintf("%v", err))
				os.Exit(1)
			}
			// dump description in machine dir (useful for debug)
			xmlMachineDescriptionPath := filepath.Join(machineDirPath, machine.Hostname+XMLMachineDescriptionSuffix)
			if err := os.WriteFile(xmlMachineDescriptionPath, xmlMachineDescription, 0660); err != nil {
				Logger.Error("cannot write the libvirt domain XML description in config dir", "machine", machine.Hostname, "path", xmlMachineDescriptionPath, "reason", err)
				os.Exit(1)
			}

			// create the machine in libvirt
			if !dryRun {
				domain, err := LibvirtConnexion.DomainDefineXML(string(xmlMachineDescription))
				if err != nil {
					Logger.Error("cannot define the machine from libvirt domain XML description", "machine", machine.Hostname, "reason", err)
					os.Exit(1)
				}
				err = LibvirtConnexion.DomainCreate(domain)
				if err != nil {
					Logger.Error("cannot start the machine", "machine", machine.Hostname, "reason", err)
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

func createOverlayImage(machine *internal.ConfigurationMachine, rootImagePath string, machineDir string) (string, error) {
	// using : https://github.com/dypflying/go-qcow2lib/blob/main/examples/backing/qcow2_backing.go
	opts := make(map[string]any)
	backingFile, err := filepath.Abs(rootImagePath)
	if err != nil {
		return "", fmt.Errorf("cannot read base image file '%s' : %w", machine.Image, err)
	}
	overlayFile := filepath.Join(machineDir, machine.Hostname+BackingImageFileSuffix)
	opts[qcow2.OPT_SIZE] = machine.Storage << 30 //qcow2 file's size is 1g
	opts[qcow2.OPT_FMT] = BackingImageFormat     //qcow2 format
	opts[qcow2.OPT_SUBCLUSTER] = true            //enable sub-cluster
	opts[qcow2.OPT_BACKING] = backingFile
	if err := qcow2.Blk_Create(overlayFile, opts); err != nil {
		return "", fmt.Errorf("failed to create overlay qcow2 file '%s' : %w", overlayFile, err)
	}

	return overlayFile, nil
}

func createLibvirtDomainXMLDescription(cm *internal.ConfigurationMachine, overlayFile string, cloudInitIsoFile string) ([]byte, error) {
	// TODO: how to generate the unique id ?
	//  - use '1', '2', ... and store the existing ids in metadata to make sure to not reuse an existing one ?
	//  - use a uuid ?
	//  - use a hash of the hostname ? -> choice 1
	mName := cm.Hostname
	// ID is the fnv hash of the machine hostname
	mUUID, err := generateMachineUUID(mName)
	if err != nil {
		return nil, fmt.Errorf("failed to generate machine UUID: %w", err)
	}
	// memory
	mMemory := internal.XMLDomainDescriptionMemory{
		Unit:  string(internal.MiBMemoryUnit),
		Value: uint64(cm.Memory),
	}
	// cpu
	//mVcpu := cm.Vcpu
	mVcpu := internal.XMLDomainDescriptionVcpu{
		Placement: string(internal.StaticVcpuPlacement),
		Value:     uint64(cm.Vcpu),
	}
	// OS description
	Arch := string(internal.X86OSArch)
	mOSType := internal.XMLDomainDescriptionOSType{
		Type: internal.DefaultOsType,
		Arch: Arch,
	}
	mOSBoot := internal.XMLDomainDescriptionOSBoot{
		Dev: string(internal.HdOsBootDev),
	}
	mOS := internal.XMLDomainDescriptionOS{
		Type: &mOSType,
		Boot: &mOSBoot,
	}
	// metadata
	// TODO : find a way to get the proper os type from golang
	//  for now, no metadata for os info part

	// live image device
	//        <disk type="file" device="disk">
	//            <driver name="qemu" type="qcow2"/>
	//            <source file="/home/kaio/Projects/thales/freyja/test/manual/debian-12-generic-amd64.qcow2"/>
	//            <target dev="hda" bus="ide"/>
	//        </disk>
	//        <disk type="file" device="cdrom">
	//            <driver name="qemu" type="raw"/>
	//            <source file="/home/kaio/Projects/thales/freyja/test/manual/debian12-cloud-init.iso"/>
	//            <target dev="hdb" bus="ide"/>
	//            <readonly/>
	//        </disk>
	liveImageDevice := internal.XMLDomainDescriptionDevicesDisk{
		Device: string(internal.DiskDeviceType),
		Type:   string(internal.FileDeviceDiskType),
		Driver: &internal.XMLDomainDescriptionDevicesDiskDriver{
			Name: string(internal.QemuDeviceDiskDriverName),
			Type: string(internal.QcowDeviceDiskDriverType),
		},
		Source: &internal.XMLDomainDescriptionDevicesDiskSource{
			File: overlayFile,
		},
		BackingStore: &internal.XMLDomainDescriptionDevicesDiskBackingStore{
			Type: string(internal.FileDeviceDiskType),
			Format: &internal.XMLDomainDescriptionDevicesDiskBackingStoreFormat{
				Type: string(internal.QcowDeviceDiskDriverType),
			},
			Source: &internal.XMLDomainDescriptionDevicesDiskBackingStoreSource{
				//File: overlayFile,
				File: os.ExpandEnv(cm.Image),
			},
		},
		Target: &internal.XMLDomainDescriptionDevicesDiskTarget{
			// Only works for 'ide' bus type
			Bus:    string(internal.IdeDeviceDiskTargetBus),
			Device: string(internal.HdaDeviceDiskTargetDev),
		},
	}

	// cloud init raw iso device
	//	    <disk device="cdrom" type="file">
	//	        <driver name="qemu" type="raw"/>
	//	        <source file="/home/kaio/freyja-workspace/build/debian12/debian12_cloud_init.iso" index="1"/>
	//	        <backingStore/>
	//	        <target bus="sata" dev="sda"/>
	//	        <readonly/>
	//	        <alias name="sata0-0-0"/>
	//	        <address bus="0" controller="0" target="0" type="drive" unit="0"/>
	//	    </disk>
	cloudInitIsoDevice := internal.XMLDomainDescriptionDevicesDisk{
		Device: string(internal.CdromDeviceType),
		Type:   string(internal.FileDeviceDiskType),
		Driver: &internal.XMLDomainDescriptionDevicesDiskDriver{
			Name: string(internal.QemuDeviceDiskDriverName),
			Type: string(internal.RawDeviceDiskDriverType),
		},
		Source: &internal.XMLDomainDescriptionDevicesDiskSource{
			File: cloudInitIsoFile,
		},
		Target: &internal.XMLDomainDescriptionDevicesDiskTarget{
			// Only works for 'ide' bus type
			Bus:    string(internal.IdeDeviceDiskTargetBus),
			Device: string(internal.HdbDeviceDiskTargetDev),
		},
	}

	// network device
	//        <interface type="network">
	//            <mac address="52:54:00:17:49:b7"/>
	//            <source network="default"/>
	//        </interface>
	var bridgeInterfaceDevices []internal.XMLDomainDescriptionDevicesInterface
	if len(cm.Networks) > 0 {
		// user defined networks
		bridgeInterfaceDevices = make([]internal.XMLDomainDescriptionDevicesInterface, len(cm.Networks))
		for _, network := range cm.Networks {
			// TODO :
			//  - add the possibility to provide the host's target interface
			bridgeInterfaceDevice := internal.XMLDomainDescriptionDevicesInterface{
				Type: string(internal.NetworkDeviceInterfaceType),
				Mac: &internal.XMLDomainDescriptionDevicesInterfaceMac{
					Address: network.Mac,
				},
				Source: &internal.XMLDomainDescriptionDevicesInterfaceSource{
					Bridge:  internal.DefaultInterfaceSourceBridge,
					Network: network.Name,
				},
				//Target: internal.XMLDomainDescriptionDevicesInterfaceTarget{}, // provide if user conf specifies a host interface
				Target: nil,
				Model: &internal.XMLDomainDescriptionDevicesInterfaceModel{
					Type: internal.DefaultInterfaceModelType,
				},
			}
			bridgeInterfaceDevices = append(bridgeInterfaceDevices, bridgeInterfaceDevice)
		}
	} else {
		// if no network define in user conf input, stick with the default one
		bridgeInterfaceDevices = []internal.XMLDomainDescriptionDevicesInterface{internal.DefaultDeviceInterface}
	}

	// console device for graphical debug
	//	        <console type='pty'>
	//	          <target type='serial' port='0'/>
	//	        </console>
	consoleDevice := internal.XMLDomainDescriptionDevicesConsole{
		Type: string(internal.PtyDeviceConsoleType),
		Target: &internal.XMLDomainDescriptionDevicesConsoleTarget{
			Type: string(internal.SerialDeviceConsoleTargetType),
		},
	}

	// xml
	deviceDisks := []internal.XMLDomainDescriptionDevicesDisk{liveImageDevice, cloudInitIsoDevice}
	xmlDescription := internal.XMLDomainDescription{
		Type:   internal.DefaultDomainType,
		Name:   mName,
		UUID:   mUUID,
		Vcpu:   &mVcpu,
		Memory: &mMemory,
		OS:     &mOS,
		Devices: &internal.XMLDomainDescriptionDevices{
			Emulator:   string(internal.QemuX86DevicesEmulator),
			Disks:      deviceDisks,
			Interfaces: bridgeInterfaceDevices,
			Console:    []internal.XMLDomainDescriptionDevicesConsole{consoleDevice},
		},
	}
	return xml.Marshal(xmlDescription)
}

func generateMachineUUID(machineName string) (string, error) {
	h := sha256.New()
	_, err := h.Write([]byte(machineName))
	if err != nil {
		return "", fmt.Errorf("cannot generate a hash based on the machine hostname '%s': %w", machineName, err)
	}
	sum := h.Sum(nil)
	mID := b64.StdEncoding.EncodeToString(sum)[:16]
	// UUID is generated from the ID
	mUUIDRaw, err := uuid.FromBytes(sum[:16])
	if err != nil {
		return "", fmt.Errorf("cannot create machine UUID based on its ID '%s': %w", mID, err)
	}
	return fmt.Sprintf("%v", mUUIDRaw), nil
}
